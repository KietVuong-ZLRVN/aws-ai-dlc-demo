package application

import (
	"context"
	"errors"
	"testing"

	"product_discovery/domain/assembler"
	"product_discovery/domain/port"
	"product_discovery/domain/query"
	"product_discovery/mocks"

	"pgregory.net/rapid"
)

func makeDetailQuery(configSku string) query.ProductDetailQuery {
	return query.ProductDetailQuery{ConfigSku: configSku}
}

func makeDetailPayload(configSku string) *port.RawProductDetailPayload {
	return &port.RawProductDetailPayload{
		Product: port.PlatformProduct{
			ConfigSku: configSku,
			Name:      "Test Product",
			Brand:     "Test Brand",
			UrlKey:    "test-product",
			Currency:  "SGD",
			Simples: []port.PlatformSimple{
				{SimpleSku: configSku + "-M-BLK", Size: "M", Color: "black", Quantity: 5},
			},
		},
	}
}

// --- 4.3.5: Success path ---

func TestProductDetailQueryHandler_Success(t *testing.T) {
	client := mocks.NewMockPlatformProductClient(t)
	payload := makeDetailPayload("PD-001")
	client.EXPECT().FetchProductDetail(context.Background(), "PD-001").Return(payload, nil)

	handler := NewProductDetailQueryHandler(client, assembler.NewProductDetailAssembler())
	result, err := handler.Handle(context.Background(), makeDetailQuery("PD-001"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ConfigSku != "PD-001" {
		t.Fatalf("ConfigSku: got %q, want %q", result.ConfigSku, "PD-001")
	}
}

// --- 4.3.6: Not found (nil, nil) ---

func TestProductDetailQueryHandler_NotFound(t *testing.T) {
	client := mocks.NewMockPlatformProductClient(t)
	client.EXPECT().FetchProductDetail(context.Background(), "MISSING").Return(nil, nil)

	handler := NewProductDetailQueryHandler(client, assembler.NewProductDetailAssembler())
	_, err := handler.Handle(context.Background(), makeDetailQuery("MISSING"))

	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got: %v", err)
	}
}

// --- 4.3.7: Platform error ---

func TestProductDetailQueryHandler_PlatformError(t *testing.T) {
	client := mocks.NewMockPlatformProductClient(t)
	client.EXPECT().FetchProductDetail(context.Background(), "PD-001").
		Return(nil, errors.New("upstream down"))

	handler := NewProductDetailQueryHandler(client, assembler.NewProductDetailAssembler())
	_, err := handler.Handle(context.Background(), makeDetailQuery("PD-001"))

	if !errors.Is(err, ErrProductDetailUnavailable) {
		t.Fatalf("expected ErrProductDetailUnavailable, got: %v", err)
	}
}

// --- 4.3.8 PBT: Identity mapping (configSku passthrough) ---

func TestProductDetailQueryHandler_IdentityMapping(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		configSku := rapid.StringMatching(`[A-Z]{2}-[0-9]{3}`).Draw(t, "configSku")
		payload := makeDetailPayload(configSku)

		client := mocks.NewMockPlatformProductClient(t)
		client.On("FetchProductDetail", context.Background(), configSku).Return(payload, nil)

		handler := NewProductDetailQueryHandler(client, assembler.NewProductDetailAssembler())
		result, err := handler.Handle(context.Background(), makeDetailQuery(configSku))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ConfigSku != configSku {
			t.Fatalf("ConfigSku: got %q, want %q", result.ConfigSku, configSku)
		}
	})
}
