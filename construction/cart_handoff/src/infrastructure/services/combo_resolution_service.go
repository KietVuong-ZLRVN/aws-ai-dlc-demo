package services

import (
	"context"
	"fmt"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/infrastructure/acl"
)

// ComboResolutionService resolves the CartItem list from a HandoffSource.
type ComboResolutionService struct {
	comboPort acl.ComboPortfolioPort
}

func NewComboResolutionService(comboPort acl.ComboPortfolioPort) *ComboResolutionService {
	return &ComboResolutionService{comboPort: comboPort}
}

// Resolve returns the CartItems for the given HandoffSource.
func (s *ComboResolutionService) Resolve(ctx context.Context, source domain.HandoffSource, sessionCookie string) ([]domain.CartItem, error) {
	switch source.Type {
	case domain.HandoffSourceTypeSavedCombo:
		items, err := s.comboPort.FetchComboItems(ctx, source.ComboId, sessionCookie)
		if err != nil {
			return nil, err
		}
		return items, nil
	case domain.HandoffSourceTypeInlineItems:
		return source.Items, nil
	default:
		return nil, domain.ErrInvalidHandoffSource
	}
}

// CartSubmissionService submits resolved CartItems to the platform cart.
type CartSubmissionService struct {
	cartPort acl.PlatformCartPort
}

func NewCartSubmissionService(cartPort acl.PlatformCartPort) *CartSubmissionService {
	return &CartSubmissionService{cartPort: cartPort}
}

// Submit sends the items to the platform cart and returns a classified result.
func (s *CartSubmissionService) Submit(ctx context.Context, items []domain.CartItem, sessionCookie string) (acl.CartSubmissionResult, error) {
	result, err := s.cartPort.BulkAddToCart(ctx, items, sessionCookie)
	if err != nil {
		return acl.CartSubmissionResult{}, fmt.Errorf("cart submission: %w", err)
	}
	return result, nil
}
