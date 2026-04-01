package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	getHandler    *WishlistGetHandler
	addHandler    *WishlistAddHandler
	removeHandler *WishlistRemoveHandler
}

func NewRouter(
	getHandler *WishlistGetHandler,
	addHandler *WishlistAddHandler,
	removeHandler *WishlistRemoveHandler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(corsMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/api/v1/wishlist", getHandler.Handle)
	r.Post("/api/v1/wishlist/items", addHandler.Handle)
	r.Delete("/api/v1/wishlist/items/{configSku}", removeHandler.Handle)

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Cookie")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
