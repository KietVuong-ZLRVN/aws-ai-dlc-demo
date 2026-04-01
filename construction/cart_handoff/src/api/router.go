package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter builds the go-chi router for the cart-handoff service.
func NewRouter(h *Handlers, sessionValidator SessionValidator) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(AuthMiddleware(sessionValidator))

	r.Post("/api/v1/cart/combo", h.AddComboToCart)

	return r
}
