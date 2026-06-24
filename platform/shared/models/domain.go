package models

import (
	"time"

	"github.com/google/uuid"
)

type DomainStatus string

const (
	DomainStatusPending  DomainStatus = "pending"
	DomainStatusActive   DomainStatus = "active"
	DomainStatusFailed   DomainStatus = "failed"
)

type Domain struct {
	ID                uuid.UUID    `json:"id"`
	AppID             uuid.UUID    `json:"app_id"`
	Domain            string       `json:"domain"`
	Status            DomainStatus `json:"status"`
	Verified          bool         `json:"verified"`
	VerificationToken string       `json:"verification_token,omitempty"`
	VerifiedAt        *time.Time   `json:"verified_at,omitempty"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}
