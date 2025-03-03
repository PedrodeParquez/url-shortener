package main

import (
	"log/slog"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/routes"
	"url-shortener/internal/http-server/server"
	"url-shortener/internal/storage/postgres"
)

func main() {
	cfg := config.MustLoad()
	log := config.SetupLogger(cfg.Env)

	log.Info("starting url-shortener", slog.String("env", cfg.Env))

	storage, err := postgres.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed init storage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("connecting to PostgreSQL database", slog.String("storage_path", cfg.StoragePath))
	
	router := routes.SetupRouter(log, storage, cfg)

	server.Start(log, cfg, router)
}
