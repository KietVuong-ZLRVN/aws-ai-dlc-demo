package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
)

// ShareTokenService generates globally unique ShareTokens.
type ShareTokenService struct {
	repo domain.ComboRepository
}

func NewShareTokenService(repo domain.ComboRepository) *ShareTokenService {
	return &ShareTokenService{repo: repo}
}

// Generate creates a unique share token, retrying on the (near-zero) chance of a collision.
func (s *ShareTokenService) Generate(ctx context.Context) (domain.ShareToken, error) {
	for i := 0; i < 3; i++ {
		token := domain.ShareToken(uuid.NewString())
		existing, err := s.repo.FindByShareToken(ctx, token)
		if err == domain.ErrComboNotFound {
			return token, nil
		}
		if err != nil {
			return "", err
		}
		if existing == nil {
			return token, nil
		}
	}
	return "", fmt.Errorf("%w", domain.ErrShareTokenConflict)
}
