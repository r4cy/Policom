package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	metrics "github.com/VictoriaMetrics/metrics"
)

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.code == 0 {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if rw.code == 0 {
		rw.code = statusCode
	}
	rw.ResponseWriter.WriteHeader(statusCode)
}

func WithMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, code: 200}

		next.ServeHTTP(rw, r)
		end := time.Since(start).Seconds()

		metrics.GetOrCreatePrometheusHistogram(
			fmt.Sprintf(`http_request_duration_seconds{status="%s", url="%s"}`, strconv.Itoa(rw.code), r.URL.Path),
		).Update(end)
	})
}
