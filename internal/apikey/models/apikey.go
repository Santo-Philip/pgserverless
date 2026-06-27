package models

import (
	"time"

	"github.com/google/uuid"
)

type KeyType string

const (
	KeyTypeSystem  KeyType = "system"
	KeyTypeService KeyType = "service"
	KeyTypeProject KeyType = "project"
)

type APIKey struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	KeyType     KeyType    `json:"key_type"`
	KeyHash     string     `json:"-"`
	KeyPrefix   string     `json:"key_prefix"`
	Scopes      []string   `json:"scopes"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	RateLimit   int        `json:"rate_limit"`
	AllowedIPs  []string   `json:"allowed_ips,omitempty"`
	Origins     []string   `json:"origins,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
