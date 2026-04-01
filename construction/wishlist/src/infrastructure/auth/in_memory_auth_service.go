package auth

import (
	"context"
	"wishlist/domain/service"
	"wishlist/domain/valueobject"
)

type InMemoryAuthService struct{}

func (s *InMemoryAuthService) ResolveShopperID(ctx context.Context, sessionToken string) (valueobject.ShopperId, error) {
	if sessionToken == "" {
		return valueobject.ShopperId{}, service.ErrUnauthenticated
	}
	if sessionToken == "demo-session-token" {
		shopperID, _ := valueobject.NewShopperId("shopper-123")
		return shopperID, nil
	}
	// Any other non-empty token maps to a shopper derived from the token value.
	shopperID, _ := valueobject.NewShopperId("shopper-" + sessionToken)
	return shopperID, nil
}
