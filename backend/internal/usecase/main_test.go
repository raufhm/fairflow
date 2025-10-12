package usecase_test

import (
	"context"
	"time"

	"github.com/raufhm/fairflow/internal/domain"
)

// Mock repositories
type mockGroupRepo struct {
	groups map[int64]*domain.Group
}

func newMockGroupRepo() *mockGroupRepo {
	return &mockGroupRepo{groups: make(map[int64]*domain.Group)}
}

func (m *mockGroupRepo) Create(ctx context.Context, group *domain.Group) error {
	group.ID = int64(len(m.groups) + 1)
	m.groups[group.ID] = group
	return nil
}

func (m *mockGroupRepo) GetByID(ctx context.Context, id int64) (*domain.Group, error) {
	if g, ok := m.groups[id]; ok {
		return g, nil
	}
	return nil, nil
}

func (m *mockGroupRepo) GetAll(ctx context.Context) ([]*domain.Group, error) {
	var groups []*domain.Group
	for _, g := range m.groups {
		groups = append(groups, g)
	}
	return groups, nil
}

func (m *mockGroupRepo) GetByUserID(ctx context.Context, userID int64) ([]*domain.Group, error) {
	var groups []*domain.Group
	for _, g := range m.groups {
		if g.UserID == userID {
			groups = append(groups, g)
		}
	}
	return groups, nil
}

func (m *mockGroupRepo) Update(ctx context.Context, group *domain.Group) error {
	m.groups[group.ID] = group
	return nil
}

func (m *mockGroupRepo) Delete(ctx context.Context, id int64) error {
	delete(m.groups, id)
	return nil
}

type mockMemberRepo struct {
	members          map[int64]*domain.Member
	nextID           int64
	assignmentCounts map[int64]int
}

func newMockMemberRepo() *mockMemberRepo {
	return &mockMemberRepo{members: make(map[int64]*domain.Member), assignmentCounts: make(map[int64]int)}
}

func (m *mockMemberRepo) Create(ctx context.Context, member *domain.Member) error {
	m.nextID++
	member.ID = m.nextID
	m.members[member.ID] = member
	m.assignmentCounts[member.ID] = 0
	return nil
}

func (m *mockMemberRepo) GetByID(ctx context.Context, id int64) (*domain.Member, error) {
	if member, ok := m.members[id]; ok {
		memberCopy := *member
		memberCopy.Assignments = m.assignmentCounts[id]
		return &memberCopy, nil
	}
	return nil, nil
}

func (r *mockMemberRepo) GetByGroupID(ctx context.Context, groupID int64) ([]*domain.Member, error) {
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

func (r *mockMemberRepo) GetActiveByGroupID(ctx context.Context, groupID int64) ([]*domain.Member, error) {
	var members []*domain.Member
	for _, member := range r.members {
		if member.GroupID == groupID && member.Active && member.Available {
			memberCopy := *member
			members = append(members, &memberCopy)
		}
	}
	return members, nil
}

func (m *mockMemberRepo) Update(ctx context.Context, member *domain.Member) error {
	m.members[member.ID] = member
	return nil
}

func (m *mockMemberRepo) Delete(ctx context.Context, id int64) error {
	delete(m.members, id)
	delete(m.assignmentCounts, id)
	return nil
}

func (m *mockMemberRepo) DecrementOpenAssignments(ctx context.Context, memberID int64) error {
	// A simple mock implementation. In a real scenario, you might want to do more here.
	return nil
}
func (m *mockMemberRepo) IncrementOpenAssignments(ctx context.Context, memberID int64) error {
	// A simple mock implementation. In a real scenario, you might want to do more here.
	return nil
}
func (m *mockMemberRepo) GetDailyAssignmentCount(ctx context.Context, memberID int64) (int, error) {
	return m.assignmentCounts[memberID], nil
}

type mockAssignmentRepo struct {
	assignments []*domain.Assignment
}

func newMockAssignmentRepo() *mockAssignmentRepo {
	return &mockAssignmentRepo{assignments: make([]*domain.Assignment, 0)}
}

func (m *mockAssignmentRepo) Create(ctx context.Context, assignment *domain.Assignment) error {
	assignment.ID = int64(len(m.assignments) + 1)
	assignment.CreatedAt = time.Now()
	m.assignments = append(m.assignments, assignment)
	return nil
}

func (m *mockAssignmentRepo) GetByGroupID(ctx context.Context, groupID int64, limit, offset int) ([]*domain.AssignmentWithMember, error) {
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

func (m *mockAssignmentRepo) GetCountByGroupID(ctx context.Context, groupID int64) (int, error) {
	count := 0
	for _, a := range m.assignments {
		if a.GroupID == groupID {
			count++
		}
	}
	return count, nil
}

func (m *mockAssignmentRepo) GetCountsByMemberIDs(ctx context.Context, memberIDs []int64) (map[int64]int, error) {
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

func (m *mockAssignmentRepo) GetByID(ctx context.Context, id int64) (*domain.Assignment, error) {
	for _, a := range m.assignments {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockAssignmentRepo) UpdateStatus(ctx context.Context, id int64, status domain.AssignmentStatus) error {
	for _, a := range m.assignments {
		if a.ID == id {
			a.Status = status
			return nil
		}
	}
	return nil
}

type mockAuditRepo struct{}

func (m *mockAuditRepo) Create(ctx context.Context, log *domain.AuditLog) error { return nil }
func (m *mockAuditRepo) GetRecent(ctx context.Context, limit int) ([]*domain.AuditLog, error) {
	return nil, nil
}

type mockUserRepo struct {
	users map[int64]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[int64]*domain.User)}
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	user.ID = int64(len(m.users) + 1)
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetAll(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, nil
}

func (m *mockUserRepo) UpdateRole(ctx context.Context, id int64, role domain.UserRole) error {
	if u, ok := m.users[id]; ok {
		u.Role = role
		return nil
	}
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id int64) error {
	delete(m.users, id)
	return nil
}

type mockAPIKeyRepo struct {
	keys map[int64]*domain.APIKey
}

func newMockAPIKeyRepo() *mockAPIKeyRepo {
	return &mockAPIKeyRepo{keys: make(map[int64]*domain.APIKey)}
}

func (m *mockAPIKeyRepo) Create(ctx context.Context, key *domain.APIKey) error {
	key.ID = int64(len(m.keys) + 1)
	m.keys[key.ID] = key
	return nil
}

func (m *mockAPIKeyRepo) GetByUserID(ctx context.Context, userID int64) ([]*domain.APIKey, error) {
	var keys []*domain.APIKey
	for _, k := range m.keys {
		if k.UserID == userID {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (m *mockAPIKeyRepo) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	for _, k := range m.keys {
		if k.KeyHash == hash {
			return k, nil
		}
	}
	return nil, nil
}

func (m *mockAPIKeyRepo) Delete(ctx context.Context, id int64) error {
	delete(m.keys, id)
	return nil
}

func (m *mockAPIKeyRepo) UpdateLastUsed(ctx context.Context, id int64) error {
	return nil
}

func (m *mockAPIKeyRepo) GetByID(ctx context.Context, id int64) (*domain.APIKey, error) {
	if k, ok := m.keys[id]; ok {
		return k, nil
	}
	return nil, nil
}

type mockWebhookRepo struct {
	webhooks map[int64]*domain.Webhook
}

func newMockWebhookRepo() *mockWebhookRepo {
	return &mockWebhookRepo{webhooks: make(map[int64]*domain.Webhook)}
}

func (m *mockWebhookRepo) Create(ctx context.Context, webhook *domain.Webhook) error {
	webhook.ID = int64(len(m.webhooks) + 1)
	m.webhooks[webhook.ID] = webhook
	return nil
}

func (m *mockWebhookRepo) GetByGroupID(ctx context.Context, groupID int64) ([]*domain.Webhook, error) {
	var webhooks []*domain.Webhook
	for _, w := range m.webhooks {
		if w.GroupID == groupID {
			webhooks = append(webhooks, w)
		}
	}
	return webhooks, nil
}

func (m *mockWebhookRepo) Update(ctx context.Context, webhook *domain.Webhook) error {
	m.webhooks[webhook.ID] = webhook
	return nil
}

func (m *mockWebhookRepo) Delete(ctx context.Context, id int64) error {
	delete(m.webhooks, id)
	return nil
}

func (m *mockWebhookRepo) GetActiveByGroupID(ctx context.Context, groupID int64) ([]*domain.Webhook, error) {
	var webhooks []*domain.Webhook
	for _, w := range m.webhooks {
		if w.GroupID == groupID && w.Active {
			webhooks = append(webhooks, w)
		}
	}
	return webhooks, nil
}

func (m *mockWebhookRepo) GetByID(ctx context.Context, id int64) (*domain.Webhook, error) {
	if w, ok := m.webhooks[id]; ok {
		return w, nil
	}
	return nil, nil
}
