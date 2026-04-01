package api

import (
	"encoding/json"
	"testing"

	"pgregory.net/rapid"
)

// requiredFields lists the nine fields Unit 3 depends on from GET /api/v1/wishlist items.
var requiredFields = []string{
	"itemId",
	"simpleSku",
	"configSku",
	"name",
	"brand",
	"price",
	"imageUrl",
	"color",
	"size",
	"inStock",
}

// 4.11.1: example — all Unit-3-required fields present
func TestContract_WishlistItemResponseContainsRequiredFields(t *testing.T) {
	item := wishlistItemResponse{
		ItemId:    "item-001",
		SimpleSku: "PD-001-M-BLK",
		ConfigSku: "PD-001",
		Name:      "Classic Tee",
		Brand:     "ZaloraBrand",
		Price:     moneyResponse{Amount: 49.90, Currency: "SGD"},
		ImageUrl:  "https://img.example.com/img.jpg",
		Color:     "black",
		Size:      "M",
		InStock:   true,
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	for _, field := range requiredFields {
		if _, ok := m[field]; !ok {
			t.Errorf("required field %q missing from JSON output", field)
		}
	}
}

// 4.11.2 PBT: required fields never omitted regardless of field values
func TestContract_RequiredFieldsNeverOmitted(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		item := wishlistItemResponse{
			ItemId:    rapid.StringMatching(`item-[0-9]+`).Draw(t, "itemId"),
			SimpleSku: rapid.StringMatching(`[A-Z]{2}-[0-9]{3}-[A-Z]{1,2}-[A-Z]{3}`).Draw(t, "simpleSku"),
			ConfigSku: rapid.StringMatching(`[A-Z]{2}-[0-9]{3}`).Draw(t, "configSku"),
			Name:      rapid.StringMatching(`[A-Za-z ]{3,15}`).Draw(t, "name"),
			Brand:     rapid.StringMatching(`[A-Za-z]{3,10}`).Draw(t, "brand"),
			Price: moneyResponse{
				Amount:   rapid.Float64Range(0, 999).Draw(t, "amount"),
				Currency: "SGD",
			},
			ImageUrl: "https://img.example.com/img.jpg",
			Color:    rapid.SampledFrom([]string{"black", "white", "red"}).Draw(t, "color"),
			Size:     rapid.SampledFrom([]string{"XS", "S", "M", "L", "XL"}).Draw(t, "size"),
			InStock:  rapid.Bool().Draw(t, "inStock"),
		}

		data, err := json.Marshal(item)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}

		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}

		for _, field := range requiredFields {
			if _, ok := m[field]; !ok {
				t.Errorf("required field %q missing from JSON output (inStock=%v)", field, item.InStock)
			}
		}
	})
}
