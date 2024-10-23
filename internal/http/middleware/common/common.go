package common

import (
	"context"
	"log"
	"net/http"
	"strings"
)

const IS_API_CONTEXT_KEY = "is_api"

func SetIsApiContextVariable(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		ctx := context.WithValue(r.Context(), IS_API_CONTEXT_KEY, strings.HasPrefix(r.RequestURI, "/api"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IsApiFromContext(ctx context.Context) (bool, bool) {
	user, ok := ctx.Value(IS_API_CONTEXT_KEY).(bool)
	return user, ok
}

func JsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
