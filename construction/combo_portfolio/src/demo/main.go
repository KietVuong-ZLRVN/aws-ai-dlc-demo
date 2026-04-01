// demo/main.go — Combo Portfolio local demo
//
// Prerequisites:
//   1. docker compose up -d   (from this directory)
//   2. Wait for MySQL to be healthy, then apply schema:
//      mysql -h 127.0.0.1 -P 3306 -u root -proot < ../infrastructure/persistence/schema.sql
//   3. go run .               (from this directory)
//
// The demo runs without hitting any real external API.

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/application"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/acl"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/persistence"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/services"
)

const dsn = "root:root@tcp(127.0.0.1:3306)/combo_portfolio?parseTime=true"

func main() {
	ctx := context.Background()

	// Connect to local MySQL (started via docker-compose).
	db := mustConnectDB()
	defer db.Close()

	// Wire dependencies (using stub catalog ACL).
	repo := persistence.NewMySQLComboRepository(db)
	enrichmentSvc := acl.NewComboEnrichmentService(&stubProductCatalogACL{})
	tokenSvc := services.NewShareTokenService(repo)

	saveHandler := application.NewSaveComboHandler(repo)
	renameHandler := application.NewRenameComboHandler(repo)
	deleteHandler := application.NewDeleteComboHandler(repo)
	shareHandler := application.NewShareComboHandler(repo, tokenSvc)
	makePrivHandler := application.NewMakePrivateHandler(repo)
	getHandler := application.NewGetComboHandler(repo, enrichmentSvc)
	listHandler := application.NewListCombosHandler(repo, enrichmentSvc)
	sharedHandler := application.NewGetSharedComboHandler(repo, enrichmentSvc)

	shopperID := domain.ShopperId("shopper-demo-001")

	// ── Scenario 1: Save a new combo ──────────────────────────────────────────
	fmt.Println("\n=== Scenario 1: Save a combo ===")
	comboID, err := saveHandler.Handle(ctx, application.SaveComboCommand{
		ShopperID:  shopperID,
		Name:       "My Summer Look",
		Items:      stubItems(),
		Visibility: domain.VisibilityPrivate,
	})
	mustOK(err, "save combo")
	fmt.Printf("Saved combo: %s\n", comboID)

	// ── Scenario 2: List combos for shopper ───────────────────────────────────
	fmt.Println("\n=== Scenario 2: List combos ===")
	combos, err := listHandler.Handle(ctx, application.ListCombosQuery{ShopperID: shopperID})
	mustOK(err, "list combos")
	fmt.Printf("Found %d combo(s)\n", len(combos))
	for _, c := range combos {
		fmt.Printf("  [%s] %s (visibility=%s, items=%d)\n", c.ID, c.Name, c.Visibility, len(c.Items))
		for _, item := range c.Items {
			fmt.Printf("    - %s | price=%.2f | inStock=%v | catalogUnavailable=%v\n",
				item.Name, item.Price, item.InStock, item.CatalogUnavailable)
		}
	}

	// ── Scenario 3: Rename the combo ─────────────────────────────────────────
	fmt.Println("\n=== Scenario 3: Rename combo ===")
	mustOK(renameHandler.Handle(ctx, application.RenameComboCommand{
		ShopperID: shopperID, ComboID: comboID, NewName: "My Updated Summer Look",
	}), "rename combo")
	fmt.Println("Renamed successfully")

	// ── Scenario 4: Share the combo (generates share token) ───────────────────
	fmt.Println("\n=== Scenario 4: Share combo ===")
	token, err := shareHandler.Handle(ctx, application.ShareComboCommand{
		ShopperID: shopperID, ComboID: comboID,
	})
	mustOK(err, "share combo")
	fmt.Printf("Share token: %s\n", token)
	fmt.Printf("Public URL: http://localhost:8080/api/v1/combos/shared/%s\n", token)

	// ── Scenario 5: Get shared combo (unauthenticated) ────────────────────────
	fmt.Println("\n=== Scenario 5: Get shared combo via share token ===")
	sharedCombo, err := sharedHandler.Handle(ctx, application.GetSharedComboQuery{ShareToken: token})
	mustOK(err, "get shared combo")
	fmt.Printf("Shared combo name: %s (visibility=%s)\n", sharedCombo.Name, sharedCombo.Visibility)

	// ── Scenario 6: Make combo private (revokes share token) ─────────────────
	fmt.Println("\n=== Scenario 6: Make combo private ===")
	mustOK(makePrivHandler.Handle(ctx, application.MakePrivateCommand{
		ShopperID: shopperID, ComboID: comboID,
	}), "make private")
	fmt.Println("Combo is now private; share token revoked")

	// Verify share token no longer works
	_, err = sharedHandler.Handle(ctx, application.GetSharedComboQuery{ShareToken: token})
	if err == domain.ErrComboNotFound {
		fmt.Println("Confirmed: share token no longer resolves (expected)")
	} else {
		fmt.Printf("Unexpected result: %v\n", err)
	}

	// ── Scenario 7: Get enriched combo ───────────────────────────────────────
	fmt.Println("\n=== Scenario 7: Get enriched combo (owner view) ===")
	enriched, err := getHandler.Handle(ctx, application.GetComboQuery{
		ShopperID: shopperID, ComboID: comboID,
	})
	mustOK(err, "get enriched combo")
	fmt.Printf("Combo: %s | visibility=%s | shareToken=%v\n", enriched.Name, enriched.Visibility, enriched.ShareToken)
	for _, item := range enriched.Items {
		fmt.Printf("  - %s | price=%.2f | inStock=%v\n", item.Name, item.Price, item.InStock)
	}

	// ── Scenario 8: Delete the combo ─────────────────────────────────────────
	fmt.Println("\n=== Scenario 8: Delete combo ===")
	mustOK(deleteHandler.Handle(ctx, application.DeleteComboCommand{
		ShopperID: shopperID, ComboID: comboID,
	}), "delete combo")
	fmt.Println("Combo deleted")

	// Verify deletion
	_, err = getHandler.Handle(ctx, application.GetComboQuery{ShopperID: shopperID, ComboID: comboID})
	if err == domain.ErrComboNotFound {
		fmt.Println("Confirmed: combo no longer exists (expected)")
	}

	fmt.Println("\n=== Demo complete ===")
}

func mustConnectDB() *sql.DB {
	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			if pingErr := db.Ping(); pingErr == nil {
				log.Println("Connected to MySQL")
				return db
			}
		}
		log.Printf("Waiting for MySQL... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("Could not connect to MySQL: %v", err)
	return nil
}

func mustOK(err error, op string) {
	if err != nil {
		log.Fatalf("FAILED [%s]: %v", op, err)
	}
}
