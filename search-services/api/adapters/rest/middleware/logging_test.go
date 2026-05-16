package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"yadro.com/course/api/core"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestLoggingUsesTraceIDFromHeader(t *testing.T) {
	expectedTraceID := "trace-123"

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := core.TraceIDFromContext(r.Context())
		require.Equal(t, expectedTraceID, traceID)

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	req.Header.Set("trace_id", expectedTraceID)
	rec := httptest.NewRecorder()

	handler := Logging(next, testLogger())
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestLoggingGeneratesTraceID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := core.TraceIDFromContext(r.Context())
		require.NotEmpty(t, traceID)

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	rec := httptest.NewRecorder()

	handler := Logging(next, testLogger())
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
