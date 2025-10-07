package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/raufhm/fairflow/internal/config"
	"github.com/raufhm/fairflow/internal/database"
	httpdelivery "github.com/raufhm/fairflow/internal/delivery/http"
	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/repository/postgres"
	"github.com/raufhm/fairflow/internal/usecase"
	"github.com/raufhm/fairflow/pkg/crypto"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Initialize slog to output to stdout
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	// Load configuration
	cfg := config.Load()

	slog.Info("Starting FairFlow API", "environment", cfg.Environment, "port", cfg.Port)

	// Initialize database
	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	slog.Info("Database connected successfully")

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
	authHandler := httpdelivery.NewAuthHandler(authUseCase)
	groupHandler := httpdelivery.NewGroupHandler(groupUseCase)
	memberHandler := httpdelivery.NewMemberHandler(memberUseCase, groupUseCase)
	assignmentHandler := httpdelivery.NewAssignmentHandler(assignmentUseCase)
	adminHandler := httpdelivery.NewAdminHandler(adminUseCase)
	webhookHandler := httpdelivery.NewWebhookHandler(webhookUseCase)
	analyticsHandler := httpdelivery.NewAnalyticsHandler(assignmentUseCase, memberUseCase, groupUseCase)

	// Initialize token service
	tokenService := crypto.NewTokenService(cfg.JWTSecret)

	// Setup router
	router := httpdelivery.NewRouter(
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
	slog.Info("Server is running on http://localhost" + addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func seedAdminUser(db *bun.DB, cfg *config.Config) error {
	userRepo := postgres.NewUserRepository(db)

	// Check if any super_admin exists
	users, err := userRepo.GetAll()
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

		if err := userRepo.Create(admin); err != nil {
			return err
		}

		log.Println("Initial Super Admin user created successfully")
		log.Println("Email: admin@rr.io, Password: password")
		log.Println("⚠️  IMPORTANT: Change the default password immediately!")
	}

	return nil
}
