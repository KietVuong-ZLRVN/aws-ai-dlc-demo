package persistence

import (
	"context"
	"fmt"
	"sync"
	"wishlist/domain/aggregate"
	"wishlist/domain/assembler"
	"wishlist/domain/valueobject"
)

// catalogSimple holds the details for a single simple SKU in the embedded product catalog.
type catalogSimple struct {
	Name      string
	Brand     string
	Price     float64
	Currency  string
	ImageUrl  string
	Color     string
	Size      string
	InStock   bool
}

// embeddedCatalog maps simpleSku -> catalogSimple.
var embeddedCatalog = map[string]catalogSimple{
	// PD-001: Classic White Tee
	"PD-001-S": {Name: "Classic White Tee", Brand: "Uniqlo", Price: 29.90, Currency: "SGD", ImageUrl: "/images/pd-001.jpg", Color: "white", Size: "S", InStock: true},
	"PD-001-M": {Name: "Classic White Tee", Brand: "Uniqlo", Price: 29.90, Currency: "SGD", ImageUrl: "/images/pd-001.jpg", Color: "white", Size: "M", InStock: true},
	"PD-001-L": {Name: "Classic White Tee", Brand: "Uniqlo", Price: 29.90, Currency: "SGD", ImageUrl: "/images/pd-001.jpg", Color: "white", Size: "L", InStock: false},

	// PD-002: Slim Fit Jeans
	"PD-002-28": {Name: "Slim Fit Jeans", Brand: "Levi's", Price: 89.90, Currency: "SGD", ImageUrl: "/images/pd-002.jpg", Color: "blue", Size: "28", InStock: true},
	"PD-002-30": {Name: "Slim Fit Jeans", Brand: "Levi's", Price: 89.90, Currency: "SGD", ImageUrl: "/images/pd-002.jpg", Color: "blue", Size: "30", InStock: true},
	"PD-002-32": {Name: "Slim Fit Jeans", Brand: "Levi's", Price: 89.90, Currency: "SGD", ImageUrl: "/images/pd-002.jpg", Color: "blue", Size: "32", InStock: true},

	// PD-003: Floral Summer Dress
	"PD-003-XS": {Name: "Floral Summer Dress", Brand: "Zalora", Price: 59.90, Currency: "SGD", ImageUrl: "/images/pd-003.jpg", Color: "pink", Size: "XS", InStock: true},
	"PD-003-S":  {Name: "Floral Summer Dress", Brand: "Zalora", Price: 59.90, Currency: "SGD", ImageUrl: "/images/pd-003.jpg", Color: "pink", Size: "S", InStock: true},
	"PD-003-M":  {Name: "Floral Summer Dress", Brand: "Zalora", Price: 59.90, Currency: "SGD", ImageUrl: "/images/pd-003.jpg", Color: "pink", Size: "M", InStock: true},

	// PD-004: Black Blazer
	"PD-004-S": {Name: "Black Blazer", Brand: "Mango", Price: 129.90, Currency: "SGD", ImageUrl: "/images/pd-004.jpg", Color: "black", Size: "S", InStock: true},
	"PD-004-M": {Name: "Black Blazer", Brand: "Mango", Price: 129.90, Currency: "SGD", ImageUrl: "/images/pd-004.jpg", Color: "black", Size: "M", InStock: true},
	"PD-004-L": {Name: "Black Blazer", Brand: "Mango", Price: 129.90, Currency: "SGD", ImageUrl: "/images/pd-004.jpg", Color: "black", Size: "L", InStock: true},

	// PD-005: Casual Polo Shirt
	"PD-005-S-NVY": {Name: "Casual Polo Shirt", Brand: "Ralph Lauren", Price: 79.90, Currency: "SGD", ImageUrl: "/images/pd-005.jpg", Color: "navy", Size: "S", InStock: true},
	"PD-005-M-NVY": {Name: "Casual Polo Shirt", Brand: "Ralph Lauren", Price: 79.90, Currency: "SGD", ImageUrl: "/images/pd-005.jpg", Color: "navy", Size: "M", InStock: true},
	"PD-005-S-WHT": {Name: "Casual Polo Shirt", Brand: "Ralph Lauren", Price: 79.90, Currency: "SGD", ImageUrl: "/images/pd-005.jpg", Color: "white", Size: "S", InStock: true},
	"PD-005-M-WHT": {Name: "Casual Polo Shirt", Brand: "Ralph Lauren", Price: 79.90, Currency: "SGD", ImageUrl: "/images/pd-005.jpg", Color: "white", Size: "M", InStock: true},

	// PD-006: High-Waist Skirt
	"PD-006-XS": {Name: "High-Waist Skirt", Brand: "H&M", Price: 39.90, Currency: "SGD", ImageUrl: "/images/pd-006.jpg", Color: "black", Size: "XS", InStock: false},
	"PD-006-S":  {Name: "High-Waist Skirt", Brand: "H&M", Price: 39.90, Currency: "SGD", ImageUrl: "/images/pd-006.jpg", Color: "black", Size: "S", InStock: true},
	"PD-006-M":  {Name: "High-Waist Skirt", Brand: "H&M", Price: 39.90, Currency: "SGD", ImageUrl: "/images/pd-006.jpg", Color: "black", Size: "M", InStock: true},

	// PD-007: Striped Linen Shirt
	"PD-007-S": {Name: "Striped Linen Shirt", Brand: "Topshop", Price: 49.90, Currency: "SGD", ImageUrl: "/images/pd-007.jpg", Color: "blue", Size: "S", InStock: true},
	"PD-007-M": {Name: "Striped Linen Shirt", Brand: "Topshop", Price: 49.90, Currency: "SGD", ImageUrl: "/images/pd-007.jpg", Color: "blue", Size: "M", InStock: true},
	"PD-007-L": {Name: "Striped Linen Shirt", Brand: "Topshop", Price: 49.90, Currency: "SGD", ImageUrl: "/images/pd-007.jpg", Color: "blue", Size: "L", InStock: false},

	// PD-008: Leopard Print Blouse
	"PD-008-XS": {Name: "Leopard Print Blouse", Brand: "Zara", Price: 55.90, Currency: "SGD", ImageUrl: "/images/pd-008.jpg", Color: "beige", Size: "XS", InStock: true},
	"PD-008-S":  {Name: "Leopard Print Blouse", Brand: "Zara", Price: 55.90, Currency: "SGD", ImageUrl: "/images/pd-008.jpg", Color: "beige", Size: "S", InStock: true},
	"PD-008-M":  {Name: "Leopard Print Blouse", Brand: "Zara", Price: 55.90, Currency: "SGD", ImageUrl: "/images/pd-008.jpg", Color: "beige", Size: "M", InStock: true},
}

// InMemoryWishlistRepository stores wishlist items in memory.
type InMemoryWishlistRepository struct {
	mu         sync.RWMutex
	store      map[string][]assembler.RawWishlistItem
	assembler  *assembler.WishlistAssembler
}

func NewInMemoryWishlistRepository(a *assembler.WishlistAssembler) *InMemoryWishlistRepository {
	return &InMemoryWishlistRepository{
		store:     make(map[string][]assembler.RawWishlistItem),
		assembler: a,
	}
}

func (r *InMemoryWishlistRepository) GetByShopperId(ctx context.Context, shopperID valueobject.ShopperId, pagination valueobject.Pagination) (*aggregate.Wishlist, error) {
	r.mu.RLock()
	items := r.store[shopperID.String()]
	// make a copy to avoid holding the lock during assembly
	copied := make([]assembler.RawWishlistItem, len(items))
	copy(copied, items)
	r.mu.RUnlock()

	// Apply pagination
	offset := pagination.Offset
	if offset > len(copied) {
		offset = len(copied)
	}
	end := offset + pagination.Limit
	if end > len(copied) {
		end = len(copied)
	}
	page := copied[offset:end]

	return r.assembler.Assemble(shopperID, page)
}

func (r *InMemoryWishlistRepository) AddItem(ctx context.Context, shopperID valueobject.ShopperId, simpleSku valueobject.SimpleSku, configSku valueobject.ConfigSku) (valueobject.WishlistItemId, error) {
	itemIdStr := fmt.Sprintf("item-%s-%s", shopperID.String(), simpleSku.String())

	// Look up catalog details
	var raw assembler.RawWishlistItem
	if details, ok := embeddedCatalog[simpleSku.String()]; ok {
		raw = assembler.RawWishlistItem{
			ItemId:    itemIdStr,
			SimpleSku: simpleSku.String(),
			ConfigSku: configSku.String(),
			Name:      details.Name,
			Brand:     details.Brand,
			Price:     details.Price,
			Currency:  details.Currency,
			ImageUrl:  details.ImageUrl,
			Color:     details.Color,
			Size:      details.Size,
			InStock:   details.InStock,
		}
	} else {
		// Placeholder for unknown SKU
		raw = assembler.RawWishlistItem{
			ItemId:    itemIdStr,
			SimpleSku: simpleSku.String(),
			ConfigSku: configSku.String(),
			Name:      "Unknown Product",
			Brand:     "Unknown",
			Price:     0,
			Currency:  "SGD",
			ImageUrl:  "",
			Color:     "",
			Size:      "",
			InStock:   false,
		}
	}

	r.mu.Lock()
	r.store[shopperID.String()] = append(r.store[shopperID.String()], raw)
	r.mu.Unlock()

	itemId, err := valueobject.NewWishlistItemId(itemIdStr)
	if err != nil {
		return valueobject.WishlistItemId{}, fmt.Errorf("creating item id: %w", err)
	}
	return itemId, nil
}

func (r *InMemoryWishlistRepository) RemoveItemByConfigSku(ctx context.Context, shopperID valueobject.ShopperId, configSku valueobject.ConfigSku) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := r.store[shopperID.String()]
	filtered := items[:0]
	for _, item := range items {
		if item.ConfigSku != configSku.String() {
			filtered = append(filtered, item)
		}
	}
	r.store[shopperID.String()] = filtered
	return nil
}
