package middleware_test

import (
	"ai-styling-engine/api/middleware"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware_ValidCookie_CallsNext(t *testing.T) {
	called := false
	handler := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, ok := middleware.SessionFromContext(r.Context())
		if !ok {
			t.Error("expected ShopperSession in context after valid auth")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestAuthMiddleware_NoCookie_Returns403(t *testing.T) {
	// TC-SEC-1
	handler := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without a valid session cookie")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestAuthMiddleware_EmptyCookieValue_Returns403(t *testing.T) {
	// TC-SEC-2
	handler := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with empty session cookie")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: ""})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

// ── Phase 8 gap tests ────────────────────────────────────────────────────────

// TC-SEC-1b: 403 response body contains {"error":"UNAUTHENTICATED"}
func TestAuthMiddleware_NoCookie_ResponseBodyIsUnauthenticated(t *testing.T) {
	handler := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["error"] != "UNAUTHENTICATED" {
		t.Errorf("expected error=UNAUTHENTICATED, got %v", body["error"])
	}
}

// TC-SEC-2b: After valid auth, SessionFromContext returns ShopperSession with matching token
func TestAuthMiddleware_ValidCookie_SessionTokenMatchesCookie(t *testing.T) {
	handler := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := middleware.SessionFromContext(r.Context())
		if !ok {
			t.Error("expected ShopperSession in context")
			return
		}
		if session.SessionToken != "my-test-token" {
			t.Errorf("expected token=my-test-token, got %q", session.SessionToken)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "my-test-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
}

// TC-SEC-13/14: LoggingMiddleware sensitive field masking
// NOTE: LoggingMiddleware writes directly to the standard logger (log.Printf) with no
// injected writer, so log output cannot be captured in unit tests without refactoring.
// These cases are documented here as a known gap requiring a future refactor to inject
// an io.Writer into LoggingMiddleware before they can be asserted programmatically.
