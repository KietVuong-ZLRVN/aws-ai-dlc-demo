package application

import (
	"context"
	"errors"
	"log/slog"

	"product_discovery/domain/assembler"
	"product_discovery/domain/port"
	"product_discovery/domain/query"
	"product_discovery/domain/readmodel"
)

// Sentinel errors for the application layer.
var (
	ErrProductNotFound          = errors.New("product not found")
	ErrProductListUnavailable   = errors.New("product list unavailable")
	ErrProductDetailUnavailable = errors.New("product detail unavailable")
)

// ProductListQueryHandler handles product list queries.
type ProductListQueryHandler struct {
	client    port.PlatformProductClient
	assembler *assembler.ProductListAssembler
}

// NewProductListQueryHandler constructs a ProductListQueryHandler.
func NewProductListQueryHandler(
	client port.PlatformProductClient,
	asm *assembler.ProductListAssembler,
) *ProductListQueryHandler {
	return &ProductListQueryHandler{
		client:    client,
		assembler: asm,
	}
}

// Handle executes the product list query and returns a ProductListReadModel.
func (h *ProductListQueryHandler) Handle(ctx context.Context, q query.ProductListQuery) (*readmodel.ProductListReadModel, error) {
	params := port.ProductListParams{
		Query:      q.Query,
		CategoryID: q.CategoryID,
		Colors:     q.Colors,
		Offset:     q.Pagination.Offset,
		Limit:      q.Pagination.Limit,
	}
	if q.PriceRange != nil {
		params.MinPrice = q.PriceRange.Min
		params.MaxPrice = q.PriceRange.Max
	}

	listPayload, err := h.client.FetchProductList(ctx, params)
	if err != nil {
		slog.ErrorContext(ctx, "FetchProductList failed", "error", err)
		return nil, ErrProductListUnavailable
	}

	filterPayload, err := h.client.FetchProductFilters(ctx, params)
	if err != nil {
		slog.ErrorContext(ctx, "FetchProductFilters failed", "error", err)
		return nil, ErrProductListUnavailable
	}

	result := h.assembler.Assemble(listPayload, filterPayload)
	return &result, nil
}
