package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/raufhm/rra/internal/domain"
	"github.com/uptrace/bun"
)

type userRepository struct {
	db *bun.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *bun.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	ctx := context.Background()
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

func (r *userRepository) GetByID(id int64) (*domain.User, error) {
	user := new(domain.User)
	err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(context.Background())
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	user := new(domain.User)
	err := r.db.NewSelect().Model(user).Where("email = ?", email).Scan(context.Background())
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetAll() ([]*domain.User, error) {
	var users []*domain.User
	err := r.db.NewSelect().Model(&users).Order("created_at DESC").Scan(context.Background())
	return users, err
}

func (r *userRepository) Update(user *domain.User) error {
	ctx := context.Background()
	user.UpdatedAt = time.Now()
	_, err := r.db.NewUpdate().Model(user).WherePK().Exec(ctx)
	return err
}

func (r *userRepository) Delete(id int64) error {
	ctx := context.Background()
	_, err := r.db.NewDelete().Model((*domain.User)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *userRepository) UpdateRole(id int64, role domain.UserRole) error {
	ctx := context.Background()
	_, err := r.db.NewUpdate().
		Model((*domain.User)(nil)).
		Set("role = ?", role).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)
	return err
}