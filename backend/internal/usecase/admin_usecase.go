package usecase

import (
	"errors"
	"fmt"

	"github.com/raufhm/rra/internal/domain"
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
func (uc *AdminUseCase) GetAllUsers() ([]*domain.User, error) {
	return uc.userRepo.GetAll()
}

// UpdateUserRole updates a user's role
func (uc *AdminUseCase) UpdateUserRole(targetUserID, adminUserID int64, adminUserName string, role domain.UserRole) error {
	// Safety checks
	if targetUserID == adminUserID {
		return errors.New("cannot modify your own account via this endpoint")
	}
	if role == domain.RoleSuperAdmin {
		return errors.New("cannot promote users to super_admin via API")
	}

	if err := uc.userRepo.UpdateRole(targetUserID, role); err != nil {
		return err
	}

	// Audit log
	_ = uc.auditRepo.Create(&domain.AuditLog{
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
func (uc *AdminUseCase) DeleteUser(targetUserID, adminUserID int64, adminUserName string) error {
	if targetUserID == adminUserID {
		return errors.New("cannot delete your own account")
	}

	if err := uc.userRepo.Delete(targetUserID); err != nil {
		return err
	}

	// Audit log
	_ = uc.auditRepo.Create(&domain.AuditLog{
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
func (uc *AdminUseCase) GetAuditLogs(limit int) ([]*domain.AuditLog, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	return uc.auditRepo.GetRecent(limit)
}