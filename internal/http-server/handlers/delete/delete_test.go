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

	"github.com/stretchr/testify/require"
)

func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name           string
		alias          string
		respError      string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Success",
			alias:          "test_alias",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Alias not found",
			alias:          "non_existent_alias",
			respError:      "alias not found",
			mockError:      storage.ErrURLNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Internal error",
			alias:          "some_alias",
			respError:      "failed to delete alias",
			mockError:      errors.New("internal error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid alias",
			alias:          "invalid alias",
			respError:      "invalid request",
			expectedStatus: http.StatusBadRequest,
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

			handler := delete.Delete(slogdiscard.NewDiscardLogger(), mockAliasRemover)

			input := fmt.Sprintf(`{"alias": "%s"}`, tc.alias)
			if tc.alias == "" {
				input = `{}`
			}

			req, err := http.NewRequest(http.MethodDelete, "/url/delete", bytes.NewReader([]byte(input)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code)

			var resp delete.Response
			if tc.respError != "" {
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				require.Equal(t, tc.respError, resp.Message)
			} else {
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				require.Equal(t, "alias deleted", resp.Message)
			}

			mockAliasRemover.AssertExpectations(t)
		})
	}
}
