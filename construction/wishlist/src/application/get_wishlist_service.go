package application

import (
	"context"
	"wishlist/domain/aggregate"
	"wishlist/domain/repository"
	"wishlist/domain/service"
	"wishlist/domain/valueobject"
)

type GetWishlistService struct {
	repo repository.WishlistRepository
	auth service.AuthSessionService
}

func NewGetWishlistService(repo repository.WishlistRepository, auth service.AuthSessionService) *GetWishlistService {
	return &GetWishlistService{repo: repo, auth: auth}
}

func (s *GetWishlistService) Execute(ctx context.Context, sessionToken string, pagination valueobject.Pagination) (*aggregate.Wishlist, error) {
	shopperID, err := s.auth.ResolveShopperID(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	wishlist, err := s.repo.GetByShopperId(ctx, shopperID, pagination)
	if err != nil {
		return nil, err
	}

	return wishlist, nil
}
