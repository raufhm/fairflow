package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/raufhm/fairflow/internal/domain"
)

type AssignmentUseCase struct {
	groupRepo      domain.GroupRepository
	memberRepo     domain.MemberRepository
	assignmentRepo domain.AssignmentRepository
	auditRepo      domain.AuditLogRepository
}

func NewAssignmentUseCase(
	groupRepo domain.GroupRepository,
	memberRepo domain.MemberRepository,
	assignmentRepo domain.AssignmentRepository,
	auditRepo domain.AuditLogRepository,
) *AssignmentUseCase {
	return &AssignmentUseCase{
		groupRepo:      groupRepo,
		memberRepo:     memberRepo,
		assignmentRepo: assignmentRepo,
		auditRepo:      auditRepo,
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
			continue // Skip this member, they're at max concurrent capacity
		}

		// Check daily assignment limit
		if member.MaxDailyAssignments != nil {
			dailyCount, err := uc.memberRepo.GetDailyAssignmentCount(ctx, member.ID)
			if err != nil {
				continue // Skip on error, don't break the whole flow
			}
			if dailyCount >= *member.MaxDailyAssignments {
				continue // Skip this member, they've reached daily limit
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
func (uc *AssignmentUseCase) RecordAssignment(ctx context.Context, groupID, userID int64, memberID *int64, metadata *string) (*domain.Member, int64, error) {
	var assignedMember *domain.Member
	var err error

	if memberID == nil {
		// Automatically determine next assignee
		assignedMember, err = uc.CalculateNextAssignee(ctx, groupID)
		if err != nil {
			return nil, 0, err
		}
	} else {
		// Validate provided member ID
		assignedMember, err = uc.memberRepo.GetByID(ctx, *memberID)
		if err != nil {
			return nil, 0, err
		}
		if assignedMember == nil || assignedMember.GroupID != groupID || !assignedMember.Active {
			return nil, 0, errors.New("invalid or inactive member ID provided")
		}
	}

	// Create assignment
	assignment := &domain.Assignment{
		GroupID:  groupID,
		MemberID: assignedMember.ID,
		Metadata: metadata,
	}

	if err := uc.assignmentRepo.Create(ctx, assignment); err != nil {
		return nil, 0, err
	}

	// Increment open assignments count
	if err := uc.memberRepo.IncrementOpenAssignments(ctx, assignedMember.ID); err != nil {
		// Log error but don't fail the assignment
		// TODO: Add proper logging
	}

	// Audit log
	group, _ := uc.groupRepo.GetByID(ctx, groupID)
	if group != nil {
		_ = uc.auditRepo.Create(ctx, &domain.AuditLog{
			UserID:       &userID,
			UserName:     "system", // Should be passed from context
			Action:       "Assignment recorded",
			ResourceType: stringPtr("assignment"),
			ResourceID:   &assignment.ID,
			Details:      stringPtr(fmt.Sprintf(`{"group":"%s","member":"%s"}`, group.Name, assignedMember.Name)),
			IPAddress:    "unknown",
		})
	}

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
	// Get all members
	members, err := uc.memberRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Get total assignments
	totalAssignments, err := uc.assignmentRepo.GetCountByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Get assignment counts per member
	memberIDs := make([]int64, len(members))
	for i, m := range members {
		memberIDs[i] = m.ID
	}

	counts, err := uc.assignmentRepo.GetCountsByMemberIDs(ctx, memberIDs)
	if err != nil {
		return nil, err
	}

	// Calculate total active weight
	totalActiveWeight := 0
	for _, m := range members {
		if m.Active {
			totalActiveWeight += m.Weight
		}
	}

	// Build distribution
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

// CompleteAssignment marks an assignment as completed and decrements open count
func (uc *AssignmentUseCase) CompleteAssignment(ctx context.Context, assignmentID int64) error {
	// Get the assignment
	assignment, err := uc.assignmentRepo.GetByID(ctx, assignmentID)
	if err != nil {
		return errors.New("assignment not found")
	}

	// Check if already completed
	if assignment.Status != domain.AssignmentStatusOpen {
		return errors.New("assignment is not open")
	}

	// Update status to completed
	if err := uc.assignmentRepo.UpdateStatus(ctx, assignmentID, domain.AssignmentStatusCompleted); err != nil {
		return err
	}

	// Decrement open assignments count
	if err := uc.memberRepo.DecrementOpenAssignments(ctx, assignment.MemberID); err != nil {
		// Log error but don't fail the completion
		// TODO: Add proper logging
	}

	return nil
}

// CancelAssignment marks an assignment as cancelled and decrements open count
func (uc *AssignmentUseCase) CancelAssignment(ctx context.Context, assignmentID int64) error {
	// Get the assignment
	assignment, err := uc.assignmentRepo.GetByID(ctx, assignmentID)
	if err != nil {
		return errors.New("assignment not found")
	}

	// Check if already completed/cancelled
	if assignment.Status != domain.AssignmentStatusOpen {
		return errors.New("assignment is not open")
	}

	// Update status to cancelled
	if err := uc.assignmentRepo.UpdateStatus(ctx, assignmentID, domain.AssignmentStatusCancelled); err != nil {
		return err
	}

	// Decrement open assignments count
	if err := uc.memberRepo.DecrementOpenAssignments(ctx, assignment.MemberID); err != nil {
		// Log error but don't fail the cancellation
		// TODO: Add proper logging
	}

	return nil
}
