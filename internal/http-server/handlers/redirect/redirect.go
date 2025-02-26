package redirect

import (
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/gin-gonic/gin"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", c.GetHeader("X-Request-Id")),
		)

		alias := c.Param("alias")
		if alias == "" {
			log.Info("alias is empty")
			c.JSON(http.StatusBadRequest, response.Error("invalid request"))
			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", slog.String("alias", alias))
			c.JSON(http.StatusNotFound, response.Error("not found"))
			return
		}
		if err != nil {
			log.Error("failed to get url", sl.Err(err), slog.String("alias", alias))
			c.JSON(http.StatusInternalServerError, response.Error("internal error"))
			return
		}

		log.Info("got url", slog.String("url", resURL))

		c.Redirect(http.StatusFound, resURL)
	}
}
