package valueobjects

type AlternativeItem struct {
	ConfigSku Sku
	SimpleSku Sku
	Name      string
	Brand     string
	Price     Money
	ImageUrl  Url
	Reason    string
}
