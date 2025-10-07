package postgres

import (
	"context"
	"time"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/uptrace/bun"
)

type assignmentRepository struct {
	db *bun.DB
}

// NewAssignmentRepository creates a new assignment repository
func NewAssignmentRepository(db *bun.DB) domain.AssignmentRepository {
	return &assignmentRepository{db: db}
}

func (r *assignmentRepository) Create(assignment *domain.Assignment) error {
	ctx := context.Background()
	assignment.CreatedAt = time.Now()
	// Set default status to open if not specified
	if assignment.Status == "" {
		assignment.Status = domain.AssignmentStatusOpen
	}
	_, err := r.db.NewInsert().Model(assignment).Exec(ctx)
	return err
}

func (r *assignmentRepository) GetByID(id int64) (*domain.Assignment, error) {
	ctx := context.Background()
	assignment := &domain.Assignment{}
	err := r.db.NewSelect().Model(assignment).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return assignment, nil
}

func (r *assignmentRepository) UpdateStatus(id int64, status domain.AssignmentStatus) error {
	ctx := context.Background()
	now := time.Now()
	update := r.db.NewUpdate().
		Model(&domain.Assignment{}).
		Set("status = ?", status).
		Where("id = ?", id)

	// Set completed_at timestamp if status is completed or cancelled
	if status == domain.AssignmentStatusCompleted || status == domain.AssignmentStatusCancelled {
		update = update.Set("completed_at = ?", now)
	}

	_, err := update.Exec(ctx)
	return err
}

func (r *assignmentRepository) GetByGroupID(groupID int64, limit, offset int) ([]*domain.AssignmentWithMember, error) {
	ctx := context.Background()
	var assignments []*domain.AssignmentWithMember
	err := r.db.NewSelect().
		ColumnExpr("a.id, a.metadata, a.created_at, m.id as member_id, m.name as member_name").
		TableExpr("assignments AS a").
		Join("JOIN members AS m ON a.member_id = m.id").
		Where("a.group_id = ?", groupID).
		Order("a.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx, &assignments)
	return assignments, err
}

func (r *assignmentRepository) GetCountByGroupID(groupID int64) (int, error) {
	ctx := context.Background()
	count, err := r.db.NewSelect().
		Model(&domain.Assignment{}).
		Where("group_id = ?", groupID).
		Count(ctx)
	return count, err
}

func (r *assignmentRepository) GetCountsByMemberIDs(memberIDs []int64) (map[int64]int, error) {
	if len(memberIDs) == 0 {
		return make(map[int64]int), nil
	}

	ctx := context.Background()
	var results []struct {
		MemberID int64 `bun:"member_id"`
		Count    int   `bun:"count"`
	}

	err := r.db.NewSelect().
		ColumnExpr("member_id, COUNT(id) as count").
		TableExpr("assignments").
		Where("member_id IN (?)", bun.In(memberIDs)).
		Group("member_id").
		Scan(ctx, &results)

	if err != nil {
		return nil, err
	}

	counts := make(map[int64]int)
	for _, result := range results {
		counts[result.MemberID] = result.Count
	}

	return counts, nil
}
