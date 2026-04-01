package platform

// platformProduct is the internal seed/storage struct for the in-memory client.
// It mirrors port.PlatformProduct but lives in the infra layer to avoid any
// import cycle (infra imports domain/port; domain/port must not import infra).
type platformProduct struct {
	ConfigSku string
	Name      string
	Brand     string
	Category  string
	Price     float64
	Currency  string
	Colors    []string
	Simples   []platformSimple
	ImageUrl  string
	UrlKey    string
	Occasions []string
}

// platformSimple is the internal variant struct.
type platformSimple struct {
	SimpleSku string
	Size      string
	Color     string
	Quantity  int
}
