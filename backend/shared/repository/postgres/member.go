package postgres

import (
	"context"
	"time"

	"github.com/raufhm/fairflow/shared/domain"
	"github.com/uptrace/bun"
)

type memberRepository struct {
	db *bun.DB
}

// NewMemberRepository creates a new member repository
func NewMemberRepository(db *bun.DB) domain.MemberRepository {
	return &memberRepository{db: db}
}

func (r *memberRepository) Create(ctx context.Context, member *domain.Member) error {
	now := time.Now()
	member.CreatedAt = now
	member.UpdatedAt = now
	// Bun automatically populates ID field after insert
	_, err := r.db.NewInsert().Model(member).Exec(ctx)
	return err
}

func (r *memberRepository) GetByID(ctx context.Context, id int64) (*domain.Member, error) {
	member := &domain.Member{}
	err := r.db.NewSelect().Model(member).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (r *memberRepository) GetByGroupID(ctx context.Context, groupID int64) ([]*domain.Member, error) {
	var members []*domain.Member
	err := r.db.NewSelect().
		Model(&members).
		Where("group_id = ?", groupID).
		Order("created_at").
		Scan(ctx)
	return members, err
}

func (r *memberRepository) GetActiveByGroupID(ctx context.Context, groupID int64) ([]*domain.Member, error) {
	var members []*domain.Member
	err := r.db.NewSelect().
		Model(&members).
		Where("group_id = ? AND active = true", groupID).
		Order("created_at").
		Scan(ctx)
	return members, err
}

func (r *memberRepository) Update(ctx context.Context, member *domain.Member) error {
	member.UpdatedAt = time.Now()
	_, err := r.db.NewUpdate().Model(member).Where("id = ?", member.ID).Exec(ctx)
	return err
}

func (r *memberRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model(&domain.Member{}).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *memberRepository) IncrementOpenAssignments(ctx context.Context, memberID int64) error {
	_, err := r.db.NewUpdate().
		Model(&domain.Member{}).
		Set("current_open_assignments = current_open_assignments + 1").
		Where("id = ?", memberID).
		Exec(ctx)
	return err
}

func (r *memberRepository) DecrementOpenAssignments(ctx context.Context, memberID int64) error {
	_, err := r.db.NewUpdate().
		Model(&domain.Member{}).
		Set("current_open_assignments = GREATEST(current_open_assignments - 1, 0)").
		Where("id = ?", memberID).
		Exec(ctx)
	return err
}

func (r *memberRepository) GetDailyAssignmentCount(ctx context.Context, memberID int64) (int, error) {
	// Get count of assignments created today
	count, err := r.db.NewSelect().
		Model(&domain.Assignment{}).
		Where("member_id = ?", memberID).
		Where("created_at >= CURRENT_DATE").
		Count(ctx)

	return count, err
}
