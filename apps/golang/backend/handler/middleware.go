package handler

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/user/micro-dp/domain"
)

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	secret := []byte(jwtSecret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				writeError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeError(w, http.StatusUnauthorized, "invalid authorization header")
				return
			}

			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secret, nil
			})
			if err != nil || !token.Valid {
				writeError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			sub, err := token.Claims.GetSubject()
			if err != nil || sub == "" {
				writeError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			ctx := domain.ContextWithUserID(r.Context(), sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TenantMiddleware(tenantRepo domain.TenantRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := r.Header.Get("X-Tenant-ID")
			if tenantID == "" {
				writeError(w, http.StatusBadRequest, "missing X-Tenant-ID header")
				return
			}

			userID, ok := domain.UserIDFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			isMember, err := tenantRepo.IsUserInTenant(r.Context(), userID, tenantID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}
			if !isMember {
				writeError(w, http.StatusForbidden, "not a member of this tenant")
				return
			}

			ctx := domain.ContextWithTenantID(r.Context(), tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
