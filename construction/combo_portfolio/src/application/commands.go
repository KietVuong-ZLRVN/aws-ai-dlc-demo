package application

import "github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"

// SaveComboCommand creates or replaces a shopper's combo.
type SaveComboCommand struct {
	ShopperID  domain.ShopperId
	Name       string
	Items      []domain.ComboItem
	Visibility domain.Visibility
}

// RenameComboCommand renames an existing combo.
type RenameComboCommand struct {
	ComboID   domain.ComboId
	ShopperID domain.ShopperId
	NewName   string
}

// DeleteComboCommand deletes a combo owned by the shopper.
type DeleteComboCommand struct {
	ComboID   domain.ComboId
	ShopperID domain.ShopperId
}

// ShareComboCommand makes a combo publicly shareable and returns its share token.
type ShareComboCommand struct {
	ComboID   domain.ComboId
	ShopperID domain.ShopperId
}

// MakePrivateCommand revokes the share link for a combo.
type MakePrivateCommand struct {
	ComboID   domain.ComboId
	ShopperID domain.ShopperId
}

// GetComboQuery fetches a single combo owned by the shopper with enrichment.
type GetComboQuery struct {
	ComboID   domain.ComboId
	ShopperID domain.ShopperId
}

// ListCombosQuery lists all combos for a shopper with enrichment.
type ListCombosQuery struct {
	ShopperID domain.ShopperId
}

// GetSharedComboQuery fetches a publicly shared combo by its share token.
type GetSharedComboQuery struct {
	ShareToken domain.ShareToken
}
