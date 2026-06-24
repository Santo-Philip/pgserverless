package models

import (
	"time"

	"github.com/google/uuid"
)

type BackupStatus string

const (
	BackupStatusRunning   BackupStatus = "running"
	BackupStatusCompleted BackupStatus = "completed"
	BackupStatusFailed    BackupStatus = "failed"
)

type Backup struct {
	ID          uuid.UUID    `json:"id"`
	AppID       uuid.UUID    `json:"app_id"`
	Status      BackupStatus `json:"status"`
	FileSize    int64        `json:"file_size,omitempty"`
	FilePath    string       `json:"file_path"`
	TriggeredBy uuid.UUID    `json:"triggered_by"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	ErrorLog    string       `json:"error_log,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}
