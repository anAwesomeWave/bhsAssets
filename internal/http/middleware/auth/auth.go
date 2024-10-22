package auth

import (
	"bhsAssets/internal/storage"
	"bhsAssets/internal/storage/models"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"log"
	"net/http"
)

type contextKey string

const userContextKey = contextKey("user")

var UNAUTHORIZED_ERR = errors.New("Unauthorized user")

func GetUserByJwtToken(strg storage.Storage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, err := jwtauth.FromContext(r.Context())
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			userID, ok := claims["user_id"].(float64)
			if !ok {
				log.Printf("GetUserByJwtToken: Cannot get userId from claims user_id - %v", claims["user_id"])
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			user, err := strg.GetUserById(int64(userID))
			if err != nil {
				http.Error(w, "User not found", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FromContext(ctx context.Context) (*models.Users, bool) {
	user, ok := ctx.Value(userContextKey).(*models.Users)
	return user, ok
}

func IdFromContext(ctx context.Context) (int64, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("no token found in context %w", UNAUTHORIZED_ERR)
	}
	userID, ok := claims["user_id"].(float64)
	if !ok {
		log.Printf("IdFromContext: Cannot get userId from claims user_id - %v", claims["user_id"])
		return 0, fmt.Errorf("invalid token %w", UNAUTHORIZED_ERR)
	}
	return int64(userID), nil
}
