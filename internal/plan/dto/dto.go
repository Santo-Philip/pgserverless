package dto

import "github.com/google/uuid"

type CreatePlanRequest struct {
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	Description    string  `json:"description,omitempty"`
	MaxDatabases   int     `json:"max_databases"`
	MaxStorageMB   int64   `json:"max_storage_mb"`
	MaxConnections int     `json:"max_connections"`
	MaxRequests    int     `json:"max_requests"`
	MaxAPIKeys     int     `json:"max_api_keys"`
	Price          float64 `json:"price"`
}

type UpdatePlanRequest struct {
	Name           *string  `json:"name,omitempty"`
	Description    *string  `json:"description,omitempty"`
	MaxDatabases   *int     `json:"max_databases,omitempty"`
	MaxStorageMB   *int64   `json:"max_storage_mb,omitempty"`
	MaxConnections *int     `json:"max_connections,omitempty"`
	MaxRequests    *int     `json:"max_requests,omitempty"`
	MaxAPIKeys     *int     `json:"max_api_keys,omitempty"`
	Price          *float64 `json:"price,omitempty"`
	IsActive       *bool    `json:"is_active,omitempty"`
}

type PlanResponse struct {
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
}
