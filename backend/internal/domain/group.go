package domain

import "time"

// AssignmentStrategy defines the strategy for assignments
type AssignmentStrategy string

const (
	StrategyWeightedRoundRobin AssignmentStrategy = "weighted_round_robin"
	StrategyStrictRotation     AssignmentStrategy = "strict_rotation"
)

// Group represents a group for round-robin assignments
type Group struct {
	ID               int64              `bun:"id,pk,autoincrement" json:"id"`
	UserID           int64              `bun:"user_id,notnull" json:"user_id"`
	OrganizationID   *int64             `bun:"organization_id" json:"organization_id,omitempty"`
	Name             string             `bun:"name,notnull" json:"name"`
	Description      *string            `bun:"description" json:"description,omitempty"`
	Strategy         AssignmentStrategy `bun:"strategy,notnull,default:'weighted_round_robin'" json:"strategy"`
	Active           bool               `bun:"active,notnull,default:true" json:"active"`
	Settings         *string            `bun:"settings" json:"settings,omitempty"`
	AssignmentPaused bool               `bun:"assignment_paused,notnull,default:false" json:"assignment_paused"`
	PauseReason      *string            `bun:"pause_reason" json:"pause_reason,omitempty"`
	PausedAt         *time.Time         `bun:"paused_at" json:"paused_at,omitempty"`
	PausedBy         *int64             `bun:"paused_by" json:"paused_by,omitempty"`
	CreatedAt        time.Time          `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt        time.Time          `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// GroupRepository defines the interface for group data access
type GroupRepository interface {
	Create(group *Group) error
	GetByID(id int64) (*Group, error)
	GetAll() ([]*Group, error)
	GetByUserID(userID int64) ([]*Group, error)
	Update(group *Group) error
	Delete(id int64) error
}