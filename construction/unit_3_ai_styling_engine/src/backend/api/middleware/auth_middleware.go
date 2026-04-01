package middleware

import (
	"ai-styling-engine/domain/valueobjects"
	"context"
	"encoding/json"
	"net/http"
)

// AuthMiddleware validates the shopper's session.
// Local mode: accepts any non-empty cookie value as valid.
// Production: replace with JWT validation using a signing key.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "UNAUTHENTICATED"})
			return
		}
		session := valueobjects.ShopperSession{SessionToken: cookie.Value}
		ctx := context.WithValue(r.Context(), ContextKeyShopperSession, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SessionFromContext retrieves the ShopperSession from the request context.
func SessionFromContext(ctx context.Context) (valueobjects.ShopperSession, bool) {
	session, ok := ctx.Value(ContextKeyShopperSession).(valueobjects.ShopperSession)
	return session, ok
}
