package models

import (
	"time"

	"github.com/google/uuid"
)

type BackupInfo struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	DatabaseName string     `json:"database_name"`
	SizeBytes    int64      `json:"size_bytes"`
	Status       string     `json:"status"`
	Type         string     `json:"type"`
	FilePath     string     `json:"file_path"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CompletedBy  uuid.UUID  `json:"completed_by"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type CreateBackupRequest struct {
	Name         string `json:"name"`
	DatabaseName string `json:"database_name"`
	Type         string `json:"type"`
}

type RestoreRequest struct {
	BackupID     string `json:"backup_id"`
	DatabaseName string `json:"database_name"`
	TargetName   string `json:"target_name"`
}

type BackupHistory struct {
	Data   []BackupInfo `json:"data"`
	Total  int          `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}
