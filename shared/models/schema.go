package models

import (
	"time"

	"github.com/google/uuid"
)

type SchemaVersion struct {
	ID          uuid.UUID `json:"id"`
	AppID       uuid.UUID `json:"app_id"`
	Version     int       `json:"version"`
	Name        string    `json:"name"`
	SQL         string    `json:"sql"`
	Checksum    string    `json:"checksum"`
	AppliedBy   uuid.UUID `json:"applied_by"`
	AppliedAt   time.Time `json:"applied_at"`
	RollbackSQL string    `json:"rollback_sql,omitempty"`
	Success     bool      `json:"success"`
	ErrorLog    string    `json:"error_log,omitempty"`
}
