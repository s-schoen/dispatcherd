package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type trackingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *trackingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracker := trackingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			startTime := time.Now()
			defer func() {
				logger.InfoContext(r.Context(), "access",
					"src", r.RemoteAddr,
					"status", tracker.statusCode,
					"method", r.Method,
					"path", r.URL.Path,
					"time", time.Since(startTime),
				)
			}()
			next.ServeHTTP(&tracker, r)
		})
	}
}
