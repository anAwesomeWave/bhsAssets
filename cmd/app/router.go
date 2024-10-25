package main

import (
	"bhsAssets/internal/http/handlers/assets"
	"bhsAssets/internal/http/handlers/auth"
	"bhsAssets/internal/http/handlers/site"
	"bhsAssets/internal/http/handlers/users"
	midauth "bhsAssets/internal/http/middleware/auth"
	"bhsAssets/internal/http/middleware/common"
	"bhsAssets/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"net/http"
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

	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.NotFound(site.NotFoundHandler)

	router.Group(func(router chi.Router) {
		// аутентификация
		router.Use(jwtauth.Verifier(auth.TokenAuth))
		router.Use(midauth.GetUserByJwtToken(*db))

		router.Route("/auth", func(r chi.Router) {
			r.Post("/register", auth.Register(*db))
			r.Get("/register", auth.GetRegisterPage)
			r.Post("/login", auth.Login(*db))
			r.Get("/login", auth.GetLoginPage)
			r.Post("/logout", auth.Logout)
		})
		router.Route("/users", func(r chi.Router) {

			r.Use(midauth.CustomAuthenticator(auth.TokenAuth))
			r.Group(func(r chi.Router) {
				r.Get("/me", users.GetUserData)
			})
		})
		router.Route("/assets", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				//r.Use(jwtauth.Verifier(auth.TokenAuth))
				//r.Use(jwtauth.Authenticator(auth.TokenAuth))
				//r.Use(midauth.GetUserByJwtToken(*db))

				r.Get("/create", assets.GetAssetsCreationPage)
				r.Post("/create", assets.CreateAsset(*db))
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
					//r.Use(jwtauth.Verifier(auth.TokenAuth))
					r.Use(midauth.CustomAuthenticator(auth.TokenAuth))
					//r.Use(midauth.GetUserByJwtToken(*db))

					r.Get("/me", users.GetUserData)
				})

				r.Route("/balance", func(r chi.Router) {
					//r.Use(jwtauth.Verifier(auth.TokenAuth))
					r.Use(midauth.CustomAuthenticator(auth.TokenAuth))

					r.Patch("/", users.UpdateBalanceInfo(*db))
				})
			})
		})
	})
	return router
}
