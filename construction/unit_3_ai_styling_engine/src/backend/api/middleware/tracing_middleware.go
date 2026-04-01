package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// TracingMiddleware generates a per-request correlation ID (StyleSessionId) and attaches
// it to the request context and response headers for distributed tracing.
func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceId := r.Header.Get("X-Correlation-ID")
		if traceId == "" {
			traceId = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), ContextKeyTraceId, traceId)
		w.Header().Set("X-Correlation-ID", traceId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TraceIdFromContext retrieves the trace correlation ID from context.
func TraceIdFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(ContextKeyTraceId).(string); ok {
		return id
	}
	return ""
}
