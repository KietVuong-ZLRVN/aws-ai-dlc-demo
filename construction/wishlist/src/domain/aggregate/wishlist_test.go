package aggregate

import (
	"fmt"
	"strings"
	"testing"
	"wishlist/domain/entity"
	"wishlist/domain/valueobject"

	"pgregory.net/rapid"
)

// --- helpers ---

func mustSimpleSku(s string) valueobject.SimpleSku {
	sku, err := valueobject.NewSimpleSku(s)
	if err != nil {
		panic(fmt.Sprintf("mustSimpleSku(%q): %v", s, err))
	}
	return sku
}

func mustConfigSku(s string) valueobject.ConfigSku {
	sku, err := valueobject.NewConfigSku(s)
	if err != nil {
		panic(fmt.Sprintf("mustConfigSku(%q): %v", s, err))
	}
	return sku
}

func emptyWishlist() *Wishlist {
	wid, _ := valueobject.NewWishlistId("wl-test")
	sid, _ := valueobject.NewShopperId("shopper-001")
	return &Wishlist{ID: wid, ShopperID: sid, Items: []entity.WishlistItem{}}
}

func wishlistWithItem(configSkuStr string) *Wishlist {
	wl := emptyWishlist()
	csku := mustConfigSku(configSkuStr)
	ssku := mustSimpleSku(configSkuStr + "-M-BLK")
	iid, _ := valueobject.NewWishlistItemId("item-001")
	wl.Items = append(wl.Items, entity.WishlistItem{
		ItemId:    iid,
		SimpleSku: ssku,
		ConfigSku: csku,
	})
	return wl
}

// genSkuParts generates prefix (2 uppercase) and num (1-999) for structured SKUs.
func genSkuParts(t *rapid.T, label string) (string, int) {
	prefix := rapid.StringMatching(`[A-Z]{2}`).Draw(t, label+".prefix")
	num := rapid.IntRange(1, 999).Draw(t, label+".num")
	return prefix, num
}

// --- 4.6.1: deriveConfigSku two-part extraction ---

func TestDeriveConfigSku_TwoParts(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		prefix, num := genSkuParts(t, "sku")
		size := rapid.SampledFrom([]string{"XS", "S", "M", "L", "XL"}).Draw(t, "size")
		color := rapid.SampledFrom([]string{"BLK", "WHT", "RED"}).Draw(t, "color")

		rawSimple := fmt.Sprintf("%s-%03d-%s-%s", prefix, num, size, color)
		expectedConfig := fmt.Sprintf("%s-%03d", prefix, num)

		simpleSku := mustSimpleSku(rawSimple)
		configSku, err := deriveConfigSku(simpleSku)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if configSku.String() != expectedConfig {
			t.Fatalf("deriveConfigSku(%q): got %q, want %q", rawSimple, configSku.String(), expectedConfig)
		}
	})
}

// --- 4.6.2: deriveConfigSku single-part fallback ---

func TestDeriveConfigSku_SinglePart(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a string with no dashes
		raw := rapid.StringMatching(`[A-Za-z0-9]{1,15}`).Draw(t, "noDash")

		simpleSku := mustSimpleSku(raw)
		configSku, err := deriveConfigSku(simpleSku)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if configSku.String() != raw {
			t.Fatalf("single-part fallback: got %q, want %q", configSku.String(), raw)
		}
	})
}

// --- 4.6.3: AddItem duplicate prevention ---

func TestWishlist_AddItem_DuplicatePrevention(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		prefix, num := genSkuParts(t, "sku")
		configSkuStr := fmt.Sprintf("%s-%03d", prefix, num)

		wl := wishlistWithItem(configSkuStr)

		// Any simpleSku that derives to the same configSku
		newSimpleRaw := fmt.Sprintf("%s-%03d-L-WHT", prefix, num)
		simpleSku := mustSimpleSku(newSimpleRaw)

		_, err := wl.AddItem(simpleSku)

		if err != ErrWishlistItemAlreadyPresent {
			t.Fatalf("expected ErrWishlistItemAlreadyPresent, got: %v", err)
		}
	})
}

// --- 4.6.4: AddItem success on empty wishlist ---

func TestWishlist_AddItem_SuccessOnEmpty(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		prefix, num := genSkuParts(t, "sku")
		raw := fmt.Sprintf("%s-%03d-M-BLK", prefix, num)
		simpleSku := mustSimpleSku(raw)

		wl := emptyWishlist()
		intent, err := wl.AddItem(simpleSku)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if intent.SimpleSku.String() != raw {
			t.Fatalf("AddItemIntent.SimpleSku: got %q, want %q", intent.SimpleSku.String(), raw)
		}
	})
}

// --- 4.6.5: AddItem success on non-conflicting wishlist ---

func TestWishlist_AddItem_SuccessNoConflict(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Existing item uses "AA-001"
		wl := wishlistWithItem("AA-001")

		// New item uses a different prefix-num pair
		prefix, num := genSkuParts(t, "sku")
		// Skip if it would collide with "AA-001"
		if fmt.Sprintf("%s-%03d", prefix, num) == "AA-001" {
			t.Skip()
		}

		raw := fmt.Sprintf("%s-%03d-M-BLK", prefix, num)
		simpleSku := mustSimpleSku(raw)

		_, err := wl.AddItem(simpleSku)

		if err != nil {
			t.Fatalf("expected success for non-conflicting sku %q, got: %v", raw, err)
		}
	})
}

// --- 4.6.6: AddItemIntent carries correct configSku ---

func TestWishlist_AddItem_IntentConfigSku(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		prefix, num := genSkuParts(t, "sku")
		rawSimple := fmt.Sprintf("%s-%03d-S-RED", prefix, num)
		expectedConfig := fmt.Sprintf("%s-%03d", prefix, num)

		wl := emptyWishlist()
		intent, err := wl.AddItem(mustSimpleSku(rawSimple))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if intent.ConfigSku.String() != expectedConfig {
			t.Fatalf("AddItemIntent.ConfigSku: got %q, want %q", intent.ConfigSku.String(), expectedConfig)
		}
	})
}

// --- 4.6.7: RemoveItem always returns intent (no error) ---

func TestWishlist_RemoveItem_AlwaysSucceeds(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Random wishlist: empty or with one item
		hasItem := rapid.Bool().Draw(t, "hasItem")

		prefix, num := genSkuParts(t, "sku")
		configSkuStr := fmt.Sprintf("%s-%03d", prefix, num)

		var wl *Wishlist
		if hasItem {
			wl = wishlistWithItem(configSkuStr)
		} else {
			wl = emptyWishlist()
		}

		configSku := mustConfigSku(configSkuStr)
		intent := wl.RemoveItem(configSku)

		if intent.ConfigSku.String() != configSkuStr {
			t.Fatalf("RemoveItemIntent.ConfigSku: got %q, want %q", intent.ConfigSku.String(), configSkuStr)
		}
	})
}

// --- 4.6.8: ToggleItem returns RemoveIntent when item present ---

func TestWishlist_ToggleItem_PresentReturnsRemoveIntent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		prefix, num := genSkuParts(t, "sku")
		configSkuStr := fmt.Sprintf("%s-%03d", prefix, num)

		wl := wishlistWithItem(configSkuStr)

		simpleSku := mustSimpleSku(fmt.Sprintf("%s-%03d-M-BLK", prefix, num))
		configSku := mustConfigSku(configSkuStr)

		result, err := wl.ToggleItem(simpleSku, configSku)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := result.(RemoveItemIntent); !ok {
			t.Fatalf("expected RemoveItemIntent, got %T", result)
		}
	})
}

// --- 4.6.9: ToggleItem returns AddIntent when item absent ---

func TestWishlist_ToggleItem_AbsentReturnsAddIntent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		prefix, num := genSkuParts(t, "sku")
		configSkuStr := fmt.Sprintf("%s-%03d", prefix, num)

		wl := emptyWishlist()

		simpleSku := mustSimpleSku(fmt.Sprintf("%s-%03d-M-BLK", prefix, num))
		configSku := mustConfigSku(configSkuStr)

		result, err := wl.ToggleItem(simpleSku, configSku)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := result.(AddItemIntent); !ok {
			t.Fatalf("expected AddItemIntent, got %T", result)
		}
	})
}

// --- 4.6.10: ToggleItem idempotency (double-toggle leaves state consistent) ---

func TestWishlist_ToggleItem_Idempotency(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		prefix, num := genSkuParts(t, "sku")
		configSkuStr := fmt.Sprintf("%s-%03d", prefix, num)

		wl := emptyWishlist()

		simpleSku := mustSimpleSku(fmt.Sprintf("%s-%03d-M-BLK", prefix, num))
		configSku := mustConfigSku(configSkuStr)

		// First toggle: absent → AddItemIntent
		result1, err := wl.ToggleItem(simpleSku, configSku)
		if err != nil {
			t.Fatalf("first toggle error: %v", err)
		}
		if _, ok := result1.(AddItemIntent); !ok {
			t.Fatalf("first toggle: expected AddItemIntent, got %T", result1)
		}

		// Simulate the item being added by inserting it into Items
		iid, _ := valueobject.NewWishlistItemId("item-x")
		wl.Items = append(wl.Items, entity.WishlistItem{
			ItemId:    iid,
			SimpleSku: simpleSku,
			ConfigSku: configSku,
		})

		// Second toggle: present → RemoveItemIntent
		result2, err := wl.ToggleItem(simpleSku, configSku)
		if err != nil {
			t.Fatalf("second toggle error: %v", err)
		}
		if _, ok := result2.(RemoveItemIntent); !ok {
			t.Fatalf("second toggle: expected RemoveItemIntent, got %T", result2)
		}

		// Remove the item from Items (simulate removal)
		remaining := wl.Items[:0]
		for _, item := range wl.Items {
			if !strings.EqualFold(item.ConfigSku.String(), configSkuStr) {
				remaining = append(remaining, item)
			}
		}
		wl.Items = remaining

		// Third toggle: absent again → AddItemIntent (no state corruption)
		result3, err := wl.ToggleItem(simpleSku, configSku)
		if err != nil {
			t.Fatalf("third toggle error: %v", err)
		}
		if _, ok := result3.(AddItemIntent); !ok {
			t.Fatalf("third toggle: expected AddItemIntent, got %T", result3)
		}
	})
}
