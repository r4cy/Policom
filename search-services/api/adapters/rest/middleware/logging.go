package middleware

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"yadro.com/course/api/core"
)

func Logging(next http.Handler, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logHeadertrace := r.Header.Get("trace_id")
		if logHeadertrace == "" {
			logHeadertrace = uuid.New().String()
		}
		ctx := core.ContextWithTraceID(r.Context(), logHeadertrace)
		r = r.WithContext(ctx)
		log.InfoContext(ctx, "request", "trace_id", logHeadertrace, "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	}
}
