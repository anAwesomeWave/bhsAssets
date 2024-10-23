package main

import (
	"bhsAssets/internal/http/handlers/auth"
	"bhsAssets/internal/http/handlers/users"
	midauth "bhsAssets/internal/http/middleware/auth"
	"bhsAssets/internal/http/middleware/common"
	"bhsAssets/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
)

func setUpRouter(db *storage.Storage) *chi.Mux {
	//subRouter := chi.NewRouter(
	//var userSubRouter *chi.Router
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer) // не падать при панике
	router.Use(middleware.URLFormat) // удобно брать из урлов данные
	router.Use(middleware.StripSlashes)
	router.Use(common.SetIsApiContextVariable)

	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", auth.Register(*db))
		r.Get("/register", auth.GetRegisterPage)
		r.Post("/login", auth.Login(*db))
		r.Get("/login", auth.GetLoginPage)
	})
	router.Route("/users", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(auth.TokenAuth))
			r.Use(jwtauth.Authenticator(auth.TokenAuth))
			r.Use(midauth.GetUserByJwtToken(*db))

			r.Get("/me", users.GetUserData)
		})

	})
	router.Route("/api", func(r chi.Router) {
		r.Use(common.JsonContentType)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", auth.Register(*db))
			r.Post("/login", auth.Login(*db))
		})
		r.Route("/users", func(r chi.Router) {

			r.Group(func(r chi.Router) {
				r.Use(jwtauth.Verifier(auth.TokenAuth))
				r.Use(jwtauth.Authenticator(auth.TokenAuth))
				r.Use(midauth.GetUserByJwtToken(*db))

				r.Get("/me", users.GetUserData)
			})

			r.Route("/balance", func(r chi.Router) {
				r.Use(jwtauth.Verifier(auth.TokenAuth))
				r.Use(jwtauth.Authenticator(auth.TokenAuth))

				r.Patch("/", users.UpdateBalanceInfo(*db))
			})
		})
	})
	return router
}
