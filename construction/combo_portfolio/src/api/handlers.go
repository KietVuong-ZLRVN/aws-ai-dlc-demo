package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/application"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
)

// Handlers holds all HTTP handler methods.
type Handlers struct {
	saveCombos     *application.SaveComboHandler
	renameCombo    *application.RenameComboHandler
	deleteCombo    *application.DeleteComboHandler
	shareCombo     *application.ShareComboHandler
	makePrivate    *application.MakePrivateHandler
	getCombo       *application.GetComboHandler
	listCombos     *application.ListCombosHandler
	getSharedCombo *application.GetSharedComboHandler
	publicBaseURL  string
}

func NewHandlers(
	save *application.SaveComboHandler,
	rename *application.RenameComboHandler,
	delete *application.DeleteComboHandler,
	share *application.ShareComboHandler,
	makePriv *application.MakePrivateHandler,
	get *application.GetComboHandler,
	list *application.ListCombosHandler,
	shared *application.GetSharedComboHandler,
	publicBaseURL string,
) *Handlers {
	return &Handlers{save, rename, delete, share, makePriv, get, list, shared, publicBaseURL}
}

func (h *Handlers) SaveCombo(w http.ResponseWriter, r *http.Request) {
	shopperID, _ := ShopperIDFromContext(r.Context())
	var req SaveComboRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	vis := domain.VisibilityPrivate
	if req.Visibility == "public" {
		vis = domain.VisibilityPublic
	}
	items := make([]domain.ComboItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = domain.ComboItem{ConfigSku: it.ConfigSku, SimpleSku: it.SimpleSku, Name: it.Name, ImageUrl: it.ImageUrl, Price: it.Price}
	}
	id, err := h.saveCombos.Handle(r.Context(), application.SaveComboCommand{
		ShopperID: shopperID, Name: req.Name, Items: items, Visibility: vis,
	})
	if err != nil {
		writeError(w, domainErrStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, SaveComboResponse{ID: id.String()})
}

func (h *Handlers) ListCombos(w http.ResponseWriter, r *http.Request) {
	shopperID, _ := ShopperIDFromContext(r.Context())
	combos, err := h.listCombos.Handle(r.Context(), application.ListCombosQuery{ShopperID: shopperID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]ComboResponse, len(combos))
	for i, c := range combos {
		resp[i] = ToComboResponse(c)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) GetCombo(w http.ResponseWriter, r *http.Request) {
	shopperID, _ := ShopperIDFromContext(r.Context())
	id := domain.ComboId(chi.URLParam(r, "id"))
	combo, err := h.getCombo.Handle(r.Context(), application.GetComboQuery{ShopperID: shopperID, ComboID: id})
	if err != nil {
		writeError(w, domainErrStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ToComboResponse(combo))
}

func (h *Handlers) UpdateCombo(w http.ResponseWriter, r *http.Request) {
	shopperID, _ := ShopperIDFromContext(r.Context())
	id := domain.ComboId(chi.URLParam(r, "id"))
	var req UpdateComboRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name != nil {
		if err := h.renameCombo.Handle(r.Context(), application.RenameComboCommand{ShopperID: shopperID, ComboID: id, NewName: *req.Name}); err != nil {
			writeError(w, domainErrStatus(err), err.Error())
			return
		}
	}
	if req.Visibility != nil && *req.Visibility == "private" {
		if err := h.makePrivate.Handle(r.Context(), application.MakePrivateCommand{ShopperID: shopperID, ComboID: id}); err != nil {
			writeError(w, domainErrStatus(err), err.Error())
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) DeleteCombo(w http.ResponseWriter, r *http.Request) {
	shopperID, _ := ShopperIDFromContext(r.Context())
	id := domain.ComboId(chi.URLParam(r, "id"))
	if err := h.deleteCombo.Handle(r.Context(), application.DeleteComboCommand{ShopperID: shopperID, ComboID: id}); err != nil {
		writeError(w, domainErrStatus(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) ShareCombo(w http.ResponseWriter, r *http.Request) {
	shopperID, _ := ShopperIDFromContext(r.Context())
	id := domain.ComboId(chi.URLParam(r, "id"))
	token, err := h.shareCombo.Handle(r.Context(), application.ShareComboCommand{ShopperID: shopperID, ComboID: id})
	if err != nil {
		writeError(w, domainErrStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ShareComboResponse{
		ShareToken: token.String(),
		ShareURL:   fmt.Sprintf("%s/shared/%s", h.publicBaseURL, token.String()),
	})
}

func (h *Handlers) GetSharedCombo(w http.ResponseWriter, r *http.Request) {
	token := domain.ShareToken(chi.URLParam(r, "token"))
	combo, err := h.getSharedCombo.Handle(r.Context(), application.GetSharedComboQuery{ShareToken: token})
	if err != nil {
		writeError(w, domainErrStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ToComboResponse(combo))
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

func domainErrStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrComboNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrComboAccessDenied):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrInvalidItemCount),
		errors.Is(err, domain.ErrDuplicateItem),
		errors.Is(err, domain.ErrInvalidComboName):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
