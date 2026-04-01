package domain

import (
	"context"
)

// ComboRepository is the persistence abstraction for the Combo aggregate.
// The domain layer depends on this interface; the infrastructure layer provides the implementation.
type ComboRepository interface {
	Save(ctx context.Context, combo *Combo) error
	FindById(ctx context.Context, id ComboId) (*Combo, error)
	FindByShopperId(ctx context.Context, shopperID ShopperId) ([]*Combo, error)
	FindByShareToken(ctx context.Context, token ShareToken) (*Combo, error)
	Delete(ctx context.Context, id ComboId) error
}
