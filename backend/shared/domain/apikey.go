package domain

import (
	"context"
	"time"
)

// APIKey represents an API key for authentication
type APIKey struct {
	ID         int64      `bun:",pk,autoincrement" json:"id"`
	UserID     int64      `bun:"user_id" json:"user_id"`
	Name       string     `bun:"name" json:"name"`
	KeyHash    string     `bun:"key_hash" json:"-"`
	ExpiresAt  *time.Time `bun:"expires_at" json:"expires_at,omitempty"`
	LastUsedAt *time.Time `bun:"last_used_at" json:"last_used_at,omitempty"`
	Active     bool       `bun:"active" json:"active"`
	CreatedAt  time.Time  `bun:"created_at" json:"created_at"`
}

// APIKeyRepository defines the interface for API key data access
type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *APIKey) error
	GetByID(ctx context.Context, id int64) (*APIKey, error)
	GetByHash(ctx context.Context, hash string) (*APIKey, error)
	GetByUserID(ctx context.Context, userID int64) ([]*APIKey, error)
	Delete(ctx context.Context, id int64) error
	UpdateLastUsed(ctx context.Context, id int64) error
}