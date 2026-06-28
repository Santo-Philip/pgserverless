package dto

import (
	"time"

	"github.com/google/uuid"
	authmodels "github.com/nexbic/platform/internal/auth/models"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	Token        string              `json:"token"`
	RefreshToken string              `json:"refresh_token,omitempty"`
	User         authmodels.User     `json:"user"`
	ExpiresAt    time.Time           `json:"expires_at"`
}

type UserResponse struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Name     string `json:"name,omitempty"`
	Role     string `json:"role,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UpdateUserPasswordRequest struct {
	NewPassword string `json:"new_password"`
}
