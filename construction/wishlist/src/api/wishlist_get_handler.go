package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"wishlist/application"
	"wishlist/domain/service"
	"wishlist/domain/valueobject"
)

type WishlistGetHandler struct {
	svc *application.GetWishlistService
}

func NewWishlistGetHandler(svc *application.GetWishlistService) *WishlistGetHandler {
	return &WishlistGetHandler{svc: svc}
}

type wishlistItemResponse struct {
	ItemId    string       `json:"itemId"`
	SimpleSku string       `json:"simpleSku"`
	ConfigSku string       `json:"configSku"`
	Name      string       `json:"name"`
	Brand     string       `json:"brand"`
	Price     moneyResponse `json:"price"`
	ImageUrl  string       `json:"imageUrl"`
	Color     string       `json:"color"`
	Size      string       `json:"size"`
	InStock   bool         `json:"inStock"`
}

type moneyResponse struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type wishlistResponse struct {
	Items []wishlistItemResponse `json:"items"`
	Total int                    `json:"total"`
}

func (h *WishlistGetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Extract session cookie
	sessionToken := ""
	if cookie, err := r.Cookie("session"); err == nil {
		sessionToken = cookie.Value
	}

	// Parse pagination query params
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	limit := 20

	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil {
			offset = v
		}
	}
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = v
		}
	}

	pagination, err := valueobject.NewPagination(offset, limit)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ctx := context.WithValue(r.Context(), application.SessionTokenKey, sessionToken)

	wishlist, err := h.svc.Execute(ctx, sessionToken, pagination)
	if err != nil {
		if err == service.ErrUnauthenticated {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "authentication required"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	items := make([]wishlistItemResponse, 0, len(wishlist.Items))
	for _, item := range wishlist.Items {
		items = append(items, wishlistItemResponse{
			ItemId:    item.ItemId.String(),
			SimpleSku: item.SimpleSku.String(),
			ConfigSku: item.ConfigSku.String(),
			Name:      item.Name,
			Brand:     item.Brand,
			Price: moneyResponse{
				Amount:   item.Price.Amount,
				Currency: item.Price.Currency,
			},
			ImageUrl: item.ImageUrl,
			Color:    item.Color,
			Size:     item.Size,
			InStock:  item.InStock,
		})
	}

	writeJSON(w, http.StatusOK, wishlistResponse{
		Items: items,
		Total: wishlist.TotalCount,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
