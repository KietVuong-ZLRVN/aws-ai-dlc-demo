package main

import (
	"context"
	"log/slog"
	"net/http"

	"wishlist/api"
	"wishlist/application"
	"wishlist/config"
	"wishlist/domain/assembler"
	"wishlist/domain/event"
	infraauth "wishlist/infrastructure/auth"
	"wishlist/infrastructure/eventbus"
	"wishlist/infrastructure/persistence"
)

func main() {
	// 1. Load config
	cfg := config.Load()

	// 2. Create InMemoryAuthService
	authService := &infraauth.InMemoryAuthService{}

	// 3. Create InMemoryWishlistRepository + WishlistAssembler
	wishlistAssembler := &assembler.WishlistAssembler{}
	wishlistRepo := persistence.NewInMemoryWishlistRepository(wishlistAssembler)

	// 4. Create InMemoryEventBus
	bus := eventbus.NewInMemoryEventBus()

	// 5. Register event handlers on the bus: slog logging handlers for each event type
	bus.Subscribe("WishlistItemAdded", func(ctx context.Context, e event.Event) {
		if ev, ok := e.(event.WishlistItemAdded); ok {
			slog.Info("event: WishlistItemAdded",
				"shopperID", ev.ShopperID,
				"simpleSku", ev.SimpleSku,
				"configSku", ev.ConfigSku,
				"itemId", ev.ItemId,
				"occurredAt", ev.OccurredAt,
			)
		}
	})

	bus.Subscribe("WishlistItemRemoved", func(ctx context.Context, e event.Event) {
		if ev, ok := e.(event.WishlistItemRemoved); ok {
			slog.Info("event: WishlistItemRemoved",
				"shopperID", ev.ShopperID,
				"configSku", ev.ConfigSku,
				"occurredAt", ev.OccurredAt,
			)
		}
	})

	bus.Subscribe("AuthenticationGateTriggered", func(ctx context.Context, e event.Event) {
		if ev, ok := e.(event.AuthenticationGateTriggered); ok {
			slog.Info("event: AuthenticationGateTriggered",
				"simpleSku", ev.SimpleSku,
				"returnPath", ev.ReturnPath,
				"occurredAt", ev.OccurredAt,
			)
		}
	})

	// Wrap the domain eventbus in an adapter satisfying application.EventBus
	appBus := eventbus.NewAppEventBusAdapter(bus)

	// 6. Create application services
	getWishlistSvc := application.NewGetWishlistService(wishlistRepo, authService)
	addWishlistItemSvc := application.NewAddWishlistItemService(wishlistRepo, authService, appBus)
	removeWishlistItemSvc := application.NewRemoveWishlistItemService(wishlistRepo, authService, appBus)

	// 7. Create API handlers
	getHandler := api.NewWishlistGetHandler(getWishlistSvc)
	addHandler := api.NewWishlistAddHandler(addWishlistItemSvc)
	removeHandler := api.NewWishlistRemoveHandler(removeWishlistItemSvc)

	// 8. Build router
	router := api.NewRouter(getHandler, addHandler, removeHandler)

	// 9. Start server
	slog.Info("wishlist service starting", "port", cfg.Port)
	if err := http.ListenAndServe(cfg.Port, router); err != nil {
		slog.Error("server error", "error", err)
	}
}
