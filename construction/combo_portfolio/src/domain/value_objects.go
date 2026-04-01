package domain

import "github.com/google/uuid"

// ComboId uniquely identifies a Combo.
type ComboId string

func NewComboId() ComboId         { return ComboId(uuid.NewString()) }
func (id ComboId) String() string { return string(id) }

// ShopperId identifies the authenticated shopper.
type ShopperId string

func (id ShopperId) String() string { return string(id) }

// ComboName is the validated display name for a Combo.
type ComboName string

func NewComboName(s string) (ComboName, error) {
	if len(s) == 0 || len(s) > 100 {
		return "", ErrInvalidComboName
	}
	return ComboName(s), nil
}

func (n ComboName) String() string { return string(n) }

// ComboItem is a saved item within a Combo.
type ComboItem struct {
	ConfigSku string
	SimpleSku string
	Name      string
	ImageUrl  string
	Price     float64
}

// ShareToken is the opaque token used to share a Combo publicly.
type ShareToken string

func (t ShareToken) String() string { return string(t) }

// Visibility controls whether a Combo is accessible via share link.
type Visibility string

const (
	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
)

// EnrichedComboItem is a ComboItem enriched with live catalog data.
type EnrichedComboItem struct {
	ConfigSku          string
	SimpleSku          string
	Name               string
	Price              float64
	InStock            bool
	ImageUrl           string
	CatalogUnavailable bool
}

// EnrichedCombo is a Combo with enriched items.
type EnrichedCombo struct {
	ID         ComboId
	ShopperID  ShopperId
	Name       ComboName
	Visibility Visibility
	ShareToken *ShareToken
	Items      []EnrichedComboItem
}
