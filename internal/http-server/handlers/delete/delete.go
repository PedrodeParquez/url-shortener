package delete

import (
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Alias string `json:"alias" validate:"required,alias"`
}

type Response struct {
	resp.Response
	Message string `json:"message,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=AliasRemover
type AliasRemover interface {
	DeleteAlias(alias string) error
}

func Delete(log *slog.Logger, aliasRemover AliasRemover) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.url.delete"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", c.GetHeader("X-Request-Id")),
		)

		var req Request

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			c.JSON(http.StatusBadRequest, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if req.Alias == "" {
			c.JSON(http.StatusBadRequest, resp.Error("invalid request"))
			return
		}

		validate := validator.New()
		validate.RegisterValidation("alias", validateAlias)

		if err := validate.Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))
			c.JSON(http.StatusBadRequest, resp.ValidationError(validateErr))
			return
		}

		err := aliasRemover.DeleteAlias(req.Alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("alias not found", slog.String("alias", req.Alias))
			c.JSON(http.StatusNotFound, resp.Error("alias not found"))
			return
		}
		if errors.Is(err, storage.ErrDBConnection) {
			log.Error("database connection error", sl.Err(err))
			c.JSON(http.StatusBadRequest, resp.Error("invalid request"))
			return
		}
		if err != nil {
			log.Error("failed to delete alias", sl.Err(err))
			c.JSON(http.StatusInternalServerError, resp.Error("failed to delete alias"))
			return
		}
		log.Info("alias deleted")
		c.JSON(http.StatusOK, Response{
			Response: resp.OK(),
			Message:  "alias deleted",
		})
	}
}

func validateAlias(fl validator.FieldLevel) bool {
	alias := fl.Field().String()
	regex := `^[a-zA-Z0-9-_]+$`
	match, _ := regexp.MatchString(regex, alias)
	return match
}
