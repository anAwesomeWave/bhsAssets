package main

import (
	"bhsAssets/internal/config"
	"bhsAssets/internal/storage"
	"errors"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	// ✔️ config (storage path, http ip:port etc.)
	// TODO: router (chi router)
	// TODO: storage (pg crud)
	// TODO: migrations
	// TODO: http server (net/http server)
	// TODO: jwt
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
	id, err := db.CreateUser("opTimus", "password")
	if err != nil {
		if errors.Is(err, storage.ErrExists) {
			log.Fatalf("non unique %v", err)
		}
		log.Fatalf("Unhandle err %v", err)
	}
	log.Printf("Inserted, id: %d\n", id)
}
