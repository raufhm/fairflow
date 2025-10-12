package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/raufhm/fairflow/shared/domain"
)

type GroupUseCase struct {
	groupRepo  domain.GroupRepository
	memberRepo domain.MemberRepository
}

func NewGroupUseCase(
	groupRepo domain.GroupRepository,
	memberRepo domain.MemberRepository,
) *GroupUseCase {
	return &GroupUseCase{
		groupRepo:  groupRepo,
		memberRepo: memberRepo,
	}
}

// CreateGroup creates a new group
func (uc *GroupUseCase) CreateGroup(ctx context.Context, userID int64, userName, name string, description *string, strategy domain.AssignmentStrategy) (*domain.Group, error) {
	group := &domain.Group{
		UserID:      userID,
		Name:        name,
		Description: description,
		Strategy:    strategy,
		Active:      true,
	}

	if err := uc.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// GetGroup retrieves a group by ID
func (uc *GroupUseCase) GetGroup(ctx context.Context, id int64) (*domain.Group, error) {
	return uc.groupRepo.GetByID(ctx, id)
}

// GetAllGroups retrieves all groups
func (uc *GroupUseCase) GetAllGroups(ctx context.Context) ([]*domain.Group, error) {
	return uc.groupRepo.GetAll(ctx)
}

// GetUserGroups retrieves groups belonging to a user
func (uc *GroupUseCase) GetUserGroups(ctx context.Context, userID int64) ([]*domain.Group, error) {
	return uc.groupRepo.GetByUserID(ctx, userID)
}

// UpdateGroup updates a group
func (uc *GroupUseCase) UpdateGroup(ctx context.Context, id, userID int64, userName string, name *string, description *string, active *bool) (*domain.Group, error) {
	group, err := uc.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New("group not found")
	}

	updated := false
	if name != nil {
		group.Name = *name
		updated = true
	}
	if description != nil {
		group.Description = description
		updated = true
	}
	if active != nil {
		group.Active = *active
		updated = true
	}

	if !updated {
		return nil, errors.New("no valid fields provided for update")
	}

	if err := uc.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// DeleteGroup deletes a group
func (uc *GroupUseCase) DeleteGroup(ctx context.Context, id, userID int64, userName string) error {
	group, err := uc.groupRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("group not found")
	}

	if err := uc.groupRepo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

// CanModifyGroup checks if a user can modify a group
func (uc *GroupUseCase) CanModifyGroup(ctx context.Context, groupID, userID int64, userRole domain.UserRole) (bool, error) {
	group, err := uc.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return false, err
	}
	if group == nil {
		return false, errors.New("group not found")
	}

	// User is owner or admin/super_admin
	isOwner := group.UserID == userID
	isAdmin := userRole == domain.RoleAdmin || userRole == domain.RoleSuperAdmin

	return isOwner || isAdmin, nil
}

// PauseGroup pauses assignments for a group
func (uc *GroupUseCase) PauseGroup(ctx context.Context, groupID, userID int64, userName string, reason *string) error {
	group, err := uc.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("group not found")
	}

	if group.AssignmentPaused {
		return errors.New("group assignments are already paused")
	}

	// Update group to paused state
	now := time.Now()
	group.AssignmentPaused = true
	group.PauseReason = reason
	group.PausedAt = &now
	group.PausedBy = &userID

	if err := uc.groupRepo.Update(ctx, group); err != nil {
		return err
	}

	return nil
}

// ResumeGroup resumes assignments for a group
func (uc *GroupUseCase) ResumeGroup(ctx context.Context, groupID, userID int64, userName string) error {
	group, err := uc.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("group not found")
	}

	if !group.AssignmentPaused {
		return errors.New("group assignments are not paused")
	}

	// Update group to resumed state
	group.AssignmentPaused = false
	group.PauseReason = nil
	group.PausedAt = nil
	group.PausedBy = nil

	if err := uc.groupRepo.Update(ctx, group); err != nil {
		return err
	}

	return nil
}
