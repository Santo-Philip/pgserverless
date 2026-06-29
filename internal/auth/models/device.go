package models

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	DeviceName     string    `json:"device_name"`
	DeviceType     string    `json:"device_type"`
	IPAddress      string    `json:"ip_address"`
	ClientDeviceID string    `json:"client_device_id,omitempty"`
	LastUsedAt     time.Time `json:"last_used_at"`
	CreatedAt      time.Time `json:"created_at"`
}
