package usecase_test

import (
	"testing"
	"time"

	"github.com/raufhm/rra/internal/domain"
	"github.com/raufhm/rra/internal/usecase"
)

// Mock repositories
type mockGroupRepo struct {
	groups map[int64]*domain.Group
}

func (m *mockGroupRepo) Create(group *domain.Group) error {
	group.ID = int64(len(m.groups) + 1)
	m.groups[group.ID] = group
	return nil
}

func (m *mockGroupRepo) GetByID(id int64) (*domain.Group, error) {
	if g, ok := m.groups[id]; ok {
		return g, nil
	}
	return nil, nil
}

func (m *mockGroupRepo) GetAll() ([]*domain.Group, error) {
	var groups []*domain.Group
	for _, g := range m.groups {
		groups = append(groups, g)
	}
	return groups, nil
}

func (m *mockGroupRepo) GetByUserID(userID int64) ([]*domain.Group, error) {
	var groups []*domain.Group
	for _, g := range m.groups {
		if g.UserID == userID {
			groups = append(groups, g)
		}
	}
	return groups, nil
}

func (m *mockGroupRepo) Update(group *domain.Group) error {
	m.groups[group.ID] = group
	return nil
}

func (m *mockGroupRepo) Delete(id int64) error {
	delete(m.groups, id)
	return nil
}

type mockMemberRepo struct {
	members        map[int64]*domain.Member
	nextID         int64
	assignmentCounts map[int64]int
}

func (m *mockMemberRepo) Create(member *domain.Member) error {
	m.nextID++
	member.ID = m.nextID
	m.members[member.ID] = member
	m.assignmentCounts[member.ID] = 0
	return nil
}

func (m *mockMemberRepo) GetByID(id int64) (*domain.Member, error) {
	if member, ok := m.members[id]; ok {
		memberCopy := *member
		memberCopy.Assignments = m.assignmentCounts[id]
		return &memberCopy, nil
	}
	return nil, nil
}

func (r *mockMemberRepo) GetByGroupID(groupID int64) ([]*domain.Member, error) {
	var members []*domain.Member
	for _, m := range r.members {
		if m.GroupID == groupID {
			memberCopy := *m
			memberCopy.Assignments = r.assignmentCounts[m.ID]
			members = append(members, &memberCopy)
		}
	}
	return members, nil
}

func (r *mockMemberRepo) GetActiveByGroupID(groupID int64) ([]*domain.Member, error) {
	var members []*domain.Member
	for _, member := range r.members {
		if member.GroupID == groupID && member.Active && member.Available {
			memberCopy := *member
			members = append(members, &memberCopy)
		}
	}
	return members, nil
}

func (m *mockMemberRepo) Update(member *domain.Member) error {
	m.members[member.ID] = member
	return nil
}

func (m *mockMemberRepo) Delete(id int64) error {
	delete(m.members, id)
	delete(m.assignmentCounts, id)
	return nil
}

type mockAssignmentRepo struct {
	assignments []*domain.Assignment
}

func (m *mockAssignmentRepo) Create(assignment *domain.Assignment) error {
	assignment.ID = int64(len(m.assignments) + 1)
	assignment.CreatedAt = time.Now()
	m.assignments = append(m.assignments, assignment)
	return nil
}

func (m *mockAssignmentRepo) GetByGroupID(groupID int64, limit, offset int) ([]*domain.AssignmentWithMember, error) {
	var assignments []*domain.AssignmentWithMember
	for _, a := range m.assignments {
		if a.GroupID == groupID {
			assignments = append(assignments, &domain.AssignmentWithMember{
				ID:         a.ID,
				MemberID:   a.MemberID,
				MemberName: "Member",
				Metadata:   a.Metadata,
				CreatedAt:  a.CreatedAt,
			})
		}
	}
	return assignments, nil
}

func (m *mockAssignmentRepo) GetCountByGroupID(groupID int64) (int, error) {
	count := 0
	for _, a := range m.assignments {
		if a.GroupID == groupID {
			count++
		}
	}
	return count, nil
}

func (m *mockAssignmentRepo) GetCountsByMemberIDs(memberIDs []int64) (map[int64]int, error) {
	counts := make(map[int64]int)
	for _, id := range memberIDs {
		counts[id] = 0
	}
	for _, a := range m.assignments {
		if _, ok := counts[a.MemberID]; ok {
			counts[a.MemberID]++
		}
	}
	return counts, nil
}

type mockAuditRepo struct{}

func (m *mockAuditRepo) Create(log *domain.AuditLog) error { return nil }
func (m *mockAuditRepo) GetRecent(limit int) ([]*domain.AuditLog, error) { return nil, nil }

func TestWeightedRoundRobin(t *testing.T) {
	// Setup
	groupRepo := &mockGroupRepo{groups: make(map[int64]*domain.Group)}
	memberRepo := &mockMemberRepo{
		members: make(map[int64]*domain.Member),
		assignmentCounts: make(map[int64]int),
	}
	assignmentRepo := &mockAssignmentRepo{}
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
	groupRepo := &mockGroupRepo{groups: make(map[int64]*domain.Group)}
	memberRepo := &mockMemberRepo{
		members: make(map[int64]*domain.Member),
		assignmentCounts: make(map[int64]int),
	}
	assignmentRepo := &mockAssignmentRepo{}
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
	groupRepo := &mockGroupRepo{groups: make(map[int64]*domain.Group)}
	memberRepo := &mockMemberRepo{
		members: make(map[int64]*domain.Member),
		assignmentCounts: make(map[int64]int),
	}
	assignmentRepo := &mockAssignmentRepo{}
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
