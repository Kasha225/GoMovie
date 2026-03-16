package handlers

import (
	"database/sql"
	"encoding/json"
	"films/auth"
	"films/movies"
	"log"
	"net/http"
	"strings"

	"github.com/lib/pq"
)

func GetStatesHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
	uid, ok := auth.UserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	var body struct {
		Items []string `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if len(body.Items) == 0 {
		json.NewEncoder(w).Encode(map[string]any{})
		return
	}

	rows, err := db.Query(`SELECT title_id, liked, watched FROM user_movie_states WHERE user_id=$1 AND title_id= ANY($2)`, uid, pq.Array(body.Items))
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	res := map[string]map[string]bool{}
	for _, item := range body.Items {
		res[item] = map[string]bool{"liked": false, "watched": false}
	}
	for rows.Next() {
		var title string
		var liked, watched bool
		if err := rows.Scan(&title, &liked, &watched); err == nil {
			res[title] = map[string]bool{"liked": liked, "watched": watched}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func SetStateHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	uid, ok := auth.UserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Item   string `json:"item"`
		Action string `json:"action"`
		Value  *bool  `json:"value,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	item := strings.TrimSpace(body.Item)
	act := strings.ToLower(strings.TrimSpace(body.Action))
	if item == "" || (act != "like" && act != "liked" && act != "watched" && act != "watch") {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	var exists bool
	if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM movies.titles WHERE id=$1)`, item).Scan(&exists); err != nil {
		log.Printf("SetStateHandler: check title exists error: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if !exists {
		if err := movies.AddTitle(db, item); err != nil {
			log.Printf("SetStateHandler: AddMovie failed for %s: %v", item, err)
			http.Error(w, "failed to add title", http.StatusInternalServerError)
			return
		}
	}

	if act == "like" || act == "liked" {
		if body.Value != nil {
			_, err := db.Exec(`INSERT INTO user_movie_states (user_id, title_id, liked)
				VALUES ($1,$2,$3)
				ON CONFLICT (user_id, title_id) DO UPDATE SET liked = EXCLUDED.liked`, uid, item, &body.Value)
			if err != nil {
				http.Error(w, "db error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"liked": *body.Value})
			return
		}

		var cur bool
		err := db.QueryRow(`SELECT liked FROM user_movie_states WHERE user_id=$1 AND title_id=$2`, uid, item).Scan(&cur)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err := db.Exec(`INSERT INTO user_movie_states (user_id, title_id, liked) VALUES ($1, $2, $3)`, uid, item, true)
				if err != nil {
					http.Error(w, "db error", http.StatusInternalServerError)
					return
				}
				json.NewEncoder(w).Encode(map[string]bool{"liked": true})
				return
			}
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		newState := !cur
		_, err = db.Exec(`UPDATE user_movie_states SET liked=$1 WHERE user_id=$2 AND title_id=$3`, newState, uid, item)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"liked": newState})
		return
	}
	if body.Value != nil {
		_, err := db.Exec(`
			INSERT INTO user_movie_states (user_id, title_id, watched)
			VALUES ($1,$2,$3)
			ON CONFLICT (user_id, title_id) DO UPDATE SET watched = EXCLUDED.watched
		`, uid, item, *body.Value)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"watched": *body.Value})
		return
	}
	var cur bool
	err := db.QueryRow(`SELECT watched FROM user_movie_states WHERE user_id=$1 AND title_id=$2`, uid, item).Scan(&cur)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err := db.Exec(`INSERT INTO user_movie_states (user_id, title_id, watched) VALUES ($1,$2,$3)`, uid, item, true)
			if err != nil {
				http.Error(w, "db error", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]bool{"watched": true})
			return
		}
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	newState := !cur
	_, err = db.Exec(`UPDATE user_movie_states SET watched=$1 WHERE user_id=$2 AND title_id=$3`, newState, uid, item)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"watched": newState})

}
