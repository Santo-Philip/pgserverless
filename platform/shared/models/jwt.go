package models

import (
	"time"

	"github.com/google/uuid"
)

type JWTSecret struct {
	ID        uuid.UUID `json:"id"`
	AppID     uuid.UUID `json:"app_id"`
	Secret    string    `json:"-"`
	IsActive  bool      `json:"is_active"`
	RotatedAt *time.Time `json:"rotated_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenClaims struct {
	Sub            string   `json:"sub"`
	Email          string   `json:"email"`
	Name           string   `json:"name"`
	Role           string   `json:"role"`
	AppID          string   `json:"app_id,omitempty"`
	OrganizationID string   `json:"org_id,omitempty"`
	Permissions    []string `json:"permissions,omitempty"`
	Type           string   `json:"type"`
	Iss            string   `json:"iss"`
	Aud            string   `json:"aud"`
	Exp            int64    `json:"exp"`
	Iat            int64    `json:"iat"`
}
