package handlers

import (
	"database/sql"
	"encoding/json"
	"films/auth"
	"films/config"
	"films/models"
	"films/movies"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func MoviesHandler(w http.ResponseWriter, r *http.Request) {
	var result models.MovieSearch
	apiKey := config.GetOMDBApiKey()
	apiUrl := fmt.Sprintf("http://www.omdbapi.com/?s=2020&page=1&apikey=%s", apiKey)
	resp, err := http.Get(apiUrl)
	if err != nil {
		log.Println("Ошибка соединения: ", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Ошибка чтения тела: ", err)
		return
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Println("Ошибка при парсинге", err)
	}
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	err = tmpl.Execute(w, result.Search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func MovieInfoHandler(w http.ResponseWriter, r *http.Request) {
	var result models.MovieInfo
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}
	movieID := pathParts[2]
	apiKey := config.GetOMDBApiKey()
	apiUrl := fmt.Sprintf("http://www.omdbapi.com/?i=%s&apikey=%s", movieID, apiKey)
	resp, err := http.Get(apiUrl)
	if err != nil {
		log.Println("Ошибка соединения: ", err)
		return
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("Ошибка парсинга", err)
		http.Error(w, "API error", http.StatusInternalServerError)
		return
	}

	data := movies.SearchPageData{
		Movie:  result,
		IsAuth: auth.IsAuthenticated(r),
	}
	tmpl, err := template.ParseFiles("templates/infofilm.html")
	if err != nil {
		log.Println("Ошибка парсинга шаблона: ", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Println("Ошибка рендеринга шаблона: ", err)
	}
}

func DbMoviesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const limit = 20
		page := 1

		if p := r.URL.Query().Get("page"); p != "" {
			if v, err := strconv.Atoi(p); err == nil && v > 0 {
				page = v
			}
		}
		offset := (page - 1) * limit

		titles, err := movies.GetTitlesPaginated(db, limit, offset)
		if err != nil {
			log.Println("Ошибка GetTitlesPaginated:", err)
			http.Error(w, "Ошибка получения данных", http.StatusInternalServerError)
			return
		}

		total, err := movies.GetTitlesCount(db)
		if err != nil {
			log.Println("Ошибка GetTitlesCount:", err)
			http.Error(w, "Ошибка подсчета", http.StatusInternalServerError)
			return
		}
		totalPages := int(math.Ceil(float64(total) / float64(limit)))

		data := struct {
			Titles     []movies.Title
			IsAuth     bool
			Page       int
			TotalPages int
		}{
			Titles:     titles,
			Page:       page,
			TotalPages: totalPages,
		}
		if _, ok := auth.UserIDFromContext(r); ok {
			data.IsAuth = true
		}
		tmpl := template.New("index.html").Funcs(template.FuncMap{
			"inc": func(i int) int { return i + 1 },
			"dec": func(i int) int { return i - 1 },
		})
		tmpl, err = tmpl.ParseFiles("templates/index.html")
		if err != nil {
			log.Println("Ошибка парсинга шаблона:", err)
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Println("Ошибка рендеринга:", err)
		}
	}
}

func WatchedMoviesHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	titles, err := movies.GetWatchedFilms(db, uid)
	if err != nil {
		http.Error(w, "Ошибка получения списка просмотренных фильмов", http.StatusInternalServerError)
	}
	tmpl, err := template.ParseFiles("templates/watchedfilms.html")
	if err != nil {
		log.Println("Ошибка парсинга шаблона")
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, titles)
	if err != nil {
		log.Println("Ошибка рендринга", err)
	}
}

func LikedMoviesHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	titles, err := movies.GetLikedFilms(db, uid)
	if err != nil {
		http.Error(w, "Ошибка получения списка лайкнутых фильмов", http.StatusInternalServerError)
		return
	}
	tmpl, err := template.ParseFiles("templates/likedfilms.html")
	if err != nil {
		log.Println("Ошибка рендеринга шаблона")
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, titles)
	if err != nil {
		log.Println("Ошибка рендеринга", err)
	}
}
func Search(w http.ResponseWriter, r *http.Request) {
	var result models.MovieInfo
	title_movie := r.FormValue("title_movie")
	encodedTitle := url.QueryEscape(title_movie)
	apiKey := config.GetOMDBApiKey()
	apiUrl := fmt.Sprintf("http://www.omdbapi.com/?t=%s&apikey=%s", encodedTitle, apiKey)
	resp, err := http.Get(apiUrl)
	if err != nil {
		log.Println("Ошибка соединения: ", err)
		return
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("Ошибка парсинга:", err)
		http.Error(w, "Parse error", http.StatusInternalServerError)
		return
	}

	data := movies.SearchPageData{
		Movie:  result,
		IsAuth: auth.IsAuthenticated(r),
	}
	tmpl := template.Must(template.ParseFiles("templates/searchfilm.html"))
	if err := tmpl.Execute(w, data); err != nil {
		log.Println("Ошибка рендеринга шаблона:", err)
	}

}
