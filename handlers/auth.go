package handlers

import (
	"database/sql"
	"encoding/json"
	"films/auth"
	"net/http"
	"strings"
	"time"
)

func RegisterHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "templates/register.html")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := ""
	email := ""
	password := ""

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		var body struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		username = strings.TrimSpace(body.Username)
		email = strings.TrimSpace(body.Email)
		password = body.Password
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		username = strings.TrimSpace(r.FormValue("username"))
		email = strings.TrimSpace(r.FormValue("email"))
		password = r.FormValue("password")
	}
	if username == "" || email == "" || len(password) < 2 {
		http.Error(w, "invalid input (password min 2, username/email required)", http.StatusBadRequest)
		return
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	var userID int
	err = db.QueryRow(`INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id`, username, email, hash).Scan(&userID)
	if err != nil {
		http.Error(w, "could not create user (maybe username/email already exists)", http.StatusBadRequest)
		return
	}
	if strings.HasPrefix(ct, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": userID})
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "templates/login.html")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	login := ""
	password := ""
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		login = body.Email
		password = body.Password
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		login = r.FormValue("email")
		password = r.FormValue("password")
	}

	if login == "" || password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return
	}

	var id int
	var username, email, hash string
	err := db.QueryRow(`SELECT id, username, email, password_hash FROM users WHERE email=$1 OR username=$1`, login).
		Scan(&id, &username, &email, &hash)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err := auth.CheckPassword(hash, password); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, err := auth.CreateAccessToken(id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	refreshToken, err := auth.CreateRefreshToken(id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	_ = auth.StoreRefreshToken(db, id, refreshToken)

	if strings.HasPrefix(ct, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"user": map[string]any{
				"id":       id,
				"username": username,
				"email":    email,
			},
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(auth.AccessTTL()),
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // в проде: true
		Expires:  time.Now().Add(auth.RefreshTTL()),
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func RefreshHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var token string
	if c, err := r.Cookie("refresh_token"); err == nil && c.Value != "" {
		token = c.Value
	} else if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			token = body.RefreshToken
		}
	} else {
		_ = r.ParseForm()
		token = r.FormValue("refresh_token")
	}
	if token == "" {
		http.Error(w, "no refresh token", http.StatusBadRequest)
		return
	}

	userID, err := auth.ParseRefreshToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	uidFromDB, err := auth.VerifyRefreshInDB(db, token)
	if err != nil || uidFromDB != userID {
		http.Error(w, "refresh token revoked or not found", http.StatusUnauthorized)
		return
	}

	_ = auth.DeleteRefreshToken(db, token)
	newRefresh, err := auth.CreateRefreshToken(userID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	_ = auth.StoreRefreshToken(db, userID, newRefresh)

	newAccess, err := auth.CreateAccessToken(userID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    newAccess,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(auth.AccessTTL()),
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(auth.RefreshTTL()),
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": newAccess,
	})
}

func LogoutHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var token string
	if c, err := r.Cookie("refresh_token"); err == nil {
		token = c.Value
	} else {
		_ = r.ParseForm()
		token = r.FormValue("refresh_token")
	}
	if token != "" {
		_ = auth.DeleteRefreshToken(db, token)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login.html?logged_out=1", http.StatusSeeOther)
}
