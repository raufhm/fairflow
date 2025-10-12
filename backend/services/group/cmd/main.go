package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/raufhm/fairflow/services/group/internal/handler"
	"github.com/raufhm/fairflow/services/group/internal/usecase"
	"github.com/raufhm/fairflow/shared/config"
	"github.com/raufhm/fairflow/shared/database"
	"github.com/raufhm/fairflow/shared/health"
	"github.com/raufhm/fairflow/shared/logger"
	"github.com/raufhm/fairflow/shared/middleware"
	"github.com/raufhm/fairflow/shared/repository/postgres"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	defer logger.Log.Sync()

	// Load configuration
	cfg := config.Load()

	logger.Log.Info("Starting Group Service",
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

	// Initialize repositories
	groupRepo := postgres.NewGroupRepository(db)
	memberRepo := postgres.NewMemberRepository(db)

	// Initialize use case
	groupUseCase := usecase.NewGroupUseCase(groupRepo, memberRepo)

	// Initialize handler
	groupHandler := handler.NewGroupHandler(groupUseCase)

	// Setup HTTP router
	mux := http.NewServeMux()

	// Health check
	healthChecker := health.NewHealthChecker(db)
	mux.HandleFunc("/health", healthChecker.Handler("group-service", "1.0.0"))

	// Group endpoints
	mux.HandleFunc("/api/v1/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			groupHandler.GetAllGroups(w, r)
		} else if r.Method == http.MethodPost {
			groupHandler.CreateGroup(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/groups/", func(w http.ResponseWriter, r *http.Request) {
		// Check for pause/resume actions
		if strings.HasSuffix(r.URL.Path, "/pause") {
			groupHandler.PauseGroup(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/resume") {
			groupHandler.ResumeGroup(w, r)
		} else if r.Method == http.MethodGet {
			groupHandler.GetGroup(w, r)
		} else if r.Method == http.MethodPut {
			groupHandler.UpdateGroup(w, r)
		} else if r.Method == http.MethodDelete {
			groupHandler.DeleteGroup(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Apply middleware
	handlerWithMiddleware := middleware.CORS(mux)

	// Start HTTP server
	port := 3002
	if cfg.Port != 0 {
		port = cfg.Port
	}
	addr := fmt.Sprintf(":%d", port)

	srv := &http.Server{
		Addr:    addr,
		Handler: handlerWithMiddleware,
	}

	go func() {
		logger.Log.Info("Group Service is running on " + addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down Group Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown:", zap.Error(err))
	}

	logger.Log.Info("Group Service exited successfully")
}
