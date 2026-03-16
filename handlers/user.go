package handlers

import (
	"database/sql"
	"encoding/json"
	"films/auth"
	"net/http"
)

func MeHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r)
	if !ok {
		http.Error(w, "no user in context", http.StatusInternalServerError)
		return
	}
	var username, email string
	err := db.QueryRow(`SELECT username, email FROM users WHERE id=$1`, uid).Scan(&username, &email)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":       uid,
		"username": username,
		"email":    email,
	})
}
