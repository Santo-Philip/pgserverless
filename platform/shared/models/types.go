package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JSON map[string]interface{}

type StringList []string

func (sl StringList) Contains(s string) bool {
	for _, v := range sl {
		if v == s {
			return true
		}
	}
	return false
}

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

func MustParseUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
}

func Now() time.Time {
	return time.Now().UTC()
}

func ToJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func FromJSON(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}
