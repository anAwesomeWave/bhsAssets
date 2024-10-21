package auth

import (
	"bhsAssets/internal/storage"
	"bhsAssets/internal/storage/models"
	"encoding/json"
	"errors"
	"github.com/go-chi/jwtauth/v5"
	"net/http"
	"time"
)

var tokenAuth *jwtauth.JWTAuth

func init() {
	// TODO: get from env
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)
}

//type registerUser struct {
//	Login    string `json:"login"`
//	Password string `json:"password"`
//}

func Register(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.Users
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		_, err := strg.CreateUser(user.Login, user.PasswordHash)
		if err != nil {
			if errors.Is(err, storage.ErrExists) {
				http.Error(w, "User exists", http.StatusBadRequest)
				return
			}
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("User registered"))
	}
}

func Login(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input models.Users
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		user, err := strg.GetUser(input.Login, input.PasswordHash)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusNotFound)
			return
		}

		_, tokenString, err := tokenAuth.Encode(map[string]interface{}{
			"user_id": user.Id,
			"exp":     time.Now().Add(time.Hour).Unix(),
		})
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		cookie := &http.Cookie{
			Name:     "jwt",
			Value:    tokenString,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			MaxAge:   3600,
		}
		http.SetCookie(w, cookie)
		w.Write([]byte(tokenString))
	}
}
