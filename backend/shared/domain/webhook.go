package domain

import (
	"context"
	"time"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID        int64     `json:"id"`
	GroupID   int64     `json:"group_id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Secret    string    `json:"-"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// WebhookRepository defines the interface for webhook data access
type WebhookRepository interface {
	Create(ctx context.Context, webhook *Webhook) error
	GetByGroupID(ctx context.Context, groupID int64) ([]*Webhook, error)
	GetActiveByGroupID(ctx context.Context, groupID int64) ([]*Webhook, error)
	Update(ctx context.Context, webhook *Webhook) error
	Delete(ctx context.Context, id int64) error
}