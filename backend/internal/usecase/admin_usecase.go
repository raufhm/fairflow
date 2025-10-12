package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/raufhm/fairflow/internal/domain"
)

type AdminUseCase struct {
	userRepo  domain.UserRepository
	auditRepo domain.AuditLogRepository
}

func NewAdminUseCase(
	userRepo domain.UserRepository,
	auditRepo domain.AuditLogRepository,
) *AdminUseCase {
	return &AdminUseCase{
		userRepo:  userRepo,
		auditRepo: auditRepo,
	}
}

// GetAllUsers retrieves all users
func (uc *AdminUseCase) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return uc.userRepo.GetAll(ctx)
}

// UpdateUserRole updates a user's role
func (uc *AdminUseCase) UpdateUserRole(ctx context.Context, targetUserID, adminUserID int64, adminUserName string, role domain.UserRole) error {
	// Safety checks
	if targetUserID == adminUserID {
		return errors.New("cannot modify your own account via this endpoint")
	}
	if role == domain.RoleSuperAdmin {
		return errors.New("cannot promote users to super_admin via API")
	}

	if err := uc.userRepo.UpdateRole(ctx, targetUserID, role); err != nil {
		return err
	}

	// Audit log
	_ = uc.auditRepo.Create(ctx, &domain.AuditLog{
		UserID:       &adminUserID,
		UserName:     adminUserName,
		Action:       "User role changed",
		ResourceType: stringPtr("user"),
		ResourceID:   &targetUserID,
		Details:      stringPtr(fmt.Sprintf(`{"new_role":"%s"}`, role)),
		IPAddress:    "unknown",
	})

	return nil
}

// DeleteUser deletes a user
func (uc *AdminUseCase) DeleteUser(ctx context.Context, targetUserID, adminUserID int64, adminUserName string) error {
	if targetUserID == adminUserID {
		return errors.New("cannot delete your own account")
	}

	if err := uc.userRepo.Delete(ctx, targetUserID); err != nil {
		return err
	}

	// Audit log
	_ = uc.auditRepo.Create(ctx, &domain.AuditLog{
		UserID:       &adminUserID,
		UserName:     adminUserName,
		Action:       "User deleted",
		ResourceType: stringPtr("user"),
		ResourceID:   &targetUserID,
		IPAddress:    "unknown",
	})

	return nil
}

// GetAuditLogs retrieves recent audit logs
func (uc *AdminUseCase) GetAuditLogs(ctx context.Context, limit int) ([]*domain.AuditLog, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	return uc.auditRepo.GetRecent(ctx, limit)
}
