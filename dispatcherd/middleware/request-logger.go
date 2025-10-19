package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type RequestLoggerMiddleware struct {
	logger *slog.Logger
}

func NewRequestLoggerMiddleware(logger *slog.Logger) *RequestLoggerMiddleware {
	return &RequestLoggerMiddleware{
		logger: logger,
	}
}

type trackingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *trackingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (h *RequestLoggerMiddleware) OnRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracker := trackingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		startTime := time.Now()
		defer func() {
			src := r.RemoteAddr
			if r.Header.Get("X-Forwarded-For") != "" {
				src = r.Header.Get("X-Forwarded-For")
			}

			h.logger.InfoContext(r.Context(), "access",
				"src", src,
				"status", tracker.statusCode,
				"method", r.Method,
				"path", r.URL.Path,
				"time", time.Since(startTime),
			)
		}()
		next.ServeHTTP(&tracker, r)
	})
}
