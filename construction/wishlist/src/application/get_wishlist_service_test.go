package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"wishlist/domain/aggregate"
	"wishlist/domain/entity"
	"wishlist/domain/service"
	"wishlist/domain/valueobject"
	"wishlist/mocks"

	"pgregory.net/rapid"
)

func makeShopperId(v string) valueobject.ShopperId {
	sid, _ := valueobject.NewShopperId(v)
	return sid
}

func makePagination(offset, limit int) valueobject.Pagination {
	p, _ := valueobject.NewPagination(offset, limit)
	return p
}

func makeWishlistWithN(n int) *aggregate.Wishlist {
	wid, _ := valueobject.NewWishlistId("wl-001")
	sid := makeShopperId("shopper-001")
	items := make([]entity.WishlistItem, n)
	for i := range items {
		iid, _ := valueobject.NewWishlistItemId(fmt.Sprintf("item-%d", i))
		ssku, _ := valueobject.NewSimpleSku(fmt.Sprintf("PD-%03d-M-BLK", i+1))
		csku, _ := valueobject.NewConfigSku(fmt.Sprintf("PD-%03d", i+1))
		m, _ := valueobject.NewMoney(float64(i+1)*10, "SGD")
		items[i] = entity.WishlistItem{
			ItemId:    iid,
			SimpleSku: ssku,
			ConfigSku: csku,
			Name:      fmt.Sprintf("Product %d", i+1),
			Brand:     "Brand",
			Price:     m,
			ImageUrl:  "https://img.example.com/img.jpg",
			Color:     "black",
			Size:      "M",
			InStock:   true,
		}
	}
	return &aggregate.Wishlist{ID: wid, ShopperID: sid, Items: items, TotalCount: n}
}

// 4.8.1: unauthenticated returns error
func TestGetWishlistService_Unauthenticated(t *testing.T) {
	repo := mocks.NewMockWishlistRepository(t)
	auth := mocks.NewMockAuthSessionService(t)
	auth.EXPECT().ResolveShopperID(context.Background(), "bad-token").
		Return(valueobject.ShopperId{}, service.ErrUnauthenticated)

	svc := NewGetWishlistService(repo, auth)
	_, err := svc.Execute(context.Background(), "bad-token", makePagination(0, 20))

	if !errors.Is(err, service.ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated, got: %v", err)
	}
}

// 4.8.2: authenticated success
func TestGetWishlistService_AuthenticatedSuccess(t *testing.T) {
	repo := mocks.NewMockWishlistRepository(t)
	auth := mocks.NewMockAuthSessionService(t)
	sid := makeShopperId("shopper-001")
	p := makePagination(0, 20)
	wl := makeWishlistWithN(3)

	auth.EXPECT().ResolveShopperID(context.Background(), "valid-token").Return(sid, nil)
	repo.EXPECT().GetByShopperId(context.Background(), sid, p).Return(wl, nil)

	svc := NewGetWishlistService(repo, auth)
	result, err := svc.Execute(context.Background(), "valid-token", p)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Items) != 3 {
		t.Fatalf("Items count: got %d, want 3", len(result.Items))
	}
}

// 4.8.3 PBT: item shape completeness
func TestGetWishlistService_ItemShapeCompleteness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 6).Draw(t, "n")
		wl := makeWishlistWithN(n)

		repo := mocks.NewMockWishlistRepository(t)
		auth := mocks.NewMockAuthSessionService(t)
		sid := makeShopperId("shopper-001")
		p := makePagination(0, 20)

		auth.On("ResolveShopperID", context.Background(), "token").Return(sid, nil)
		repo.On("GetByShopperId", context.Background(), sid, p).Return(wl, nil)

		svc := NewGetWishlistService(repo, auth)
		result, err := svc.Execute(context.Background(), "token", p)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Items) != n {
			t.Fatalf("Items count: got %d, want %d", len(result.Items), n)
		}
		for i, item := range result.Items {
			if item.ItemId.String() == "" {
				t.Fatalf("item[%d].ItemId is empty", i)
			}
			if item.SimpleSku.String() == "" {
				t.Fatalf("item[%d].SimpleSku is empty", i)
			}
			if item.ConfigSku.String() == "" {
				t.Fatalf("item[%d].ConfigSku is empty", i)
			}
			if item.Name == "" {
				t.Fatalf("item[%d].Name is empty", i)
			}
			if item.Brand == "" {
				t.Fatalf("item[%d].Brand is empty", i)
			}
			if item.ImageUrl == "" {
				t.Fatalf("item[%d].ImageUrl is empty", i)
			}
		}
	})
}
