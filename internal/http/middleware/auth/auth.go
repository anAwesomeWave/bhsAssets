package auth

import (
	"bhsAssets/internal/storage"
	"bhsAssets/internal/storage/models"
	"context"
	"encoding/json"
	"github.com/go-chi/jwtauth/v5"
	"github.com/ydb-platform/ydb-go-sdk/v3/log"
	"net/http"
)

type contextKey string

const userContextKey = contextKey("user")

func GetUserByJwtToken(strg storage.Storage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, err := jwtauth.FromContext(r.Context())
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			userID, err := (claims["user_id"].(json.Number)).Int64()
			if err != nil {
				log.Error(err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			user, err := strg.GetUserById(userID)
			if err != nil {
				http.Error(w, "User not found", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FromContext(ctx context.Context) (*models.Users, bool) {
	user, ok := ctx.Value(userContextKey).(*models.Users)
	return user, ok
}
