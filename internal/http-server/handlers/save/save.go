package save

import (
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/config"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) error
}

func New(log *slog.Logger, urlSaver URLSaver, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", c.Request.Header.Get("X-Request-Id")),
		)

		var req Request

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			c.JSON(http.StatusBadRequest, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			c.JSON(http.StatusBadRequest, resp.ValidationError(validateErr))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(cfg.AliasLength)
		}

		err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))
			c.JSON(http.StatusConflict, resp.Error("url already exists"))
			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))
			c.JSON(http.StatusInternalServerError, resp.Error("failed to add url"))
			return
		}

		log.Info("url added")

		c.JSON(http.StatusOK, Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}
