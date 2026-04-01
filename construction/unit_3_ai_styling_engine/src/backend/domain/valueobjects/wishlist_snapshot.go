package valueobjects

type WishlistSnapshot struct {
	Items      []WishlistItem
	TotalCount int
}

// InStockItems returns only items that are currently in stock.
func (ws WishlistSnapshot) InStockItems() []WishlistItem {
	var result []WishlistItem
	for _, item := range ws.Items {
		if item.InStock {
			result = append(result, item)
		}
	}
	return result
}
