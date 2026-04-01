package usecases

import (
	"ai-styling-engine/application/commands"
	"ai-styling-engine/domain/apperrors"
	"ai-styling-engine/domain/aggregates"
	"ai-styling-engine/domain/entities"
	"ai-styling-engine/domain/events"
	"ai-styling-engine/domain/repositories"
	"ai-styling-engine/domain/services"
	"ai-styling-engine/domain/valueobjects"
	"fmt"

	"github.com/google/uuid"
)

// GenerateCombosUseCase orchestrates the full combo generation pipeline.
type GenerateCombosUseCase struct {
	wishlistRepo    repositories.WishlistRepository
	catalogRepo     repositories.ProductCatalogRepository
	completeLookRepo repositories.CompleteLookRepository
	scoringSvc      services.ComboCompatibilityScoringService
	reasoningSvc    services.ComboReasoningGenerationService
	dispatcher      events.EventDispatcher
}

func NewGenerateCombosUseCase(
	wishlistRepo repositories.WishlistRepository,
	catalogRepo repositories.ProductCatalogRepository,
	completeLookRepo repositories.CompleteLookRepository,
	scoringSvc services.ComboCompatibilityScoringService,
	reasoningSvc services.ComboReasoningGenerationService,
	dispatcher events.EventDispatcher,
) *GenerateCombosUseCase {
	return &GenerateCombosUseCase{
		wishlistRepo:     wishlistRepo,
		catalogRepo:      catalogRepo,
		completeLookRepo: completeLookRepo,
		scoringSvc:       scoringSvc,
		reasoningSvc:     reasoningSvc,
		dispatcher:       dispatcher,
	}
}

func (uc *GenerateCombosUseCase) Execute(cmd commands.GenerateCombosCommand) (ComboGenerationResult, error) {
	sessionId := valueobjects.StyleSessionId(uuid.New().String())

	// Step 1: Initialise session — raises ComboGenerationRequested.
	session := aggregates.NewStyleSession(sessionId, cmd.ShopperSession, cmd.Preferences, cmd.ExcludedIds, uc.dispatcher)

	// Step 2: Fetch wishlist — raises WishlistFetchCompleted, which triggers WishlistSupplementationPolicy.
	snapshot, err := uc.wishlistRepo.FetchForSession(cmd.ShopperSession)
	if err != nil {
		return ComboGenerationResult{}, fmt.Errorf("%w: %v", apperrors.ErrWishlistUnavailable, err)
	}
	session.LoadWishlist(snapshot)

	// Step 3: Fetch supplementary items when the wishlist alone cannot form a combo.
	// Evaluated directly from the snapshot — mirrors the WishlistSupplementationPolicy rule.
	var completeLookSignals []valueobjects.ComboItem
	var catalogItems []valueobjects.ComboItem

	if len(snapshot.InStockItems()) < 2 {
		// Step 3a: Fetch Complete-the-Look signals for each in-stock wishlist item (explicit separate step).
		for _, item := range snapshot.InStockItems() {
			signals, _ := uc.completeLookRepo.FetchCompleteLookSignals(item.ConfigSku)
			completeLookSignals = append(completeLookSignals, signals...)
		}

		// Step 3b: Fetch supplementary catalog items.
		filters := valueobjects.CatalogSearchFiltersFromPreferences(cmd.Preferences)
		catalogItems, _ = uc.catalogRepo.SearchSupplementaryItems(filters)

		merged := deduplicateByConfigSku(append(completeLookSignals, catalogItems...))
		session.LoadCatalogItems(merged)
	}

	// Step 4: AI compatibility scoring.
	scoringInput := services.ScoringInput{
		WishlistItems:       snapshot.InStockItems(),
		SupplementaryItems:  catalogItems,
		CompleteLookSignals: completeLookSignals,
		Preferences:         cmd.Preferences,
		ExcludedComboIds:    cmd.ExcludedIds,
		QuickGenerate:       session.QuickGenerate(),
	}

	result, err := uc.scoringSvc.Score(scoringInput)
	if err != nil {
		return ComboGenerationResult{}, fmt.Errorf("%w: %v", apperrors.ErrAIUnavailable, err)
	}

	// Step 5: Handle fallback.
	if result.IsFallback() {
		fb := entities.NewFallbackResult(result.Fallback.Message, result.Fallback.Alternatives)
		session.TriggerFallback(fb)
		return ComboGenerationResult{Fallback: &ComboGenerationFallback{FallbackResult: fb}}, nil
	}

	// Step 6: Generate reasoning for each candidate.
	var combos []entities.Combo
	for i, candidate := range result.Candidates {
		combo := entities.NewCombo(candidate.Id, candidate.Items, i+1)
		reasoning, err := uc.reasoningSvc.GenerateReasoning(candidate, cmd.Preferences)
		if err != nil {
			reasoning, _ = valueobjects.NewComboReasoning("A great combination of pieces from your wishlist.")
		}
		combo.AttachReasoning(reasoning)
		combos = append(combos, combo)
	}

	// Step 7: Finalise — raises CombosGenerated; ComboExclusionPolicy filters excluded IDs.
	if err := session.CompleteCombos(combos); err != nil {
		return ComboGenerationResult{}, err
	}

	// One retry if exclusion left us with zero combos.
	if session.IsExhausted() && len(result.Candidates) > 0 {
		return ComboGenerationResult{
			Success: &ComboGenerationSuccess{Combos: []entities.Combo{}, Exhausted: true},
		}, nil
	}

	return ComboGenerationResult{
		Success: &ComboGenerationSuccess{
			Combos:    session.Combos(),
			Exhausted: session.IsExhausted(),
		},
	}, nil
}

// deduplicateByConfigSku removes duplicate catalog items by ConfigSku.
func deduplicateByConfigSku(items []valueobjects.ComboItem) []valueobjects.ComboItem {
	seen := make(map[valueobjects.Sku]bool)
	var result []valueobjects.ComboItem
	for _, item := range items {
		if !seen[item.ConfigSku] {
			seen[item.ConfigSku] = true
			result = append(result, item)
		}
	}
	return result
}
