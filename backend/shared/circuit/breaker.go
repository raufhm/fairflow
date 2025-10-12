package circuit

import (
	"time"

	"github.com/failsafe-go/failsafe-go/circuitbreaker"
	"github.com/raufhm/fairflow/shared/logger"
	"go.uber.org/zap"
)

// NewDatabaseCircuitBreaker creates a circuit breaker for database operations
func NewDatabaseCircuitBreaker() circuitbreaker.CircuitBreaker[any] {
	cb := circuitbreaker.Builder[any]().
		WithFailureThreshold(5).
		WithSuccessThreshold(2).
		WithDelay(30 * time.Second)

	cb.OnOpen(func(event circuitbreaker.StateChangedEvent) {
		logger.Log.Warn("Database circuit breaker opened")
	})

	cb.OnClose(func(event circuitbreaker.StateChangedEvent) {
		logger.Log.Info("Database circuit breaker closed")
	})

	return cb.Build()
}

// NewHTTPCircuitBreaker creates a circuit breaker for HTTP calls
func NewHTTPCircuitBreaker(serviceName string) circuitbreaker.CircuitBreaker[any] {
	cb := circuitbreaker.Builder[any]().
		WithFailureThresholdRatio(50, 100).
		WithDelay(30 * time.Second)

	cb.OnOpen(func(event circuitbreaker.StateChangedEvent) {
		logger.Log.Warn("HTTP circuit breaker opened",
			zap.String("service", serviceName),
		)
	})

	cb.OnClose(func(event circuitbreaker.StateChangedEvent) {
		logger.Log.Info("HTTP circuit breaker closed",
			zap.String("service", serviceName),
		)
	})

	return cb.Build()
}
