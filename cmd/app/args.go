package main

import "flag"

type AppArgs struct {
	configPath     *string
	storageEnvPath *string
}

func parseFlags() *AppArgs {
	confPath := flag.String("confPath", "./config/local.yaml", "path to config file")
	storageEnv := flag.String(
		"storageEnv",
		"./config/.storage_env",
		"path to env file with postgres attributes (POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB)",
	)
	flag.Parse()
	return &AppArgs{configPath: confPath, storageEnvPath: storageEnv}
}
