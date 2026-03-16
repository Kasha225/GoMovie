package movies

import "films/models"

type Title struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Year      int    `json:"year"`
	Genres    string `json:"genre"`
	PosterURL string `json:"poster_url"`
}
type TitleJSON struct {
	ID        string `json:"imdbID"`
	Title     string `json:"title"`
	Year      string `json:"year"`
	Genres    string `json:"genre"`
	PosterURL string `json:"poster"`
}

type SearchPageData struct {
	Movie  models.MovieInfo
	IsAuth bool
}
