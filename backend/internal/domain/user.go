package domain

import (
	"context"
	"time"
)

// UserRole defines the available user roles
type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleAdmin      UserRole = "admin"
	RoleManager    UserRole = "manager"
	RoleUser       UserRole = "user"
)

// User represents a system user
type User struct {
	ID             int64     `bun:"id,pk,autoincrement" json:"id"`
	Email          string    `bun:"email,notnull,unique" json:"email"`
	PasswordHash   string    `bun:"password_hash,notnull" json:"-"`
	Name           string    `bun:"name,notnull" json:"name"`
	Role           UserRole  `bun:"role,notnull,default:'user'" json:"role"`
	OrganizationID *int64    `bun:"organization_id" json:"organization_id,omitempty"`
	CreatedAt      time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
	UpdateRole(ctx context.Context, id int64, role UserRole) error
}