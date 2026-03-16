-- ===== Создание схемы =====
CREATE SCHEMA IF NOT EXISTS movies;

-- ===== Последовательности =====
CREATE SEQUENCE IF NOT EXISTS movies.refresh_tokens_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS movies.user_movie_states_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS movies.users_id_seq START 1;

-- ===== Таблицы =====
CREATE TABLE IF NOT EXISTS movies.favorite_status (
                                                      movie_id VARCHAR(50) NOT NULL,
    favorite BOOLEAN DEFAULT FALSE NOT NULL,
    watched BOOLEAN DEFAULT FALSE NOT NULL
    );

CREATE TABLE IF NOT EXISTS movies.refresh_tokens (
                                                     id INTEGER NOT NULL DEFAULT nextval('movies.refresh_tokens_id_seq'),
    user_id INTEGER,
    token TEXT NOT NULL,
    issued_at TIMESTAMPTZ DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS movies.titles (
                                             id VARCHAR(20) NOT NULL PRIMARY KEY,
    title VARCHAR(1000),
    year INTEGER,
    genres VARCHAR(500),
    poster_url VARCHAR(500)
    );

CREATE TABLE IF NOT EXISTS movies.user_movie_states (
                                                        id INTEGER NOT NULL DEFAULT nextval('movies.user_movie_states_id_seq'),
    user_id INTEGER NOT NULL,
    title_id TEXT NOT NULL,
    liked BOOLEAN DEFAULT FALSE NOT NULL,
    watched BOOLEAN DEFAULT FALSE NOT NULL,
    PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS movies.users (
                                            id INTEGER NOT NULL DEFAULT nextval('movies.users_id_seq'),
    username VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
    );

-- ===== Загрузка фильмов из CSV =====
-- Важно: путь должен быть таким же, как в docker-compose volume
COPY movies.titles (id, title, year, genres, poster_url)
    FROM '/docker-entrypoint-initdb.d/movies.csv'
    DELIMITER ','
    CSV HEADER;