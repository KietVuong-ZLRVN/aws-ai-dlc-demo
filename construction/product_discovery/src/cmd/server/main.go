package main

import (
	"log/slog"
	"net/http"

	"product_discovery/api"
	"product_discovery/application"
	"product_discovery/config"
	"product_discovery/domain/assembler"
	infra "product_discovery/infrastructure/platform"
)

func main() {
	cfg := config.Load()

	// Infrastructure
	platformClient := infra.NewInMemoryProductClient()

	// Assemblers
	listAssembler := assembler.NewProductListAssembler()
	detailAssembler := assembler.NewProductDetailAssembler()

	// Application handlers
	listQueryHandler := application.NewProductListQueryHandler(platformClient, listAssembler)
	detailQueryHandler := application.NewProductDetailQueryHandler(platformClient, detailAssembler)

	// API handlers
	productListHandler := api.NewProductListHandler(listQueryHandler)
	productDetailHandler := api.NewProductDetailHandler(detailQueryHandler)

	// Router
	router := api.NewRouter(productListHandler, productDetailHandler)

	addr := ":" + cfg.Port
	slog.Info("starting server", "addr", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		slog.Error("server stopped", "error", err)
	}
}
