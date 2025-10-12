package postgres

import (
	"context"
	"time"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/uptrace/bun"
)

type groupRepository struct {
	db *bun.DB
}

// NewGroupRepository creates a new group repository
func NewGroupRepository(db *bun.DB) domain.GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) Create(ctx context.Context, group *domain.Group) error {
	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now
	_, err := r.db.NewInsert().Model(group).Exec(ctx)
	return err
}

func (r *groupRepository) GetByID(ctx context.Context, id int64) (*domain.Group, error) {
	group := &domain.Group{}
	err := r.db.NewSelect().Model(group).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (r *groupRepository) GetAll(ctx context.Context) ([]*domain.Group, error) {
	var groups []*domain.Group
	err := r.db.NewSelect().Model(&groups).Order("created_at DESC").Scan(ctx)
	return groups, err
}

func (r *groupRepository) GetByUserID(ctx context.Context, userID int64) ([]*domain.Group, error) {
	var groups []*domain.Group
	err := r.db.NewSelect().Model(&groups).Where("user_id = ?", userID).Order("created_at DESC").Scan(ctx)
	return groups, err
}

func (r *groupRepository) Update(ctx context.Context, group *domain.Group) error {
	group.UpdatedAt = time.Now()
	_, err := r.db.NewUpdate().Model(group).Where("id = ?", group.ID).Exec(ctx)
	return err
}

func (r *groupRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model(&domain.Group{}).Where("id = ?", id).Exec(ctx)
	return err
}
