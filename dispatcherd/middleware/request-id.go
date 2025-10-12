package middleware

import (
	"context"
	dispatcherContext "dispatcherd/context"
	"net/http"

	"github.com/google/uuid"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: only allow certain origins
		requestID := r.Header.Get("x-request-id")

		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), dispatcherContext.KeyRequestID, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
