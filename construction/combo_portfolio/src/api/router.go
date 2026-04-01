package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter builds the go-chi router with all routes and middleware.
func NewRouter(h *Handlers, sessionValidator SessionValidator) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	// Public routes (no auth)
	r.Get("/api/v1/combos/shared/{token}", h.GetSharedCombo)

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(sessionValidator))
		r.Post("/api/v1/combos", h.SaveCombo)
		r.Get("/api/v1/combos", h.ListCombos)
		r.Get("/api/v1/combos/{id}", h.GetCombo)
		r.Put("/api/v1/combos/{id}", h.UpdateCombo)
		r.Delete("/api/v1/combos/{id}", h.DeleteCombo)
		r.Post("/api/v1/combos/{id}/share", h.ShareCombo)
	})

	return r
}
