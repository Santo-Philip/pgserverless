package models

import (
	"time"

	"github.com/google/uuid"
)

type AppStatus string

const (
	AppStatusActive     AppStatus = "active"
	AppStatusInactive   AppStatus = "inactive"
	AppStatusSuspended  AppStatus = "suspended"
	AppStatusDeleted    AppStatus = "deleted"
)

type Visibility string

const (
	VisibilityPublic   Visibility = "public"
	VisibilityPrivate  Visibility = "private"
)

type App struct {
	ID           uuid.UUID  `json:"id"`
	OrgID        *uuid.UUID `json:"org_id,omitempty"`
	OwnerID      *uuid.UUID `json:"owner_id,omitempty"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	Description  string     `json:"description,omitempty"`
	SchemaName   string     `json:"schema_name"`
	Status       AppStatus  `json:"status"`
	Region       string     `json:"region"`
	Visibility   Visibility `json:"visibility"`
	Settings     JSON       `json:"settings,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}
