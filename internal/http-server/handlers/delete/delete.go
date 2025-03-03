package delete

import (
	"errors"
	"log/slog"
	"net/http"

	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/gin-gonic/gin"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=AliasRemover
type AliasRemover interface {
	DeleteAlias(alias string) error
}

func Delete(log *slog.Logger, aliasRemover AliasRemover) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.url.delete"

		requestID := c.GetString("request_id")
		alias := c.Param("alias")

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", requestID),
			slog.String("alias", alias),
		)

		if alias == "" {
			log.Error("empty alias in URL")
			c.JSON(http.StatusBadRequest, resp.Error("invalid request"))
			return
		}

		err := aliasRemover.DeleteAlias(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("alias not found")
			c.JSON(http.StatusNotFound, resp.Error("alias not found"))
			return
		}
		if errors.Is(err, storage.ErrDBConnection) {
			log.Error("database connection error", sl.Err(err))
			c.JSON(http.StatusBadRequest, resp.Error("database connection error"))
			return
		}
		if err != nil {
			log.Error("failed to delete alias", sl.Err(err))
			c.JSON(http.StatusInternalServerError, resp.Error("failed to delete alias"))
			return
		}

		log.Info("alias deleted")

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "alias deleted",
		})
	}
}
