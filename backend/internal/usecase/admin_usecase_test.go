package usecase_test

import (
	"testing"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestGetAllUsers(t *testing.T) {
	userRepo := newMockUserRepo()
	userRepo.users[1] = &domain.User{ID: 1, Name: "Admin", Email: "admin@example.com", Role: domain.RoleAdmin}
	userRepo.users[2] = &domain.User{ID: 2, Name: "User", Email: "user@example.com", Role: domain.RoleUser}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewAdminUseCase(userRepo, auditRepo)

	users, err := uc.GetAllUsers()

	assert.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUpdateUserRole(t *testing.T) {
	userRepo := newMockUserRepo()
	userRepo.users[1] = &domain.User{ID: 1, Name: "Admin", Email: "admin@example.com", Role: domain.RoleAdmin}
	userRepo.users[2] = &domain.User{ID: 2, Name: "User", Email: "user@example.com", Role: domain.RoleUser}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewAdminUseCase(userRepo, auditRepo)

	err := uc.UpdateUserRole(2, 1, "Admin", domain.RoleAdmin)

	assert.NoError(t, err)
	user, _ := userRepo.GetByID(2)
	assert.Equal(t, domain.RoleAdmin, user.Role)
}

func TestDeleteUser(t *testing.T) {
	userRepo := newMockUserRepo()
	userRepo.users[1] = &domain.User{ID: 1, Name: "Admin", Email: "admin@example.com", Role: domain.RoleAdmin}
	userRepo.users[2] = &domain.User{ID: 2, Name: "User", Email: "user@example.com", Role: domain.RoleUser}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewAdminUseCase(userRepo, auditRepo)

	err := uc.DeleteUser(2, 1, "Admin")

	assert.NoError(t, err)
	user, _ := userRepo.GetByID(2)
	assert.Nil(t, user)
}

func TestGetAuditLogs(t *testing.T) {
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewAdminUseCase(nil, auditRepo)

	_, err := uc.GetAuditLogs(10)

	assert.NoError(t, err)
}
