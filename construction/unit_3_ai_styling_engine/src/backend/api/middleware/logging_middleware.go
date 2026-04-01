package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs each request's method, path, trace ID, status, and latency.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		traceId := TraceIdFromContext(r.Context())
		log.Printf("[REQ] %s %s trace=%s", r.Method, r.URL.Path, traceId)

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		log.Printf("[RES] %s %s trace=%s status=%d latency=%dms",
			r.Method, r.URL.Path, traceId, rw.statusCode, time.Since(start).Milliseconds())
	})
}
