package models

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description,omitempty"`
	MaxDatabases   int       `json:"max_databases"`
	MaxStorageMB   int64     `json:"max_storage_mb"`
	MaxConnections int       `json:"max_connections"`
	MaxRequests    int       `json:"max_requests"`
	MaxAPIKeys     int       `json:"max_api_keys"`
	Price          float64   `json:"price"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
