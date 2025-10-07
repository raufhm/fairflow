package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/raufhm/fairflow/internal/domain"
)

type GroupUseCase struct {
	groupRepo  domain.GroupRepository
	memberRepo domain.MemberRepository
	auditRepo  domain.AuditLogRepository
}

func NewGroupUseCase(
	groupRepo domain.GroupRepository,
	memberRepo domain.MemberRepository,
	auditRepo domain.AuditLogRepository,
) *GroupUseCase {
	return &GroupUseCase{
		groupRepo:  groupRepo,
		memberRepo: memberRepo,
		auditRepo:  auditRepo,
	}
}

// CreateGroup creates a new group
func (uc *GroupUseCase) CreateGroup(userID int64, userName, name string, description *string, strategy domain.AssignmentStrategy) (*domain.Group, error) {
	group := &domain.Group{
		UserID:      userID,
		Name:        name,
		Description: description,
		Strategy:    strategy,
		Active:      true,
	}

	if err := uc.groupRepo.Create(group); err != nil {
		return nil, err
	}

	// Audit log
	_ = uc.auditRepo.Create(&domain.AuditLog{
		UserID:       &userID,
		UserName:     userName,
		Action:       "Group created",
		ResourceType: stringPtr("group"),
		ResourceID:   &group.ID,
		Details:      stringPtr(fmt.Sprintf(`{"name":"%s"}`, name)),
		IPAddress:    "unknown",
	})

	return group, nil
}

// GetGroup retrieves a group by ID
func (uc *GroupUseCase) GetGroup(id int64) (*domain.Group, error) {
	return uc.groupRepo.GetByID(id)
}

// GetAllGroups retrieves all groups
func (uc *GroupUseCase) GetAllGroups() ([]*domain.Group, error) {
	return uc.groupRepo.GetAll()
}

// GetUserGroups retrieves groups belonging to a user
func (uc *GroupUseCase) GetUserGroups(userID int64) ([]*domain.Group, error) {
	return uc.groupRepo.GetByUserID(userID)
}

// UpdateGroup updates a group
func (uc *GroupUseCase) UpdateGroup(id, userID int64, userName string, name *string, description *string, active *bool) (*domain.Group, error) {
	group, err := uc.groupRepo.GetByID(id)
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

	if err := uc.groupRepo.Update(group); err != nil {
		return nil, err
	}

	// Audit log
	_ = uc.auditRepo.Create(&domain.AuditLog{
		UserID:       &userID,
		UserName:     userName,
		Action:       "Group updated",
		ResourceType: stringPtr("group"),
		ResourceID:   &id,
		Details:      stringPtr(fmt.Sprintf(`{"name":"%s"}`, group.Name)),
		IPAddress:    "unknown",
	})

	return group, nil
}

// DeleteGroup deletes a group
func (uc *GroupUseCase) DeleteGroup(id, userID int64, userName string) error {
	group, err := uc.groupRepo.GetByID(id)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("group not found")
	}

	if err := uc.groupRepo.Delete(id); err != nil {
		return err
	}

	// Audit log
	_ = uc.auditRepo.Create(&domain.AuditLog{
		UserID:       &userID,
		UserName:     userName,
		Action:       "Group deleted",
		ResourceType: stringPtr("group"),
		ResourceID:   &id,
		Details:      stringPtr(fmt.Sprintf(`{"groupName":"%s"}`, group.Name)),
		IPAddress:    "unknown",
	})

	return nil
}

// CanModifyGroup checks if a user can modify a group
func (uc *GroupUseCase) CanModifyGroup(groupID, userID int64, userRole domain.UserRole) (bool, error) {
	group, err := uc.groupRepo.GetByID(groupID)
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
func (uc *GroupUseCase) PauseGroup(groupID, userID int64, userName string, reason *string) error {
	group, err := uc.groupRepo.GetByID(groupID)
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

	if err := uc.groupRepo.Update(group); err != nil {
		return err
	}

	// Audit log
	reasonStr := "No reason provided"
	if reason != nil {
		reasonStr = *reason
	}
	_ = uc.auditRepo.Create(&domain.AuditLog{
		UserID:       &userID,
		UserName:     userName,
		Action:       "Group assignments paused",
		ResourceType: stringPtr("group"),
		ResourceID:   &groupID,
		Details:      stringPtr(fmt.Sprintf(`{"name":"%s","reason":"%s"}`, group.Name, reasonStr)),
		IPAddress:    "unknown",
	})

	return nil
}

// ResumeGroup resumes assignments for a group
func (uc *GroupUseCase) ResumeGroup(groupID, userID int64, userName string) error {
	group, err := uc.groupRepo.GetByID(groupID)
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

	if err := uc.groupRepo.Update(group); err != nil {
		return err
	}

	// Audit log
	_ = uc.auditRepo.Create(&domain.AuditLog{
		UserID:       &userID,
		UserName:     userName,
		Action:       "Group assignments resumed",
		ResourceType: stringPtr("group"),
		ResourceID:   &groupID,
		Details:      stringPtr(fmt.Sprintf(`{"name":"%s"}`, group.Name)),
		IPAddress:    "unknown",
	})

	return nil
}
