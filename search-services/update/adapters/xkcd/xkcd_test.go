package xkcd

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"yadro.com/course/update/core"
)

func TestXKCD(t *testing.T) {
	tablecase := []struct {
		name           string
		serverStatus   int
		serverResponse string
		expectedError  error
	}{
		{
			name:           "success",
			serverStatus:   200,
			serverResponse: `{"num":42,"title":"Gravity","alt":"","img":""}`,
		},
		{
			name:           "server returns 404",
			serverStatus:   404,
			serverResponse: "",
			expectedError:  core.ErrNotFound,
		},
		{
			name:           "invalid json",
			serverStatus:   200,
			serverResponse: "not a json {{{{",
			expectedError:  core.ErrNotFound,
		},
	}

	for _, tc := range tablecase {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.serverStatus)
			_, _ = w.Write([]byte(tc.serverResponse))
		}))
		defer testServer.Close()

		client, err := NewXKCDClient(testServer.URL, 10*time.Second, slog.Default())
		require.NoError(t, err)

		comics, err := client.Get(t.Context(), 42)

		if tc.expectedError != nil {
			require.ErrorIs(t, err, tc.expectedError)
		} else {
			require.NoError(t, err)
			require.Equal(t, 42, comics.ID)
		}
	}
}

func TestXKCDLastID(t *testing.T) {
	tablecase := []struct {
		name             string
		serverStatus     int
		serverResponse   string
		expectedError    error
		expectedResponse int
	}{
		{
			name:             "success",
			serverStatus:     200,
			serverResponse:   `{"num":1000}`,
			expectedResponse: 1000,
		},
		{
			name:             "server returns 404",
			serverStatus:     404,
			serverResponse:   `{"num":0}`,
			expectedResponse: 0,
			expectedError:    core.ErrNotFound,
		},
	}

	for _, tc := range tablecase {
		t.Run(tc.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.serverStatus)
				_, _ = w.Write([]byte(tc.serverResponse))
			}))
			defer testServer.Close()

			client, err := NewXKCDClient(testServer.URL, 10*time.Second, slog.Default())
			require.NoError(t, err)

			comics, err := client.LastID(t.Context())

			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, comics, tc.expectedResponse)
			}
		})
	}
}
