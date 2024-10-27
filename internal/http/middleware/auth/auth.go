package auth

import (
	"bhsAssets/internal/http/middleware/common"
	"bhsAssets/internal/storage"
	"bhsAssets/internal/storage/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"html/template"
	"log"
	"net/http"
)

type contextKey string

const userContextKey = contextKey("user")

var UnauthorizedErr = errors.New("Unauthorized user")

func GetUserByJwtToken(strg storage.Storage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		const fn = "Middleware.Auth.GetUserByJwtToken"
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, err := jwtauth.FromContext(r.Context())
			if err != nil {
				// за это отвечает jwtauth.Authenticator
				//http.Error(w, "Unauthorized", http.StatusUnauthorized)
				next.ServeHTTP(w, r)
				return
			}
			userID, ok := claims["user_id"].(float64)
			if !ok {
				log.Printf("%v: Cannot get userId from claims user_id - %v", fn, claims["user_id"])
				http.Error(w, "Invalid token", http.StatusInternalServerError)
				return
			}
			user, err := strg.GetUserById(int64(userID))
			if err != nil {
				user = nil
				// за это отвечает jwtauth.Authenticator
				log.Printf("%s: User with id %v not found in database: %v\n", fn, userID, err)
				http.Error(w, "User not found in database", http.StatusInternalServerError)
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
		return 0, fmt.Errorf("no token found in context %w", UnauthorizedErr)
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		log.Printf("IdFromContext: Cannot get userId from claims user_id - %v", claims["user_id"])
		return 0, fmt.Errorf("invalid token %w", UnauthorizedErr)
	}
	return int64(userID), nil
}

func CustomAuthenticator(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, _, err := jwtauth.FromContext(r.Context())

			if err != nil {
				isApi, ok := common.IsApiFromContext(r.Context())
				if !ok {
					http.Error(w, "Failed to get context", http.StatusInternalServerError)
					return
				}
				if isApi {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]string{"error": "401"})
					return
				}
				t, err := template.ParseFiles("./templates/common/base.html", "./templates/common/401_error.html")
				if err != nil {
					http.Error(w, "Error loading template", http.StatusInternalServerError)
					log.Println(err)
					return
				}
				w.WriteHeader(http.StatusUnauthorized)
				if err := t.Execute(w, map[string]interface{}{"isLogined": false}); err != nil {
					http.Error(w, "Error templating", http.StatusInternalServerError)
					log.Println(err)
					return
				}
				return
			}
			if token == nil || jwt.Validate(token) != nil {
				isApi, ok := common.IsApiFromContext(r.Context())
				if !ok {
					http.Error(w, "Failed to get context", http.StatusInternalServerError)
					return
				}
				if isApi {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]string{"error": "401"})
					return
				}
				t, err := template.ParseFiles("./templates/common/base.html", "./templates/common/401_error.html")
				if err != nil {
					http.Error(w, "Error loading template", http.StatusInternalServerError)
					log.Println(err)
					return
				}
				w.WriteHeader(http.StatusUnauthorized)
				if err := t.Execute(w, map[string]interface{}{"isLogined": false}); err != nil {
					http.Error(w, "Error templating", http.StatusInternalServerError)
					log.Println(err)
					return
				}
				return
			}

			// Token is authenticated, pass it through
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}
