package models

import (
	"time"

	"github.com/google/uuid"
)

type MembershipRole string

const (
	RoleOwner  MembershipRole = "owner"
	RoleAdmin  MembershipRole = "admin"
	RoleEditor MembershipRole = "editor"
	RoleViewer MembershipRole = "viewer"
)

type Member struct {
	ID             uuid.UUID      `json:"id"`
	AppID          uuid.UUID      `json:"app_id"`
	UserID         uuid.UUID      `json:"user_id"`
	Role           MembershipRole `json:"role"`
	InvitedBy      *uuid.UUID     `json:"invited_by,omitempty"`
	AcceptedAt     *time.Time     `json:"accepted_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}
