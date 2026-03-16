package main

import (
	"films/auth"
	"films/connectdatabase"
	"films/handlers"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env не найден, используем системные переменные")
	}
	db := connectdatabase.Connect()
	defer db.Close()
	auth.InitAuth()
	http.Handle("/", auth.OptionalAuth(handlers.DbMoviesHandler(db)))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/register", handlers.MakeDBHandler(db, handlers.RegisterHandler))
	http.HandleFunc("/login", handlers.MakeDBHandler(db, handlers.LoginHandler))
	http.HandleFunc("/refresh", handlers.MakeDBHandler(db, handlers.RefreshHandler))
	http.HandleFunc("/logout", handlers.MakeDBHandler(db, handlers.LogoutHandler))
	http.Handle("/me", auth.AuthMiddleware(handlers.MakeDBHandler(db, handlers.MeHandler)))
	http.Handle("/states", auth.AuthMiddleware(handlers.MakeDBHandler(db, handlers.GetStatesHandler)))
	http.Handle("/state", auth.AuthMiddleware(handlers.MakeDBHandler(db, handlers.SetStateHandler)))
	http.Handle("/watched", auth.AuthMiddleware(handlers.MakeDBHandler(db, handlers.WatchedMoviesHandler)))
	http.Handle("/liked", auth.AuthMiddleware(handlers.MakeDBHandler(db, handlers.LikedMoviesHandler)))
	http.Handle("/movie/", auth.OptionalAuth(http.HandlerFunc(handlers.MovieInfoHandler)))
	http.Handle("/search", auth.OptionalAuth(http.HandlerFunc(handlers.Search)))
	http.ListenAndServe(":8080", nil)
}
