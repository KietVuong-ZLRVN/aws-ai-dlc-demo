package api_test

import (
	api "ai-styling-engine/api"
	"ai-styling-engine/api/controllers"
	"ai-styling-engine/application/usecases"
	infraACL "ai-styling-engine/infrastructure/acl"
	infraAI "ai-styling-engine/infrastructure/ai"
	"ai-styling-engine/infrastructure/dispatcher"
	"ai-styling-engine/domain/events"
	"ai-styling-engine/domain/policies"
	"ai-styling-engine/domain/services"
	"ai-styling-engine/domain/valueobjects"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ── E2E helpers ───────────────────────────────────────────────────────────────

func newE2ERouter() http.Handler {
	d := dispatcher.NewInProcessEventDispatcher()
	d.Register(events.EventTypeComboGenerationRequested, policies.NewPreferenceDefaultPolicy().Handle)
	d.Register(events.EventTypeWishlistFetchCompleted, policies.NewWishlistSupplementationPolicy(d).Handle)
	d.Register(events.EventTypeFallbackTriggered, policies.NewFallbackPolicy().Handle)
	d.Register(events.EventTypeCombosGenerated, policies.NewComboExclusionPolicy().Handle)

	generateUC := usecases.NewGenerateCombosUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraACL.NewInMemoryProductCatalogRepository(),
		infraACL.NewInMemoryCompleteLookRepository(),
		infraAI.NewMockComboCompatibilityScoringService(),
		infraAI.NewMockComboReasoningGenerationService(),
		d,
	)

	d2 := dispatcher.NewInProcessEventDispatcher()
	confirmUC := usecases.NewConfirmPreferencesUseCase(infraAI.NewMockPreferenceInterpretationService(), d2)
	optionsUC := usecases.NewGetPreferenceOptionsUseCase()

	return api.NewRouter(
		controllers.NewStylePreferencesController(optionsUC),
		controllers.NewPreferenceConfirmationController(confirmUC),
		controllers.NewComboGenerationController(generateUC),
	)
}

func newE2ERouterWithFallbackScoring() http.Handler {
	d := dispatcher.NewInProcessEventDispatcher()
	d.Register(events.EventTypeComboGenerationRequested, policies.NewPreferenceDefaultPolicy().Handle)
	d.Register(events.EventTypeWishlistFetchCompleted, policies.NewWishlistSupplementationPolicy(d).Handle)
	d.Register(events.EventTypeFallbackTriggered, policies.NewFallbackPolicy().Handle)
	d.Register(events.EventTypeCombosGenerated, policies.NewComboExclusionPolicy().Handle)

	generateUC := usecases.NewGenerateCombosUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraACL.NewInMemoryProductCatalogRepository(),
		infraACL.NewInMemoryCompleteLookRepository(),
		&e2eAlwaysFallbackScorer{},
		infraAI.NewMockComboReasoningGenerationService(),
		d,
	)

	d2 := dispatcher.NewInProcessEventDispatcher()
	confirmUC := usecases.NewConfirmPreferencesUseCase(infraAI.NewMockPreferenceInterpretationService(), d2)
	optionsUC := usecases.NewGetPreferenceOptionsUseCase()

	return api.NewRouter(
		controllers.NewStylePreferencesController(optionsUC),
		controllers.NewPreferenceConfirmationController(confirmUC),
		controllers.NewComboGenerationController(generateUC),
	)
}

func decodeE2E(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return body
}

// ── TC-E2E-1: Quick-generate success — full response shape ────────────────────

func TestE2E_QuickGenerate_FullResponseShape(t *testing.T) {
	router := newE2ERouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-token"})
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	body := decodeE2E(t, rr)
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", body["status"])
	}
	combos, ok := body["combos"].([]interface{})
	if !ok || len(combos) == 0 {
		t.Fatal("expected non-empty combos array")
	}
	for i, c := range combos {
		combo := c.(map[string]interface{})
		if combo["id"] == "" || combo["id"] == nil {
			t.Errorf("combo[%d] missing id", i)
		}
		if combo["reasoning"] == "" || combo["reasoning"] == nil {
			t.Errorf("combo[%d] missing reasoning", i)
		}
		items, _ := combo["items"].([]interface{})
		if len(items) < 2 {
			t.Errorf("combo[%d] has fewer than 2 items", i)
		}
		for j, it := range items {
			item := it.(map[string]interface{})
			src, _ := item["source"].(string)
			if src != "wishlist" && src != "catalog" {
				t.Errorf("combo[%d].items[%d] has invalid source %q", i, j, src)
			}
		}
	}
}

// ── TC-E2E-2: Preference-guided success ──────────────────────────────────────

func TestE2E_PreferenceGuided_Returns200Ok(t *testing.T) {
	router := newE2ERouter()

	body := `{"preferences":{"occasions":["casual"],"styles":["minimalist"],"budget":{"min":0,"max":300}}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-token"})
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	resp := decodeE2E(t, rr)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", resp["status"])
	}
}

// ── TC-E2E-3: Preference confirmation flow ────────────────────────────────────

func TestE2E_PreferenceConfirmation_SummaryAndEcho(t *testing.T) {
	router := newE2ERouter()

	body := `{"occasions":["casual","beach"],"styles":["minimalist"],"budget":{"min":50,"max":200},"freeText":"summer trip"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/preferences/confirm", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-token"})
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	resp := decodeE2E(t, rr)
	if summary, _ := resp["summary"].(string); summary == "" {
		t.Error("expected non-empty summary")
	}
	prefs, ok := resp["preferences"].(map[string]interface{})
	if !ok {
		t.Fatal("expected preferences object in response")
	}
	if prefs["freeText"] != "summer trip" {
		t.Errorf("expected freeText echoed, got %v", prefs["freeText"])
	}
}

// ── TC-E2E-4: Unauthenticated request → 403 ───────────────────────────────────

func TestE2E_Unauthenticated_Returns403(t *testing.T) {
	router := newE2ERouter()

	endpoints := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/v1/style/preferences/options", ""},
		{http.MethodPost, "/api/v1/style/preferences/confirm", "{}"},
		{http.MethodPost, "/api/v1/style/combos/generate", "{}"},
	}

	for _, ep := range endpoints {
		var bodyReader *strings.Reader
		if ep.body != "" {
			bodyReader = strings.NewReader(ep.body)
		} else {
			bodyReader = strings.NewReader("")
		}
		req := httptest.NewRequest(ep.method, ep.path, bodyReader)
		req.Header.Set("Content-Type", "application/json")
		// No session cookie.
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("%s %s: expected 403, got %d", ep.method, ep.path, rr.Code)
		}
		resp := decodeE2E(t, rr)
		if resp["error"] != "UNAUTHENTICATED" {
			t.Errorf("%s %s: expected UNAUTHENTICATED, got %v", ep.method, ep.path, resp["error"])
		}
	}
}

// ── TC-E2E-5: Fallback path ───────────────────────────────────────────────────

func TestE2E_FallbackPath_StatusFallbackWithAlternatives(t *testing.T) {
	router := newE2ERouterWithFallbackScoring()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-token"})
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	resp := decodeE2E(t, rr)
	if resp["status"] != "fallback" {
		t.Fatalf("expected status=fallback, got %v", resp["status"])
	}
	if msg, _ := resp["message"].(string); msg == "" {
		t.Error("expected non-empty message in fallback response")
	}
	alts, _ := resp["alternatives"].([]interface{})
	if len(alts) == 0 {
		t.Error("expected non-empty alternatives in fallback response")
	}
}

// ── TC-E2E-6: Combo exclusion across two calls ────────────────────────────────

func TestE2E_ComboExclusion_ExcludedIdsAbsentFromSecondResponse(t *testing.T) {
	router := newE2ERouter()

	// First call — get combo IDs.
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req1.Header.Set("Content-Type", "application/json")
	req1.AddCookie(&http.Cookie{Name: "session", Value: "test-token"})
	rr1 := httptest.NewRecorder()
	router.ServeHTTP(rr1, req1)

	resp1 := decodeE2E(t, rr1)
	combos1, ok := resp1["combos"].([]interface{})
	if !ok || len(combos1) == 0 {
		t.Skip("no combos returned from first request")
	}
	firstId := combos1[0].(map[string]interface{})["id"].(string)

	// Second call — exclude the first combo ID.
	body2 := `{"excludeComboIds":["` + firstId + `"]}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: "session", Value: "test-token"})
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
	resp2 := decodeE2E(t, rr2)
	if combos2, ok := resp2["combos"].([]interface{}); ok {
		for _, c := range combos2 {
			id := c.(map[string]interface{})["id"].(string)
			if id == firstId {
				t.Errorf("excluded combo %q still appeared in second response", firstId)
			}
		}
	}
}

// ── E2E stubs ─────────────────────────────────────────────────────────────────

type e2eAlwaysFallbackScorer struct{}

func (s *e2eAlwaysFallbackScorer) Score(_ services.ScoringInput) (services.ScoringResult, error) {
	return services.ScoringResult{
		Fallback: &services.ScoringFallback{
			Message: "No suitable combo could be formed.",
			Alternatives: []valueobjects.AlternativeItem{
				{
					ConfigSku: "CFG-ALT-1",
					SimpleSku: "SKU-ALT-1",
					Name:      "Classic White Shirt",
					Brand:     "Uniqlo",
					Price:     39.90,
					ImageUrl:  "https://example.com/shirt.jpg",
					Reason:    "A versatile piece that pairs with most items",
				},
			},
		},
	}, nil
}

// Ensure unused imports don't cause compile errors.
var _ = errors.New
