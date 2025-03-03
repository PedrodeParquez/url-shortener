package main

import (
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/delete"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/save"
	middleware "url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/postgres"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.MustLoad()

	log := config.SetupLogger(cfg.Env)

	log.Info(
		"starting url-shortener",
		slog.String("env", cfg.Env),
	)

	storage, err := postgres.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed init storage", sl.Err(err))
		os.Exit(1)
	}

	log.Info(
		"connecting to PostgreSQL database",
		slog.String("storage_path", cfg.StoragePath),
	)

	gin.SetMode(cfg.GinMode)
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware(log))

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to the URL Shortener!")
	})

	api := router.Group("/api")
	{
		api.GET("/:alias", redirect.New(log, storage))

		auth := gin.BasicAuth(gin.Accounts{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		})

		apiWithAuth := api.Group("/", auth)
		{
			apiWithAuth.POST("/save", save.New(log, storage, cfg))
			apiWithAuth.DELETE("/link/:alias", delete.Delete(log, storage))
		}
	}

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}
