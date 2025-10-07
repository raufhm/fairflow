package domain

import "time"

// Member represents a group member who can be assigned
type Member struct {
	ID                    int64     `bun:",pk,autoincrement" json:"id"`
	GroupID               int64     `bun:"group_id" json:"group_id"`
	Name                  string    `bun:"name" json:"name"`
	Email                 *string   `bun:"email" json:"email,omitempty"`
	Weight                int       `bun:"weight" json:"weight"`
	Active                bool      `bun:"active" json:"active"`
	Available             bool      `bun:"available" json:"available"`                        // Availability status
	WorkingHours          *string   `bun:"working_hours" json:"working_hours,omitempty"`          // JSON: {"monday": "09:00-17:00", ...}
	Timezone              *string   `bun:"timezone" json:"timezone,omitempty"`               // IANA timezone e.g. "America/New_York"
	Metadata              *string   `bun:"metadata" json:"metadata,omitempty"`
	MaxDailyAssignments   *int      `bun:"max_daily_assignments" json:"max_daily_assignments,omitempty"`   // Maximum assignments per day
	MaxConcurrentOpen     *int      `bun:"max_concurrent_open" json:"max_concurrent_open,omitempty"`     // Maximum open assignments at once
	CurrentOpenAssignments int      `bun:"current_open_assignments" json:"current_open_assignments"`          // Current number of open assignments
	CreatedAt             time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt             time.Time `bun:"updated_at" json:"updated_at"`
	Assignments           int       `bun:"-" json:"assignments,omitempty"`     // Calculated field, not stored in DB
}

// MemberRepository defines the interface for member data access
type MemberRepository interface {
	Create(member *Member) error
	GetByID(id int64) (*Member, error)
	GetByGroupID(groupID int64) ([]*Member, error)
	GetActiveByGroupID(groupID int64) ([]*Member, error)
	Update(member *Member) error
	Delete(id int64) error
	IncrementOpenAssignments(memberID int64) error
	DecrementOpenAssignments(memberID int64) error
	GetDailyAssignmentCount(memberID int64) (int, error)
}