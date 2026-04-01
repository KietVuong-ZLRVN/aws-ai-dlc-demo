package ai

import (
	"ai-styling-engine/domain/services"
	"ai-styling-engine/domain/valueobjects"
	"fmt"
)

// MockComboCompatibilityScoringService returns realistic hardcoded combo candidates.
// Replaces AWS Bedrock for local development.
type MockComboCompatibilityScoringService struct{}

func NewMockComboCompatibilityScoringService() *MockComboCompatibilityScoringService {
	return &MockComboCompatibilityScoringService{}
}

func (s *MockComboCompatibilityScoringService) Score(input services.ScoringInput) (services.ScoringResult, error) {
	// Collect all available items (wishlist-sourced + supplementary).
	var allItems []valueobjects.ComboItem
	for _, wi := range input.WishlistItems {
		if wi.InStock {
			allItems = append(allItems, valueobjects.ComboItem{
				ConfigSku: wi.ConfigSku,
				SimpleSku: wi.SimpleSku,
				Name:      wi.Name,
				Brand:     wi.Brand,
				Price:     wi.Price,
				ImageUrl:  wi.ImageUrl,
				Source:    valueobjects.ItemSourceWishlist,
			})
		}
	}
	allItems = append(allItems, input.SupplementaryItems...)
	allItems = append(allItems, input.CompleteLookSignals...)

	if len(allItems) < 2 {
		return services.ScoringResult{
			Fallback: &services.ScoringFallback{
				Message: "Your wishlist doesn't have enough items to form a complete combo. Here are some suggestions:",
				Alternatives: []valueobjects.AlternativeItem{
					{
						ConfigSku: "CFG-SHIRT-WHT",
						SimpleSku: "SKU-SHIRT-WHT-M",
						Name:      "Classic White Shirt",
						Brand:     "Uniqlo",
						Price:     39.90,
						ImageUrl:  "https://example.com/images/white-shirt.jpg",
						Reason:    "A versatile shirt that pairs well with most items",
					},
				},
			},
		}, nil
	}

	// Build mock combos from the available items.
	// Combo 1: first item leads (e.g. blazer-forward look).
	// Combo 2: second item leads (e.g. trouser-forward look) — distinct ID for exclusion.
	// Combo 3: if a third item exists, pair it with the first item.
	candidates := []services.ComboCandidate{
		{
			Id:    fmt.Sprintf("combo-%s-%s", allItems[0].ConfigSku, allItems[1].ConfigSku),
			Items: []valueobjects.ComboItem{allItems[0], allItems[1]},
			Score: 0.92,
		},
		{
			Id:    fmt.Sprintf("combo-%s-%s", allItems[1].ConfigSku, allItems[0].ConfigSku),
			Items: []valueobjects.ComboItem{allItems[1], allItems[0]},
			Score: 0.88,
		},
	}
	if len(allItems) >= 3 {
		candidates = append(candidates, services.ComboCandidate{
			Id:    fmt.Sprintf("combo-%s-%s", allItems[0].ConfigSku, allItems[2].ConfigSku),
			Items: []valueobjects.ComboItem{allItems[0], allItems[2]},
			Score: 0.85,
		})
	}

	// Filter out excluded combos.
	var filtered []services.ComboCandidate
	for _, c := range candidates {
		if !input.ExcludedComboIds.Contains(c.Id) {
			filtered = append(filtered, c)
		}
	}
	if len(filtered) == 0 {
		return services.ScoringResult{
			Fallback: &services.ScoringFallback{
				Message:      "We've shown you all available combos. Try adjusting your preferences.",
				Alternatives: []valueobjects.AlternativeItem{},
			},
		}, nil
	}

	return services.ScoringResult{Candidates: filtered}, nil
}
