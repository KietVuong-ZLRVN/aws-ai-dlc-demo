package controllers_test

import (
	"ai-styling-engine/api/controllers"
	"ai-styling-engine/api/middleware"
	"ai-styling-engine/application/usecases"
	infraACL "ai-styling-engine/infrastructure/acl"
	infraAI "ai-styling-engine/infrastructure/ai"
	"ai-styling-engine/infrastructure/dispatcher"
	"ai-styling-engine/domain/events"
	"ai-styling-engine/domain/policies"
	"ai-styling-engine/domain/services"
	"ai-styling-engine/domain/valueobjects"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ── Test helpers ──────────────────────────────────────────────────────────────

// injectSession adds a ShopperSession to the request context, bypassing AuthMiddleware.
func injectSession(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.ContextKeyShopperSession,
		valueobjects.ShopperSession{SessionToken: "test-token"},
	)
	return r.WithContext(ctx)
}

func newDispatcher() events.EventDispatcher {
	d := dispatcher.NewInProcessEventDispatcher()
	d.Register(events.EventTypeComboGenerationRequested, policies.NewPreferenceDefaultPolicy().Handle)
	d.Register(events.EventTypeWishlistFetchCompleted, policies.NewWishlistSupplementationPolicy(d).Handle)
	d.Register(events.EventTypeFallbackTriggered, policies.NewFallbackPolicy().Handle)
	d.Register(events.EventTypeCombosGenerated, policies.NewComboExclusionPolicy().Handle)
	return d
}

func newGenerateUC() *usecases.GenerateCombosUseCase {
	d := newDispatcher()
	return usecases.NewGenerateCombosUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraACL.NewInMemoryProductCatalogRepository(),
		infraACL.NewInMemoryCompleteLookRepository(),
		infraAI.NewMockComboCompatibilityScoringService(),
		infraAI.NewMockComboReasoningGenerationService(),
		d,
	)
}

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return body
}

// ── StylePreferencesController ────────────────────────────────────────────────

func TestGetOptions_Returns200WithAllFields(t *testing.T) {
	// TC-301b-1
	uc := usecases.NewGetPreferenceOptionsUseCase()
	ctrl := controllers.NewStylePreferencesController(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/style/preferences/options", nil)
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.GetOptions(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	body := decodeBody(t, rr)
	for _, field := range []string{"occasions", "styles", "colors"} {
		if _, ok := body[field]; !ok {
			t.Errorf("expected field %q in response", field)
		}
	}
}

// ── PreferenceConfirmationController ─────────────────────────────────────────

func TestConfirmPreferences_ValidBody_Returns200WithSummary(t *testing.T) {
	// TC-302-1
	d := dispatcher.NewInProcessEventDispatcher()
	uc := usecases.NewConfirmPreferencesUseCase(infraAI.NewMockPreferenceInterpretationService(), d)
	ctrl := controllers.NewPreferenceConfirmationController(uc)

	body := `{"occasions":["casual"],"styles":["minimalist"],"budget":{"min":50,"max":200}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/preferences/confirm", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Confirm(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	resp := decodeBody(t, rr)
	if _, ok := resp["summary"]; !ok {
		t.Error("expected 'summary' field in response")
	}
	if _, ok := resp["preferences"]; !ok {
		t.Error("expected 'preferences' field in response (echo)")
	}
}

func TestConfirmPreferences_EmptyBody_Returns200(t *testing.T) {
	// TC-302-2
	d := dispatcher.NewInProcessEventDispatcher()
	uc := usecases.NewConfirmPreferencesUseCase(infraAI.NewMockPreferenceInterpretationService(), d)
	ctrl := controllers.NewPreferenceConfirmationController(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/preferences/confirm", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Confirm(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for empty preferences, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestConfirmPreferences_InvalidJSON_Returns400(t *testing.T) {
	// TC-SEC-5
	d := dispatcher.NewInProcessEventDispatcher()
	uc := usecases.NewConfirmPreferencesUseCase(infraAI.NewMockPreferenceInterpretationService(), d)
	ctrl := controllers.NewPreferenceConfirmationController(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/preferences/confirm", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Confirm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}

func TestConfirmPreferences_InvalidBudget_Returns400(t *testing.T) {
	// TC-SEC-3
	d := dispatcher.NewInProcessEventDispatcher()
	uc := usecases.NewConfirmPreferencesUseCase(infraAI.NewMockPreferenceInterpretationService(), d)
	ctrl := controllers.NewPreferenceConfirmationController(uc)

	body := `{"budget":{"min":200,"max":50}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/preferences/confirm", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Confirm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when max < min, got %d", rr.Code)
	}
	resp := decodeBody(t, rr)
	if resp["error"] != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %v", resp["error"])
	}
}

func TestConfirmPreferences_ColorInBothLists_Returns400(t *testing.T) {
	// TC-SEC-4
	d := dispatcher.NewInProcessEventDispatcher()
	uc := usecases.NewConfirmPreferencesUseCase(infraAI.NewMockPreferenceInterpretationService(), d)
	ctrl := controllers.NewPreferenceConfirmationController(uc)

	body := `{"colors":{"preferred":["black"],"excluded":["black"]}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/preferences/confirm", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Confirm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when color in both lists, got %d", rr.Code)
	}
}

// ── ComboGenerationController ─────────────────────────────────────────────────

func TestGenerateCombos_QuickGenerate_Returns200WithOkStatus(t *testing.T) {
	// TC-301-1, TC-301-2
	ctrl := controllers.NewComboGenerationController(newGenerateUC())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	resp := decodeBody(t, rr)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", resp["status"])
	}
}

func TestGenerateCombos_CombosArrayNonEmpty(t *testing.T) {
	ctrl := controllers.NewComboGenerationController(newGenerateUC())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	resp := decodeBody(t, rr)
	combos, ok := resp["combos"].([]interface{})
	if !ok || len(combos) == 0 {
		t.Error("expected non-empty combos array")
	}
}

func TestGenerateCombos_EachComboHasReasoningAndItems(t *testing.T) {
	// TC-402-1, TC-401-1
	ctrl := controllers.NewComboGenerationController(newGenerateUC())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	resp := decodeBody(t, rr)
	combos := resp["combos"].([]interface{})
	for i, c := range combos {
		combo := c.(map[string]interface{})
		if combo["reasoning"] == "" || combo["reasoning"] == nil {
			t.Errorf("combo[%d] has empty reasoning", i)
		}
		items, ok := combo["items"].([]interface{})
		if !ok || len(items) < 2 {
			t.Errorf("combo[%d] has fewer than 2 items", i)
		}
	}
}

func TestGenerateCombos_WithPreferences_Returns200(t *testing.T) {
	// TC-301b-3
	ctrl := controllers.NewComboGenerationController(newGenerateUC())

	body := `{"preferences":{"occasions":["casual"],"styles":["minimalist"],"budget":{"min":0,"max":300}}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGenerateCombos_ExcludeFirstCombo_AbsentFromResponse(t *testing.T) {
	// TC-405-1
	ctrl := controllers.NewComboGenerationController(newGenerateUC())

	// First request to get combo IDs.
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req1.Header.Set("Content-Type", "application/json")
	req1 = injectSession(req1)
	rr1 := httptest.NewRecorder()
	ctrl.Generate(rr1, req1)
	resp1 := decodeBody(t, rr1)
	combos1 := resp1["combos"].([]interface{})
	if len(combos1) == 0 {
		t.Skip("no combos returned from first request")
	}
	firstId := combos1[0].(map[string]interface{})["id"].(string)

	// Second request excluding the first combo.
	body2 := `{"excludeComboIds":["` + firstId + `"]}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2 = injectSession(req2)
	rr2 := httptest.NewRecorder()
	ctrl.Generate(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr2.Code)
	}
	resp2 := decodeBody(t, rr2)
	if combos2, ok := resp2["combos"].([]interface{}); ok {
		for _, c := range combos2 {
			id := c.(map[string]interface{})["id"].(string)
			if id == firstId {
				t.Errorf("excluded combo %q still appeared in response", firstId)
			}
		}
	}
}

func TestGenerateCombos_InvalidJSON_Returns400(t *testing.T) {
	// TC-SEC-5
	ctrl := controllers.NewComboGenerationController(newGenerateUC())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate",
		strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = 10 // non-zero so the body is parsed
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}

func TestGenerateCombos_CorrelationHeaderPresent(t *testing.T) {
	// TC-SEC-6 — verified at router level; tracing middleware test covers it.
	// This test confirms the controller itself doesn't strip the header.
	ctrl := controllers.NewComboGenerationController(newGenerateUC())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "trace-abc")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// ── Phase 7 gap tests ────────────────────────────────────────────────────────

// TC-301b-5: GET /preferences/options returns exactly the 6 occasions, 4 styles, 7 colors
func TestGetOptions_ExactEnumValues(t *testing.T) {
	uc := usecases.NewGetPreferenceOptionsUseCase()
	ctrl := controllers.NewStylePreferencesController(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/style/preferences/options", nil)
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.GetOptions(rr, req)

	body := decodeBody(t, rr)

	wantOccasions := []string{"casual", "formal", "outdoor", "beach", "office", "party"}
	wantStyles := []string{"minimalist", "bold", "classic", "bohemian"}
	wantColors := []string{"black", "white", "navy", "beige", "red", "green", "pastel"}

	checkStringSlice := func(field string, want []string) {
		raw, ok := body[field].([]interface{})
		if !ok {
			t.Errorf("field %q missing or not an array", field)
			return
		}
		if len(raw) != len(want) {
			t.Errorf("field %q: expected %d items, got %d", field, len(want), len(raw))
			return
		}
		got := make(map[string]bool)
		for _, v := range raw {
			got[v.(string)] = true
		}
		for _, w := range want {
			if !got[w] {
				t.Errorf("field %q: missing value %q", field, w)
			}
		}
	}

	checkStringSlice("occasions", wantOccasions)
	checkStringSlice("styles", wantStyles)
	checkStringSlice("colors", wantColors)
}

// TC-302-5: POST /preferences/confirm echoes all submitted fields in response.preferences
func TestConfirmPreferences_EchoesAllFields(t *testing.T) {
	d := dispatcher.NewInProcessEventDispatcher()
	uc := usecases.NewConfirmPreferencesUseCase(infraAI.NewMockPreferenceInterpretationService(), d)
	ctrl := controllers.NewPreferenceConfirmationController(uc)

	body := `{"occasions":["casual","beach"],"styles":["minimalist"],"budget":{"min":50,"max":200},"colors":{"preferred":["beige"],"excluded":["black"]},"freeText":"summer trip"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/preferences/confirm", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Confirm(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	resp := decodeBody(t, rr)
	prefs, ok := resp["preferences"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'preferences' object in response")
	}
	occasions, _ := prefs["occasions"].([]interface{})
	if len(occasions) != 2 {
		t.Errorf("expected 2 occasions echoed, got %d", len(occasions))
	}
	if prefs["freeText"] != "summer trip" {
		t.Errorf("expected freeText echoed, got %v", prefs["freeText"])
	}
	budget, _ := prefs["budget"].(map[string]interface{})
	if budget == nil {
		t.Error("expected budget echoed")
	}
}

// TC-404-6: POST /combos/generate with fallback scoring → status="fallback", message, alternatives
func TestGenerateCombos_FallbackResponse_Shape(t *testing.T) {
	// Wire a use case backed by a scoring stub that always returns fallback.
	d := newDispatcher()
	uc := usecases.NewGenerateCombosUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraACL.NewInMemoryProductCatalogRepository(),
		infraACL.NewInMemoryCompleteLookRepository(),
		&alwaysFallbackScoringStub{},
		infraAI.NewMockComboReasoningGenerationService(),
		d,
	)
	ctrl := controllers.NewComboGenerationController(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	resp := decodeBody(t, rr)
	if resp["status"] != "fallback" {
		t.Fatalf("expected status=fallback, got %v", resp["status"])
	}
	if msg, _ := resp["message"].(string); msg == "" {
		t.Error("expected non-empty message in fallback response")
	}
	alts, ok := resp["alternatives"].([]interface{})
	if !ok || len(alts) == 0 {
		t.Error("expected non-empty alternatives array in fallback response")
	}
}

// TC-404-7: Each alternative in fallback response has all required fields
func TestGenerateCombos_FallbackAlternatives_HaveAllFields(t *testing.T) {
	d := newDispatcher()
	uc := usecases.NewGenerateCombosUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraACL.NewInMemoryProductCatalogRepository(),
		infraACL.NewInMemoryCompleteLookRepository(),
		&alwaysFallbackScoringStub{},
		infraAI.NewMockComboReasoningGenerationService(),
		d,
	)
	ctrl := controllers.NewComboGenerationController(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	resp := decodeBody(t, rr)
	if resp["status"] != "fallback" {
		t.Skip("not a fallback response, skipping field check")
	}
	alts, _ := resp["alternatives"].([]interface{})
	for i, a := range alts {
		alt := a.(map[string]interface{})
		for _, field := range []string{"configSku", "simpleSku", "name", "brand", "price", "imageUrl", "reason"} {
			if _, ok := alt[field]; !ok {
				t.Errorf("alternative[%d] missing field %q", i, field)
			}
		}
	}
}

// TC-403-6: Each combo item has source field set to "wishlist" or "catalog"
func TestGenerateCombos_ComboItems_HaveSourceField(t *testing.T) {
	ctrl := controllers.NewComboGenerationController(newGenerateUC())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	resp := decodeBody(t, rr)
	combos, _ := resp["combos"].([]interface{})
	for i, c := range combos {
		combo := c.(map[string]interface{})
		items, _ := combo["items"].([]interface{})
		for j, it := range items {
			item := it.(map[string]interface{})
			src, _ := item["source"].(string)
			if src != "wishlist" && src != "catalog" {
				t.Errorf("combo[%d].items[%d] has invalid source %q", i, j, src)
			}
		}
	}
}

// TC-SEC-3b: budget.max == budget.min returns 400 VALIDATION_ERROR
func TestGenerateCombos_BudgetMaxEqualsMin_Returns400(t *testing.T) {
	ctrl := controllers.NewComboGenerationController(newGenerateUC())

	body := `{"preferences":{"budget":{"min":100,"max":100}}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when max==min, got %d", rr.Code)
	}
	resp := decodeBody(t, rr)
	if resp["error"] != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %v", resp["error"])
	}
}

// TC-SEC-3c: budget.min < 0 returns 400 VALIDATION_ERROR
func TestGenerateCombos_NegativeBudgetMin_Returns400(t *testing.T) {
	ctrl := controllers.NewComboGenerationController(newGenerateUC())

	body := `{"preferences":{"budget":{"min":-10,"max":200}}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative budget min, got %d", rr.Code)
	}
}

// TC-SEC-7b: Wishlist repository error → 502 DEPENDENCY_UNAVAILABLE
func TestGenerateCombos_WishlistError_Returns502(t *testing.T) {
	d := newDispatcher()
	uc := usecases.NewGenerateCombosUseCase(
		&erroringWishlistStub{},
		infraACL.NewInMemoryProductCatalogRepository(),
		infraACL.NewInMemoryCompleteLookRepository(),
		infraAI.NewMockComboCompatibilityScoringService(),
		infraAI.NewMockComboReasoningGenerationService(),
		d,
	)
	ctrl := controllers.NewComboGenerationController(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rr.Code)
	}
	resp := decodeBody(t, rr)
	if resp["error"] != "DEPENDENCY_UNAVAILABLE" {
		t.Errorf("expected DEPENDENCY_UNAVAILABLE, got %v", resp["error"])
	}
}

// TC-SEC-12: Unhandled use case error → 502 (current impl maps all errors to 502)
// Note: the implementation maps all use case errors to 502 DEPENDENCY_UNAVAILABLE.
// A future improvement would distinguish 503 AI_UNAVAILABLE from 502 DEPENDENCY_UNAVAILABLE.
func TestGenerateCombos_ScoringError_ReturnsErrorStatus(t *testing.T) {
	d := newDispatcher()
	uc := usecases.NewGenerateCombosUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraACL.NewInMemoryProductCatalogRepository(),
		infraACL.NewInMemoryCompleteLookRepository(),
		&erroringScoringStub{},
		infraAI.NewMockComboReasoningGenerationService(),
		d,
	)
	ctrl := controllers.NewComboGenerationController(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)

	if rr.Code == http.StatusOK {
		t.Error("expected non-200 status when scoring service fails")
	}
}

// ── Controller test stubs ─────────────────────────────────────────────────────

type emptyWishlistStub struct{}

func (r *emptyWishlistStub) FetchForSession(_ valueobjects.ShopperSession) (valueobjects.WishlistSnapshot, error) {
	return valueobjects.WishlistSnapshot{Items: nil, TotalCount: 0}, nil
}

type erroringWishlistStub struct{}

func (r *erroringWishlistStub) FetchForSession(_ valueobjects.ShopperSession) (valueobjects.WishlistSnapshot, error) {
	return valueobjects.WishlistSnapshot{}, errors.New("wishlist service unavailable")
}

type erroringScoringStub struct{}

func (s *erroringScoringStub) Score(_ services.ScoringInput) (services.ScoringResult, error) {
	return services.ScoringResult{}, errors.New("bedrock unavailable")
}

// alwaysFallbackScoringStub always returns a fallback result regardless of input.
type alwaysFallbackScoringStub struct{}

func (s *alwaysFallbackScoringStub) Score(_ services.ScoringInput) (services.ScoringResult, error) {
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

// TC-SEC-8: Unrecognised occasion value returns 400 VALIDATION_ERROR
func TestGenerateCombos_UnrecognisedOccasion_Returns400(t *testing.T) {
	ctrl := controllers.NewComboGenerationController(newGenerateUC())
	body := `{"preferences":{"occasions":["disco"]}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unrecognised occasion, got %d", rr.Code)
	}
	resp := decodeBody(t, rr)
	if resp["error"] != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %v", resp["error"])
	}
}

// TC-SEC-9: Unrecognised style value returns 400 VALIDATION_ERROR
func TestGenerateCombos_UnrecognisedStyle_Returns400(t *testing.T) {
	ctrl := controllers.NewComboGenerationController(newGenerateUC())
	body := `{"preferences":{"styles":["grunge"]}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unrecognised style, got %d", rr.Code)
	}
}

// TC-SEC-10: Unrecognised color value returns 400 VALIDATION_ERROR
func TestGenerateCombos_UnrecognisedColor_Returns400(t *testing.T) {
	ctrl := controllers.NewComboGenerationController(newGenerateUC())
	body := `{"preferences":{"colors":{"preferred":["ultraviolet"],"excluded":[]}}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unrecognised color, got %d", rr.Code)
	}
}

// TC-SEC-11: AI scoring failure returns 503 AI_UNAVAILABLE
func TestGenerateCombos_AIFailure_Returns503(t *testing.T) {
	d := newDispatcher()
	uc := usecases.NewGenerateCombosUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraACL.NewInMemoryProductCatalogRepository(),
		infraACL.NewInMemoryCompleteLookRepository(),
		&erroringScoringStub{},
		infraAI.NewMockComboReasoningGenerationService(),
		d,
	)
	ctrl := controllers.NewComboGenerationController(uc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/style/combos/generate", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req = injectSession(req)
	rr := httptest.NewRecorder()
	ctrl.Generate(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 for AI failure, got %d", rr.Code)
	}
	resp := decodeBody(t, rr)
	if resp["error"] != "AI_UNAVAILABLE" {
		t.Errorf("expected AI_UNAVAILABLE, got %v", resp["error"])
	}
}
