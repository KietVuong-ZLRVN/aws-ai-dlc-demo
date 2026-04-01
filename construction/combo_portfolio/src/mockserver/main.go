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
	backendPort    = ":8081"
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
		frontendOrigin,
	)
	inner := api.NewRouter(handlers, &alwaysAuthValidator{})
	inner.Post("/api/v1/cart/combo", handleCartCombo)
	log.Printf("Mock backend  -> http://localhost%s", backendPort)
	log.Printf("Shopper ID    -> %s  (fixed, no cookie needed)", mockShopperID)
	log.Printf("Vite frontend -> %s  (run in a separate terminal)", frontendOrigin)
	log.Fatal(http.ListenAndServe(backendPort, corsMiddleware(inner)))
}

// corsMiddleware adds CORS headers so the Vite dev server can call this backend.
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

// alwaysAuthValidator satisfies api.SessionValidator without requiring a real session cookie.
type alwaysAuthValidator struct{}

func (v *alwaysAuthValidator) ValidateSession(_ *http.Request) (domain.ShopperId, error) {
	return domain.ShopperId(mockShopperID), nil
}

// handleCartCombo is a Unit-5 stub: pretends to add all combo items to the cart.
func handleCartCombo(w http.ResponseWriter, r *http.Request) {
	type item struct {
		SimpleSku string `json:"simpleSku"`
		Quantity  int    `json:"quantity"`
	}
	type resp struct {
		Status       string `json:"status"`
		AddedItems   []item `json:"addedItems"`
		SkippedItems []item `json:"skippedItems"`
	}
	body := resp{
		Status: "ok",
		AddedItems: []item{
			{"SIMPLE-SKU-001-M", 1},
			{"SIMPLE-SKU-002-30", 1},
			{"SIMPLE-SKU-003-42", 1},
		},
		SkippedItems: []item{},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(body)
}

// ---------------------------------------------------------------------------
// In-memory combo repository
// ---------------------------------------------------------------------------

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
	var out []*domain.Combo
	for _, c := range r.combos {
		if c.OwnedBy(shopperID) {
			out = append(out, c)
		}
	}
	return out, nil
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

func (r *inMemoryComboRepository) seed() {
	items1 := []domain.ComboItem{
		{
			ConfigSku: "CONFIG-SKU-001",
			SimpleSku: "SIMPLE-SKU-001-M",
			Name:      "Classic White T-Shirt",
			Price:     19.99,
			ImageUrl:  "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=400",
		},
		{
			ConfigSku: "CONFIG-SKU-002",
			SimpleSku: "SIMPLE-SKU-002-30",
			Name:      "Slim Fit Blue Jeans",
			Price:     59.99,
			ImageUrl:  "https://images.unsplash.com/photo-1542272604-787c3835535d?w=400",
		},
		{
			ConfigSku: "CONFIG-SKU-003",
			SimpleSku: "SIMPLE-SKU-003-42",
			Name:      "White Canvas Sneakers",
			Price:     89.99,
			ImageUrl:  "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=400",
		},
	}
	items2 := []domain.ComboItem{
		{
			ConfigSku: "CONFIG-SKU-001",
			SimpleSku: "SIMPLE-SKU-001-S",
			Name:      "Classic White T-Shirt",
			Price:     19.99,
			ImageUrl:  "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=400",
		},
		{
			ConfigSku: "CONFIG-SKU-002",
			SimpleSku: "SIMPLE-SKU-002-32",
			Name:      "Slim Fit Blue Jeans",
			Price:     59.99,
			ImageUrl:  "https://images.unsplash.com/photo-1542272604-787c3835535d?w=400",
		},
	}
	name1, _ := domain.NewComboName("Summer Casual Look")
	name2, _ := domain.NewComboName("Weekend Outfit")
	id1 := domain.NewComboId()
	id2 := domain.NewComboId()
	now := time.Now().UTC()
	c1 := domain.ReconstitueCombo(id1, mockShopperID, name1, items1, domain.VisibilityPrivate, nil, now.Add(-48*time.Hour), now.Add(-48*time.Hour))
	c2 := domain.ReconstitueCombo(id2, mockShopperID, name2, items2, domain.VisibilityPrivate, nil, now.Add(-2*time.Hour), now.Add(-2*time.Hour))
	r.combos[id1.String()] = c1
	r.combos[id2.String()] = c2
	fmt.Printf("[seed] %q (%s)\n", name1, id1)
	fmt.Printf("[seed] %q (%s)\n", name2, id2)
}

// ---------------------------------------------------------------------------
// Stub product-catalog ACL
// ---------------------------------------------------------------------------

type stubProductCatalogACL struct{}

func (s *stubProductCatalogACL) FetchProduct(_ context.Context, configSku string) (*acl.CatalogProduct, error) {
	stubs := map[string]*acl.CatalogProduct{
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
	if p, ok := stubs[configSku]; ok {
		return p, nil
	}
	return nil, nil
}
