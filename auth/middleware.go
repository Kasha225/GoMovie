package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		var tokenStr string

		if auth != "" {
			parts := strings.Fields(auth)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid auth header", http.StatusUnauthorized)
				return
			}
			tokenStr = parts[1]
		} else {
			if c, err := r.Cookie("access_token"); err == nil && c.Value != "" {
				tokenStr = c.Value
			} else {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected alg")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
			return
		}
		typ, ok := claims["typ"].(string)
		if !ok || typ != "access" {
			http.Error(w, "invalid token type", http.StatusUnauthorized)
			return
		}
		sub, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, "invalid sub", http.StatusUnauthorized)
			return
		}
		uid, err := strconv.Atoi(sub)
		if err != nil {
			http.Error(w, "invalid sub", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
