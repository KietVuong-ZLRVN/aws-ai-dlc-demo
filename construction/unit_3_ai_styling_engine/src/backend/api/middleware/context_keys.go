package middleware

type contextKey string

const (
	ContextKeyShopperSession contextKey = "shopperSession"
	ContextKeyTraceId        contextKey = "traceId"
)
