package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"product_discovery/application"
	"product_discovery/domain/query"
	"product_discovery/domain/valueobject"
)

// ProductListHandler handles GET /api/v1/products.
type ProductListHandler struct {
	handler *application.ProductListQueryHandler
}

// NewProductListHandler constructs a ProductListHandler.
func NewProductListHandler(h *application.ProductListQueryHandler) *ProductListHandler {
	return &ProductListHandler{handler: h}
}

// Handle processes a product list request.
func (h *ProductListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := r.URL.Query()

	// Parse offset and limit.
	offset := 0
	limit := 0
	if v := params.Get("offset"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "invalid offset parameter")
			return
		}
		offset = parsed
	}
	if v := params.Get("limit"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "invalid limit parameter")
			return
		}
		limit = parsed
	}

	pagination, err := valueobject.NewPagination(offset, limit)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Parse optional price range.
	var priceRange *valueobject.PriceRange
	if priceStr := params.Get("price"); priceStr != "" {
		pr, err := valueobject.NewPriceRange(priceStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		priceRange = &pr
	}

	q := query.ProductListQuery{
		Query:      params.Get("q"),
		CategoryID: params.Get("category"),
		Colors:     params["colors"],
		PriceRange: priceRange,
		Pagination: pagination,
	}

	result, err := h.handler.Handle(ctx, q)
	if err != nil {
		if errors.Is(err, application.ErrProductListUnavailable) {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// writeJSON serialises v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a standard {"error": "..."} JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
