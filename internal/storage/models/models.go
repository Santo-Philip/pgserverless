package models

import (
	"time"

	"github.com/google/uuid"
)

type ProviderType string

const (
	ProviderTypeLocal ProviderType = "local"
	ProviderTypeS3    ProviderType = "s3"
	ProviderTypeGCS   ProviderType = "gcs"
	ProviderTypeAzure ProviderType = "azure"
)

type StorageProvider struct {
	ID           uuid.UUID       `json:"id"`
	Name         string          `json:"name"`
	ProviderType ProviderType    `json:"provider_type"`
	Config       map[string]any  `json:"config"`
	IsDefault    bool            `json:"is_default"`
	IsEnabled    bool            `json:"is_enabled"`
	CreatedBy    *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type CreateProviderRequest struct {
	Name         string          `json:"name"`
	ProviderType ProviderType    `json:"provider_type"`
	Config       map[string]any  `json:"config"`
	IsDefault    bool            `json:"is_default,omitempty"`
}

type UpdateProviderRequest struct {
	Name      *string          `json:"name,omitempty"`
	Config    *map[string]any  `json:"config,omitempty"`
	IsDefault *bool            `json:"is_default,omitempty"`
	IsEnabled *bool            `json:"is_enabled,omitempty"`
}

type StorageBucket struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	ProviderID uuid.UUID  `json:"provider_id"`
	Path       string     `json:"path"`
	IsPublic   bool       `json:"is_public"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type CreateBucketRequest struct {
	Name       string `json:"name"`
	ProviderID string `json:"provider_id"`
	Path       string `json:"path,omitempty"`
	IsPublic   bool   `json:"is_public,omitempty"`
}

type StorageFile struct {
	ID        uuid.UUID       `json:"id"`
	BucketID  uuid.UUID       `json:"bucket_id"`
	Name      string          `json:"name"`
	Path      string          `json:"path"`
	MimeType  string          `json:"mime_type"`
	SizeBytes int64           `json:"size_bytes"`
	MD5Hash   string          `json:"md5_hash,omitempty"`
	Metadata  map[string]any  `json:"metadata,omitempty"`
	CreatedBy *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	URL       string          `json:"url,omitempty"`
}
