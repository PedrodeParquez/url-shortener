package tests

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/save"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/random"
)

const (
	host = "localhost:8080"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	e.POST("/api/save").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(10),
		}).
		WithBasicAuth("pedro", "d123").
		Expect().
		Status(200).
		JSON().Object().
		ContainsKey("alias")
}

func TestURLShortener_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name       string
		url        string
		alias      string
		error      string
		statusCode int
	}{
		{
			name:       "Valid URL",
			url:        gofakeit.URL(),
			alias:      gofakeit.Word() + gofakeit.Word(),
			statusCode: http.StatusOK,
		},
		{
			name:       "Invalid URL",
			url:        "invalid_url",
			alias:      gofakeit.Word(),
			error:      "field URL is not a valid URL",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Empty Alias",
			url:        gofakeit.URL(),
			alias:      "",
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			resp := e.POST("/api/save").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth("pedro", "d123").
				Expect().
				Status(tc.statusCode).
				JSON().Object()

			if tc.error != "" {
				resp.NotContainsKey("alias")
				resp.Value("error").String().IsEqual(tc.error)
				return
			}

			if tc.statusCode == http.StatusOK {
				alias := tc.alias

				if tc.alias != "" {
					resp.Value("alias").String().IsEqual(tc.alias)
				} else {
					resp.Value("alias").String().NotEmpty()
					alias = resp.Value("alias").String().Raw()
				}

				testRedirect(t, alias, tc.url)
			}
		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   "/api/" + alias,
	}

	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)

	require.Equal(t, urlToRedirect, redirectedToURL)
}
