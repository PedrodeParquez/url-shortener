package delete_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"url-shortener/internal/http-server/handlers/delete"
	"url-shortener/internal/http-server/handlers/delete/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDeleteHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name      string
		alias     string
		respError string
		mockError error
		status    int
	}{
		{
			name:      "Success",
			alias:     "test_alias",
			respError: "",
			mockError: nil,
			status:    http.StatusOK,
		},
		{
			name:      "Alias not found",
			alias:     "non_existent_alias",
			respError: "alias not found",
			mockError: storage.ErrURLNotFound,
			status:    http.StatusNotFound,
		},
		{
			name:      "Internal error",
			alias:     "some_alias",
			respError: "failed to delete alias",
			mockError: errors.New("internal error"),
			status:    http.StatusInternalServerError,
		},
		{
			name:      "Empty alias",
			alias:     "",
			respError: "404 page not found",
			mockError: nil,
			status:    http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockAliasRemover := mocks.NewAliasRemover(t)

			if tc.alias != "" && tc.mockError != nil {
				mockAliasRemover.On("DeleteAlias", tc.alias).Return(tc.mockError).Once()
			} else if tc.alias != "" {
				mockAliasRemover.On("DeleteAlias", tc.alias).Return(nil).Once()
			}

			router := gin.New()
			router.DELETE("/api/link/:alias", delete.Delete(slogdiscard.NewDiscardLogger(), mockAliasRemover))

			req, err := http.NewRequest(http.MethodDelete, "/api/link/"+tc.alias, nil)
			if tc.name == "Empty alias" {
				req, err = http.NewRequest(http.MethodDelete, "/api/link/", nil)
			}
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(t, tc.status, rr.Code)

			if tc.respError != "" {
				require.Contains(t, rr.Body.String(), tc.respError)
			} else {
				require.Contains(t, rr.Body.String(), "alias deleted")
			}

			mockAliasRemover.AssertExpectations(t)
		})
	}
}
