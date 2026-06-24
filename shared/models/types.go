package models

import (
	"time"

	"github.com/google/uuid"
)

type JSON map[string]interface{}

type PaginationParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewID() uuid.UUID {
	return uuid.New()
}

func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func Now() time.Time {
	return time.Now().UTC()
}
