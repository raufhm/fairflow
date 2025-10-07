package domain

import "time"

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
	Create(webhook *Webhook) error
	GetByGroupID(groupID int64) ([]*Webhook, error)
	GetActiveByGroupID(groupID int64) ([]*Webhook, error)
	Update(webhook *Webhook) error
	Delete(id int64) error
}