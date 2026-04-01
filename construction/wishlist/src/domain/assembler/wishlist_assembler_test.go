package assembler

import (
	"fmt"
	"testing"
	"wishlist/domain/valueobject"

	"pgregory.net/rapid"
)

func makeShopperId(t *testing.T) valueobject.ShopperId {
	t.Helper()
	sid, err := valueobject.NewShopperId("shopper-test")
	if err != nil {
		t.Fatalf("NewShopperId: %v", err)
	}
	return sid
}

func genRawItem(t *rapid.T, i int) RawWishlistItem {
	prefix := rapid.StringMatching(`[A-Z]{2}`).Draw(t, fmt.Sprintf("item[%d].prefix", i))
	num := rapid.IntRange(1, 999).Draw(t, fmt.Sprintf("item[%d].num", i))
	return RawWishlistItem{
		ItemId:    fmt.Sprintf("item-%d", i),
		SimpleSku: fmt.Sprintf("%s-%03d-M-BLK", prefix, num),
		ConfigSku: fmt.Sprintf("%s-%03d", prefix, num),
		Name:      rapid.StringMatching(`[A-Za-z ]{3,10}`).Draw(t, fmt.Sprintf("item[%d].name", i)),
		Brand:     rapid.StringMatching(`[A-Za-z]{3,8}`).Draw(t, fmt.Sprintf("item[%d].brand", i)),
		Price:     rapid.Float64Range(0, 999).Draw(t, fmt.Sprintf("item[%d].price", i)),
		Currency:  "SGD",
		ImageUrl:  "https://img.example.com/img.jpg",
		Color:     "black",
		Size:      "M",
		InStock:   rapid.Bool().Draw(t, fmt.Sprintf("item[%d].inStock", i)),
	}
}

// 4.7.1: item count preservation
func TestWishlistAssembler_ItemCountPreservation(t *testing.T) {
	asm := &WishlistAssembler{}
	sid := makeShopperId(t)

	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 8).Draw(t, "n")
		rawItems := make([]RawWishlistItem, n)
		for i := range rawItems {
			rawItems[i] = genRawItem(t, i)
		}

		wl, err := asm.Assemble(sid, rawItems)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(wl.Items) != n {
			t.Fatalf("Items count: got %d, want %d", len(wl.Items), n)
		}
	})
}

// 4.7.2: TotalCount passthrough
func TestWishlistAssembler_TotalCountPassthrough(t *testing.T) {
	asm := &WishlistAssembler{}
	sid := makeShopperId(t)

	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 5).Draw(t, "n")
		rawItems := make([]RawWishlistItem, n)
		for i := range rawItems {
			rawItems[i] = genRawItem(t, i)
		}

		wl, err := asm.Assemble(sid, rawItems)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wl.TotalCount != n {
			t.Fatalf("TotalCount: got %d, want %d", wl.TotalCount, n)
		}
	})
}

// 4.7.3: field mapping correctness
func TestWishlistAssembler_FieldMapping(t *testing.T) {
	asm := &WishlistAssembler{}
	sid := makeShopperId(t)

	rapid.Check(t, func(t *rapid.T) {
		raw := genRawItem(t, 0)

		wl, err := asm.Assemble(sid, []RawWishlistItem{raw})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		item := wl.Items[0]

		if item.ItemId.String() != raw.ItemId {
			t.Fatalf("ItemId: got %q, want %q", item.ItemId.String(), raw.ItemId)
		}
		if item.SimpleSku.String() != raw.SimpleSku {
			t.Fatalf("SimpleSku: got %q, want %q", item.SimpleSku.String(), raw.SimpleSku)
		}
		if item.ConfigSku.String() != raw.ConfigSku {
			t.Fatalf("ConfigSku: got %q, want %q", item.ConfigSku.String(), raw.ConfigSku)
		}
		if item.Name != raw.Name {
			t.Fatalf("Name: got %q, want %q", item.Name, raw.Name)
		}
		if item.Brand != raw.Brand {
			t.Fatalf("Brand: got %q, want %q", item.Brand, raw.Brand)
		}
		if item.Price.Amount != raw.Price {
			t.Fatalf("Price.Amount: got %v, want %v", item.Price.Amount, raw.Price)
		}
		if item.ImageUrl != raw.ImageUrl {
			t.Fatalf("ImageUrl: got %q, want %q", item.ImageUrl, raw.ImageUrl)
		}
	})
}

// 4.7.4: inStock passthrough
func TestWishlistAssembler_InStockPassthrough(t *testing.T) {
	asm := &WishlistAssembler{}
	sid := makeShopperId(t)

	rapid.Check(t, func(t *rapid.T) {
		raw := genRawItem(t, 0)

		wl, err := asm.Assemble(sid, []RawWishlistItem{raw})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wl.Items[0].InStock != raw.InStock {
			t.Fatalf("InStock: got %v, want %v", wl.Items[0].InStock, raw.InStock)
		}
	})
}

// error path: invalid itemId returns error
func TestWishlistAssembler_InvalidItemIdReturnsError(t *testing.T) {
	asm := &WishlistAssembler{}
	sid := makeShopperId(t)

	raw := RawWishlistItem{
		ItemId:    "", // invalid
		SimpleSku: "PD-001-M-BLK",
		ConfigSku: "PD-001",
		Price:     10,
		Currency:  "SGD",
	}
	_, err := asm.Assemble(sid, []RawWishlistItem{raw})
	if err == nil {
		t.Fatal("expected error for empty ItemId, got nil")
	}
}

// error path: invalid simpleSku returns error
func TestWishlistAssembler_InvalidSimpleSkuReturnsError(t *testing.T) {
	asm := &WishlistAssembler{}
	sid := makeShopperId(t)

	raw := RawWishlistItem{
		ItemId:    "item-001",
		SimpleSku: "", // invalid
		ConfigSku: "PD-001",
		Price:     10,
		Currency:  "SGD",
	}
	_, err := asm.Assemble(sid, []RawWishlistItem{raw})
	if err == nil {
		t.Fatal("expected error for empty SimpleSku, got nil")
	}
}

// error path: invalid configSku returns error
func TestWishlistAssembler_InvalidConfigSkuReturnsError(t *testing.T) {
	asm := &WishlistAssembler{}
	sid := makeShopperId(t)

	raw := RawWishlistItem{
		ItemId:    "item-001",
		SimpleSku: "PD-001-M-BLK",
		ConfigSku: "", // invalid
		Price:     10,
		Currency:  "SGD",
	}
	_, err := asm.Assemble(sid, []RawWishlistItem{raw})
	if err == nil {
		t.Fatal("expected error for empty ConfigSku, got nil")
	}
}

// 4.7.5: empty items array
func TestWishlistAssembler_EmptyItems(t *testing.T) {
	asm := &WishlistAssembler{}
	sid := makeShopperId(t)

	wl, err := asm.Assemble(sid, []RawWishlistItem{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wl.Items) != 0 {
		t.Fatalf("expected empty Items, got %d", len(wl.Items))
	}
}
