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

func MethodOverride(next http.Handler) http.Handler {
	// нужно для html форм, т.к. они поддержива.т только get и post по умолч.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the form has a _method parameter
		log.Println("ОТРАБОТАЛ")
		if r.Method == http.MethodPost {
			if r.FormValue("_method") == "PATCH" {
				// Override the request method
				log.Println("TYT BIL")
				r.Method = http.MethodPatch
			}
		}
		log.Println(r.Method)
		next.ServeHTTP(w, r)
	})
}
