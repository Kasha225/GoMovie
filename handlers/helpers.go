package handlers

import (
	"database/sql"
	"net/http"
)

func MakeDBHandler(db *sql.DB, fn func(db *sql.DB, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(db, w, r)
	}
}
