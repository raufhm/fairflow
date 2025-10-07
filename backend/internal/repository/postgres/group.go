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

func (r *groupRepository) Create(group *domain.Group) error {
	ctx := context.Background()
	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now
	_, err := r.db.NewInsert().Model(group).Exec(ctx)
	return err
}

func (r *groupRepository) GetByID(id int64) (*domain.Group, error) {
	ctx := context.Background()
	group := &domain.Group{}
	err := r.db.NewSelect().Model(group).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (r *groupRepository) GetAll() ([]*domain.Group, error) {
	ctx := context.Background()
	var groups []*domain.Group
	err := r.db.NewSelect().Model(&groups).Order("created_at DESC").Scan(ctx)
	return groups, err
}

func (r *groupRepository) GetByUserID(userID int64) ([]*domain.Group, error) {
	ctx := context.Background()
	var groups []*domain.Group
	err := r.db.NewSelect().Model(&groups).Where("user_id = ?", userID).Order("created_at DESC").Scan(ctx)
	return groups, err
}

func (r *groupRepository) Update(group *domain.Group) error {
	ctx := context.Background()
	group.UpdatedAt = time.Now()
	_, err := r.db.NewUpdate().Model(group).Where("id = ?", group.ID).Exec(ctx)
	return err
}

func (r *groupRepository) Delete(id int64) error {
	ctx := context.Background()
	_, err := r.db.NewDelete().Model(&domain.Group{}).Where("id = ?", id).Exec(ctx)
	return err
}
