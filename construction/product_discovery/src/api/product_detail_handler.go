package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"product_discovery/application"
	"product_discovery/domain/query"
)

// ProductDetailHandler handles GET /api/v1/products/{configSku}.
type ProductDetailHandler struct {
	handler *application.ProductDetailQueryHandler
}

// NewProductDetailHandler constructs a ProductDetailHandler.
func NewProductDetailHandler(h *application.ProductDetailQueryHandler) *ProductDetailHandler {
	return &ProductDetailHandler{handler: h}
}

// Handle processes a product detail request.
func (h *ProductDetailHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	configSku := chi.URLParam(r, "configSku")

	q := query.ProductDetailQuery{ConfigSku: configSku}

	result, err := h.handler.Handle(ctx, q)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrProductNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, application.ErrProductDetailUnavailable):
			writeError(w, http.StatusBadGateway, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}
