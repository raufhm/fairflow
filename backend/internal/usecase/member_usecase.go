package usecase

import (
	"errors"
	"fmt"

	"github.com/raufhm/fairflow/internal/domain"
)

type MemberUseCase struct {
	memberRepo domain.MemberRepository
	groupRepo  domain.GroupRepository
	auditRepo  domain.AuditLogRepository
}

func NewMemberUseCase(
	memberRepo domain.MemberRepository,
	groupRepo domain.GroupRepository,
	auditRepo domain.AuditLogRepository,
) *MemberUseCase {
	return &MemberUseCase{
		memberRepo: memberRepo,
		groupRepo:  groupRepo,
		auditRepo:  auditRepo,
	}
}

// CreateMember creates a new member in a group
func (uc *MemberUseCase) CreateMember(groupID, userID int64, userName, name string, email *string, weight int) (*domain.Member, error) {
	member := &domain.Member{
		GroupID: groupID,
		Name:    name,
		Email:   email,
		Weight:  weight,
		Active:  true,
	}

	if err := uc.memberRepo.Create(member); err != nil {
		return nil, err
	}

	// Audit log
	group, _ := uc.groupRepo.GetByID(groupID)
	groupName := "unknown"
	if group != nil {
		groupName = group.Name
	}

	_ = uc.auditRepo.Create(&domain.AuditLog{
		UserID:       &userID,
		UserName:     userName,
		Action:       "Member added",
		ResourceType: stringPtr("member"),
		ResourceID:   &member.ID,
		Details:      stringPtr(fmt.Sprintf(`{"group":"%s","name":"%s"}`, groupName, name)),
		IPAddress:    "unknown",
	})

	return member, nil
}

// GetMembers retrieves all members of a group
func (uc *MemberUseCase) GetMembers(groupID int64) ([]*domain.Member, error) {
	return uc.memberRepo.GetByGroupID(groupID)
}

// GetMember retrieves a member by ID
func (uc *MemberUseCase) GetMember(id int64) (*domain.Member, error) {
	return uc.memberRepo.GetByID(id)
}

// UpdateMember updates a member
func (uc *MemberUseCase) UpdateMember(id, userID int64, userName string, name *string, email *string, weight *int, active *bool) error {
	member, err := uc.memberRepo.GetByID(id)
	if err != nil {
		return err
	}
	if member == nil {
		return errors.New("member not found")
	}

	updated := false
	if name != nil {
		member.Name = *name
		updated = true
	}
	if email != nil {
		member.Email = email
		updated = true
	}
	if weight != nil {
		member.Weight = *weight
		updated = true
	}
	if active != nil {
		member.Active = *active
		updated = true
	}

	if !updated {
		return errors.New("no valid fields provided for update")
	}

	if err := uc.memberRepo.Update(member); err != nil {
		return err
	}

	// Audit log
	group, _ := uc.groupRepo.GetByID(member.GroupID)
	groupName := "unknown"
	if group != nil {
		groupName = group.Name
	}

	_ = uc.auditRepo.Create(&domain.AuditLog{
		UserID:       &userID,
		UserName:     userName,
		Action:       "Member updated",
		ResourceType: stringPtr("member"),
		ResourceID:   &id,
		Details:      stringPtr(fmt.Sprintf(`{"group":"%s"}`, groupName)),
		IPAddress:    "unknown",
	})

	return nil
}

// DeleteMember deletes a member
func (uc *MemberUseCase) DeleteMember(id, userID int64, userName string) error {
	member, err := uc.memberRepo.GetByID(id)
	if err != nil {
		return err
	}
	if member == nil {
		return errors.New("member not found")
	}

	groupID := member.GroupID

	if err := uc.memberRepo.Delete(id); err != nil {
		return err
	}

	// Audit log
	group, _ := uc.groupRepo.GetByID(groupID)
	groupName := "unknown"
	if group != nil {
		groupName = group.Name
	}

	_ = uc.auditRepo.Create(&domain.AuditLog{
		UserID:       &userID,
		UserName:     userName,
		Action:       "Member removed",
		ResourceType: stringPtr("member"),
		ResourceID:   &id,
		Details:      stringPtr(fmt.Sprintf(`{"group":"%s"}`, groupName)),
		IPAddress:    "unknown",
	})

	return nil
}

// CapacityStatus represents the current capacity status of a member
type CapacityStatus struct {
	MemberID                    int64  `json:"member_id"`
	Name                        string `json:"name"`
	MaxDailyAssignments         *int   `json:"max_daily_assignments,omitempty"`
	DailyAssignments            int    `json:"daily_assignments"`
	DailyCapacityRemaining      *int   `json:"daily_capacity_remaining,omitempty"`
	MaxConcurrentOpen           *int   `json:"max_concurrent_open,omitempty"`
	CurrentOpenAssignments      int    `json:"current_open_assignments"`
	ConcurrentCapacityRemaining *int   `json:"concurrent_capacity_remaining,omitempty"`
	HasCapacity                 bool   `json:"has_capacity"`
}

// GetMemberCapacity returns the capacity status of a member
func (uc *MemberUseCase) GetMemberCapacity(memberID int64) (*CapacityStatus, error) {
	member, err := uc.memberRepo.GetByID(memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("member not found")
	}

	// Get daily assignment count
	dailyCount, err := uc.memberRepo.GetDailyAssignmentCount(memberID)
	if err != nil {
		return nil, err
	}

	status := &CapacityStatus{
		MemberID:               member.ID,
		Name:                   member.Name,
		MaxDailyAssignments:    member.MaxDailyAssignments,
		DailyAssignments:       dailyCount,
		MaxConcurrentOpen:      member.MaxConcurrentOpen,
		CurrentOpenAssignments: member.CurrentOpenAssignments,
		HasCapacity:            true,
	}

	// Calculate remaining capacity
	if member.MaxDailyAssignments != nil {
		remaining := *member.MaxDailyAssignments - dailyCount
		if remaining < 0 {
			remaining = 0
		}
		status.DailyCapacityRemaining = &remaining
		if dailyCount >= *member.MaxDailyAssignments {
			status.HasCapacity = false
		}
	}

	if member.MaxConcurrentOpen != nil {
		remaining := *member.MaxConcurrentOpen - member.CurrentOpenAssignments
		if remaining < 0 {
			remaining = 0
		}
		status.ConcurrentCapacityRemaining = &remaining
		if member.CurrentOpenAssignments >= *member.MaxConcurrentOpen {
			status.HasCapacity = false
		}
	}

	return status, nil
}
