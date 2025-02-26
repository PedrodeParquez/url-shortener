package redirect_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/redirect/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"
)

func TestSaveHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
		status    int
	}{
		{
			name:      "Success",
			alias:     "test_alias",
			url:       "https://www.google.com/",
			respError: "",
			mockError: nil,
			status:    http.StatusFound,
		},
		{
			name:      "Not Found",
			alias:     "unknown_alias",
			respError: "not found",
			mockError: storage.ErrURLNotFound,
			status:    http.StatusNotFound,
		},
		{
			name:      "Internal Error",
			alias:     "internal_error",
			respError: "internal error",
			mockError: storage.ErrDBConnection,
			status:    http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlGetterMock := mocks.NewURLGetter(t)

			urlGetterMock.On("GetURL", tc.alias).
				Return(tc.url, tc.mockError).Once()

			router := gin.Default()
			router.GET("/:alias", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))

			req, _ := http.NewRequest(http.MethodGet, "/"+tc.alias, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)

			if tc.status == http.StatusFound {
				assert.Equal(t, tc.url, rec.Header().Get("Location"))
			} else {
				require.Contains(t, rec.Body.String(), tc.respError)
			}
		})
	}
}
