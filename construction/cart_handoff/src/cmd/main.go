package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/api"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/application"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/infrastructure/acl"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/infrastructure/persistence"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/infrastructure/services"
)

func main() {
	dsn := mustEnv("DB_DSN")
	comboPortfolioBaseURL := mustEnv("COMBO_PORTFOLIO_BASE_URL")
	platformCartBaseURL := mustEnv("PLATFORM_CART_BASE_URL")
	contentLanguage := getEnv("CONTENT_LANGUAGE", "en-SG")
	comboTimeoutMs := getEnvInt("COMBO_PORTFOLIO_TIMEOUT_MS", 3000)
	cartTimeoutMs := getEnvInt("PLATFORM_CART_TIMEOUT_MS", 5000)
	port := getEnv("PORT", "8081")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	db.SetMaxOpenConns(getEnvInt("DB_MAX_OPEN_CONNS", 25))
	db.SetMaxIdleConns(getEnvInt("DB_MAX_IDLE_CONNS", 10))

	repo := persistence.NewMySQLCartHandoffRepository(db)
	comboACL := acl.NewHTTPComboPortfolioACL(comboPortfolioBaseURL, comboTimeoutMs)
	cartACL := acl.NewHTTPPlatformCartACL(platformCartBaseURL, contentLanguage, cartTimeoutMs)

	resolutionSvc := services.NewComboResolutionService(comboACL)
	submissionSvc := services.NewCartSubmissionService(cartACL)

	addToCartHandler := application.NewAddComboToCartHandler(repo, resolutionSvc, submissionSvc)

	handlers := api.NewHandlers(addToCartHandler)
	sessionValidator := &platformSessionValidator{}
	r := api.NewRouter(handlers, sessionValidator)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("cart-handoff service starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// platformSessionValidator is a compile-time placeholder.
// Replace with the real platform session adapter before deploying to ECS.
type platformSessionValidator struct{}

func (v *platformSessionValidator) ValidateSession(r *http.Request) (domain.ShopperId, error) {
	panic("platformSessionValidator: not implemented — wire the real platform session service")
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return val
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
