package routes

import (
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/delete"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/save"
	middleware "url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/storage/postgres"

	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(log *slog.Logger, storage *postgres.Storage, cfg *config.Config) *gin.Engine {
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

	return router
}
