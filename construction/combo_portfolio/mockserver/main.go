// mockserver — combo-portfolio local mock HTTP server
//
// Runs the full HTTP API with in-memory storage (no MySQL, no Docker).
// A fixed shopper ID is injected so no real session cookie is needed.
// Two sample combos are pre-seeded so the UI opens with data.
//
// Usage:  go run ./mockserver
// Then:   cd frontend && npm run dev

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/api"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/application"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/acl"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/services"
)

const (
	mockShopperID  = "demo-shopper-001"
	backendPort    = ":8080"
	frontendOrigin = "http://localhost:5173"
)

func main() {
	repo := newInMemoryComboRepository()

	enrichmentSvc := acl.NewComboEnrichmentService(&stubProductCatalogACL{})
	tokenSvc := services.NewShareTokenService(repo)

	handlers := api.NewHandlers(
		application.NewSaveComboHandler(repo),
		application.NewRenameComboHandler(repo),
		application.NewDeleteComboHandler(repo),
		application.NewShareComboHandler(repo, tokenSvc),
		application.NewMakePrivateHandler(repo),
		application.NewGetComboHandler(repo, enrichmentSvc),
		application.NewListCombosHandler(repo, enrichmentSvc),
		application.NewGetSharedComboHandler(repo, enrichmentSvc),
		frontendOrigin, // publicBaseURL — share links point to the Vite dev server
	)

	// chi router from application layer
	inner := api.NewRouter(handlers, &alwaysAuthValidator{})

	// Add the Unit 5 cart stub on the same server so the frontend can call it
	inner.Post("/api/v1/cart/combo", handleCartCombo)

	// Wrap with CORS + logging
	mux := http.NewServeMux()
	mux.Handle("/", corsMiddleware(inner))

	log.Printf("Mock backend  → http://localhost%s", backendPort)
	log.Printf("Shopper ID    → %s (fixed; no cookie required)", mockShopperID)
	log.Printf("Vite frontend → %s  (run: cd frontend && npm run dev)", frontendOrigin)
	log.Fatal(http.ListenAndServe(backendPort, mux))
}

// ── CORS ─────────────────────────────────────────────────────────────────────

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,Cookie")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── Auth — always passes (no real session) ────────────────────────────────────

type alwaysAuthValidator struct{}

func (v *alwaysAuthValidator) ValidateSession(_ *http.Request) (domain.ShopperId, error) {
	return domain.ShopperId(mockShopperID), nil
}

// ── Cart combo stub (simulates Unit 5 POST /api/v1/cart/combo) ───────────────

func handleCartCombo(w http.ResponseWriter, r *http.Request) {
	type addedItem struct {
		SimpleSku string `json:"simpleSku"`
		Quantity  int    `json:"quantity"`
	}
	type cartResp struct {
		Status       string      `json:"status"`
		AddedItems   []addedItem `json:"addedItems"`
		SkippedItems []addedItem `json:"skippedItems"`
	}

	resp := cartResp{
		Status: "ok",
		AddedItems: []addedItem{
			{SimpleSku: "SIMPLE-SKU-001-M", Quantity: 1},
			{SimpleSku: "SIMPLE-SKU-002-30", Quantity: 1},
			{SimpleSku: "SIMPLE-SKU-003-42", Quantity: 1},
		},
		SkippedItems: []addedItem{},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ── In-memory ComboRepository ─────────────────────────────────────────────────

type inMemoryComboRepository struct {
	mu     sync.RWMutex
	combos map[string]*domain.Combo
}

func newInMemoryComboRepository() *inMemoryComboRepository {
	r := &inMemoryComboRepository{combos: make(map[string]*domain.Combo)}
	r.seed()
	return r
}

func (r *inMemoryComboRepository) Save(_ context.Context, combo *domain.Combo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.combos[combo.ID().String()] = combo
	return nil
}

func (r *inMemoryComboRepository) FindById(_ context.Context, id domain.ComboId) (*domain.Combo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.combos[id.String()]; ok {
		return c, nil
	}
	return nil, domain.ErrComboNotFound
}

func (r *inMemoryComboRepository) FindByShopperId(_ context.Context, shopperID domain.ShopperId) ([]*domain.Combo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Combo
	for _, c := range r.combos {
		if c.OwnedBy(shopperID) {
			result = append(result, c)
		}
	}
	return result, nil
}

func (r *inMemoryComboRepository) FindByShareToken(_ context.Context, token domain.ShareToken) (*domain.Combo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.combos {
		if c.ShareToken() != nil && *c.ShareToken() == token {
			return c, nil
		}
	}
	return nil, domain.ErrComboNotFound
}

func (r *inMemoryComboRepository) Delete(_ context.Context, id domain.ComboId) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.combos, id.String())
	return nil
}

// seed pre-populates two sample combos so the UI opens with data.
func (r *inMemoryComboRepository) seed() {
	items1 := []domain.ComboItem{
		{ConfigSku: "CONFIG-SKU-001", SimpleSku: "SIMPLE-SKU-001-M", Name: "Classic White T-Shirt", Price: 19.99,
			ImageUrl: "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=400"},
		{ConfigSku: "CONFIG-SKU-002", SimpleSku: "SIMPLE-SKU-002-30", Name: "Slim Fit Blue Jeans", Price: 59.99,
			ImageUrl: "https://images.unsplash.com/photo-1542272604-787c3835535d?w=400"},
		{ConfigSku: "CONFIG-SKU-003", SimpleSku: "SIMPLE-SKU-003-42", Name: "White Canvas Sneakers", Price: 89.99,
			ImageUrl: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=400"},
	}
	items2 := []domain.ComboItem{
		{ConfigSku: "CONFIG-SKU-001", SimpleSku: "SIMPLE-SKU-001-S", Name: "Classic White T-Shirt", Price: 19.99,
			ImageUrl: "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=400"},
		{ConfigSku: "CONFIG-SKU-002", SimpleSku: "SIMPLE-SKU-002-32", Name: "Slim Fit Blue Jeans", Price: 59.99,
			ImageUrl: "https://images.unsplash.com/photo-1542272604-787c3835535d?w=400"},
	}

	name1, _ := domain.NewComboName("Summer Casual Look")
	name2, _ := domain.NewComboName("Weekend Outfit")
	id1, id2 := domain.NewComboId(), domain.NewComboId()
	now := time.Now().UTC()

	c1 := domain.ReconstitueCombo(id1, mockShopperID, name1, items1,
		domain.VisibilityPrivate, nil, now.Add(-48*time.Hour), now.Add(-48*time.Hour))
	c2 := domain.ReconstitueCombo(id2, mockShopperID, name2, items2,
		domain.VisibilityPrivate, nil, now.Add(-2*time.Hour), now.Add(-2*time.Hour))

	r.combos[id1.String()] = c1
	r.combos[id2.String()] = c2

	fmt.Printf("Seeded 2 combos: %q and %q\n", name1, name2)
}

// ── Stub Product Catalog ACL ──────────────────────────────────────────────────

type stubProductCatalogACL struct{}

var stubProducts = map[string]*acl.CatalogProduct{
	"CONFIG-SKU-001": {
		ConfigSku: "CONFIG-SKU-001",
		Name:      "Classic White T-Shirt",
		Price:     19.99,
		ImageURL:  "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=400",
		Variants: []acl.CatalogVariant{
			{SimpleSku: "SIMPLE-SKU-001-S", InStock: true},
			{SimpleSku: "SIMPLE-SKU-001-M", InStock: true},
		},
	},
	"CONFIG-SKU-002": {
		ConfigSku: "CONFIG-SKU-002",
		Name:      "Slim Fit Blue Jeans",
		Price:     59.99,
		ImageURL:  "https://images.unsplash.com/photo-1542272604-787c3835535d?w=400",
		Variants: []acl.CatalogVariant{
			{SimpleSku: "SIMPLE-SKU-002-30", InStock: true},
			{SimpleSku: "SIMPLE-SKU-002-32", InStock: true},
		},
	},
	"CONFIG-SKU-003": {
		ConfigSku: "CONFIG-SKU-003",
		Name:      "White Canvas Sneakers",
		Price:     89.99,
		ImageURL:  "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=400",
		Variants: []acl.CatalogVariant{
			{SimpleSku: "SIMPLE-SKU-003-42", InStock: true},
			{SimpleSku: "SIMPLE-SKU-003-43", InStock: false},
		},
	},
}

func (s *stubProductCatalogACL) FetchProduct(_ context.Context, configSku string) (*acl.CatalogProduct, error) {
	if p, ok := stubProducts[configSku]; ok {
		return p, nil
	}
	return nil, nil
}
