package application

import (
	"context"
	"errors"
	"testing"
	"wishlist/domain/event"
	"wishlist/domain/service"
	"wishlist/domain/valueobject"
	"wishlist/mocks"

	"pgregory.net/rapid"
)

// 4.8.9: unauthenticated returns error; no repo call
func TestRemoveWishlistItemService_Unauthenticated(t *testing.T) {
	repo := mocks.NewMockWishlistRepository(t)
	auth := mocks.NewMockAuthSessionService(t)
	bus := &testEventBus{}

	auth.EXPECT().ResolveShopperID(context.Background(), "bad-token").
		Return(valueobject.ShopperId{}, service.ErrUnauthenticated)

	svc := NewRemoveWishlistItemService(repo, auth, bus)
	configSku, _ := valueobject.NewConfigSku("PD-001")
	err := svc.Execute(context.Background(), "bad-token", configSku)

	if !errors.Is(err, service.ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated, got: %v", err)
	}
	repo.AssertNotCalled(t, "RemoveItemByConfigSku")
}

// 4.8.10: success emits WishlistItemRemoved
func TestRemoveWishlistItemService_SuccessEmitsEvent(t *testing.T) {
	repo := mocks.NewMockWishlistRepository(t)
	auth := mocks.NewMockAuthSessionService(t)
	bus := &testEventBus{}

	sid := makeShopperId("shopper-001")
	auth.EXPECT().ResolveShopperID(context.Background(), "valid-token").Return(sid, nil)

	configSku, _ := valueobject.NewConfigSku("PD-001")
	repo.EXPECT().RemoveItemByConfigSku(context.Background(), sid, configSku).Return(nil)

	svc := NewRemoveWishlistItemService(repo, auth, bus)
	err := svc.Execute(context.Background(), "valid-token", configSku)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bus.published) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(bus.published))
	}
	ev, ok := bus.published[0].(event.WishlistItemRemoved)
	if !ok {
		t.Fatalf("expected WishlistItemRemoved event, got %T", bus.published[0])
	}
	if ev.ShopperID != "shopper-001" {
		t.Fatalf("event.ShopperID: got %q, want shopper-001", ev.ShopperID)
	}
	if ev.ConfigSku != "PD-001" {
		t.Fatalf("event.ConfigSku: got %q, want PD-001", ev.ConfigSku)
	}
}

// 4.8.11 PBT: event ConfigSku equals input configSku
func TestRemoveWishlistItemService_EventPayloadCorrectness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		configSkuRaw := rapid.StringMatching(`[A-Z]{2}-[0-9]{3}`).Draw(t, "configSku")

		repo := mocks.NewMockWishlistRepository(t)
		auth := mocks.NewMockAuthSessionService(t)
		bus := &testEventBus{}

		sid := makeShopperId("shopper-001")
		auth.On("ResolveShopperID", context.Background(), "tok").Return(sid, nil)
		configSku, _ := valueobject.NewConfigSku(configSkuRaw)
		repo.On("RemoveItemByConfigSku", context.Background(), sid, configSku).Return(nil)

		svc := NewRemoveWishlistItemService(repo, auth, bus)
		err := svc.Execute(context.Background(), "tok", configSku)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bus.published) == 0 {
			t.Fatal("expected event to be published")
		}
		ev, ok := bus.published[0].(event.WishlistItemRemoved)
		if !ok {
			t.Fatalf("expected WishlistItemRemoved, got %T", bus.published[0])
		}
		if ev.ConfigSku != configSkuRaw {
			t.Fatalf("event.ConfigSku: got %q, want %q", ev.ConfigSku, configSkuRaw)
		}
	})
}
