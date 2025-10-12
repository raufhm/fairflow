package postgres

import (
	"context"

	"github.com/uptrace/bun"
)

// BaseRepository provides common functionality for all repositories
type BaseRepository struct {
	db *bun.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *bun.DB) BaseRepository {
	return BaseRepository{
		db: db,
	}
}

// Execute runs a database operation
func (r *BaseRepository) Execute(ctx context.Context, fn func() error) error {
	return fn()
}

// GetDB returns the underlying database connection
func (r *BaseRepository) GetDB() *bun.DB {
	return r.db
}
