package main

import (
	"bhsAssets/internal/config"
	"bhsAssets/internal/storage"
	"context"
	"github.com/go-chi/docgen"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := serv.ListenAndServe(); err != nil {
			log.Fatalln("failed to start server")
		}
	}()

	<-done

	log.Println("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serv.Shutdown(ctx); err != nil {

		log.Printf("failed to stop server: %v\n", err)
		return
	}
	log.Println("server stopped")
}
