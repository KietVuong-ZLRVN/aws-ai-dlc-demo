package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"wishlist/domain/event"
	"wishlist/domain/service"
	"wishlist/domain/valueobject"
	"wishlist/mocks"

	"pgregory.net/rapid"
)

// testEventBus is a hand-rolled test double for EventBus that captures published events.
type testEventBus struct {
	published []Event
}

func (b *testEventBus) Publish(_ context.Context, e Event) {
	b.published = append(b.published, e)
}

func mustSimpleSku(s string) valueobject.SimpleSku {
	sku, _ := valueobject.NewSimpleSku(s)
	return sku
}

func mustConfigSku(s string) valueobject.ConfigSku {
	sku, _ := valueobject.NewConfigSku(s)
	return sku
}

func mustItemId(s string) valueobject.WishlistItemId {
	id, _ := valueobject.NewWishlistItemId(s)
	return id
}

// 4.8.4: unauthenticated triggers AuthenticationGateTriggered event
func TestAddWishlistItemService_UnauthenticatedTriggersEvent(t *testing.T) {
	repo := mocks.NewMockWishlistRepository(t)
	auth := mocks.NewMockAuthSessionService(t)
	bus := &testEventBus{}

	auth.EXPECT().ResolveShopperID(context.Background(), "bad-token").
		Return(valueobject.ShopperId{}, service.ErrUnauthenticated)

	svc := NewAddWishlistItemService(repo, auth, bus)
	simpleSku := mustSimpleSku("PD-001-M-BLK")
	_, err := svc.Execute(context.Background(), "bad-token", simpleSku, "/products/pd-001")

	if !errors.Is(err, service.ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated, got: %v", err)
	}
	if len(bus.published) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(bus.published))
	}
	if bus.published[0].EventName() != "AuthenticationGateTriggered" {
		t.Fatalf("event name: got %q, want AuthenticationGateTriggered", bus.published[0].EventName())
	}
}

// 4.8.5: duplicate configSku returns conflict error; no repo AddItem call
func TestAddWishlistItemService_DuplicateReturnsConflict(t *testing.T) {
	repo := mocks.NewMockWishlistRepository(t)
	auth := mocks.NewMockAuthSessionService(t)
	bus := &testEventBus{}

	sid := makeShopperId("shopper-001")
	auth.EXPECT().ResolveShopperID(context.Background(), "valid-token").Return(sid, nil)

	// Wishlist already contains PD-001
	wl := makeWishlistWithN(1) // item[0] has configSku PD-001
	pagination, _ := valueobject.NewPagination(0, 100)
	repo.EXPECT().GetByShopperId(context.Background(), sid, pagination).Return(wl, nil)

	svc := NewAddWishlistItemService(repo, auth, bus)
	simpleSku := mustSimpleSku("PD-001-L-WHT") // derives to PD-001
	_, err := svc.Execute(context.Background(), "valid-token", simpleSku, "")

	if !errors.Is(err, ErrWishlistItemAlreadyPresent) {
		t.Fatalf("expected ErrWishlistItemAlreadyPresent, got: %v", err)
	}
	repo.AssertNotCalled(t, "AddItem")
}

// 4.8.6: success emits WishlistItemAdded with correct fields
func TestAddWishlistItemService_SuccessEmitsEvent(t *testing.T) {
	repo := mocks.NewMockWishlistRepository(t)
	auth := mocks.NewMockAuthSessionService(t)
	bus := &testEventBus{}

	sid := makeShopperId("shopper-001")
	auth.EXPECT().ResolveShopperID(context.Background(), "valid-token").Return(sid, nil)

	wl := makeWishlistWithN(0) // empty
	pagination, _ := valueobject.NewPagination(0, 100)
	repo.EXPECT().GetByShopperId(context.Background(), sid, pagination).Return(wl, nil)

	simpleSku := mustSimpleSku("PD-002-M-BLK")
	configSku := mustConfigSku("PD-002")
	newItemId := mustItemId("item-new")

	repo.EXPECT().AddItem(context.Background(), sid, simpleSku, configSku).Return(newItemId, nil)

	svc := NewAddWishlistItemService(repo, auth, bus)
	result, err := svc.Execute(context.Background(), "valid-token", simpleSku, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ItemId != "item-new" {
		t.Fatalf("ItemId: got %q, want item-new", result.ItemId)
	}
	if len(bus.published) != 1 {
		t.Fatalf("expected 1 event, got %d", len(bus.published))
	}
	ev, ok := bus.published[0].(event.WishlistItemAdded)
	if !ok {
		t.Fatalf("expected WishlistItemAdded event, got %T", bus.published[0])
	}
	if ev.ShopperID != "shopper-001" {
		t.Fatalf("event.ShopperID: got %q, want shopper-001", ev.ShopperID)
	}
	if ev.ConfigSku != "PD-002" {
		t.Fatalf("event.ConfigSku: got %q, want PD-002", ev.ConfigSku)
	}
}

// 4.8.7 PBT: no repo.AddItem call when configSku already in wishlist
func TestAddWishlistItemService_NoPlatformCallOnDuplicate(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		i := rapid.IntRange(1, 5).Draw(t, "existingCount")
		wl := makeWishlistWithN(i)
		conflictIdx := rapid.IntRange(1, i).Draw(t, "conflictIdx")
		simpleSkuRaw := fmt.Sprintf("PD-%03d-L-RED", conflictIdx)

		repo := mocks.NewMockWishlistRepository(t)
		auth := mocks.NewMockAuthSessionService(t)
		bus := &testEventBus{}

		sid := makeShopperId("shopper-001")
		auth.On("ResolveShopperID", context.Background(), "tok").Return(sid, nil)
		pagination, _ := valueobject.NewPagination(0, 100)
		repo.On("GetByShopperId", context.Background(), sid, pagination).Return(wl, nil)

		svc := NewAddWishlistItemService(repo, auth, bus)
		_, err := svc.Execute(context.Background(), "tok", mustSimpleSku(simpleSkuRaw), "")

		if !errors.Is(err, ErrWishlistItemAlreadyPresent) {
			t.Fatalf("expected ErrWishlistItemAlreadyPresent, got: %v", err)
		}
		repo.AssertNotCalled(t, "AddItem")
	})
}

// 4.8.8 PBT: response itemId matches repo return
func TestAddWishlistItemService_ResponseItemIdMatchesRepo(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		returnedId := rapid.StringMatching(`item-[0-9]{1,6}`).Draw(t, "itemId")

		repo := mocks.NewMockWishlistRepository(t)
		auth := mocks.NewMockAuthSessionService(t)
		bus := &testEventBus{}

		sid := makeShopperId("shopper-001")
		auth.On("ResolveShopperID", context.Background(), "tok").Return(sid, nil)
		wl := makeWishlistWithN(0)
		pagination, _ := valueobject.NewPagination(0, 100)
		repo.On("GetByShopperId", context.Background(), sid, pagination).Return(wl, nil)

		simpleSku := mustSimpleSku("PD-010-M-BLK")
		configSku := mustConfigSku("PD-010")
		repoItemId := mustItemId(returnedId)
		repo.On("AddItem", context.Background(), sid, simpleSku, configSku).Return(repoItemId, nil)

		svc := NewAddWishlistItemService(repo, auth, bus)
		result, err := svc.Execute(context.Background(), "tok", simpleSku, "")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ItemId != returnedId {
			t.Fatalf("result.ItemId: got %q, want %q", result.ItemId, returnedId)
		}
	})
}
