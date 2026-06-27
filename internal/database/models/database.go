package models

import (
	"time"

	"github.com/google/uuid"
)

type Database struct {
	ID           uuid.UUID  `json:"id"`
	ProjectID    uuid.UUID  `json:"project_id"`
	Name         string     `json:"name"`
	SchemaName   string     `json:"schema_name"`
	DBUser       string     `json:"db_user"`
	DBPassword   string     `json:"-"`
	ConnString   string     `json:"connection_string,omitempty"`
	Status       string     `json:"status"`
	SizeBytes    int64      `json:"size_bytes"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type DatabaseUser struct {
	ID         uuid.UUID `json:"id"`
	DatabaseID uuid.UUID `json:"database_id"`
	Name       string    `json:"name"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type Backup struct {
	ID         uuid.UUID  `json:"id"`
	DatabaseID uuid.UUID  `json:"database_id"`
	SizeBytes  int64      `json:"size_bytes"`
	Status     string     `json:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Extension struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
}

type TableInfo struct {
	Name    string       `json:"name"`
	Columns []TableColumn `json:"columns"`
}

type TableColumn struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	IsPK         bool   `json:"is_pk"`
	DefaultValue string `json:"default_value,omitempty"`
}
