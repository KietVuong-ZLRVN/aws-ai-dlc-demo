package acl_test

import (
	"ai-styling-engine/infrastructure/acl"
	"ai-styling-engine/domain/valueobjects"
	"testing"
)

// TC-INFRA-5: InMemoryWishlistRepository returns ≥ 2 in-stock items
func TestInMemoryWishlistRepository_ReturnsAtLeastTwoInStockItems(t *testing.T) {
	repo := acl.NewInMemoryWishlistRepository()
	snap, err := repo.FetchForSession(valueobjects.ShopperSession{SessionToken: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	inStock := snap.InStockItems()
	if len(inStock) < 2 {
		t.Errorf("expected ≥ 2 in-stock items, got %d — default test path would produce fallback", len(inStock))
	}
}

// TC-INFRA-6: InMemoryProductCatalogRepository returns a non-empty list
func TestInMemoryProductCatalogRepository_ReturnsNonEmptyList(t *testing.T) {
	repo := acl.NewInMemoryProductCatalogRepository()
	items, err := repo.SearchSupplementaryItems(valueobjects.CatalogSearchFilters{Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) == 0 {
		t.Error("expected non-empty catalog items from in-memory stub")
	}
	for _, item := range items {
		if item.Source != valueobjects.ItemSourceCatalog {
			t.Errorf("expected source=catalog for all catalog items, got %q", item.Source)
		}
	}
}

// TC-INFRA-7: InMemoryCompleteLookRepository returns a non-empty list for any SKU
func TestInMemoryCompleteLookRepository_ReturnsNonEmptyList(t *testing.T) {
	repo := acl.NewInMemoryCompleteLookRepository()
	items, err := repo.FetchCompleteLookSignals("CFG-BLAZER-BLK")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) == 0 {
		t.Error("expected non-empty complete-look signals from in-memory stub")
	}
	for _, item := range items {
		if item.Source != valueobjects.ItemSourceCatalog {
			t.Errorf("expected source=catalog for complete-look items, got %q", item.Source)
		}
	}
}
