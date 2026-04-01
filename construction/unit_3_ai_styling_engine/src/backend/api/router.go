package api

import (
	"ai-styling-engine/api/controllers"
	"ai-styling-engine/api/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	prefController *controllers.StylePreferencesController,
	confirmController *controllers.PreferenceConfirmationController,
	comboController *controllers.ComboGenerationController,
) http.Handler {
	r := chi.NewRouter()

	// Global middleware.
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.TracingMiddleware)
	r.Use(middleware.LoggingMiddleware)

	// CORS for local React dev server.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Cookie")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Authenticated routes.
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Get("/api/v1/style/preferences/options", prefController.GetOptions)
		r.Post("/api/v1/style/preferences/confirm", confirmController.Confirm)
		r.Post("/api/v1/style/combos/generate", comboController.Generate)
	})

	return r
}
