// demo/main.go — Cart Handoff local demo
//
// Prerequisites:
//   1. docker compose up -d   (from this directory)
//   2. Wait for MySQL to be healthy, then apply schema:
//      mysql -h 127.0.0.1 -P 3307 -u root -proot < ../infrastructure/persistence/schema.sql
//   3. go run .               (from this directory)
//
// The demo runs four scenarios without hitting any real external API.

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/application"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/infrastructure/persistence"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/infrastructure/services"
)

const dsn = "root:root@tcp(127.0.0.1:3307)/cart_handoff?parseTime=true"

func main() {
	ctx := context.Background()
	db := mustConnectDB()
	defer db.Close()

	repo := persistence.NewMySQLCartHandoffRepository(db)
	shopperID := domain.ShopperId("shopper-demo-001")
	sessionCookie := "session=stub-session-token"

	// ── Scenario 1: Add saved combo by comboId (all items in stock) ───────────
	fmt.Println("\n=== Scenario 1: Add saved combo by comboId (ok) ===")
	handler := buildHandler(repo, "ok")
	result, err := handler.Handle(ctx, application.AddComboToCartCommand{
		ShopperID: shopperID, SessionCookie: sessionCookie, ComboId: "combo-uuid-abc123",
	})
	mustOK(err, "add by comboId")
	printResult(result)

	// ── Scenario 2: Add inline items (unsaved combo from AI engine) ───────────
	fmt.Println("\n=== Scenario 2: Add inline items (ok) ===")
	handler = buildHandler(repo, "ok")
	result, err = handler.Handle(ctx, application.AddComboToCartCommand{
		ShopperID:     shopperID,
		SessionCookie: sessionCookie,
		InlineItems: []domain.CartItem{
			{SimpleSku: "INLINE-SKU-001", Quantity: 1, Size: "S"},
			{SimpleSku: "INLINE-SKU-002", Quantity: 1, Size: ""},
		},
	})
	mustOK(err, "add inline items")
	printResult(result)

	// ── Scenario 3: Partial out-of-stock ─────────────────────────────────────
	fmt.Println("\n=== Scenario 3: Partial — some items out of stock ===")
	handler = buildHandler(repo, "partial")
	result, err = handler.Handle(ctx, application.AddComboToCartCommand{
		ShopperID: shopperID, SessionCookie: sessionCookie, ComboId: "combo-uuid-partial",
	})
	mustOK(err, "partial add")
	printResult(result)

	// ── Scenario 4: Platform cart API unavailable ─────────────────────────────
	fmt.Println("\n=== Scenario 4: Platform cart API failure ===")
	handler = buildHandler(repo, "failed")
	_, err = handler.Handle(ctx, application.AddComboToCartCommand{
		ShopperID: shopperID, SessionCookie: sessionCookie, ComboId: "combo-uuid-fail",
	})
	if errors.Is(err, domain.ErrPlatformCartUnavailable) {
		fmt.Println("Received expected error: platform cart unavailable")
		fmt.Println("Audit record persisted with status=failed")
	} else {
		fmt.Printf("Unexpected error: %v\n", err)
	}

	// ── Scenario 5: Invalid source (both comboId and items) ───────────────────
	fmt.Println("\n=== Scenario 5: Invalid request — both comboId and items provided ===")
	handler = buildHandler(repo, "ok")
	_, err = handler.Handle(ctx, application.AddComboToCartCommand{
		ShopperID:     shopperID,
		SessionCookie: sessionCookie,
		ComboId:       "combo-uuid-abc123",
		InlineItems:   []domain.CartItem{{SimpleSku: "SKU-001", Quantity: 1}},
	})
	if errors.Is(err, domain.ErrInvalidHandoffSource) {
		fmt.Println("Received expected error: invalid handoff source")
	} else {
		fmt.Printf("Unexpected error: %v\n", err)
	}

	// ── Verify audit records in DB ────────────────────────────────────────────
	fmt.Println("\n=== Audit records for shopper ===")
	records, err := repo.FindByShopperId(ctx, shopperID)
	mustOK(err, "find records")
	fmt.Printf("Total audit records: %d\n", len(records))
	for _, rec := range records {
		fmt.Printf("  [%s] status=%s | added=%d | skipped=%d | at=%s\n",
			rec.ID(), rec.Status(), len(rec.AddedItems()), len(rec.SkippedItems()),
			rec.RecordedAt().Format(time.RFC3339))
	}

	fmt.Println("\n=== Demo complete ===")
}

func buildHandler(repo domain.CartHandoffRecordRepository, scenario string) *application.AddComboToCartHandler {
	comboACL := &stubComboPortfolioACL{}
	cartACL := &stubPlatformCartACL{scenario: scenario}
	resolutionSvc := services.NewComboResolutionService(comboACL)
	submissionSvc := services.NewCartSubmissionService(cartACL)
	return application.NewAddComboToCartHandler(repo, resolutionSvc, submissionSvc)
}

func printResult(r application.HandoffResult) {
	fmt.Printf("Status: %s\n", r.Status)
	fmt.Printf("Added (%d):", len(r.AddedItems))
	for _, item := range r.AddedItems {
		fmt.Printf(" %s", item.SimpleSku)
	}
	fmt.Println()
	if len(r.SkippedItems) > 0 {
		fmt.Printf("Skipped (%d):", len(r.SkippedItems))
		for _, item := range r.SkippedItems {
			fmt.Printf(" %s(%s)", item.SimpleSku, item.Reason)
		}
		fmt.Println()
	}
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
