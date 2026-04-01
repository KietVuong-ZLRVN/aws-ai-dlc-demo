package domain

import (
	"time"
)

// Combo is the aggregate root of the Combo Portfolio bounded context.
// It enforces all invariants related to a saved combo's lifecycle.
type Combo struct {
	id         ComboId
	shopperID  ShopperId
	name       ComboName
	items      []ComboItem
	visibility Visibility
	shareToken *ShareToken
	createdAt  time.Time
	updatedAt  time.Time
	events     []DomainEvent
}

// NewCombo creates a new Combo aggregate, enforcing all creation invariants.
func NewCombo(id ComboId, shopperID ShopperId, name ComboName, items []ComboItem, visibility Visibility) (*Combo, error) {
	if err := validateItems(items); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	c := &Combo{
		id:         id,
		shopperID:  shopperID,
		name:       name,
		items:      items,
		visibility: visibility,
		shareToken: nil,
		createdAt:  now,
		updatedAt:  now,
	}
	c.addEvent(ComboCreated{
		ComboId:   id.String(),
		ShopperId: shopperID.String(),
		Name:      name.String(),
		ItemCount: len(items),
		CreatedAt: now,
	})
	return c, nil
}

// ReconstitueCombo rebuilds a Combo from persistence without firing events.
func ReconstitueCombo(id ComboId, shopperID ShopperId, name ComboName, items []ComboItem, visibility Visibility, shareToken *ShareToken, createdAt, updatedAt time.Time) *Combo {
	return &Combo{
		id:         id,
		shopperID:  shopperID,
		name:       name,
		items:      items,
		visibility: visibility,
		shareToken: shareToken,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

// Rename changes the combo's name.
func (c *Combo) Rename(newName ComboName) {
	old := c.name
	c.name = newName
	c.updatedAt = time.Now().UTC()
	c.addEvent(ComboRenamed{
		ComboId:   c.id.String(),
		ShopperId: c.shopperID.String(),
		OldName:   old.String(),
		NewName:   newName.String(),
		RenamedAt: c.updatedAt,
	})
}

// Share generates a share token and sets visibility to public.
func (c *Combo) Share(token ShareToken) {
	c.shareToken = &token
	c.visibility = VisibilityPublic
	c.updatedAt = time.Now().UTC()
	c.addEvent(ComboShared{
		ComboId:    c.id.String(),
		ShopperId:  c.shopperID.String(),
		ShareToken: token.String(),
		SharedAt:   c.updatedAt,
	})
}

// MakePrivate sets visibility to private and atomically revokes the share token.
func (c *Combo) MakePrivate() {
	revokedToken := ""
	if c.shareToken != nil {
		revokedToken = c.shareToken.String()
	}
	c.visibility = VisibilityPrivate
	c.shareToken = nil
	c.updatedAt = time.Now().UTC()
	c.addEvent(ComboMadePrivate{
		ComboId:           c.id.String(),
		ShopperId:         c.shopperID.String(),
		RevokedShareToken: revokedToken,
		MadePrivateAt:     c.updatedAt,
	})
}

// OwnedBy verifies the combo belongs to the given shopper.
func (c *Combo) OwnedBy(shopperID ShopperId) bool {
	return c.shopperID == shopperID
}

// PopEvents returns all pending domain events and clears the internal list.
func (c *Combo) PopEvents() []DomainEvent {
	evts := c.events
	c.events = nil
	return evts
}

// Getters

func (c *Combo) ID() ComboId             { return c.id }
func (c *Combo) ShopperID() ShopperId    { return c.shopperID }
func (c *Combo) Name() ComboName         { return c.name }
func (c *Combo) Items() []ComboItem      { return c.items }
func (c *Combo) Visibility() Visibility  { return c.visibility }
func (c *Combo) ShareToken() *ShareToken { return c.shareToken }
func (c *Combo) CreatedAt() time.Time    { return c.createdAt }
func (c *Combo) UpdatedAt() time.Time    { return c.updatedAt }

// addEvent appends a domain event to the pending list.
func (c *Combo) addEvent(e DomainEvent) {
	c.events = append(c.events, e)
}

// validateItems enforces the ComboItemCountPolicy and UniqueItemPolicy.
func validateItems(items []ComboItem) error {
	if len(items) < 2 || len(items) > 10 {
		return ErrInvalidItemCount
	}
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if _, exists := seen[item.SimpleSku]; exists {
			return ErrDuplicateItem
		}
		seen[item.SimpleSku] = struct{}{}
	}
	return nil
}
