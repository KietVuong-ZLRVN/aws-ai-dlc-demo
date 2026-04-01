package application

import (
	"context"
	"time"
	"wishlist/domain/event"
	"wishlist/domain/repository"
	"wishlist/domain/service"
	"wishlist/domain/valueobject"
)

type RemoveWishlistItemService struct {
	repo     repository.WishlistRepository
	auth     service.AuthSessionService
	eventBus EventBus
}

func NewRemoveWishlistItemService(repo repository.WishlistRepository, auth service.AuthSessionService, eventBus EventBus) *RemoveWishlistItemService {
	return &RemoveWishlistItemService{repo: repo, auth: auth, eventBus: eventBus}
}

func (s *RemoveWishlistItemService) Execute(ctx context.Context, sessionToken string, configSku valueobject.ConfigSku) error {
	shopperID, err := s.auth.ResolveShopperID(ctx, sessionToken)
	if err != nil {
		return service.ErrUnauthenticated
	}

	if err := s.repo.RemoveItemByConfigSku(ctx, shopperID, configSku); err != nil {
		return err
	}

	s.eventBus.Publish(ctx, event.WishlistItemRemoved{
		ShopperID:  shopperID.String(),
		ConfigSku:  configSku.String(),
		OccurredAt: time.Now(),
	})

	return nil
}
