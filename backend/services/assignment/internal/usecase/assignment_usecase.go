package usecase

import (
	"context"
	"errors"
	"math"

	"github.com/raufhm/fairflow/shared/domain"
)

type AssignmentUseCase struct {
	groupRepo      domain.GroupRepository
	memberRepo     domain.MemberRepository
	assignmentRepo domain.AssignmentRepository
}

func NewAssignmentUseCase(
	groupRepo domain.GroupRepository,
	memberRepo domain.MemberRepository,
	assignmentRepo domain.AssignmentRepository,
) *AssignmentUseCase {
	return &AssignmentUseCase{
		groupRepo:      groupRepo,
		memberRepo:     memberRepo,
		assignmentRepo: assignmentRepo,
	}
}

// CalculateNextAssignee calculates the next assignee using weighted round robin
func (uc *AssignmentUseCase) CalculateNextAssignee(ctx context.Context, groupID int64) (*domain.Member, error) {
	// Check if group is paused
	group, err := uc.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New("group not found")
	}
	if group.AssignmentPaused {
		return nil, errors.New("assignments are paused for this group")
	}

	// Get active members
	members, err := uc.memberRepo.GetActiveByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, errors.New("no active members available for assignment")
	}

	// Filter members based on capacity limits
	eligibleMembers := []*domain.Member{}
	for _, member := range members {
		// Check concurrent open assignments limit
		if member.MaxConcurrentOpen != nil && member.CurrentOpenAssignments >= *member.MaxConcurrentOpen {
			continue
		}

		// Check daily assignment limit
		if member.MaxDailyAssignments != nil {
			dailyCount, err := uc.memberRepo.GetDailyAssignmentCount(ctx, member.ID)
			if err != nil {
				continue
			}
			if dailyCount >= *member.MaxDailyAssignments {
				continue
			}
		}

		eligibleMembers = append(eligibleMembers, member)
	}

	if len(eligibleMembers) == 0 {
		return nil, errors.New("no members available with capacity for assignment")
	}

	// Get member IDs for assignment count query
	memberIDs := make([]int64, len(eligibleMembers))
	for i, m := range eligibleMembers {
		memberIDs[i] = m.ID
	}

	// Get assignment counts
	counts, err := uc.assignmentRepo.GetCountsByMemberIDs(ctx, memberIDs)
	if err != nil {
		return nil, err
	}

	// Calculate weighted round robin
	var lowestRatio float64 = math.MaxFloat64
	var nextAssignee *domain.Member

	for _, member := range eligibleMembers {
		actual := float64(counts[member.ID])
		expected := float64(member.Weight) / 100.0

		var ratio float64
		if expected > 0 {
			ratio = actual / expected
		} else {
			ratio = actual
		}

		if ratio < lowestRatio {
			lowestRatio = ratio
			nextAssignee = member
		}
	}

	return nextAssignee, nil
}

// RecordAssignment creates a new assignment record
func (uc *AssignmentUseCase) RecordAssignment(ctx context.Context, groupID, userID int64, userName string, memberID *int64, metadata *string) (*domain.Member, int64, error) {
	var assignedMember *domain.Member
	var err error

	if memberID == nil {
		assignedMember, err = uc.CalculateNextAssignee(ctx, groupID)
		if err != nil {
			return nil, 0, err
		}
	} else {
		assignedMember, err = uc.memberRepo.GetByID(ctx, *memberID)
		if err != nil {
			return nil, 0, err
		}
		if assignedMember == nil || assignedMember.GroupID != groupID || !assignedMember.Active {
			return nil, 0, errors.New("invalid or inactive member ID provided")
		}
	}

	assignment := &domain.Assignment{
		GroupID:  groupID,
		MemberID: assignedMember.ID,
		Metadata: metadata,
	}

	if err := uc.assignmentRepo.Create(ctx, assignment); err != nil {
		return nil, 0, err
	}

	_ = uc.memberRepo.IncrementOpenAssignments(ctx, assignedMember.ID)

	return assignedMember, assignment.ID, nil
}

// GetAssignments retrieves assignments for a group with pagination
func (uc *AssignmentUseCase) GetAssignments(ctx context.Context, groupID int64, limit, offset int) ([]*domain.AssignmentWithMember, int, error) {
	assignments, err := uc.assignmentRepo.GetByGroupID(ctx, groupID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := uc.assignmentRepo.GetCountByGroupID(ctx, groupID)
	if err != nil {
		return nil, 0, err
	}

	return assignments, total, nil
}

// GetStats calculates assignment statistics for a group
func (uc *AssignmentUseCase) GetStats(ctx context.Context, groupID int64) (*domain.AssignmentStats, error) {
	members, err := uc.memberRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	totalAssignments, err := uc.assignmentRepo.GetCountByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	memberIDs := make([]int64, len(members))
	for i, m := range members {
		memberIDs[i] = m.ID
	}

	counts, err := uc.assignmentRepo.GetCountsByMemberIDs(ctx, memberIDs)
	if err != nil {
		return nil, err
	}

	totalActiveWeight := 0
	for _, m := range members {
		if m.Active {
			totalActiveWeight += m.Weight
		}
	}

	distribution := make([]domain.MemberDistribution, 0, len(members))
	for _, member := range members {
		actualAssignments := counts[member.ID]
		var expectedAssignments float64
		var variance float64

		if member.Active && totalActiveWeight > 0 {
			share := float64(member.Weight) / float64(totalActiveWeight)
			expectedAssignments = share * float64(totalAssignments)
			variance = float64(actualAssignments) - expectedAssignments
		}

		distribution = append(distribution, domain.MemberDistribution{
			MemberID:    member.ID,
			Name:        member.Name,
			Weight:      member.Weight,
			Assignments: actualAssignments,
			Expected:    math.Round(expectedAssignments*100) / 100,
			Variance:    math.Round(variance*100) / 100,
		})
	}

	return &domain.AssignmentStats{
		TotalAssignments: totalAssignments,
		Distribution:     distribution,
	}, nil
}
