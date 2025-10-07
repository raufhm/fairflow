package usecase_test

import (
	"testing"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestCreateGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewGroupUseCase(groupRepo, nil, auditRepo)

	group, err := uc.CreateGroup(1, "Admin", "Test Group", nil, domain.StrategyWeightedRoundRobin)

	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, "Test Group", group.Name)
}

func TestGetGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group"}
	uc := usecase.NewGroupUseCase(groupRepo, nil, nil)

	group, err := uc.GetGroup(1)

	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, "Test Group", group.Name)
}

func TestGetAllGroups(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group 1"}
	groupRepo.groups[2] = &domain.Group{ID: 2, UserID: 1, Name: "Test Group 2"}
	uc := usecase.NewGroupUseCase(groupRepo, nil, nil)

	groups, err := uc.GetAllGroups()

	assert.NoError(t, err)
	assert.Len(t, groups, 2)
}

func TestGetUserGroups(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group 1"}
	groupRepo.groups[2] = &domain.Group{ID: 2, UserID: 2, Name: "Test Group 2"}
	uc := usecase.NewGroupUseCase(groupRepo, nil, nil)

	groups, err := uc.GetUserGroups(1)

	assert.NoError(t, err)
	assert.Len(t, groups, 1)
}

func TestUpdateGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group"}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewGroupUseCase(groupRepo, nil, auditRepo)

	newName := "New Name"
	group, err := uc.UpdateGroup(1, 1, "Admin", &newName, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", group.Name)
}

func TestDeleteGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group"}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewGroupUseCase(groupRepo, nil, auditRepo)

	err := uc.DeleteGroup(1, 1, "Admin")

	assert.NoError(t, err)
	group, _ := groupRepo.GetByID(1)
	assert.Nil(t, group)
}

func TestCanModifyGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group"}
	uc := usecase.NewGroupUseCase(groupRepo, nil, nil)

	can, err := uc.CanModifyGroup(1, 1, domain.RoleUser)
	assert.NoError(t, err)
	assert.True(t, can)

	can, err = uc.CanModifyGroup(1, 2, domain.RoleAdmin)
	assert.NoError(t, err)
	assert.True(t, can)

	can, err = uc.CanModifyGroup(1, 2, domain.RoleUser)
	assert.NoError(t, err)
	assert.False(t, can)
}

func TestPauseGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group"}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewGroupUseCase(groupRepo, nil, auditRepo)

	err := uc.PauseGroup(1, 1, "Admin", nil)

	assert.NoError(t, err)
	group, _ := groupRepo.GetByID(1)
	assert.True(t, group.AssignmentPaused)
}

func TestResumeGroup(t *testing.T) {
	groupRepo := newMockGroupRepo()
	groupRepo.groups[1] = &domain.Group{ID: 1, UserID: 1, Name: "Test Group", AssignmentPaused: true}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewGroupUseCase(groupRepo, nil, auditRepo)

	err := uc.ResumeGroup(1, 1, "Admin")

	assert.NoError(t, err)
	group, _ := groupRepo.GetByID(1)
	assert.False(t, group.AssignmentPaused)
}
