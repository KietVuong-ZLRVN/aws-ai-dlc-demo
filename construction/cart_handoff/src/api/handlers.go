package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/application"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"
)

// Handlers holds all HTTP handler methods.
type Handlers struct {
	addToCart *application.AddComboToCartHandler
}

func NewHandlers(addToCart *application.AddComboToCartHandler) *Handlers {
	return &Handlers{addToCart: addToCart}
}

func (h *Handlers) AddComboToCart(w http.ResponseWriter, r *http.Request) {
	shopperID, _ := ShopperIDFromContext(r.Context())
	sessionCookie := SessionCookieFromContext(r.Context())

	var req AddComboToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Map DTO items to domain CartItems.
	var inlineItems []domain.CartItem
	for _, it := range req.Items {
		inlineItems = append(inlineItems, domain.CartItem{SimpleSku: it.SimpleSku, Quantity: it.Quantity, Size: it.Size})
	}

	result, err := h.addToCart.Handle(r.Context(), application.AddComboToCartCommand{
		ShopperID:     shopperID,
		SessionCookie: sessionCookie,
		ComboId:       req.ComboId,
		InlineItems:   inlineItems,
	})
	if err != nil {
		writeError(w, handoffErrStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toHandoffResponse(result.Status, result.AddedItems, result.SkippedItems))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

func handoffErrStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidHandoffSource):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrComboNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrComboAccessDenied):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrComboPortfolioUnavailable),
		errors.Is(err, domain.ErrPlatformCartUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
