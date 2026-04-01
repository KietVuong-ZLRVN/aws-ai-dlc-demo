package main

import (
	"context"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/infrastructure/acl"
)

// stubComboPortfolioACL returns a fixed set of CartItems regardless of comboId.
type stubComboPortfolioACL struct{}

func (s *stubComboPortfolioACL) FetchComboItems(_ context.Context, _ string, _ string) ([]domain.CartItem, error) {
	return []domain.CartItem{
		{SimpleSku: "STUB-SKU-001", Quantity: 1, Size: "M"},
		{SimpleSku: "STUB-SKU-002", Quantity: 1, Size: ""},
		{SimpleSku: "STUB-SKU-003", Quantity: 1, Size: "42"},
	}, nil
}

// stubPlatformCartACL simulates different platform cart outcomes.
type stubPlatformCartACL struct {
	scenario string // "ok" | "partial" | "failed"
}

func (s *stubPlatformCartACL) BulkAddToCart(_ context.Context, items []domain.CartItem, _ string) (acl.CartSubmissionResult, error) {
	switch s.scenario {
	case "ok":
		return acl.CartSubmissionResult{AddedItems: items}, nil

	case "partial":
		if len(items) == 0 {
			return acl.CartSubmissionResult{}, nil
		}
		added := items[:1]
		skipped := make([]domain.SkippedItem, 0, len(items)-1)
		for _, it := range items[1:] {
			skipped = append(skipped, domain.SkippedItem{SimpleSku: it.SimpleSku, Reason: "out_of_stock"})
		}
		return acl.CartSubmissionResult{AddedItems: added, SkippedItems: skipped}, nil

	case "failed":
		return acl.CartSubmissionResult{}, domain.ErrPlatformCartUnavailable

	default:
		return acl.CartSubmissionResult{AddedItems: items}, nil
	}
}
