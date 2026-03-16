package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

func OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxIsAuth, false)

		cookie, err := r.Cookie("access_token")
		if err == nil && cookie.Value != "" {

			token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
				if t.Method != jwt.SigningMethodHS256 {
					return nil, errors.New("unexpected alg")
				}
				return jwtSecret, nil
			})

			if err == nil && token.Valid {

				claims, ok := token.Claims.(jwt.MapClaims)
				if ok {

					typ, ok := claims["typ"].(string)
					if ok && typ == "access" {

						sub, ok := claims["sub"].(string)
						if ok {
							uid, err := strconv.Atoi(sub)
							if err == nil {
								ctx = context.WithValue(ctx, ctxIsAuth, true)
								ctx = context.WithValue(ctx, ctxUserID, uid)
							}
						}
					}
				}
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
