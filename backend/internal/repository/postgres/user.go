package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/uptrace/bun"
)

type userRepository struct {
	db *bun.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *bun.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user := new(domain.User)
	err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := new(domain.User)
	err := r.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User
	err := r.db.NewSelect().Model(&users).Order("created_at DESC").Scan(ctx)
	return users, err
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()
	_, err := r.db.NewUpdate().Model(user).WherePK().Exec(ctx)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*domain.User)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *userRepository) UpdateRole(ctx context.Context, id int64, role domain.UserRole) error {
	_, err := r.db.NewUpdate().
		Model((*domain.User)(nil)).
		Set("role = ?", role).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)
	return err
}
