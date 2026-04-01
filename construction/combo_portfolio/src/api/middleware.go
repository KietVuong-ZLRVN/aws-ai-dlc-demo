package api

import (
	"context"
	"net/http"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
)

type contextKey string

const shopperIDKey contextKey = "shopperID"

// SessionValidator is the interface for validating session cookies.
// Implemented by the existing platform session service adapter.
type SessionValidator interface {
	ValidateSession(r *http.Request) (domain.ShopperId, error)
}

// AuthMiddleware validates the session cookie and injects ShopperId into context.
func AuthMiddleware(validator SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			shopperID, err := validator.ValidateSession(r)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			ctx := context.WithValue(r.Context(), shopperIDKey, shopperID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ShopperIDFromContext extracts the ShopperId from request context.
func ShopperIDFromContext(ctx context.Context) (domain.ShopperId, bool) {
	v, ok := ctx.Value(shopperIDKey).(domain.ShopperId)
	return v, ok
}
