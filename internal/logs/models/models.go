package models

import "time"

type LogEntry struct {
	Timestamp      time.Time `json:"timestamp"`
	ProcessID      int       `json:"process_id"`
	SessionID      string    `json:"session_id,omitempty"`
	SessionLineNum int       `json:"session_line_num,omitempty"`
	Database       string    `json:"database,omitempty"`
	User           string    `json:"user,omitempty"`
	Severity       string    `json:"severity"`
	LogMessage     string    `json:"log_message"`
	Detail         string    `json:"detail,omitempty"`
	Hint           string    `json:"hint,omitempty"`
	Query          string    `json:"query,omitempty"`
	Context        string    `json:"context,omitempty"`
	ErrorCode      string    `json:"error_code,omitempty"`
	File           string    `json:"file,omitempty"`
	Line           int       `json:"line,omitempty"`
	Routine        string    `json:"routine,omitempty"`
}

type LogQuery struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Severity  string     `json:"severity,omitempty"`
	Database  string     `json:"database,omitempty"`
	User      string     `json:"user,omitempty"`
	Search    string     `json:"search,omitempty"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

type LogResponse struct {
	Entries []LogEntry `json:"entries"`
	Total   int        `json:"total"`
	Limit   int        `json:"limit"`
	Offset  int        `json:"offset"`
}
