package models

import (
	"time"

	"github.com/google/uuid"
)

type KeyType string

const (
	KeyTypePublishable KeyType = "publishable"
	KeyTypeSecret      KeyType = "secret"
	KeyTypeService     KeyType = "service"
	KeyTypeAdmin       KeyType = "admin"
)

type APIKey struct {
	ID            uuid.UUID  `json:"id"`
	AppID         uuid.UUID  `json:"app_id"`
	UserID        *uuid.UUID `json:"user_id,omitempty"`
	Name          string     `json:"name"`
	KeyType       KeyType    `json:"key_type"`
	KeyHash       string     `json:"-"`
	KeyPrefix     string     `json:"key_prefix"`
	Scopes        []string   `json:"scopes"`
	RateLimit     int        `json:"rate_limit"`
	AllowedIPs    []string   `json:"allowed_ips,omitempty"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CreateAPIKeyRequest struct {
	Name      string     `json:"name"`
	KeyType   KeyType    `json:"key_type"`
	Scopes    []string   `json:"scopes"`
	RateLimit int        `json:"rate_limit"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type APIKeyResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	KeyType   KeyType   `json:"key_type"`
	KeyPrefix string    `json:"key_prefix"`
	RawKey    string    `json:"raw_key,omitempty"`
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"created_at"`
}
