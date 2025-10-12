package middleware

import (
	"context"
	"net/http"

	"github.com/raufhm/fairflow/shared/domain"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// AdminOnly middleware ensures user is admin or super_admin
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, `{"message":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		if user.Role != domain.RoleAdmin && user.Role != domain.RoleSuperAdmin {
			http.Error(w, `{"message":"Forbidden: Admin access required"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SuperAdminOnly middleware ensures user is super_admin
func SuperAdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, `{"message":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		if user.Role != domain.RoleSuperAdmin {
			http.Error(w, `{"message":"Forbidden: Super Admin access required"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) *domain.User {
	user, ok := ctx.Value(UserContextKey).(*domain.User)
	if !ok {
		return nil
	}
	return user
}
