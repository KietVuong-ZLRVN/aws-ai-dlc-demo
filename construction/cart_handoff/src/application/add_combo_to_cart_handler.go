package application

import (
	"context"
	"fmt"
	"log"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/infrastructure/services"
)

// AddComboToCartCommand contains all input for adding a combo to the cart.
type AddComboToCartCommand struct {
	ShopperID     domain.ShopperId
	SessionCookie string
	ComboId       string            // set when source is saved_combo
	InlineItems   []domain.CartItem // set when source is inline_items
}

// HandoffResult is returned to the API layer.
type HandoffResult struct {
	Status       domain.HandoffStatus
	AddedItems   []domain.CartItem
	SkippedItems []domain.SkippedItem
}

// AddComboToCartHandler orchestrates: resolve → submit → persist → emit event.
type AddComboToCartHandler struct {
	repo          domain.CartHandoffRecordRepository
	resolutionSvc *services.ComboResolutionService
	submissionSvc *services.CartSubmissionService
}

func NewAddComboToCartHandler(
	repo domain.CartHandoffRecordRepository,
	resolutionSvc *services.ComboResolutionService,
	submissionSvc *services.CartSubmissionService,
) *AddComboToCartHandler {
	return &AddComboToCartHandler{repo: repo, resolutionSvc: resolutionSvc, submissionSvc: submissionSvc}
}

func (h *AddComboToCartHandler) Handle(ctx context.Context, cmd AddComboToCartCommand) (HandoffResult, error) {
	// Build HandoffSource discriminated union.
	var source domain.HandoffSource
	hasBoth := cmd.ComboId != "" && len(cmd.InlineItems) > 0
	if hasBoth {
		return HandoffResult{}, domain.ErrInvalidHandoffSource
	} else if cmd.ComboId != "" {
		source = domain.HandoffSource{
			Type:    domain.HandoffSourceTypeSavedCombo,
			ComboId: cmd.ComboId,
		}
	} else if len(cmd.InlineItems) > 0 {
		source = domain.HandoffSource{
			Type:  domain.HandoffSourceTypeInlineItems,
			Items: cmd.InlineItems,
		}
	} else {
		return HandoffResult{}, domain.ErrInvalidHandoffSource
	}

	// Step 1: Resolve CartItems from the source.
	items, err := h.resolutionSvc.Resolve(ctx, source, cmd.SessionCookie)
	if err != nil {
		return HandoffResult{}, fmt.Errorf("resolve items: %w", err)
	}

	// Step 2: Submit to platform cart.
	submissionResult, err := h.submissionSvc.Submit(ctx, items, cmd.SessionCookie)
	if err != nil {
		// On total platform failure, still persist the audit record with empty added items.
		emptyAdded := []domain.CartItem{}
		skippedAll := make([]domain.SkippedItem, len(items))
		for i, it := range items {
			skippedAll[i] = domain.SkippedItem{SimpleSku: it.SimpleSku, Reason: "platform_error"}
		}
		h.persistAuditRecord(ctx, cmd.ShopperID, source, emptyAdded, skippedAll)
		return HandoffResult{}, fmt.Errorf("submit to cart: %w", err)
	}

	// Step 3: Persist audit record (always).
	h.persistAuditRecord(ctx, cmd.ShopperID, source, submissionResult.AddedItems, submissionResult.SkippedItems)

	return HandoffResult{
		Status:       deriveStatus(submissionResult.AddedItems, submissionResult.SkippedItems),
		AddedItems:   submissionResult.AddedItems,
		SkippedItems: submissionResult.SkippedItems,
	}, nil
}

func (h *AddComboToCartHandler) persistAuditRecord(
	ctx context.Context,
	shopperID domain.ShopperId,
	source domain.HandoffSource,
	added []domain.CartItem,
	skipped []domain.SkippedItem,
) {
	rec, err := domain.NewCartHandoffRecord(shopperID, source, added, skipped)
	if err != nil {
		log.Printf("[WARN] could not create audit record: %v", err)
		return
	}
	if err := h.repo.Save(ctx, rec); err != nil {
		log.Printf("[WARN] could not persist audit record: %v", err)
	}
	for _, e := range rec.PopEvents() {
		log.Printf("[DomainEvent] %s: %+v", e.EventName(), e)
	}
}

func deriveStatus(added []domain.CartItem, skipped []domain.SkippedItem) domain.HandoffStatus {
	switch {
	case len(added) == 0:
		return domain.HandoffStatusFailed
	case len(skipped) > 0:
		return domain.HandoffStatusPartial
	default:
		return domain.HandoffStatusOk
	}
}
