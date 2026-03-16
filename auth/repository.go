package auth

import (
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func StoreRefreshToken(db *sql.DB, userID int, token string) error {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return err
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid claims")
	}
	expF, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("invalid exp")
	}
	expiresAt := time.Unix(int64(expF), 0)
	_, err = db.Exec(`INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`, userID, token, expiresAt)
	return err
}

func DeleteRefreshToken(db *sql.DB, token string) error {
	_, err := db.Exec(`DELETE FROM refresh_tokens WHERE token=$1`, token)
	return err
}
func DeleteUserRefreshTokens(db *sql.DB, userID int) error {
	_, err := db.Exec(`DELETE FROM refresh_tokens WHERE user_id=$1`, userID)
	return err
}

func VerifyRefreshInDB(db *sql.DB, token string) (int, error) {
	var userID int
	var expiresAt time.Time
	err := db.QueryRow(`SELECT user_id, expires_at FROM refresh_tokens WHERE token=$1`, token).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("not found")
		}
		return 0, err
	}
	if time.Now().After(expiresAt) {
		_ = DeleteRefreshToken(db, token)
		return 0, errors.New("expired")
	}
	return userID, nil
}
