package domain

import "time"

// AssignmentStatus represents the status of an assignment
type AssignmentStatus string

const (
	AssignmentStatusOpen      AssignmentStatus = "open"
	AssignmentStatusCompleted AssignmentStatus = "completed"
	AssignmentStatusCancelled AssignmentStatus = "cancelled"
)

// Assignment represents a recorded assignment
type Assignment struct {
	ID          int64            `bun:",pk,autoincrement" json:"id"`
	GroupID     int64            `bun:"group_id" json:"group_id"`
	MemberID    int64            `bun:"member_id" json:"member_id"`
	Metadata    *string          `bun:"metadata" json:"metadata,omitempty"`
	Status      AssignmentStatus `bun:"status" json:"status"`
	CompletedAt *time.Time       `bun:"completed_at" json:"completed_at,omitempty"`
	CreatedAt   time.Time        `bun:"created_at" json:"created_at"`
}

// AssignmentWithMember represents an assignment with member details
type AssignmentWithMember struct {
	ID         int64     `json:"id"`
	MemberID   int64     `json:"member_id"`
	MemberName string    `json:"member_name"`
	Metadata   *string   `json:"metadata,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// AssignmentStats represents statistics for a group
type AssignmentStats struct {
	TotalAssignments int                   `json:"total_assignments"`
	Distribution     []MemberDistribution  `json:"distribution"`
}

// MemberDistribution represents assignment distribution for a member
type MemberDistribution struct {
	MemberID    int64   `json:"member_id"`
	Name        string  `json:"name"`
	Weight      int     `json:"weight"`
	Assignments int     `json:"assignments"`
	Expected    float64 `json:"expected"`
	Variance    float64 `json:"variance"`
}

// AssignmentRepository defines the interface for assignment data access
type AssignmentRepository interface {
	Create(assignment *Assignment) error
	GetByID(id int64) (*Assignment, error)
	GetByGroupID(groupID int64, limit, offset int) ([]*AssignmentWithMember, error)
	GetCountByGroupID(groupID int64) (int, error)
	GetCountsByMemberIDs(memberIDs []int64) (map[int64]int, error)
	UpdateStatus(id int64, status AssignmentStatus) error
}