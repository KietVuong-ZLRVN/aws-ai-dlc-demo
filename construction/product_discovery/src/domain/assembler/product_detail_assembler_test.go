package assembler

import (
	"testing"

	"pgregory.net/rapid"
)

func TestProductDetailAssembler_VariantInStockDerivation(t *testing.T) {
	// 4.2.7: each variant InStock == (quantity > 0)
	asm := NewProductDetailAssembler()
	rapid.Check(t, func(t *rapid.T) {
		product := genPlatformProduct(t, "product")

		result := asm.Assemble(product)

		for i, s := range product.Simples {
			expected := s.Quantity > 0
			if result.Variants[i].InStock != expected {
				t.Fatalf("variant[%d] InStock: got %v, want %v (quantity=%d)",
					i, result.Variants[i].InStock, expected, s.Quantity)
			}
		}
	})
}

func TestProductDetailAssembler_FieldNameMapping(t *testing.T) {
	// 4.2.8: config_sku→ConfigSku, url_key→Slug, name→Name, brand→Brand
	asm := NewProductDetailAssembler()
	rapid.Check(t, func(t *rapid.T) {
		product := genPlatformProduct(t, "product")

		result := asm.Assemble(product)

		if result.ConfigSku != product.ConfigSku {
			t.Fatalf("ConfigSku: got %q, want %q", result.ConfigSku, product.ConfigSku)
		}
		if result.Slug != product.UrlKey {
			t.Fatalf("Slug (url_key): got %q, want %q", result.Slug, product.UrlKey)
		}
		if result.Name != product.Name {
			t.Fatalf("Name: got %q, want %q", result.Name, product.Name)
		}
		if result.Brand != product.Brand {
			t.Fatalf("Brand: got %q, want %q", result.Brand, product.Brand)
		}
	})
}

func TestProductDetailAssembler_VariantCountPreservation(t *testing.T) {
	// 4.2.9: len(Variants) == len(simples)
	asm := NewProductDetailAssembler()
	rapid.Check(t, func(t *rapid.T) {
		product := genPlatformProduct(t, "product")

		result := asm.Assemble(product)

		if len(result.Variants) != len(product.Simples) {
			t.Fatalf("Variants count: got %d, want %d", len(result.Variants), len(product.Simples))
		}
	})
}

func TestProductDetailAssembler_OccasionsPassthrough(t *testing.T) {
	// 4.2.10: assembled Occasions matches raw Occasions (Images not present in this model)
	asm := NewProductDetailAssembler()
	rapid.Check(t, func(t *rapid.T) {
		product := genPlatformProduct(t, "product")

		result := asm.Assemble(product)

		if len(result.Occasions) != len(product.Occasions) {
			t.Fatalf("Occasions count: got %d, want %d", len(result.Occasions), len(product.Occasions))
		}
		for i, o := range product.Occasions {
			if result.Occasions[i] != o {
				t.Fatalf("Occasions[%d]: got %q, want %q", i, result.Occasions[i], o)
			}
		}
	})
}
