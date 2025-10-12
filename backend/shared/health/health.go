package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/uptrace/bun"
)

// HealthChecker provides health check functionality
type HealthChecker struct {
	db *bun.DB
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *bun.DB) *HealthChecker {
	return &HealthChecker{db: db}
}

// Handler returns an HTTP handler for health checks
func (hc *HealthChecker) Handler(serviceName, version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := map[string]interface{}{
			"status":  "healthy",
			"service": serviceName,
			"version": version,
			"checks":  map[string]string{},
		}

		// Check database
		if hc.db != nil {
			if err := hc.db.PingContext(ctx); err != nil {
				status["status"] = "unhealthy"
				status["checks"].(map[string]string)["database"] = "down"
				w.WriteHeader(http.StatusServiceUnavailable)
			} else {
				status["checks"].(map[string]string)["database"] = "up"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}

// SimpleHandler returns a simple health check that doesn't check dependencies
func SimpleHandler(serviceName, version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"status":  "healthy",
			"service": serviceName,
			"version": version,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}
