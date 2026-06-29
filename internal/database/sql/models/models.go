package models

import (
	"time"

	"github.com/google/uuid"
)

type ExecuteRequest struct {
	Query  string `json:"query"`
	Params []any  `json:"params,omitempty"`
}

type ExecuteResponse struct {
	Columns      []string         `json:"columns,omitempty"`
	Rows         []map[string]any `json:"rows,omitempty"`
	RowCount     int              `json:"row_count,omitempty"`
	RowsAffected int64            `json:"rows_affected,omitempty"`
	DurationMs   int64            `json:"duration_ms"`
	Error        string           `json:"error,omitempty"`
}

type QueryResult struct {
	Columns      []string         `json:"columns"`
	Rows         []map[string]any `json:"rows"`
	RowsAffected int64            `json:"rows_affected"`
}

type SavedQuery struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	QueryText   string    `json:"query_text"`
	Description string    `json:"description"`
	IsShared    bool      `json:"is_shared"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type QueryHistory struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	QueryText    string    `json:"query_text"`
	DurationMs   int       `json:"duration_ms"`
	RowsAffected int       `json:"rows_affected"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
