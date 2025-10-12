package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raufhm/fairflow/shared/config"
	"github.com/raufhm/fairflow/shared/database"
	"github.com/raufhm/fairflow/shared/health"
	"github.com/raufhm/fairflow/shared/logger"
	"github.com/raufhm/fairflow/shared/middleware"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	defer logger.Log.Sync()

	// Load configuration
	cfg := config.Load()

	logger.Log.Info("Starting Analytics Service",
		zap.String("environment", cfg.Environment),
		zap.Int("port", cfg.Port),
	)

	// Initialize database
	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		logger.Log.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	logger.Log.Info("Database connected successfully")

	// Setup HTTP router
	mux := http.NewServeMux()

	// Health check
	healthChecker := health.NewHealthChecker(db)
	mux.HandleFunc("/health", healthChecker.Handler("analytics-service", "1.0.0"))

	// TODO: Add analytics endpoints

	// Apply middleware
	handlerWithMiddleware := middleware.CORS(mux)

	// Start HTTP server
	port := 3007
	if cfg.Port != 0 {
		port = cfg.Port
	}
	addr := fmt.Sprintf(":%d", port)

	srv := &http.Server{
		Addr:    addr,
		Handler: handlerWithMiddleware,
	}

	go func() {
		logger.Log.Info("Analytics Service is running on " + addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down Analytics Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown:", zap.Error(err))
	}

	logger.Log.Info("Analytics Service exited successfully")
}
