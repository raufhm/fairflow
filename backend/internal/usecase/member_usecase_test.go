package usecase_test

import (
	"context"
	"testing"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestCreateMember(t *testing.T) {
	memberRepo := newMockMemberRepo()
	groupRepo := newMockGroupRepo()
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewMemberUseCase(memberRepo, groupRepo, auditRepo)

	member, err := uc.CreateMember(context.Background(), 1, 1, "Admin", "Test Member", nil, 100)

	assert.NoError(t, err)
	assert.NotNil(t, member)
	assert.Equal(t, "Test Member", member.Name)
}

func TestGetMembers(t *testing.T) {
	memberRepo := newMockMemberRepo()
	memberRepo.members[1] = &domain.Member{ID: 1, GroupID: 1, Name: "Test Member 1"}
	memberRepo.members[2] = &domain.Member{ID: 2, GroupID: 1, Name: "Test Member 2"}
	uc := usecase.NewMemberUseCase(memberRepo, nil, nil)

	members, err := uc.GetMembers(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, members, 2)
}

func TestGetMember(t *testing.T) {
	memberRepo := newMockMemberRepo()
	memberRepo.members[1] = &domain.Member{ID: 1, GroupID: 1, Name: "Test Member"}
	uc := usecase.NewMemberUseCase(memberRepo, nil, nil)

	member, err := uc.GetMember(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, member)
	assert.Equal(t, "Test Member", member.Name)
}

func TestUpdateMember(t *testing.T) {
	memberRepo := newMockMemberRepo()
	memberRepo.members[1] = &domain.Member{ID: 1, GroupID: 1, Name: "Test Member"}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewMemberUseCase(memberRepo, newMockGroupRepo(), auditRepo)

	newName := "New Name"
	err := uc.UpdateMember(context.Background(), 1, 1, "Admin", &newName, nil, nil, nil)

	assert.NoError(t, err)
	member, _ := memberRepo.GetByID(context.Background(), 1)
	assert.Equal(t, "New Name", member.Name)
}

func TestDeleteMember(t *testing.T) {
	memberRepo := newMockMemberRepo()
	memberRepo.members[1] = &domain.Member{ID: 1, GroupID: 1, Name: "Test Member"}
	auditRepo := &mockAuditRepo{}
	uc := usecase.NewMemberUseCase(memberRepo, newMockGroupRepo(), auditRepo)

	err := uc.DeleteMember(context.Background(), 1, 1, "Admin")

	assert.NoError(t, err)
	member, _ := memberRepo.GetByID(context.Background(), 1)
	assert.Nil(t, member)
}

func TestGetMemberCapacity(t *testing.T) {
	maxDaily := 10
	maxConcurrent := 5
	memberRepo := newMockMemberRepo()
	memberRepo.members[1] = &domain.Member{ID: 1, GroupID: 1, Name: "Test Member", MaxDailyAssignments: &maxDaily, MaxConcurrentOpen: &maxConcurrent, CurrentOpenAssignments: 2}
	memberRepo.assignmentCounts[1] = 5
	uc := usecase.NewMemberUseCase(memberRepo, nil, nil)

	capacity, err := uc.GetMemberCapacity(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, capacity)
	assert.Equal(t, 5, *capacity.DailyCapacityRemaining)
	assert.Equal(t, 3, *capacity.ConcurrentCapacityRemaining)
	assert.True(t, capacity.HasCapacity)
}
