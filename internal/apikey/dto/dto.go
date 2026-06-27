package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/apikey/models"
)

type CreateKeyRequest struct {
	Name      string         `json:"name"`
	KeyType   models.KeyType `json:"key_type"`
	ProjectID string         `json:"project_id,omitempty"`
	Scopes    []string       `json:"scopes"`
	RateLimit int            `json:"rate_limit"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
	IPs       []string       `json:"allowed_ips,omitempty"`
	Origins   []string       `json:"origins,omitempty"`
}

type KeyResponse struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	KeyType   models.KeyType `json:"key_type"`
	KeyPrefix string         `json:"key_prefix"`
	RawKey    string         `json:"raw_key,omitempty"`
	Scopes    []string       `json:"scopes"`
	CreatedAt time.Time      `json:"created_at"`
}

type KeyListResponse struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	KeyType   models.KeyType `json:"key_type"`
	KeyPrefix string         `json:"key_prefix"`
	Scopes    []string       `json:"scopes"`
	ProjectID *uuid.UUID     `json:"project_id,omitempty"`
	IsActive  bool           `json:"is_active"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}
