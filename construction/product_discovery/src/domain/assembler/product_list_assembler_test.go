package assembler

import (
	"testing"

	"pgregory.net/rapid"
)

func TestProductListAssembler_InStockDerivation(t *testing.T) {
	// 4.2.1: inStock == true iff at least one simple has Quantity > 0
	asm := NewProductListAssembler()
	rapid.Check(t, func(t *rapid.T) {
		payload := genRawProductListPayload(t)
		filter := genRawFilterPayload(t)

		result := asm.Assemble(payload, filter)

		for i, product := range payload.Products {
			expectedInStock := false
			for _, s := range product.Simples {
				if s.Quantity > 0 {
					expectedInStock = true
					break
				}
			}
			got := result.Products[i].InStock
			if got != expectedInStock {
				t.Fatalf("product[%d] InStock: got %v, want %v (simples=%v)",
					i, got, expectedInStock, product.Simples)
			}
		}
	})
}

func TestProductListAssembler_TotalCountPassthrough(t *testing.T) {
	// 4.2.2: Total equals numProductFound from raw payload
	asm := NewProductListAssembler()
	rapid.Check(t, func(t *rapid.T) {
		payload := genRawProductListPayload(t)
		filter := genRawFilterPayload(t)

		result := asm.Assemble(payload, filter)

		if result.Total != payload.TotalCount {
			t.Fatalf("Total: got %d, want %d", result.Total, payload.TotalCount)
		}
	})
}

func TestProductListAssembler_ItemCountPreservation(t *testing.T) {
	// 4.2.3: len(Items) == len(products in payload)
	asm := NewProductListAssembler()
	rapid.Check(t, func(t *rapid.T) {
		payload := genRawProductListPayload(t)
		filter := genRawFilterPayload(t)

		result := asm.Assemble(payload, filter)

		if len(result.Products) != len(payload.Products) {
			t.Fatalf("len(Products): got %d, want %d", len(result.Products), len(payload.Products))
		}
	})
}

func TestProductListAssembler_ColorFacetsPassthrough(t *testing.T) {
	// 4.2.4: assembled Colors slice equals filter payload Colors
	asm := NewProductListAssembler()
	rapid.Check(t, func(t *rapid.T) {
		payload := genRawProductListPayload(t)
		filter := genRawFilterPayload(t)

		result := asm.Assemble(payload, filter)

		if len(result.Filters.Colors) != len(filter.Colors) {
			t.Fatalf("Filters.Colors length: got %d, want %d",
				len(result.Filters.Colors), len(filter.Colors))
		}
		for i, c := range filter.Colors {
			if result.Filters.Colors[i] != c {
				t.Fatalf("Filters.Colors[%d]: got %q, want %q", i, result.Filters.Colors[i], c)
			}
		}
	})
}

func TestProductListAssembler_PriceRangeFacetPassthrough(t *testing.T) {
	// 4.2.6: assembled PriceRange matches filter payload min/max
	asm := NewProductListAssembler()
	rapid.Check(t, func(t *rapid.T) {
		payload := genRawProductListPayload(t)
		filter := genRawFilterPayload(t)

		result := asm.Assemble(payload, filter)

		if result.Filters.PriceRange.Min != filter.MinPrice {
			t.Fatalf("PriceRange.Min: got %v, want %v", result.Filters.PriceRange.Min, filter.MinPrice)
		}
		if result.Filters.PriceRange.Max != filter.MaxPrice {
			t.Fatalf("PriceRange.Max: got %v, want %v", result.Filters.PriceRange.Max, filter.MaxPrice)
		}
	})
}
