package domain

import "context"

type contextKey int

const (
	contextKeyUserID contextKey = iota
	contextKeyTenantID
)

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(contextKeyUserID).(string)
	return id, ok
}

func ContextWithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, contextKeyTenantID, tenantID)
}

func TenantIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(contextKeyTenantID).(string)
	return id, ok
}
