package usecase

import (
	"context"
	"errors"

	"github.com/raufhm/fairflow/shared/domain"
)

type MemberUseCase struct {
	memberRepo domain.MemberRepository
	groupRepo  domain.GroupRepository
}

func NewMemberUseCase(
	memberRepo domain.MemberRepository,
	groupRepo domain.GroupRepository,
) *MemberUseCase {
	return &MemberUseCase{
		memberRepo: memberRepo,
		groupRepo:  groupRepo,
	}
}

// CreateMember creates a new member in a group
func (uc *MemberUseCase) CreateMember(ctx context.Context, groupID, userID int64, userName, name string, email *string, weight int) (*domain.Member, error) {
	member := &domain.Member{
		GroupID: groupID,
		Name:    name,
		Email:   email,
		Weight:  weight,
		Active:  true,
	}

	if err := uc.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// GetMembers retrieves all members of a group
func (uc *MemberUseCase) GetMembers(ctx context.Context, groupID int64) ([]*domain.Member, error) {
	return uc.memberRepo.GetByGroupID(ctx, groupID)
}

// GetMember retrieves a member by ID
func (uc *MemberUseCase) GetMember(ctx context.Context, id int64) (*domain.Member, error) {
	return uc.memberRepo.GetByID(ctx, id)
}

// UpdateMember updates a member
func (uc *MemberUseCase) UpdateMember(ctx context.Context, id, userID int64, userName string, name *string, email *string, weight *int, active *bool) error {
	member, err := uc.memberRepo.GetByID(ctx, id)
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

	if err := uc.memberRepo.Update(ctx, member); err != nil {
		return err
	}

	return nil
}

// DeleteMember deletes a member
func (uc *MemberUseCase) DeleteMember(ctx context.Context, id, userID int64, userName string) error {
	member, err := uc.memberRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if member == nil {
		return errors.New("member not found")
	}

	if err := uc.memberRepo.Delete(ctx, id); err != nil {
		return err
	}

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
func (uc *MemberUseCase) GetMemberCapacity(ctx context.Context, memberID int64) (*CapacityStatus, error) {
	member, err := uc.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("member not found")
	}

	// Get daily assignment count
	dailyCount, err := uc.memberRepo.GetDailyAssignmentCount(ctx, memberID)
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
