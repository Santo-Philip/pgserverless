package models

import (
	"time"

	"github.com/google/uuid"
)

type Quota struct {
	ID             uuid.UUID `json:"id"`
	ProjectID      uuid.UUID `json:"project_id"`
	DatabasesUsed  int       `json:"databases_used"`
	StorageBytes   int64     `json:"storage_bytes"`
	RequestsUsed   int64     `json:"requests_used"`
	APIKeysUsed    int       `json:"api_keys_used"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type QuotaLimit struct {
	MaxDatabases   int   `json:"max_databases"`
	MaxStorageMB   int64 `json:"max_storage_mb"`
	MaxConnections int   `json:"max_connections"`
	MaxRequests    int   `json:"max_requests"`
	MaxAPIKeys     int   `json:"max_api_keys"`
}
