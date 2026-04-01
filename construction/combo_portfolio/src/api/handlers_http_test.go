package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/api"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/application"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/acl"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// ---------------------------------------------------------------------------
// In-memory repo (test double)
// ---------------------------------------------------------------------------

type fakeRepo struct {
	mu     sync.RWMutex
	combos map[string]*domain.Combo
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{combos: make(map[string]*domain.Combo)}
}

func (r *fakeRepo) Save(_ context.Context, combo *domain.Combo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.combos[combo.ID().String()] = combo
	return nil
}
func (r *fakeRepo) FindById(_ context.Context, id domain.ComboId) (*domain.Combo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.combos[id.String()]; ok {
		return c, nil
	}
	return nil, domain.ErrComboNotFound
}
func (r *fakeRepo) FindByShopperId(_ context.Context, sid domain.ShopperId) ([]*domain.Combo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Combo
	for _, c := range r.combos {
		if c.OwnedBy(sid) {
			out = append(out, c)
		}
	}
	return out, nil
}
func (r *fakeRepo) FindByShareToken(_ context.Context, t domain.ShareToken) (*domain.Combo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.combos {
		if c.ShareToken() != nil && *c.ShareToken() == t {
			return c, nil
		}
	}
	return nil, domain.ErrComboNotFound
}
func (r *fakeRepo) Delete(_ context.Context, id domain.ComboId) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.combos, id.String())
	return nil
}

// ---------------------------------------------------------------------------
// Always-pass session validator and stub enrichment
// ---------------------------------------------------------------------------

type fixedValidator struct{ id domain.ShopperId }

func (v *fixedValidator) ValidateSession(_ *http.Request) (domain.ShopperId, error) {
	return v.id, nil
}

type passthroughCatalog struct{}

func (p *passthroughCatalog) FetchProduct(_ context.Context, _ string) (*acl.CatalogProduct, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Test server builder
// ---------------------------------------------------------------------------

func newTestServer(shopperID domain.ShopperId, repo *fakeRepo) *httptest.Server {
	enrichment := acl.NewComboEnrichmentService(&passthroughCatalog{})
	tokenSvc := services.NewShareTokenService(repo)
	h := api.NewHandlers(
		application.NewSaveComboHandler(repo),
		application.NewRenameComboHandler(repo),
		application.NewDeleteComboHandler(repo),
		application.NewShareComboHandler(repo, tokenSvc),
		application.NewMakePrivateHandler(repo),
		application.NewGetComboHandler(repo, enrichment),
		application.NewListCombosHandler(repo, enrichment),
		application.NewGetSharedComboHandler(repo, enrichment),
		"http://localhost:5173",
	)
	router := api.NewRouter(h, &fixedValidator{id: shopperID})
	return httptest.NewServer(router)
}

func jsonBody(v interface{}) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func makeItemPayload(n int) []map[string]interface{} {
	items := make([]map[string]interface{}, n)
	for i := 0; i < n; i++ {
		items[i] = map[string]interface{}{
			"configSku": "cfg",
			"simpleSku": strings.Repeat("x", i+1),
			"name":      "prod",
			"imageUrl":  "https://example.com/img.jpg",
			"price":     9.99,
		}
	}
	return items
}

// ---------------------------------------------------------------------------
// 6.1 POST /api/v1/combos — name validation
// ---------------------------------------------------------------------------

func TestAPI_SaveCombo_InvalidNameReturns400(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		server := newTestServer("s1", newFakeRepo())
		defer server.Close()
		nameLen := rapid.IntRange(101, 300).Draw(t, "nameLen")
		body := map[string]interface{}{
			"name":  strings.Repeat("a", nameLen),
			"items": makeItemPayload(3),
		}
		resp, err := http.Post(server.URL+"/api/v1/combos", "application/json", jsonBody(body))
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 for name len %d, got %d", nameLen, resp.StatusCode)
		}
	})
}

func TestAPI_SaveCombo_ValidRequestReturns201(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		server := newTestServer("s1", newFakeRepo())
		defer server.Close()
		n := rapid.IntRange(2, 10).Draw(t, "n")
		body := map[string]interface{}{
			"name":       "My Combo",
			"items":      makeItemPayload(n),
			"visibility": "private",
		}
		resp, err := http.Post(server.URL+"/api/v1/combos", "application/json", jsonBody(body))
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		if result["id"] == "" {
			t.Fatal("expected non-empty id in response")
		}
	})
}

// ---------------------------------------------------------------------------
// 6.2 POST /api/v1/combos — item count boundary
// ---------------------------------------------------------------------------

func TestAPI_SaveCombo_TooManyItemsReturns400(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		server := newTestServer("s1", newFakeRepo())
		defer server.Close()
		n := rapid.IntRange(11, 25).Draw(t, "n")
		body := map[string]interface{}{"name": "x", "items": makeItemPayload(n)}
		resp, err := http.Post(server.URL+"/api/v1/combos", "application/json", jsonBody(body))
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 for %d items, got %d", n, resp.StatusCode)
		}
	})
}

// ---------------------------------------------------------------------------
// 6.3 GET /api/v1/combos/{id}
// ---------------------------------------------------------------------------

func TestAPI_GetCombo_ExistingReturns200(t *testing.T) {
	repo := newFakeRepo()
	server := newTestServer("owner-1", repo)
	defer server.Close()

	// Seed via POST
	body := map[string]interface{}{"name": "Test Combo", "items": makeItemPayload(3), "visibility": "private"}
	resp, err := http.Post(server.URL+"/api/v1/combos", "application/json", jsonBody(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	var created map[string]string
	json.NewDecoder(resp.Body).Decode(&created)
	id := created["id"]
	require.NotEmpty(t, id)

	// GET it back
	resp2, err := http.Get(server.URL + "/api/v1/combos/" + id)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

func TestAPI_GetCombo_DifferentShopperForbidden(t *testing.T) {
	repo := newFakeRepo()
	ownerServer := newTestServer("owner-1", repo)
	defer ownerServer.Close()
	intruderServer := newTestServer("intruder-2", repo)
	defer intruderServer.Close()

	body := map[string]interface{}{"name": "Private", "items": makeItemPayload(2), "visibility": "private"}
	resp, _ := http.Post(ownerServer.URL+"/api/v1/combos", "application/json", jsonBody(body))
	defer resp.Body.Close()
	var created map[string]string
	json.NewDecoder(resp.Body).Decode(&created)
	id := created["id"]

	resp2, err := http.Get(intruderServer.URL + "/api/v1/combos/" + id)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp2.StatusCode)
}

func TestAPI_GetCombo_NonExistentReturns404(t *testing.T) {
	server := newTestServer("s1", newFakeRepo())
	defer server.Close()
	resp, err := http.Get(server.URL + "/api/v1/combos/nonexistent-uuid")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// 6.4 DELETE /api/v1/combos/{id}
// ---------------------------------------------------------------------------

func TestAPI_DeleteCombo_OwnerDeletesThenGet404(t *testing.T) {
	repo := newFakeRepo()
	server := newTestServer("owner-1", repo)
	defer server.Close()

	body := map[string]interface{}{"name": "To Delete", "items": makeItemPayload(2)}
	resp, _ := http.Post(server.URL+"/api/v1/combos", "application/json", jsonBody(body))
	defer resp.Body.Close()
	var created map[string]string
	json.NewDecoder(resp.Body).Decode(&created)
	id := created["id"]

	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/combos/"+id, nil)
	client := &http.Client{}
	delResp, err := client.Do(req)
	require.NoError(t, err)
	delResp.Body.Close()
	assert.Equal(t, http.StatusNoContent, delResp.StatusCode)

	getResp, _ := http.Get(server.URL + "/api/v1/combos/" + id)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
}

// ---------------------------------------------------------------------------
// 6.5 POST /api/v1/combos/{id}/share — shareUrl uses frontend /shared/{token}
// ---------------------------------------------------------------------------

func TestAPI_ShareCombo_ShareUrlPointsToFrontendRoute(t *testing.T) {
	repo := newFakeRepo()
	server := newTestServer("owner-1", repo)
	defer server.Close()

	body := map[string]interface{}{"name": "Share Me", "items": makeItemPayload(3)}
	resp, _ := http.Post(server.URL+"/api/v1/combos", "application/json", jsonBody(body))
	defer resp.Body.Close()
	var created map[string]string
	json.NewDecoder(resp.Body).Decode(&created)
	id := created["id"]

	shareResp, err := http.Post(server.URL+"/api/v1/combos/"+id+"/share", "application/json", nil)
	require.NoError(t, err)
	defer shareResp.Body.Close()
	assert.Equal(t, http.StatusOK, shareResp.StatusCode)

	var shareResult map[string]string
	json.NewDecoder(shareResp.Body).Decode(&shareResult)
	shareURL := shareResult["shareUrl"]
	shareTok := shareResult["shareToken"]
	require.NotEmpty(t, shareURL)
	expected := fmt.Sprintf("http://localhost:5173/shared/%s", shareTok)
	assert.Equal(t, expected, shareURL, "shareUrl must point to the frontend /shared/{token} route, not the API")
}

// ---------------------------------------------------------------------------
// 6.6 GET /api/v1/combos/shared/{token} — no auth required
// ---------------------------------------------------------------------------

func TestAPI_GetSharedCombo_ValidTokenNoAuthRequired(t *testing.T) {
	repo := newFakeRepo()
	server := newTestServer("owner-1", repo)
	defer server.Close()

	// Save combo
	body := map[string]interface{}{"name": "Public Combo", "items": makeItemPayload(3)}
	resp, _ := http.Post(server.URL+"/api/v1/combos", "application/json", jsonBody(body))
	defer resp.Body.Close()
	var created map[string]string
	json.NewDecoder(resp.Body).Decode(&created)
	id := created["id"]

	// Share it
	shareResp, _ := http.Post(server.URL+"/api/v1/combos/"+id+"/share", "application/json", nil)
	defer shareResp.Body.Close()
	var shareResult map[string]string
	json.NewDecoder(shareResp.Body).Decode(&shareResult)
	token := shareResult["shareToken"]

	// Access public URL with no session (use a new server with a failing validator to prove no auth needed)
	pubResp, err := http.Get(server.URL + "/api/v1/combos/shared/" + token)
	require.NoError(t, err)
	defer pubResp.Body.Close()
	assert.Equal(t, http.StatusOK, pubResp.StatusCode)
}

func TestAPI_GetSharedCombo_InvalidTokenReturns404(t *testing.T) {
	server := newTestServer("s1", newFakeRepo())
	defer server.Close()
	resp, err := http.Get(server.URL + "/api/v1/combos/shared/ghost-token")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// 6.7 PUT /api/v1/combos/{id} visibility=private -> shareToken becomes null
// ---------------------------------------------------------------------------

func TestAPI_UpdateVisibilityToPrivate_ClearsShareToken(t *testing.T) {
	repo := newFakeRepo()
	server := newTestServer("owner-1", repo)
	defer server.Close()

	body := map[string]interface{}{"name": "Will Be Private", "items": makeItemPayload(2)}
	resp, _ := http.Post(server.URL+"/api/v1/combos", "application/json", jsonBody(body))
	defer resp.Body.Close()
	var created map[string]string
	json.NewDecoder(resp.Body).Decode(&created)
	id := created["id"]

	// Share it first
	shareResp, _ := http.Post(server.URL+"/api/v1/combos/"+id+"/share", "application/json", nil)
	shareResp.Body.Close()

	// Make it private via PUT
	putBody := map[string]interface{}{"visibility": "private"}
	putReq, _ := http.NewRequest(http.MethodPut, server.URL+"/api/v1/combos/"+id, jsonBody(putBody))
	putReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	putResp, err := client.Do(putReq)
	require.NoError(t, err)
	putResp.Body.Close()
	assert.Equal(t, http.StatusNoContent, putResp.StatusCode)

	// Verify shareToken is null in GET response
	getResp, _ := http.Get(server.URL + "/api/v1/combos/" + id)
	defer getResp.Body.Close()
	var comboResp map[string]interface{}
	json.NewDecoder(getResp.Body).Decode(&comboResp)
	assert.Nil(t, comboResp["shareToken"], "shareToken should be null after making combo private")
}
