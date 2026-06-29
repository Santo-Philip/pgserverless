package dto

import (
	"time"

	"github.com/google/uuid"
	authmodels "github.com/nexbic/platform/internal/auth/models"
)

// ── Auth ────────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code,omitempty"`
}

type LoginResponse struct {
	RequiresTOTP bool   `json:"requires_totp,omitempty"`
	SessionToken string `json:"session_token,omitempty"`
}

type CompleteTOTPLoginRequest struct {
	SessionToken string `json:"session_token"`
	TOTPCode     string `json:"totp_code"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	Token        string          `json:"token"`
	RefreshToken string          `json:"refresh_token,omitempty"`
	User         authmodels.User `json:"user"`
	ExpiresAt    time.Time       `json:"expires_at"`
}

// ── Registration ────────────────────────────────────────

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// ── Email Verification ──────────────────────────────────

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// ── Password Reset ──────────────────────────────────────

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// ── TOTP ────────────────────────────────────────────────

type EnableTOTPResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

type VerifyTOTPRequest struct {
	Code string `json:"code"`
}

type DisableTOTPRequest struct {
	Code string `json:"code"`
}

// ── API Keys ────────────────────────────────────────────

type CreateAPIKeyRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type CreateAPIKeyResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"`
	Prefix    string     `json:"prefix"`
	Status    string     `json:"status"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type APIKeyResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Prefix    string     `json:"prefix"`
	Status    string     `json:"status"`
	LastUsed  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// ── Sessions / Devices ──────────────────────────────────

type DeviceResponse struct {
	ID           uuid.UUID `json:"id"`
	DeviceName   string    `json:"device_name"`
	DeviceType   string    `json:"device_type"`
	IPAddress    string    `json:"ip_address"`
	LastUsedAt   time.Time `json:"last_used_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// ── Security Events ─────────────────────────────────────

type SecurityEventResponse struct {
	ID        uuid.UUID `json:"id"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	IPAddress string    `json:"ip_address"`
	Timestamp time.Time `json:"timestamp"`
}

// ── User Management ─────────────────────────────────────

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
