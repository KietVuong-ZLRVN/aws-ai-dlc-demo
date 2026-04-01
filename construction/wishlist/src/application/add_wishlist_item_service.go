package application

import (
	"context"
	"errors"
	"time"
	"wishlist/domain/aggregate"
	"wishlist/domain/event"
	"wishlist/domain/repository"
	"wishlist/domain/service"
	"wishlist/domain/valueobject"
)

// ErrWishlistItemAlreadyPresent is returned when an item is already in the wishlist.
var ErrWishlistItemAlreadyPresent = errors.New("item already in wishlist")

// AddResult holds the result of a successful add operation.
type AddResult struct {
	ItemId    string
	SimpleSku string
	ConfigSku string
}

type AddWishlistItemService struct {
	repo     repository.WishlistRepository
	auth     service.AuthSessionService
	eventBus EventBus
}

func NewAddWishlistItemService(repo repository.WishlistRepository, auth service.AuthSessionService, eventBus EventBus) *AddWishlistItemService {
	return &AddWishlistItemService{repo: repo, auth: auth, eventBus: eventBus}
}

func (s *AddWishlistItemService) Execute(ctx context.Context, sessionToken string, simpleSku valueobject.SimpleSku, returnPath string) (*AddResult, error) {
	shopperID, err := s.auth.ResolveShopperID(ctx, sessionToken)
	if err != nil {
		// Publish authentication gate triggered event
		s.eventBus.Publish(ctx, event.AuthenticationGateTriggered{
			SimpleSku:  simpleSku.String(),
			ReturnPath: returnPath,
			OccurredAt: time.Now(),
		})
		return nil, service.ErrUnauthenticated
	}

	// Load current wishlist for duplicate check (fetch up to 100 items)
	pagination, _ := valueobject.NewPagination(0, 100)
	wishlist, err := s.repo.GetByShopperId(ctx, shopperID, pagination)
	if err != nil {
		return nil, err
	}

	// Check for duplicates via the aggregate
	intent, err := wishlist.AddItem(simpleSku)
	if err != nil {
		if errors.Is(err, aggregate.ErrWishlistItemAlreadyPresent) {
			return nil, ErrWishlistItemAlreadyPresent
		}
		return nil, err
	}

	// Persist the new item
	itemId, err := s.repo.AddItem(ctx, shopperID, simpleSku, intent.ConfigSku)
	if err != nil {
		return nil, err
	}

	// Publish event
	s.eventBus.Publish(ctx, event.WishlistItemAdded{
		ShopperID:  shopperID.String(),
		SimpleSku:  simpleSku.String(),
		ConfigSku:  intent.ConfigSku.String(),
		ItemId:     itemId.String(),
		OccurredAt: time.Now(),
	})

	return &AddResult{
		ItemId:    itemId.String(),
		SimpleSku: simpleSku.String(),
		ConfigSku: intent.ConfigSku.String(),
	}, nil
}
