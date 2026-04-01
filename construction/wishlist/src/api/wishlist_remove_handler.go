package api

import (
	"net/http"
	"wishlist/application"
	"wishlist/domain/service"
	"wishlist/domain/valueobject"

	"github.com/go-chi/chi/v5"
)

type WishlistRemoveHandler struct {
	svc *application.RemoveWishlistItemService
}

func NewWishlistRemoveHandler(svc *application.RemoveWishlistItemService) *WishlistRemoveHandler {
	return &WishlistRemoveHandler{svc: svc}
}

func (h *WishlistRemoveHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Extract session cookie
	sessionToken := ""
	if cookie, err := r.Cookie("session"); err == nil {
		sessionToken = cookie.Value
	}

	// Extract configSku from URL params
	configSkuStr := chi.URLParam(r, "configSku")

	configSku, err := valueobject.NewConfigSku(configSkuStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := h.svc.Execute(r.Context(), sessionToken, configSku); err != nil {
		if err == service.ErrUnauthenticated {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "authentication required"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusOK)
}
