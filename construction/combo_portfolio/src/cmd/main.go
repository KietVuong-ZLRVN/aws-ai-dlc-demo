package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/api"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/application"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/acl"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/persistence"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/services"
)

func main() {
	dsn := mustEnv("DB_DSN")
	productCatalogBaseURL := mustEnv("PRODUCT_CATALOG_BASE_URL")
	contentLanguage := getEnv("CONTENT_LANGUAGE", "en-SG")
	timeoutMs := getEnvInt("PRODUCT_CATALOG_TIMEOUT_MS", 2000)
	publicBaseURL := getEnv("PUBLIC_BASE_URL", "http://localhost:8080")
	port := getEnv("PORT", "8080")

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

	repo := persistence.NewMySQLComboRepository(db)
	catalogACL := acl.NewHTTPProductCatalogACL(productCatalogBaseURL, contentLanguage, timeoutMs)
	enrichmentSvc := acl.NewComboEnrichmentService(catalogACL)
	tokenSvc := services.NewShareTokenService(repo)

	saveHandler := application.NewSaveComboHandler(repo)
	renameHandler := application.NewRenameComboHandler(repo)
	deleteHandler := application.NewDeleteComboHandler(repo)
	shareHandler := application.NewShareComboHandler(repo, tokenSvc)
	makePrivHandler := application.NewMakePrivateHandler(repo)
	getHandler := application.NewGetComboHandler(repo, enrichmentSvc)
	listHandler := application.NewListCombosHandler(repo, enrichmentSvc)
	sharedHandler := application.NewGetSharedComboHandler(repo, enrichmentSvc)

	handlers := api.NewHandlers(
		saveHandler, renameHandler, deleteHandler, shareHandler,
		makePrivHandler, getHandler, listHandler, sharedHandler,
		publicBaseURL,
	)

	sessionValidator := &platformSessionValidator{}
	r := api.NewRouter(handlers, sessionValidator)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("combo-portfolio service starting on %s", addr)
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
