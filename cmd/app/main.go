package main

import (
	"bhsAssets/internal/config"
	"bhsAssets/internal/storage"
	"github.com/go-chi/docgen"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func main() {
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

	router := setUpRouter(db)

	serv := &http.Server{
		Addr:         appConfig.HTTPServerCfg.Address,
		Handler:      router,
		ReadTimeout:  appConfig.HTTPServerCfg.Timeout,
		WriteTimeout: appConfig.HTTPServerCfg.Timeout,
		IdleTimeout:  appConfig.HTTPServerCfg.IdleTimeout,
	}
	log.Println("Server routes")
	docgen.PrintRoutes(router)
	log.Println("Starting server...")
	if err := serv.ListenAndServe(); err != nil {
		log.Fatalln("failed to start server")
	}
}
