package api

import "github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"

// --- Request DTOs ---

type AddComboToCartRequest struct {
	ComboId string          `json:"comboId"`
	Items   []CartItemInput `json:"items"`
}

type CartItemInput struct {
	SimpleSku string `json:"simpleSku"`
	Quantity  int    `json:"quantity"`
	Size      string `json:"size,omitempty"`
}

// --- Response DTOs ---

type ErrorResponse struct {
	Error string `json:"error"`
}

type HandoffResponse struct {
	Status       string                `json:"status"`
	AddedItems   []CartItemResponse    `json:"addedItems"`
	SkippedItems []SkippedItemResponse `json:"skippedItems"`
}

type CartItemResponse struct {
	SimpleSku string `json:"simpleSku"`
	Quantity  int    `json:"quantity"`
}

type SkippedItemResponse struct {
	SimpleSku string `json:"simpleSku"`
	Reason    string `json:"reason"`
}

// toHandoffResponse maps domain types to the HTTP response.
func toHandoffResponse(status domain.HandoffStatus, added []domain.CartItem, skipped []domain.SkippedItem) HandoffResponse {
	addedResp := make([]CartItemResponse, len(added))
	for i, it := range added {
		addedResp[i] = CartItemResponse{SimpleSku: it.SimpleSku, Quantity: it.Quantity}
	}
	skippedResp := make([]SkippedItemResponse, len(skipped))
	for i, it := range skipped {
		skippedResp[i] = SkippedItemResponse{SimpleSku: it.SimpleSku, Reason: it.Reason}
	}
	return HandoffResponse{
		Status:       string(status),
		AddedItems:   addedResp,
		SkippedItems: skippedResp,
	}
}
