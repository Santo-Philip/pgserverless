package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleSuperAdmin = "super_admin"
	RoleDBA        = "dba"
	RoleDeveloper  = "developer"
	RoleReadOnly   = "read_only"
)

type User struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	PasswordHash    string     `json:"-"`
	Name            string     `json:"name"`
	Image           string     `json:"image,omitempty"`
	Role            string     `json:"role"`
	IsActive        bool       `json:"is_active"`
	EmailVerified   bool       `json:"email_verified"`
	TOTPEnabled     bool       `json:"totp_enabled"`
	TOTPSecret      string     `json:"-"`
	RecoveryCodes   []string   `json:"-"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
