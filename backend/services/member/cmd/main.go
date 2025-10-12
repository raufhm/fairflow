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

	"github.com/raufhm/fairflow/services/member/internal/handler"
	"github.com/raufhm/fairflow/services/member/internal/usecase"
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

	logger.Log.Info("Starting Member Service",
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
	memberRepo := postgres.NewMemberRepository(db)
	groupRepo := postgres.NewGroupRepository(db)

	// Initialize use case
	memberUseCase := usecase.NewMemberUseCase(memberRepo, groupRepo)

	// Initialize handler
	memberHandler := handler.NewMemberHandler(memberUseCase)

	// Setup HTTP router
	mux := http.NewServeMux()

	// Health check
	healthChecker := health.NewHealthChecker(db)
	mux.HandleFunc("/health", healthChecker.Handler("member-service", "1.0.0"))

	// Member endpoints
	mux.HandleFunc("/api/v1/groups/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/members") {
			if r.Method == http.MethodGet {
				memberHandler.GetMembers(w, r)
			} else if r.Method == http.MethodPost {
				memberHandler.CreateMember(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	mux.HandleFunc("/api/v1/members/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capacity") {
			memberHandler.GetMemberCapacity(w, r)
		} else if r.Method == http.MethodGet {
			memberHandler.GetMember(w, r)
		} else if r.Method == http.MethodPut {
			memberHandler.UpdateMember(w, r)
		} else if r.Method == http.MethodDelete {
			memberHandler.DeleteMember(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Apply middleware
	handlerWithMiddleware := middleware.CORS(mux)

	// Start HTTP server
	port := 3003
	if cfg.Port != 0 {
		port = cfg.Port
	}
	addr := fmt.Sprintf(":%d", port)

	srv := &http.Server{
		Addr:    addr,
		Handler: handlerWithMiddleware,
	}

	go func() {
		logger.Log.Info("Member Service is running on " + addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down Member Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown:", zap.Error(err))
	}

	logger.Log.Info("Member Service exited successfully")
}
