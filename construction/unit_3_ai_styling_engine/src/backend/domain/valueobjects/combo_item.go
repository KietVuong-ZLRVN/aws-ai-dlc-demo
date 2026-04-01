package valueobjects

type ComboItem struct {
	ConfigSku Sku
	SimpleSku Sku
	Name      string
	Brand     string
	Price     Money
	ImageUrl  Url
	Source    ItemSource
}
