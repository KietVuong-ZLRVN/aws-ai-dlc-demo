package middleware_test

import (
	"ai-styling-engine/api/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTracingMiddleware_SetsCorrelationHeader(t *testing.T) {
	// TC-SEC-6
	handler := middleware.TracingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceId := middleware.TraceIdFromContext(r.Context())
		if traceId == "" {
			t.Error("expected non-empty trace ID in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("X-Correlation-ID") == "" {
		t.Error("expected X-Correlation-ID header in response")
	}
}

func TestTracingMiddleware_ForwardsExistingCorrelationId(t *testing.T) {
	handler := middleware.TracingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceId := middleware.TraceIdFromContext(r.Context())
		if traceId != "existing-trace-id" {
			t.Errorf("expected trace ID to be forwarded, got %q", traceId)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "existing-trace-id")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
}

// TC-SEC-6b: X-Correlation-ID is set on response even when no incoming header present
func TestTracingMiddleware_NoIncomingHeader_GeneratesCorrelationId(t *testing.T) {
	handler := middleware.TracingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Explicitly no X-Correlation-ID header set.
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	correlationId := rr.Header().Get("X-Correlation-ID")
	if correlationId == "" {
		t.Error("expected X-Correlation-ID to be generated and set on response when not provided")
	}
}
