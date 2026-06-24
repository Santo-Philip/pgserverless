package models

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID           uuid.UUID      `json:"id"`
	AppID        *uuid.UUID     `json:"app_id,omitempty"`
	UserID       *uuid.UUID     `json:"user_id,omitempty"`
	APIKeyID     *uuid.UUID     `json:"api_key_id,omitempty"`
	Method       string         `json:"method"`
	Path         string         `json:"path"`
	StatusCode   int            `json:"status_code"`
	IPAddress    net.IP         `json:"ip_address,omitempty"`
	UserAgent    string         `json:"user_agent,omitempty"`
	ResponseTime int            `json:"response_time_ms"`
	RequestBody  string         `json:"request_body,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}
