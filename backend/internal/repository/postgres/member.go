package postgres

import (
	"context"
	"time"

	"github.com/raufhm/rra/internal/domain"
	"github.com/uptrace/bun"
)

type memberRepository struct {
	db *bun.DB
}

// NewMemberRepository creates a new member repository
func NewMemberRepository(db *bun.DB) domain.MemberRepository {
	return &memberRepository{db: db}
}

func (r *memberRepository) Create(member *domain.Member) error {
	ctx := context.Background()
	now := time.Now()
	member.CreatedAt = now
	member.UpdatedAt = now
	// Bun automatically populates ID field after insert
	_, err := r.db.NewInsert().Model(member).Exec(ctx)
	return err
}

func (r *memberRepository) GetByID(id int64) (*domain.Member, error) {
	ctx := context.Background()
	member := &domain.Member{}
	err := r.db.NewSelect().Model(member).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (r *memberRepository) GetByGroupID(groupID int64) ([]*domain.Member, error) {
	ctx := context.Background()
	var members []*domain.Member
	err := r.db.NewSelect().
		Model(&members).
		Where("group_id = ?", groupID).
		Order("created_at").
		Scan(ctx)
	return members, err
}

func (r *memberRepository) GetActiveByGroupID(groupID int64) ([]*domain.Member, error) {
	ctx := context.Background()
	var members []*domain.Member
	err := r.db.NewSelect().
		Model(&members).
		Where("group_id = ? AND active = true", groupID).
		Order("created_at").
		Scan(ctx)
	return members, err
}

func (r *memberRepository) Update(member *domain.Member) error {
	ctx := context.Background()
	member.UpdatedAt = time.Now()
	_, err := r.db.NewUpdate().Model(member).Where("id = ?", member.ID).Exec(ctx)
	return err
}

func (r *memberRepository) Delete(id int64) error {
	ctx := context.Background()
	_, err := r.db.NewDelete().Model(&domain.Member{}).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *memberRepository) IncrementOpenAssignments(memberID int64) error {
	ctx := context.Background()
	_, err := r.db.NewUpdate().
		Model(&domain.Member{}).
		Set("current_open_assignments = current_open_assignments + 1").
		Where("id = ?", memberID).
		Exec(ctx)
	return err
}

func (r *memberRepository) DecrementOpenAssignments(memberID int64) error {
	ctx := context.Background()
	_, err := r.db.NewUpdate().
		Model(&domain.Member{}).
		Set("current_open_assignments = GREATEST(current_open_assignments - 1, 0)").
		Where("id = ?", memberID).
		Exec(ctx)
	return err
}

func (r *memberRepository) GetDailyAssignmentCount(memberID int64) (int, error) {
	ctx := context.Background()

	// Get count of assignments created today
	count, err := r.db.NewSelect().
		Model(&domain.Assignment{}).
		Where("member_id = ?", memberID).
		Where("created_at >= CURRENT_DATE").
		Count(ctx)

	return count, err
}