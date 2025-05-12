package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/save"
	"url-shortener/internal/http-server/handlers/save/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AliasLength: 8,
	}

	cases := []struct {
		name      string
		request   save.Request
		respError string
		mockError error
		status    int
	}{
		{
			name: "Success",
			request: save.Request{
				URL:   "https://example.com",
				Alias: "test_alias",
			},
			respError: "",
			mockError: nil,
			status:    http.StatusOK,
		},
		{
			name: "URL already exists",
			request: save.Request{
				URL:   "https://existing.com",
				Alias: "existing",
			},
			respError: "url already exists",
			mockError: storage.ErrURLExists,
			status:    http.StatusConflict,
		},
		{
			name: "Invalid URL",
			request: save.Request{
				URL:   "invalid-url",
				Alias: "test",
			},
			respError: "field URL is not a valid URL",
			mockError: nil,
			status:    http.StatusBadRequest,
		},
		{
			name: "Internal error",
			request: save.Request{
				URL:   "https://example.com",
				Alias: "test",
			},
			respError: "failed to add url",
			mockError: errors.New("internal error"),
			status:    http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			urlSaverMock := mocks.NewURLSaver(t)

			if tc.mockError != nil && tc.name != "Invalid URL" {
				urlSaverMock.On("SaveURL", tc.request.URL, tc.request.Alias).Return(tc.mockError).Once()
			} else if tc.mockError == nil && tc.name != "Invalid URL" {
				urlSaverMock.On("SaveURL", tc.request.URL, tc.request.Alias).Return(nil).Once()
			}

			router := gin.New()
			router.POST("/api/save", save.New(slogdiscard.NewDiscardLogger(), urlSaverMock, cfg))

			body, err := json.Marshal(tc.request)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/api/save", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(t, tc.status, rr.Code)

			if tc.respError != "" {
				require.Contains(t, rr.Body.String(), tc.respError)
			} else {
				require.Contains(t, rr.Body.String(), "alias")
			}

			urlSaverMock.AssertExpectations(t)
		})
	}
}
