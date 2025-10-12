package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/usecase"
	"github.com/raufhm/fairflow/pkg/crypto"
	"github.com/raufhm/fairflow/pkg/logger"
	"go.uber.org/zap"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// AuthMiddleware handles JWT and API key authentication
func AuthMiddleware(authUseCase *usecase.AuthUseCase, tokenService *crypto.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try JWT token first
			authHeader := r.Header.Get("Authorization")
			logger.Log.Info("Auth attempt", zap.String("path", r.URL.Path), zap.String("authHeader", authHeader))
			fmt.Fprintf(os.Stdout, "AUTH MW: %s %s | Header: %s\n", r.Method, r.URL.Path, authHeader)
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					userID, err := tokenService.VerifyToken(parts[1])
					logger.Log.Info("Token verification", zap.Error(err), zap.Int64("userID", userID))
					if err == nil {
						user, err := authUseCase.GetUserByID(r.Context(), userID)
						logger.Log.Info("User lookup", zap.Error(err), zap.Any("user", user))
						if err == nil && user != nil {
							logger.Log.Info("Auth success", zap.String("userRole", string(user.Role)))
							ctx := context.WithValue(r.Context(), UserContextKey, user)
							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}
					}
				}
			}

			// Try API key
			apiKey := r.Header.Get("X-Api-Key")
			if apiKey != "" {
				user, err := authUseCase.VerifyAPIKey(r.Context(), apiKey)
				if err == nil && user != nil {
					ctx := context.WithValue(r.Context(), UserContextKey, user)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// No valid authentication
			logger.Log.Info("Auth failed - no valid credentials")
			http.Error(w, `{"message":"Authentication required"}`, http.StatusUnauthorized)
		})
	}
}

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
