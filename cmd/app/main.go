package main

import (
	"bhsAssets/internal/config"
	"fmt"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	// ✔️ config (storage path, http ip:port etc.)
	// TODO: router (chi router)
	// TODO: storage (pg crud)
	// TODO: http server (net/http server)
	// TODO: jwt
	args := *parseFlags()
	if err := godotenv.Load(*args.storageEnvPath); err != nil {
		log.Fatalf("Error with loading StorageEnv file: %v", err)
	}
	appConfig := config.Load(*args.configPath)

	fmt.Println(appConfig)
}
