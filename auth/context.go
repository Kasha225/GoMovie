package auth

import "net/http"

type ctxKey string

const (
	ctxUserID ctxKey = "user_id"
	ctxIsAuth ctxKey = "isAuth"
)

func UserIDFromContext(r *http.Request) (int, bool) {
	uid, ok := r.Context().Value(ctxUserID).(int)
	return uid, ok
}

func IsAuthenticated(r *http.Request) bool {
	isAuth, ok := r.Context().Value(ctxIsAuth).(bool)
	return ok && isAuth
}
