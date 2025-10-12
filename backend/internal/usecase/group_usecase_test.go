package usecase_test

import (
	"context"
	"testing"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestCreateGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewGroupUseCase(groupRepo, nil, auditRepo)

	group, err := uc.CreateGroup(context.Background(), 1, "Admin", "Test Group", nil, domain.StrategyWeightedRoundRobin)

	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, "Test Group", group.Name)
}

func TestGetGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group"}
	uc := usecase.NewGroupUseCase(groupRepo, nil, nil)

	group, err := uc.GetGroup(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, "Test Group", group.Name)
}

func TestGetAllGroups(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group 1"}
	groupRepo.groups[2] = &domain.Group{ID: 2, UserID: 1, Name: "Test Group 2"}
	uc := usecase.NewGroupUseCase(groupRepo, nil, nil)

	groups, err := uc.GetAllGroups(context.Background())

	assert.NoError(t, err)
	assert.Len(t, groups, 2)
}

func TestGetUserGroups(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group 1"}
	groupRepo.groups[2] = &domain.Group{ID: 2, UserID: 2, Name: "Test Group 2"}
	uc := usecase.NewGroupUseCase(groupRepo, nil, nil)

	groups, err := uc.GetUserGroups(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, groups, 1)
}

func TestUpdateGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group"}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewGroupUseCase(groupRepo, nil, auditRepo)

	newName := "New Name"
	group, err := uc.UpdateGroup(context.Background(), 1, 1, "Admin", &newName, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", group.Name)
}

func TestDeleteGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group"}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewGroupUseCase(groupRepo, nil, auditRepo)

	err := uc.DeleteGroup(context.Background(), 1, 1, "Admin")

	assert.NoError(t, err)
	group, _ := groupRepo.GetByID(context.Background(), 1)
	assert.Nil(t, group)
}

func TestCanModifyGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group"}
	uc := usecase.NewGroupUseCase(groupRepo, nil, nil)

	can, err := uc.CanModifyGroup(context.Background(), 1, 1, domain.RoleUser)
	assert.NoError(t, err)
	assert.True(t, can)

	can, err = uc.CanModifyGroup(context.Background(), 1, 2, domain.RoleAdmin)
	assert.NoError(t, err)
	assert.True(t, can)

	can, err = uc.CanModifyGroup(context.Background(), 1, 2, domain.RoleUser)
	assert.NoError(t, err)
	assert.False(t, can)
}

func TestPauseGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group"}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewGroupUseCase(groupRepo, nil, auditRepo)

	err := uc.PauseGroup(context.Background(), 1, 1, "Admin", nil)

	assert.NoError(t, err)
	group, _ := groupRepo.GetByID(context.Background(), 1)
	assert.True(t, group.AssignmentPaused)
}

func TestResumeGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group", AssignmentPaused: true}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewGroupUseCase(groupRepo, nil, auditRepo)

	err := uc.ResumeGroup(context.Background(), 1, 1, "Admin")

	assert.NoError(t, err)
	group, _ := groupRepo.GetByID(context.Background(), 1)
	assert.False(t, group.AssignmentPaused)
}
