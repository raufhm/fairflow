package usecase_test

import (
	"testing"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/usecase"
)

func TestWeightedRoundRobin(t *testing.T) {
	// Setup
	groupRepo := newMockGroupRepo()
	memberRepo := newMockMemberRepo()
	assignmentRepo := newMockAssignmentRepo()
	auditRepo := &mockAuditRepo{}

	// Create use case
	uc := usecase.NewAssignmentUseCase(groupRepo, memberRepo, assignmentRepo, auditRepo)

	// Create group
	group := &domain.Group{
		UserID:   1,
		Name:     "Test Group",
		Strategy: domain.StrategyWeightedRoundRobin,
		Active:   true,
	}
	groupRepo.Create(group)

	// Create members with different weights
	members := []*domain.Member{
		{GroupID: group.ID, Name: "Alice", Weight: 2, Active: true, Available: true},
		{GroupID: group.ID, Name: "Bob", Weight: 1, Active: true, Available: true},
		{GroupID: group.ID, Name: "Charlie", Weight: 3, Active: true, Available: true},
	}

	for _, m := range members {
		memberRepo.Create(m)
	}

	// Test: Get next assignee multiple times
	assignmentCounts := make(map[string]int)
	totalAssignments := 12 // 2+1+3 = 6, so 12 should distribute evenly

	for i := 0; i < totalAssignments; i++ {
		next, err := uc.CalculateNextAssignee(group.ID)
		if err != nil {
			t.Fatalf("Failed to get next assignee: %v", err)
		}
		if next == nil {
			t.Fatal("Expected assignee but got nil")
		}

		assignmentCounts[next.Name]++

		// Record assignment
		_, _, err = uc.RecordAssignment(group.ID, 1, &next.ID, nil)
		if err != nil {
			t.Fatalf("Failed to record assignment: %v", err)
		}
		memberRepo.assignmentCounts[next.ID]++
	}

	// Verify distribution matches weights
	// Alice (weight 2) should get 4 assignments (2/6 * 12)
	// Bob (weight 1) should get 2 assignments (1/6 * 12)
	// Charlie (weight 3) should get 6 assignments (3/6 * 12)

	if assignmentCounts["Alice"] != 4 {
		t.Errorf("Alice should have 4 assignments, got %d", assignmentCounts["Alice"])
	}
	if assignmentCounts["Bob"] != 2 {
		t.Errorf("Bob should have 2 assignments, got %d", assignmentCounts["Bob"])
	}
	if assignmentCounts["Charlie"] != 6 {
		t.Errorf("Charlie should have 6 assignments, got %d", assignmentCounts["Charlie"])
	}
}

func TestAvailabilityRespecting(t *testing.T) {
	// Setup
	groupRepo := newMockGroupRepo()
	memberRepo := newMockMemberRepo()
	assignmentRepo := newMockAssignmentRepo()
	auditRepo := &mockAuditRepo{}

	uc := usecase.NewAssignmentUseCase(groupRepo, memberRepo, assignmentRepo, auditRepo)

	// Create group
	group := &domain.Group{
		UserID:   1,
		Name:     "Test Group",
		Strategy: domain.StrategyWeightedRoundRobin,
		Active:   true,
	}
	groupRepo.Create(group)

	// Create members - one unavailable
	members := []*domain.Member{
		{GroupID: group.ID, Name: "Alice", Weight: 1, Active: true, Available: true},
		{GroupID: group.ID, Name: "Bob", Weight: 1, Active: true, Available: false}, // Unavailable
		{GroupID: group.ID, Name: "Charlie", Weight: 1, Active: true, Available: true},
	}

	for _, m := range members {
		memberRepo.Create(m)
	}

	// Test: Get next assignee - should only return available members
	for i := 0; i < 10; i++ {
		next, err := uc.CalculateNextAssignee(group.ID)
		if err != nil {
			t.Fatalf("Failed to get next assignee: %v", err)
		}

		if next.Name == "Bob" {
			t.Fatal("Bob should not be assigned (unavailable)")
		}

		_, _, err = uc.RecordAssignment(group.ID, 1, &next.ID, nil)
		if err != nil {
			t.Fatalf("Failed to record assignment: %v", err)
		}
		memberRepo.assignmentCounts[next.ID]++
	}
}

func TestGetAssignments(t *testing.T) {
	// Setup
	groupRepo := newMockGroupRepo()
	memberRepo := newMockMemberRepo()
	assignmentRepo := newMockAssignmentRepo()
	auditRepo := &mockAuditRepo{}

	uc := usecase.NewAssignmentUseCase(groupRepo, memberRepo, assignmentRepo, auditRepo)

	// Create group and member
	group := &domain.Group{UserID: 1, Name: "Test", Strategy: domain.StrategyWeightedRoundRobin, Active: true}
	groupRepo.Create(group)

	member := &domain.Member{GroupID: group.ID, Name: "Alice", Weight: 1, Active: true, Available: true}
	memberRepo.Create(member)

	// Record assignments
	for i := 0; i < 5; i++ {
		_, _, err := uc.RecordAssignment(group.ID, 1, &member.ID, nil)
		if err != nil {
			t.Fatalf("Failed to record assignment: %v", err)
		}
	}

	// Get assignments
	assignments, total, err := uc.GetAssignments(group.ID, 100, 0)
	if err != nil {
		t.Fatalf("Failed to get assignments: %v", err)
	}

	if len(assignments) != 5 {
		t.Errorf("Expected 5 assignments, got %d", len(assignments))
	}
	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
}
