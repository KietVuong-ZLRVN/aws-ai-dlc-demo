package readmodel

// ProductVariantReadModel represents a single purchasable variant (simple product).
type ProductVariantReadModel struct {
	SimpleSku string `json:"simpleSku"`
	Size      string `json:"size"`
	Color     string `json:"color"`
	InStock   bool   `json:"inStock"`
}

// ProductDetailReadModel is the full product detail response.
type ProductDetailReadModel struct {
	ConfigSku string                    `json:"configSku"`
	Name      string                    `json:"name"`
	Brand     string                    `json:"brand"`
	Price     Money                     `json:"price"`
	ImageUrl  string                    `json:"imageUrl"`
	Slug      string                    `json:"slug"`
	InStock   bool                      `json:"inStock"`
	Colors    []string                  `json:"colors"`
	Variants  []ProductVariantReadModel `json:"variants"`
	Occasions []string                  `json:"occasions"`
}
