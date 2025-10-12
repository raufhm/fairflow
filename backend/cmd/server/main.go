package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raufhm/fairflow/internal/config"
	"github.com/raufhm/fairflow/internal/database"
	"github.com/raufhm/fairflow/internal/delivery/restful"
	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/repository/postgres"
	"github.com/raufhm/fairflow/internal/usecase"
	"github.com/raufhm/fairflow/pkg/crypto"
	"github.com/raufhm/fairflow/pkg/logger"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Initialize zap logger
	defer logger.Log.Sync() // flushes buffer, if any

	// Load configuration
	cfg := config.Load()

	logger.Log.Info("Starting FairFlow API", zap.String("environment", cfg.Environment), zap.Int("port", cfg.Port))

	// Initialize database
	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		logger.Log.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	logger.Log.Info("Database connected successfully")

	// Seed initial admin user if needed
	if err := seedAdminUser(db, cfg); err != nil {
		log.Printf("Warning: Failed to seed admin user: %v", err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	groupRepo := postgres.NewGroupRepository(db)
	memberRepo := postgres.NewMemberRepository(db)
	assignmentRepo := postgres.NewAssignmentRepository(db)
	apiKeyRepo := postgres.NewAPIKeyRepository(db)
	auditRepo := postgres.NewAuditLogRepository(db)
	webhookRepo := postgres.NewWebhookRepository(db)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(userRepo, apiKeyRepo, auditRepo, cfg.JWTSecret)
	groupUseCase := usecase.NewGroupUseCase(groupRepo, memberRepo, auditRepo)
	memberUseCase := usecase.NewMemberUseCase(memberRepo, groupRepo, auditRepo)
	assignmentUseCase := usecase.NewAssignmentUseCase(groupRepo, memberRepo, assignmentRepo, auditRepo)
	adminUseCase := usecase.NewAdminUseCase(userRepo, auditRepo)
	webhookUseCase := usecase.NewWebhookUseCase(webhookRepo, auditRepo)

	// Initialize handlers
	authHandler := restful.NewAuthHandler(authUseCase)
	groupHandler := restful.NewGroupHandler(groupUseCase)
	memberHandler := restful.NewMemberHandler(memberUseCase, groupUseCase)
	assignmentHandler := restful.NewAssignmentHandler(assignmentUseCase)
	adminHandler := restful.NewAdminHandler(adminUseCase)
	webhookHandler := restful.NewWebhookHandler(webhookUseCase)
	analyticsHandler := restful.NewAnalyticsHandler(assignmentUseCase, memberUseCase, groupUseCase)

	// Initialize token service
	tokenService := crypto.NewTokenService(cfg.JWTSecret)

	// Setup router
	router := restful.NewRouter(
		authHandler,
		groupHandler,
		memberHandler,
		assignmentHandler,
		adminHandler,
		webhookHandler,
		analyticsHandler,
		authUseCase,
		tokenService,
	)

	handler := router.SetupRoutes()

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		logger.Log.Info("Server is running on " + addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown:", zap.Error(err))
	}

	logger.Log.Info("Server exiting")
}

func seedAdminUser(db *bun.DB, cfg *config.Config) error {
	ctx := context.Background()
	userRepo := postgres.NewUserRepository(db)

	// Check if any super_admin exists
	users, err := userRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	hasSuperAdmin := false
	for _, user := range users {
		if user.Role == domain.RoleSuperAdmin {
			hasSuperAdmin = true
			break
		}
	}

	if !hasSuperAdmin {
		log.Println("No admin found. Seeding initial Super Admin user.")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		admin := &domain.User{
			Name:         "Super Admin",
			Email:        "admin@rr.io",
			PasswordHash: string(hashedPassword),
			Role:         domain.RoleSuperAdmin,
		}

		if err := userRepo.Create(ctx, admin); err != nil {
			return err
		}

		log.Println("Initial Super Admin user created successfully")
		log.Println("Email: admin@rr.io, Password: password")
		log.Println("⚠️  IMPORTANT: Change the default password immediately!")
	}

	return nil
}
