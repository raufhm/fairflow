package domain

import (
	"context"
	"time"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           int64     `bun:",pk,autoincrement" json:"id"`
	UserID       *int64    `bun:"user_id" json:"user_id,omitempty"`
	UserName     string    `bun:"user_name" json:"user_name"`
	Action       string    `bun:"action" json:"action"`
	ResourceType *string   `bun:"resource_type" json:"resource_type,omitempty"`
	ResourceID   *int64    `bun:"resource_id" json:"resource_id,omitempty"`
	Details      *string   `bun:"details" json:"details,omitempty"`
	IPAddress    string    `bun:"ip_address" json:"ip_address"`
	CreatedAt    time.Time `bun:"created_at" json:"created_at"`
}

// AuditLogRepository defines the interface for audit log data access
type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetRecent(ctx context.Context, limit int) ([]*AuditLog, error)
}