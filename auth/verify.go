package auth

import (
	"errors"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

func ParseRefreshToken(token string) (int, error) {
	tok, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !tok.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	if typ, ok := claims["typ"].(string); !ok || typ != "refresh" {
		return 0, errors.New("invalid token type")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return 0, errors.New("invalid sub")
	}

	userID, err := strconv.Atoi(sub)
	if err != nil {
		return 0, errors.New("invalid sub")
	}

	return userID, nil
}
