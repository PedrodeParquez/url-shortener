package server

import (
	"log/slog"
	"net/http"
	"url-shortener/internal/config"
)

func Start(log *slog.Logger, cfg *config.Config, router http.Handler) {
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	log.Info("starting server", slog.String("address", cfg.Address))

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}
