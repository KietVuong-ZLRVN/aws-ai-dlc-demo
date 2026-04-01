package service

import (
	"context"
	"errors"
	"wishlist/domain/valueobject"
)

var ErrUnauthenticated = errors.New("unauthenticated")

type AuthSessionService interface {
	ResolveShopperID(ctx context.Context, sessionToken string) (valueobject.ShopperId, error)
}
