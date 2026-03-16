package movies

import (
	"encoding/json"
	"films/config"
	"fmt"
	"net/http"
	"strconv"
)

func fetchTitleFromApi(id string) (*Title, error) {
	apiKey := config.GetOMDBApiKey()
	apiUrl := fmt.Sprintf("http://www.omdbapi.com/?i=%s&apikey=%s", id, apiKey)
	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к API: %w", err)
	}
	defer resp.Body.Close()
	var titleJS TitleJSON
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела запроса: %w", err)
	}
	if err := json.NewDecoder(resp.Body).Decode(&titleJS); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа API: %w", err)
	}
	yearInt, err := strconv.Atoi(titleJS.Year)
	if err != nil {
		return nil, fmt.Errorf("ошибка при конвертации: %w", err)
	}
	return &Title{
		ID:        titleJS.ID,
		Title:     titleJS.Title,
		Year:      yearInt,
		Genres:    titleJS.Genres,
		PosterURL: titleJS.PosterURL,
	}, nil
}
