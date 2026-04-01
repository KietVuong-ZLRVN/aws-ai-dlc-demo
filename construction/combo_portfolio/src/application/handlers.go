package application

import (
	"context"
	"fmt"
	"log"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/acl"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/services"
)

// SaveComboHandler handles the SaveComboCommand.
type SaveComboHandler struct {
	repo domain.ComboRepository
}

func NewSaveComboHandler(repo domain.ComboRepository) *SaveComboHandler {
	return &SaveComboHandler{repo: repo}
}

func (h *SaveComboHandler) Handle(ctx context.Context, cmd SaveComboCommand) (domain.ComboId, error) {
	name, err := domain.NewComboName(cmd.Name)
	if err != nil {
		return "", err
	}
	id := domain.NewComboId()
	combo, err := domain.NewCombo(id, cmd.ShopperID, name, cmd.Items, cmd.Visibility)
	if err != nil {
		return "", err
	}
	if err := h.repo.Save(ctx, combo); err != nil {
		return "", fmt.Errorf("save combo: %w", err)
	}
	logEvents(combo.PopEvents())
	return id, nil
}

// RenameComboHandler handles the RenameComboCommand.
type RenameComboHandler struct {
	repo domain.ComboRepository
}

func NewRenameComboHandler(repo domain.ComboRepository) *RenameComboHandler {
	return &RenameComboHandler{repo: repo}
}

func (h *RenameComboHandler) Handle(ctx context.Context, cmd RenameComboCommand) error {
	combo, err := loadOwnedCombo(ctx, h.repo, cmd.ComboID, cmd.ShopperID)
	if err != nil {
		return err
	}
	newName, err := domain.NewComboName(cmd.NewName)
	if err != nil {
		return err
	}
	combo.Rename(newName)
	if err := h.repo.Save(ctx, combo); err != nil {
		return fmt.Errorf("save renamed combo: %w", err)
	}
	logEvents(combo.PopEvents())
	return nil
}

// DeleteComboHandler handles the DeleteComboCommand.
type DeleteComboHandler struct {
	repo domain.ComboRepository
}

func NewDeleteComboHandler(repo domain.ComboRepository) *DeleteComboHandler {
	return &DeleteComboHandler{repo: repo}
}

func (h *DeleteComboHandler) Handle(ctx context.Context, cmd DeleteComboCommand) error {
	combo, err := loadOwnedCombo(ctx, h.repo, cmd.ComboID, cmd.ShopperID)
	if err != nil {
		return err
	}
	if err := h.repo.Delete(ctx, combo.ID()); err != nil {
		return fmt.Errorf("delete combo: %w", err)
	}
	logEvents([]domain.DomainEvent{domain.ComboDeleted{
		ComboId:   combo.ID().String(),
		ShopperId: combo.ShopperID().String(),
	}})
	return nil
}

// ShareComboHandler handles the ShareComboCommand.
type ShareComboHandler struct {
	repo         domain.ComboRepository
	tokenService *services.ShareTokenService
}

func NewShareComboHandler(repo domain.ComboRepository, tokenService *services.ShareTokenService) *ShareComboHandler {
	return &ShareComboHandler{repo: repo, tokenService: tokenService}
}

func (h *ShareComboHandler) Handle(ctx context.Context, cmd ShareComboCommand) (domain.ShareToken, error) {
	combo, err := loadOwnedCombo(ctx, h.repo, cmd.ComboID, cmd.ShopperID)
	if err != nil {
		return "", err
	}
	token, err := h.tokenService.Generate(ctx)
	if err != nil {
		return "", fmt.Errorf("generate share token: %w", err)
	}
	combo.Share(token)
	if err := h.repo.Save(ctx, combo); err != nil {
		return "", fmt.Errorf("save shared combo: %w", err)
	}
	logEvents(combo.PopEvents())
	return token, nil
}

// MakePrivateHandler handles the MakePrivateCommand.
type MakePrivateHandler struct {
	repo domain.ComboRepository
}

func NewMakePrivateHandler(repo domain.ComboRepository) *MakePrivateHandler {
	return &MakePrivateHandler{repo: repo}
}

func (h *MakePrivateHandler) Handle(ctx context.Context, cmd MakePrivateCommand) error {
	combo, err := loadOwnedCombo(ctx, h.repo, cmd.ComboID, cmd.ShopperID)
	if err != nil {
		return err
	}
	combo.MakePrivate()
	if err := h.repo.Save(ctx, combo); err != nil {
		return fmt.Errorf("save private combo: %w", err)
	}
	logEvents(combo.PopEvents())
	return nil
}

// GetComboHandler handles the GetComboQuery.
type GetComboHandler struct {
	repo       domain.ComboRepository
	enrichment *acl.ComboEnrichmentService
}

func NewGetComboHandler(repo domain.ComboRepository, enrichment *acl.ComboEnrichmentService) *GetComboHandler {
	return &GetComboHandler{repo: repo, enrichment: enrichment}
}

func (h *GetComboHandler) Handle(ctx context.Context, q GetComboQuery) (*domain.EnrichedCombo, error) {
	combo, err := loadOwnedCombo(ctx, h.repo, q.ComboID, q.ShopperID)
	if err != nil {
		return nil, err
	}
	return h.enrichment.Enrich(ctx, combo)
}

// ListCombosHandler handles the ListCombosQuery.
type ListCombosHandler struct {
	repo       domain.ComboRepository
	enrichment *acl.ComboEnrichmentService
}

func NewListCombosHandler(repo domain.ComboRepository, enrichment *acl.ComboEnrichmentService) *ListCombosHandler {
	return &ListCombosHandler{repo: repo, enrichment: enrichment}
}

func (h *ListCombosHandler) Handle(ctx context.Context, q ListCombosQuery) ([]*domain.EnrichedCombo, error) {
	combos, err := h.repo.FindByShopperId(ctx, q.ShopperID)
	if err != nil {
		return nil, fmt.Errorf("list combos: %w", err)
	}
	result := make([]*domain.EnrichedCombo, 0, len(combos))
	for _, c := range combos {
		enriched, err := h.enrichment.Enrich(ctx, c)
		if err != nil {
			return nil, err
		}
		result = append(result, enriched)
	}
	return result, nil
}

// GetSharedComboHandler handles the GetSharedComboQuery (unauthenticated).
type GetSharedComboHandler struct {
	repo       domain.ComboRepository
	enrichment *acl.ComboEnrichmentService
}

func NewGetSharedComboHandler(repo domain.ComboRepository, enrichment *acl.ComboEnrichmentService) *GetSharedComboHandler {
	return &GetSharedComboHandler{repo: repo, enrichment: enrichment}
}

func (h *GetSharedComboHandler) Handle(ctx context.Context, q GetSharedComboQuery) (*domain.EnrichedCombo, error) {
	combo, err := h.repo.FindByShareToken(ctx, q.ShareToken)
	if err != nil {
		return nil, err
	}
	return h.enrichment.Enrich(ctx, combo)
}

// loadOwnedCombo loads a combo and asserts ownership.
func loadOwnedCombo(ctx context.Context, repo domain.ComboRepository, id domain.ComboId, shopperID domain.ShopperId) (*domain.Combo, error) {
	combo, err := repo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	if !combo.OwnedBy(shopperID) {
		return nil, domain.ErrComboAccessDenied
	}
	return combo, nil
}

// logEvents logs domain events in-process (no external bus at this stage).
func logEvents(events []domain.DomainEvent) {
	for _, e := range events {
		log.Printf("[DomainEvent] %s: %+v", e.EventName(), e)
	}
}
