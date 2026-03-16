package movies

import (
	"database/sql"
	"fmt"
)

func GetTitlesPaginated(db *sql.DB, limit, offset int) ([]Title, error) {
	rows, err := db.Query(`
		SELECT id, title, year, genres, poster_url
		FROM titles
		ORDER BY id
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var titles []Title
	for rows.Next() {
		var t Title
		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Year,
			&t.Genres,
			&t.PosterURL,
		); err != nil {
			return nil, err
		}
		titles = append(titles, t)
	}

	return titles, nil
}

func GetTitlesCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM titles`).Scan(&count)
	return count, err
}

func GetTitles(db *sql.DB) ([]Title, error) {
	query := `SELECT id, title, year, genres, poster_url FROM titles`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var titles []Title
	for rows.Next() {
		var t Title
		err := rows.Scan(&t.ID, &t.Title, &t.Year, &t.Genres, &t.PosterURL)
		if err != nil {
			return nil, err
		}
		titles = append(titles, t)
	}
	return titles, nil
}

func MovieExists(db *sql.DB, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM titles WHERE id = $1)`
	var exists bool
	err := db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func AddTitle(db *sql.DB, id string) error {
	exists, err := MovieExists(db, id)
	if err != nil {
		return fmt.Errorf("Ошибка проверка фильма в бд : %w", err)
	}
	if exists {
		return fmt.Errorf("Фильм уже существует в базе данных: %w", err)
	}

	title, err := fetchTitleFromApi(id)
	if err != nil {
		return fmt.Errorf("не удалось получить информацию о фильме: %w", err)
	}
	query := `
		INSERT INTO titles (id, title, year, genres, poster_url)
		VALUES ($1, $2, $3, $4, $5)
		`
	_, err = db.Exec(query, title.ID, title.Title, title.Year, title.Genres, title.PosterURL)
	if err != nil {
		return fmt.Errorf("Ошибка добавления фильма в базу данных: %w", err)
	}
	return nil
}

func GetWatchedFilms(db *sql.DB, id int) ([]Title, error) {
	query := `SELECT t.id, t.title, t.year, t.genres, t.poster_url FROM user_movie_states ums JOIN titles t ON ums.title_id = t.id WHERE ums.user_id=$1 AND ums.watched=true`
	rows, err := db.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении списка просмотренных фильмов", err)
	}
	defer rows.Close()
	var titles []Title
	for rows.Next() {
		var t Title
		err := rows.Scan(&t.ID, &t.Title, &t.Year, &t.Genres, &t.PosterURL)
		if err != nil {
			return nil, err
		}
		titles = append(titles, t)
	}
	return titles, nil
}

func GetLikedFilms(db *sql.DB, id int) ([]Title, error) {
	query := `SELECT t.id, t.title, t.year, t.genres, t.poster_url FROM user_movie_states ums JOIN titles t ON ums.title_id = t.id WHERE ums.user_id=$1 AND ums.liked=true`
	rows, err := db.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении списка просмотренных фильмов", err)
	}
	defer rows.Close()
	var titles []Title
	for rows.Next() {
		var t Title
		err := rows.Scan(&t.ID, &t.Title, &t.Year, &t.Genres, &t.PosterURL)
		if err != nil {
			return nil, err
		}
		titles = append(titles, t)
	}
	return titles, nil
}
