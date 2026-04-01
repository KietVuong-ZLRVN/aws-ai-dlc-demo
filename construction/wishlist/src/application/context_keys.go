package application

import "context"

type contextKey string

const SessionTokenKey contextKey = "sessionToken"

func SessionTokenFromContext(ctx context.Context) string {
	v, _ := ctx.Value(SessionTokenKey).(string)
	return v
}
