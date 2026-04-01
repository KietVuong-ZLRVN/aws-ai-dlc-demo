package application

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"product_discovery/domain/assembler"
	"product_discovery/domain/port"
	"product_discovery/domain/query"
	"product_discovery/domain/valueobject"
	"product_discovery/mocks"

	"pgregory.net/rapid"
)

// --- helpers ---

func makeListQuery() query.ProductListQuery {
	p, _ := valueobject.NewPagination(0, 20)
	return query.ProductListQuery{Pagination: p}
}

func makeListPayload(n, total int) *port.RawProductListPayload {
	products := make([]port.PlatformProduct, n)
	for i := range products {
		products[i] = port.PlatformProduct{
			ConfigSku: fmt.Sprintf("PD-%03d", i+1),
			Name:      "Product",
			Brand:     "Brand",
			Currency:  "SGD",
			Simples:   []port.PlatformSimple{{SimpleSku: "SKU", Quantity: 1}},
		}
	}
	return &port.RawProductListPayload{Products: products, TotalCount: total}
}

func makeFilterPayload() *port.RawFilterPayload {
	return &port.RawFilterPayload{Colors: []string{"black"}, MinPrice: 0, MaxPrice: 500}
}

// --- 4.3.1: Success path ---

func TestProductListQueryHandler_Success(t *testing.T) {
	client := mocks.NewMockPlatformProductClient(t)
	listPayload := makeListPayload(3, 50)
	filterPayload := makeFilterPayload()

	client.EXPECT().FetchProductList(context.Background(), port.ProductListParams{Limit: 20}).
		Return(listPayload, nil)
	client.EXPECT().FetchProductFilters(context.Background(), port.ProductListParams{Limit: 20}).
		Return(filterPayload, nil)

	handler := NewProductListQueryHandler(client, assembler.NewProductListAssembler())
	result, err := handler.Handle(context.Background(), makeListQuery())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Total != 50 {
		t.Fatalf("Total: got %d, want 50", result.Total)
	}
	if len(result.Products) != 3 {
		t.Fatalf("Products count: got %d, want 3", len(result.Products))
	}
}

// --- 4.3.2: List fetch failure ---

func TestProductListQueryHandler_ListFetchFailure(t *testing.T) {
	client := mocks.NewMockPlatformProductClient(t)
	client.EXPECT().FetchProductList(context.Background(), port.ProductListParams{Limit: 20}).
		Return(nil, errors.New("platform error"))
	// FetchProductFilters may or may not be called depending on concurrency; either is acceptable.
	client.On("FetchProductFilters", context.Background(), port.ProductListParams{Limit: 20}).
		Return(makeFilterPayload(), nil).Maybe()

	handler := NewProductListQueryHandler(client, assembler.NewProductListAssembler())
	_, err := handler.Handle(context.Background(), makeListQuery())

	if !errors.Is(err, ErrProductListUnavailable) {
		t.Fatalf("expected ErrProductListUnavailable, got: %v", err)
	}
}

// --- 4.3.3: Filter fetch failure ---

func TestProductListQueryHandler_FilterFetchFailure(t *testing.T) {
	client := mocks.NewMockPlatformProductClient(t)
	client.EXPECT().FetchProductList(context.Background(), port.ProductListParams{Limit: 20}).
		Return(makeListPayload(2, 2), nil)
	client.EXPECT().FetchProductFilters(context.Background(), port.ProductListParams{Limit: 20}).
		Return(nil, errors.New("filter error"))

	handler := NewProductListQueryHandler(client, assembler.NewProductListAssembler())
	_, err := handler.Handle(context.Background(), makeListQuery())

	if !errors.Is(err, ErrProductListUnavailable) {
		t.Fatalf("expected ErrProductListUnavailable, got: %v", err)
	}
}

// --- 4.3.4 PBT: Payload routing correctness ---

func TestProductListQueryHandler_PayloadRoutingCorrectness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		nProducts := rapid.IntRange(0, 6).Draw(t, "nProducts")
		total := rapid.IntRange(nProducts, nProducts+50).Draw(t, "total")
		minPrice := rapid.Float64Range(0, 500).Draw(t, "minPrice")
		maxPrice := rapid.Float64Range(minPrice, 1000).Draw(t, "maxPrice")
		nColors := rapid.IntRange(0, 4).Draw(t, "nColors")

		listPayload := makeListPayload(nProducts, total)
		filterPayload := &port.RawFilterPayload{
			Colors:   make([]string, nColors),
			MinPrice: minPrice,
			MaxPrice: maxPrice,
		}
		for i := range filterPayload.Colors {
			filterPayload.Colors[i] = fmt.Sprintf("color-%d", i)
		}

		client := mocks.NewMockPlatformProductClient(t)
		client.On("FetchProductList", context.Background(), port.ProductListParams{Limit: 20}).
			Return(listPayload, nil)
		client.On("FetchProductFilters", context.Background(), port.ProductListParams{Limit: 20}).
			Return(filterPayload, nil)

		handler := NewProductListQueryHandler(client, assembler.NewProductListAssembler())
		result, err := handler.Handle(context.Background(), makeListQuery())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Items come from list payload
		if len(result.Products) != nProducts {
			t.Fatalf("Products count: got %d, want %d", len(result.Products), nProducts)
		}
		if result.Total != total {
			t.Fatalf("Total: got %d, want %d", result.Total, total)
		}
		// Filters come from filter payload
		if len(result.Filters.Colors) != nColors {
			t.Fatalf("Filters.Colors count: got %d, want %d", len(result.Filters.Colors), nColors)
		}
		if result.Filters.PriceRange.Min != minPrice {
			t.Fatalf("PriceRange.Min: got %v, want %v", result.Filters.PriceRange.Min, minPrice)
		}
	})
}
