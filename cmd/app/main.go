package main

import (
	"bhsAssets/internal/config"
	"bhsAssets/internal/http/handlers/auth"
	"bhsAssets/internal/http/handlers/users"
	midauth "bhsAssets/internal/http/middleware/auth"
	"bhsAssets/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func main() {
	// ✔️ config (storage path, http ip:port etc.)
	// TODO: router (chi router) ✔️
	// TODO: storage (pg crud) ✔️
	// TODO: migrations ✔️
	// TODO: http server (net/http server) ✔️
	// TODO: jwt ✔️
	// TODO: jwt expiration
	args := *parseFlags()
	if err := godotenv.Load(*args.storageEnvPath); err != nil {
		log.Fatalf("Error with loading StorageEnv file: %v", err)
	}
	appConfig := config.Load(*args.configPath)

	db, err := storage.New(appConfig.StorageCfg)

	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("db opened")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer) // не падать при панике
	router.Use(middleware.URLFormat) // удобно брать из урлов данные
	router.Use(middleware.StripSlashes)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	router.Route("/users", func(r chi.Router) {
		r.Post("/auth", auth.Register(*db))
		r.Post("/login", auth.Login(*db))

		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(auth.TokenAuth))
			r.Use(jwtauth.Authenticator(auth.TokenAuth))
			r.Use(midauth.GetUserByJwtToken(*db))

			r.Get("/me", users.GetUserData)
		})
	})
	serv := &http.Server{
		Addr:         appConfig.HTTPServerCfg.Address,
		Handler:      router,
		ReadTimeout:  appConfig.HTTPServerCfg.Timeout,
		WriteTimeout: appConfig.HTTPServerCfg.Timeout,
		IdleTimeout:  appConfig.HTTPServerCfg.IdleTimeout,
	}

	log.Println("Starting server...")
	log.Printf("%v\n", router.Routes())
	if err := serv.ListenAndServe(); err != nil {
		log.Fatalln("failed to start server")
	}
}
