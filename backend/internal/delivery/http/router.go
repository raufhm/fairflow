package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/raufhm/rra/internal/middleware"
	"github.com/raufhm/rra/internal/usecase"
	"github.com/raufhm/rra/pkg/crypto"
)

type Router struct {
	authHandler       *AuthHandler
	groupHandler      *GroupHandler
	memberHandler     *MemberHandler
	assignmentHandler *AssignmentHandler
	adminHandler      *AdminHandler
	webhookHandler    *WebhookHandler
	analyticsHandler  *AnalyticsHandler
	authUseCase       *usecase.AuthUseCase
	tokenService      *crypto.TokenService
}

func NewRouter(
	authHandler *AuthHandler,
	groupHandler *GroupHandler,
	memberHandler *MemberHandler,
	assignmentHandler *AssignmentHandler,
	adminHandler *AdminHandler,
	webhookHandler *WebhookHandler,
	analyticsHandler *AnalyticsHandler,
	authUseCase *usecase.AuthUseCase,
	tokenService *crypto.TokenService,
) *Router {
	return &Router{
		authHandler:       authHandler,
		groupHandler:      groupHandler,
		memberHandler:     memberHandler,
		assignmentHandler: assignmentHandler,
		adminHandler:      adminHandler,
		webhookHandler:    webhookHandler,
		analyticsHandler:  analyticsHandler,
		authUseCase:       authUseCase,
		tokenService:      tokenService,
	}
}

func (r *Router) SetupRoutes() http.Handler {
	router := chi.NewRouter()

	// Apply CORS middleware
	router.Use(middleware.CORS)

	// Auth middleware wrapper
	authMiddleware := middleware.AuthMiddleware(r.authUseCase, r.tokenService)

	// Public routes (no auth required)
	router.Post("/api/v1/auth/register", r.authHandler.Register)
	router.Post("/api/v1/auth/login", r.authHandler.Login)
	router.Post("/api/v1/auth/forgot-password", r.authHandler.ForgotPassword)

	// Protected routes
	router.Group(func(rt chi.Router) {
		rt.Use(authMiddleware)

		// Auth routes
		rt.Patch("/api/v1/auth/user-settings", r.authHandler.UpdateUserSettings)
		rt.Get("/api/v1/auth/api-keys", r.authHandler.GetAPIKeys)
		rt.Post("/api/v1/auth/api-keys", r.authHandler.CreateAPIKey)
		rt.Delete("/api/v1/auth/api-keys/{id}", r.authHandler.RevokeAPIKey)

		// Group routes
		rt.Get("/api/v1/groups", r.groupHandler.GetAllGroups)
		rt.Post("/api/v1/groups", r.groupHandler.CreateGroup)
		rt.Get("/api/v1/groups/{id}", r.groupHandler.GetGroup)
		rt.Patch("/api/v1/groups/{id}", r.groupHandler.UpdateGroup)
		rt.Delete("/api/v1/groups/{id}", r.groupHandler.DeleteGroup)
		rt.Post("/api/v1/groups/{id}/pause", r.groupHandler.PauseGroup)
		rt.Post("/api/v1/groups/{id}/resume", r.groupHandler.ResumeGroup)

		// Group member routes
		rt.Get("/api/v1/groups/{groupId}/members", r.memberHandler.GetMembers)
		rt.Post("/api/v1/groups/{groupId}/members", r.memberHandler.CreateMember)
		rt.Patch("/api/v1/groups/{groupId}/members/{id}", r.memberHandler.UpdateMember)
		rt.Delete("/api/v1/groups/{groupId}/members/{id}", r.memberHandler.DeleteMember)
		rt.Get("/api/v1/members/{id}/capacity", r.memberHandler.GetMemberCapacity)

		// Assignment routes
		rt.Get("/api/v1/groups/{id}/next", r.assignmentHandler.GetNextAssignee)
		rt.Post("/api/v1/groups/{id}/assign", r.assignmentHandler.RecordAssignment)
		rt.Get("/api/v1/groups/{id}/assignments", r.assignmentHandler.GetAssignments)
		rt.Get("/api/v1/groups/{id}/stats", r.assignmentHandler.GetStats)
		rt.Post("/api/v1/assignments/{id}/complete", r.assignmentHandler.CompleteAssignment)
		rt.Post("/api/v1/assignments/{id}/cancel", r.assignmentHandler.CancelAssignment)

		// Webhook routes
		rt.Get("/api/v1/groups/{groupId}/webhooks", r.webhookHandler.GetWebhooks)
		rt.Post("/api/v1/groups/{groupId}/webhooks", r.webhookHandler.CreateWebhook)
		rt.Delete("/api/v1/webhooks/{id}", r.webhookHandler.DeleteWebhook)

		// Analytics routes
		rt.Get("/api/v1/groups/{id}/analytics/fairness", r.analyticsHandler.GetFairnessMetrics)
		rt.Get("/api/v1/groups/{id}/analytics/trends", r.analyticsHandler.GetWorkloadTrends)
		rt.Get("/api/v1/groups/{id}/analytics/performance", r.analyticsHandler.GetMemberPerformance)

		// Admin routes
		rt.Group(func(admin chi.Router) {
			admin.Use(middleware.AdminOnly)

			admin.Get("/api/v1/admin/users", r.adminHandler.GetAllUsers)
			admin.Patch("/api/v1/admin/users/{id}", r.adminHandler.UpdateUserRole)
			admin.Delete("/api/v1/admin/users/{id}", r.adminHandler.DeleteUser)
			admin.Get("/api/v1/admin/audit-logs", r.adminHandler.GetAuditLogs)
			admin.Post("/api/v1/admin/backup", r.adminHandler.Backup)
			admin.Get("/api/v1/admin/backups", r.adminHandler.GetBackups)
			admin.Post("/api/v1/admin/restore", r.adminHandler.Restore)
			admin.Get("/api/v1/admin/export", r.adminHandler.ExportData)
		})
	})

	return router
}
