package main

import (
	"ai-styling-engine/api"
	"ai-styling-engine/api/controllers"
	"ai-styling-engine/application/usecases"
	infraAI "ai-styling-engine/infrastructure/ai"
	"ai-styling-engine/infrastructure/acl"
	"ai-styling-engine/infrastructure/dispatcher"
	"ai-styling-engine/domain/events"
	"ai-styling-engine/domain/policies"
	"log"
	"net/http"
)

func main() {
	// ── Infrastructure ────────────────────────────────────────────────────────

	// In-process event dispatcher (manual wiring — no DI framework).
	d := dispatcher.NewInProcessEventDispatcher()

	// Register policies with the dispatcher.
	registerPolicies(d)

	// In-memory repository stubs (replace with HTTP ACL implementations in production).
	wishlistRepo := acl.NewInMemoryWishlistRepository()
	catalogRepo := acl.NewInMemoryProductCatalogRepository()
	completeLookRepo := acl.NewInMemoryCompleteLookRepository()

	// Mock AI service implementations (replace with Bedrock implementations in production).
	scoringSvc := infraAI.NewMockComboCompatibilityScoringService()
	reasoningSvc := infraAI.NewMockComboReasoningGenerationService()
	interpretSvc := infraAI.NewMockPreferenceInterpretationService()

	// ── Application ───────────────────────────────────────────────────────────

	getOptionsUC := usecases.NewGetPreferenceOptionsUseCase()
	confirmUC := usecases.NewConfirmPreferencesUseCase(interpretSvc, d)
	generateUC := usecases.NewGenerateCombosUseCase(
		wishlistRepo, catalogRepo, completeLookRepo,
		scoringSvc, reasoningSvc, d,
	)

	// ── API ───────────────────────────────────────────────────────────────────

	prefController := controllers.NewStylePreferencesController(getOptionsUC)
	confirmController := controllers.NewPreferenceConfirmationController(confirmUC)
	comboController := controllers.NewComboGenerationController(generateUC)

	router := api.NewRouter(prefController, confirmController, comboController)

	// ── Server ────────────────────────────────────────────────────────────────

	addr := ":8080"
	log.Printf("AI Styling Engine listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func registerPolicies(d events.EventDispatcher) {
	preferencePolicy := policies.NewPreferenceDefaultPolicy()
	d.Register(events.EventTypeComboGenerationRequested, preferencePolicy.Handle)

	wishlistPolicy := policies.NewWishlistSupplementationPolicy(d)
	d.Register(events.EventTypeWishlistFetchCompleted, wishlistPolicy.Handle)

	fallbackPolicy := policies.NewFallbackPolicy()
	d.Register(events.EventTypeFallbackTriggered, fallbackPolicy.Handle)

	exclusionPolicy := policies.NewComboExclusionPolicy()
	d.Register(events.EventTypeCombosGenerated, exclusionPolicy.Handle)
}
