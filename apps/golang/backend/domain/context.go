package domain

import "context"

type contextKey int

const (
	contextKeyUserID contextKey = iota
	contextKeyTenantID
	contextKeyPlatformRole
	contextKeyTenantRole
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

func ContextWithPlatformRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, contextKeyPlatformRole, role)
}

func PlatformRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(contextKeyPlatformRole).(string)
	return role, ok
}

func ContextWithTenantRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, contextKeyTenantRole, role)
}

func TenantRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(contextKeyTenantRole).(string)
	return role, ok
}
