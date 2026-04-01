package api

import (
	"context"
	"net/http"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"
)

type contextKey string

const (
	shopperIDKey     contextKey = "shopperID"
	sessionCookieKey contextKey = "sessionCookie"
)

// SessionValidator validates session cookies.
type SessionValidator interface {
	ValidateSession(r *http.Request) (domain.ShopperId, error)
}

// AuthMiddleware validates the session and injects ShopperId + raw cookie into context.
func AuthMiddleware(validator SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			shopperID, err := validator.ValidateSession(r)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			// Forward the raw Cookie header so ACL adapters can pass it to upstream services.
			rawCookie := r.Header.Get("Cookie")
			ctx := context.WithValue(r.Context(), shopperIDKey, shopperID)
			ctx = context.WithValue(ctx, sessionCookieKey, rawCookie)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ShopperIDFromContext(ctx context.Context) (domain.ShopperId, bool) {
	v, ok := ctx.Value(shopperIDKey).(domain.ShopperId)
	return v, ok
}

func SessionCookieFromContext(ctx context.Context) string {
	v, _ := ctx.Value(sessionCookieKey).(string)
	return v
}
