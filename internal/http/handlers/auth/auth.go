package auth

import (
	mauth "bhsAssets/internal/http/middleware/auth"
	"bhsAssets/internal/http/middleware/common"
	"bhsAssets/internal/storage"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"html/template"
	"log"
	"net/http"
	"time"
)

var TokenAuth *jwtauth.JWTAuth

const INVAILD_CREDENTIALS_QUERY = "isInvalid"

func init() {
	// TODO: get from env
	TokenAuth = jwtauth.New("HS256", []byte("secret"), nil)
}

type authUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func Register(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user authUser
		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)

		}
		if isApi {
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil || len(user.Login) == 0 || len(user.Password) == 0 {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
		} else {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Unable to parse form", http.StatusInternalServerError)
				return
			}
			if _, ok := r.Form["login"]; !ok {
				http.Error(w, "Bad form. Unable to find `login` field", http.StatusBadRequest)
				return
			}
			if _, ok := r.Form["password"]; !ok {
				http.Error(w, "Bad form. Unable to find `password` field", http.StatusBadRequest)
				return
			}
			user.Login = r.Form["login"][0]
			user.Password = r.Form["password"][0]
		}
		if len(user.Login) == 0 || len(user.Password) == 0 {
			log.Println(user)
			http.Error(w, "user Parsing error. fields are empty", http.StatusInternalServerError)
			return
		}
		_, err := strg.CreateUser(user.Login, user.Password)
		if err != nil {
			if errors.Is(err, storage.ErrExists) {
				http.Error(w, "User exists", http.StatusBadRequest)
				return
			}
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		if isApi {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("User registered"))
		} else {
			http.Redirect(
				w,
				r,
				"/auth/login",
				http.StatusSeeOther,
			)
		}
	}
}

func GetRegisterPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/common/base.html", "./templates/auth/register.html")

	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	userId, authErr := mauth.IdFromContext(r.Context())
	if err != nil && !errors.Is(err, mauth.UnauthorizedErr) {
		http.Error(w, "Auth token internal error", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, map[string]interface{}{"isLogined": authErr == nil && userId > 0}); err != nil {
		http.Error(w, "Error templating", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func Login(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input authUser
		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)

		}
		if isApi {
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
		} else {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Unable to parse form", http.StatusInternalServerError)
				return
			}
			if _, ok := r.Form["login"]; !ok {
				http.Error(w, "Bad form. Unable to find `login` field", http.StatusBadRequest)
				return
			}
			if _, ok := r.Form["password"]; !ok {
				http.Error(w, "Bad form. Unable to find `password` field", http.StatusBadRequest)
				return
			}
			input.Login = r.Form["login"][0]
			input.Password = r.Form["password"][0]
		}
		user, err := strg.GetUser(input.Login, input.Password)
		if err != nil {
			if isApi {
				http.Error(w, "Invalid credentials", http.StatusNotFound)
			} else {
				http.Redirect(
					w,
					r,
					fmt.Sprintf("/auth/login?%s", INVAILD_CREDENTIALS_QUERY),
					http.StatusSeeOther,
				)
			}
			return
		}

		_, tokenString, err := TokenAuth.Encode(map[string]interface{}{
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
		if isApi {
			w.Write([]byte(tokenString))
		} else {
			http.Redirect(w, r, "/users/me", http.StatusSeeOther)
		}
	}
}

func GetLoginPage(w http.ResponseWriter, r *http.Request) {
	_, isInvalid := r.URL.Query()[INVAILD_CREDENTIALS_QUERY]
	tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/auth/login.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	userId, authErr := mauth.IdFromContext(r.Context())
	if authErr != nil && !errors.Is(authErr, mauth.UnauthorizedErr) {
		http.Error(w, "Auth token internal error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, map[string]interface{}{"isLogined": authErr == nil && userId > 0, "isInvalid": isInvalid}); err != nil {
		http.Error(w, "Error templating", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	c := &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
	}

	http.SetCookie(w, c)

	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}
