package api

import (
	"encoding/json"
	"net/http"
	"wishlist/application"
	"wishlist/domain/service"
	"wishlist/domain/valueobject"
)

type WishlistAddHandler struct {
	svc *application.AddWishlistItemService
}

func NewWishlistAddHandler(svc *application.AddWishlistItemService) *WishlistAddHandler {
	return &WishlistAddHandler{svc: svc}
}

type addItemRequest struct {
	SimpleSku string `json:"simpleSku"`
}

type addItemResponse struct {
	ItemId    string `json:"itemId"`
	SimpleSku string `json:"simpleSku"`
	ConfigSku string `json:"configSku"`
}

type addItemAuthErrorResponse struct {
	Error        string `json:"error"`
	RequiresAuth bool   `json:"requiresAuth"`
	ReturnPath   string `json:"returnPath"`
}

func (h *WishlistAddHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Extract session cookie
	sessionToken := ""
	if cookie, err := r.Cookie("session"); err == nil {
		sessionToken = cookie.Value
	}

	// Parse JSON body
	var req addItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	simpleSku, err := valueobject.NewSimpleSku(req.SimpleSku)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Read Referer header as returnPath
	returnPath := r.Header.Get("Referer")

	result, err := h.svc.Execute(r.Context(), sessionToken, simpleSku, returnPath)
	if err != nil {
		if err == service.ErrUnauthenticated {
			writeJSON(w, http.StatusForbidden, addItemAuthErrorResponse{
				Error:        "authentication required",
				RequiresAuth: true,
				ReturnPath:   returnPath,
			})
			return
		}
		if err == application.ErrWishlistItemAlreadyPresent {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "item already in wishlist"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, addItemResponse{
		ItemId:    result.ItemId,
		SimpleSku: result.SimpleSku,
		ConfigSku: result.ConfigSku,
	})
}
