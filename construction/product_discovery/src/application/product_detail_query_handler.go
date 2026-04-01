package application

import (
	"context"
	"log/slog"

	"product_discovery/domain/assembler"
	"product_discovery/domain/port"
	"product_discovery/domain/query"
	"product_discovery/domain/readmodel"
)

// ProductDetailQueryHandler handles product detail queries.
type ProductDetailQueryHandler struct {
	client    port.PlatformProductClient
	assembler *assembler.ProductDetailAssembler
}

// NewProductDetailQueryHandler constructs a ProductDetailQueryHandler.
func NewProductDetailQueryHandler(
	client port.PlatformProductClient,
	asm *assembler.ProductDetailAssembler,
) *ProductDetailQueryHandler {
	return &ProductDetailQueryHandler{
		client:    client,
		assembler: asm,
	}
}

// Handle executes the product detail query and returns a ProductDetailReadModel.
// Returns ErrProductNotFound when no product matches the configSku.
// Returns ErrProductDetailUnavailable on upstream errors.
func (h *ProductDetailQueryHandler) Handle(ctx context.Context, q query.ProductDetailQuery) (*readmodel.ProductDetailReadModel, error) {
	payload, err := h.client.FetchProductDetail(ctx, q.ConfigSku)
	if err != nil {
		slog.ErrorContext(ctx, "FetchProductDetail failed", "configSku", q.ConfigSku, "error", err)
		return nil, ErrProductDetailUnavailable
	}
	if payload == nil {
		return nil, ErrProductNotFound
	}

	result := h.assembler.Assemble(payload.Product)
	return &result, nil
}
