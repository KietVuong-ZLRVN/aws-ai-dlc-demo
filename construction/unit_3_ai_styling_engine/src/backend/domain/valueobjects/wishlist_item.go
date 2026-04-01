package valueobjects

type WishlistItem struct {
	ItemId    string
	SimpleSku Sku
	ConfigSku Sku
	Name      string
	Brand     string
	Price     Money
	ImageUrl  Url
	Color     string
	Size      string
	InStock   bool
}
