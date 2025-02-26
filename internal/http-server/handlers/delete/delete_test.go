package delete_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
			name:      "Invalid alias",
			alias:     "invalid_alias",
			respError: "invalid request",
			mockError: storage.ErrDBConnection,
			status:    http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockAliasRemover := mocks.NewAliasRemover(t)

			if tc.mockError != nil {
				mockAliasRemover.On("DeleteAlias", tc.alias).Return(tc.mockError).Once()
			} else {
				mockAliasRemover.On("DeleteAlias", tc.alias).Return(nil).Once()
			}

			router := gin.New()
			router.DELETE("/api/delete", delete.Delete(slogdiscard.NewDiscardLogger(), mockAliasRemover))

			input := fmt.Sprintf(`{"alias": "%s"}`, tc.alias)
			if tc.alias == "" {
				input = `{}`
			}

			req, err := http.NewRequest(http.MethodDelete, "/api/delete", bytes.NewReader([]byte(input)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(t, tc.status, rr.Code)

			var resp delete.Response
			if tc.respError != "" {
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				require.Equal(t, tc.respError, resp.Error)
			} else {
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				require.Equal(t, "alias deleted", resp.Message)
			}

			mockAliasRemover.AssertExpectations(t)
		})
	}
}
