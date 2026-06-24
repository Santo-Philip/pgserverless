package models

import (
	"time"

	"github.com/google/uuid"
)

type OrgStatus string

const (
	OrgStatusActive   OrgStatus = "active"
	OrgStatusInactive OrgStatus = "inactive"
	OrgStatusSuspended OrgStatus = "suspended"
)

type Organization struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description,omitempty"`
	Status      OrgStatus  `json:"status"`
	Settings    JSON       `json:"settings,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}


