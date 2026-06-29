package service

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/database/logs/models"
	"github.com/nexbic/platform/pkg/database"
)

type LogsService struct {
	db *database.DB
}

func NewLogsService(db *database.DB) *LogsService {
	return &LogsService{db: db}
}

func (s *LogsService) GetLogs(ctx context.Context, q models.LogQuery) (*models.LogResponse, error) {
	entries, err := s.readFileLogs(ctx, q)
	if err == nil && len(entries) > 0 {
		return s.paginate(entries, q.Limit, q.Offset), nil
	}
	if err != nil {
		slog.Debug("file log read failed, falling back to database", "error", err)
	}
	return s.GetQueryLogs(ctx, q.Limit, q.Offset)
}

func (s *LogsService) GetQueryLogs(ctx context.Context, limit, offset int) (*models.LogResponse, error) {
	var total int
	err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM query_history`).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, user_id, query_text, database_name, duration_ms, rows_affected, status,
		       COALESCE(error_message, ''), created_at
		FROM query_history
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LogEntry
	for rows.Next() {
		var id, userID uuid.UUID
		var queryText, dbName, status, errMsg string
		var durMs, rowsAff int
		var createdAt time.Time

		if err := rows.Scan(&id, &userID, &queryText, &dbName, &durMs, &rowsAff, &status, &errMsg, &createdAt); err != nil {
			return nil, err
		}

		msg := fmt.Sprintf("duration: %dms, rows: %d, status: %s", durMs, rowsAff, status)
		if errMsg != "" {
			msg = errMsg
		}

		entries = append(entries, models.LogEntry{
			Timestamp:  createdAt,
			Database:   dbName,
			User:       userID.String(),
			Severity:   mapStatusToSeverity(status),
			LogMessage: msg,
			Query:      queryText,
		})
	}
	if entries == nil {
		entries = []models.LogEntry{}
	}
	return &models.LogResponse{Entries: entries, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *LogsService) GetErrorLogs(ctx context.Context, severity string, limit, offset int) (*models.LogResponse, error) {
	entries, err := s.readFileLogs(ctx, models.LogQuery{
		Severity: severity,
		Limit:    limit,
		Offset:   offset,
	})
	if err == nil && len(entries) > 0 {
		return s.paginate(entries, limit, offset), nil
	}

	if severity == "" {
		severity = "error"
	}

	var total int
	err = s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM query_history WHERE status = 'error'
		AND ($1 = '' OR LOWER(error_message) LIKE LOWER('%' || $1 || '%'))`, severity).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, user_id, query_text, database_name, duration_ms, rows_affected, status,
		       COALESCE(error_message, ''), created_at
		FROM query_history
		WHERE status = 'error'
		AND ($1 = '' OR LOWER(error_message) LIKE LOWER('%' || $1 || '%'))
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, severity, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries = []models.LogEntry{}
	for rows.Next() {
		var id, userID uuid.UUID
		var queryText, dbName, status, errMsg string
		var durMs, rowsAff int
		var createdAt time.Time

		if err := rows.Scan(&id, &userID, &queryText, &dbName, &durMs, &rowsAff, &status, &errMsg, &createdAt); err != nil {
			return nil, err
		}

		entries = append(entries, models.LogEntry{
			Timestamp:  createdAt,
			Database:   dbName,
			User:       userID.String(),
			Severity:   "ERROR",
			LogMessage: errMsg,
			Query:      queryText,
		})
	}
	if entries == nil {
		entries = []models.LogEntry{}
	}
	return &models.LogResponse{Entries: entries, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *LogsService) GetAuthLogs(ctx context.Context, limit, offset int) (*models.LogResponse, error) {
	var total int
	err := s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_logs
		WHERE action IN ('login', 'logout', 'password', 'revoke')`).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, actor_id, action, resource, COALESCE(resource_id, ''), metadata,
		       COALESCE(ip_address, ''), COALESCE(user_agent, ''), created_at
		FROM audit_logs
		WHERE action IN ('login', 'logout', 'password', 'revoke')
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LogEntry
	for rows.Next() {
		var id, actorID uuid.UUID
		var action, resource, resourceID, ip, ua string
		var meta map[string]any
		var createdAt time.Time

		if err := rows.Scan(&id, &actorID, &action, &resource, &resourceID, &meta, &ip, &ua, &createdAt); err != nil {
			return nil, err
		}

		msg := fmt.Sprintf("%s on %s", action, resource)
		if resourceID != "" {
			msg = fmt.Sprintf("%s on %s/%s", action, resource, resourceID)
		}

		entries = append(entries, models.LogEntry{
			Timestamp:  createdAt,
			User:       actorID.String(),
			Severity:   "LOG",
			LogMessage: msg,
			Context:    fmt.Sprintf("ip: %s, user_agent: %s", ip, ua),
		})
	}
	if entries == nil {
		entries = []models.LogEntry{}
	}
	return &models.LogResponse{Entries: entries, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *LogsService) GetConnectionLogs(ctx context.Context, limit, offset int) (*models.LogResponse, error) {
	entries, err := s.readFileLogs(ctx, models.LogQuery{
		Limit:  limit,
		Offset: offset,
	})
	if err == nil {
		connEntries := filterConnectionEntries(entries)
		if len(connEntries) > 0 {
			return s.paginate(connEntries, limit, offset), nil
		}
	}

	var total int
	err = s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_logs
		WHERE action IN ('create', 'delete', 'enable', 'disable')
		AND resource = 'connection'`).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, actor_id, action, resource, COALESCE(resource_id, ''), metadata,
		       COALESCE(ip_address, ''), COALESCE(user_agent, ''), created_at
		FROM audit_logs
		WHERE action IN ('create', 'delete', 'enable', 'disable')
		AND resource = 'connection'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dbEntries := []models.LogEntry{}
	for rows.Next() {
		var id, actorID uuid.UUID
		var action, resource, resourceID, ip, ua string
		var meta map[string]any
		var createdAt time.Time

		if err := rows.Scan(&id, &actorID, &action, &resource, &resourceID, &meta, &ip, &ua, &createdAt); err != nil {
			return nil, err
		}

		dbEntries = append(dbEntries, models.LogEntry{
			Timestamp:  createdAt,
			User:       actorID.String(),
			Severity:   "LOG",
			LogMessage: fmt.Sprintf("connection %s: %s", action, resourceID),
			Context:    fmt.Sprintf("ip: %s", ip),
		})
	}
	if dbEntries == nil {
		dbEntries = []models.LogEntry{}
	}
	return &models.LogResponse{Entries: dbEntries, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *LogsService) readFileLogs(ctx context.Context, q models.LogQuery) ([]models.LogEntry, error) {
	logDir, err := s.getLogDirectory(ctx)
	if err != nil {
		return nil, err
	}

	logDest, err := s.getLogSetting(ctx, "log_destination")
	if err != nil {
		return nil, err
	}

	entries, err := s.readLogFiles(logDir, logDest, q)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func (s *LogsService) getLogDirectory(ctx context.Context) (string, error) {
	var dir string
	err := s.db.Pool.QueryRow(ctx, `SELECT setting FROM pg_settings WHERE name = 'log_directory'`).Scan(&dir)
	if err != nil {
		return "", fmt.Errorf("query log_directory: %w", err)
	}

	dataDir := ""
	err = s.db.Pool.QueryRow(ctx, `SELECT setting FROM pg_settings WHERE name = 'data_directory'`).Scan(&dataDir)
	if err == nil && dataDir != "" {
		dir = filepath.Join(dataDir, dir)
	}

	return dir, nil
}

func (s *LogsService) getLogSetting(ctx context.Context, name string) (string, error) {
	var val string
	err := s.db.Pool.QueryRow(ctx, `SELECT setting FROM pg_settings WHERE name = $1`, name).Scan(&val)
	if err != nil {
		return "", fmt.Errorf("query %s: %w", name, err)
	}
	return strings.ToLower(val), nil
}

func (s *LogsService) readLogFiles(logDir, logDest string, q models.LogQuery) ([]models.LogEntry, error) {
	fi, err := os.Stat(logDir)
	if err != nil {
		return nil, fmt.Errorf("stat log directory: %w", err)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("log path is not a directory: %s", logDir)
	}

	files, err := os.ReadDir(logDir)
	if err != nil {
		return nil, fmt.Errorf("read log directory: %w", err)
	}

	var allEntries []models.LogEntry
	isCSV := strings.Contains(logDest, "csvlog")

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		fpath := filepath.Join(logDir, f.Name())

		var entries []models.LogEntry
		if isCSV && strings.HasSuffix(strings.ToLower(f.Name()), ".csv") {
			entries, err = parseCSVLogFile(fpath)
		} else {
			entries, err = parseStderrLogFile(fpath)
		}
		if err != nil {
			slog.Debug("skipping unparseable log file", "file", f.Name(), "error", err)
			continue
		}

		entries = filterLogEntries(entries, q)
		allEntries = append(allEntries, entries...)
	}

	sortLogEntriesDesc(allEntries)

	return allEntries, nil
}

func parseCSVLogFile(fpath string) ([]models.LogEntry, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.ReuseRecord = true

	var entries []models.LogEntry
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if len(record) < 12 {
			continue
		}

		entry := models.LogEntry{}
		if t, err := time.Parse("2006-01-02 15:04:05.999 MST", strings.TrimSpace(record[0])); err == nil {
			entry.Timestamp = t
		} else if t, err := time.Parse("2006-01-02 15:04:05 MST", strings.TrimSpace(record[0])); err == nil {
			entry.Timestamp = t
		}

		entry.User = strings.TrimSpace(record[1])
		entry.Database = strings.TrimSpace(record[2])

		if pid, err := strconv.Atoi(strings.TrimSpace(record[3])); err == nil {
			entry.ProcessID = pid
		}

		entry.SessionID = strings.TrimSpace(record[6])
		if sn, err := strconv.Atoi(strings.TrimSpace(record[7])); err == nil {
			entry.SessionLineNum = sn
		}

		entry.Severity = strings.TrimSpace(record[11])
		entry.ErrorCode = strings.TrimSpace(record[12])
		entry.LogMessage = strings.TrimSpace(record[13])

		if len(record) > 14 {
			entry.Detail = strings.TrimSpace(record[14])
		}
		if len(record) > 15 {
			entry.Hint = strings.TrimSpace(record[15])
		}
		if len(record) > 18 {
			entry.Context = strings.TrimSpace(record[18])
		}
		if len(record) > 19 {
			entry.Query = strings.TrimSpace(record[19])
		}

		entries = append(entries, entry)
	}
	return entries, nil
}

var stderrLogRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?:\.\d+)?(?:\s+\w+)?)\s+\[(\d+)\](?:\s+(\w+)@(\w+))?\s+(\w+):\s+(.+)$`)

func parseStderrLogFile(fpath string) ([]models.LogEntry, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []models.LogEntry
	scanner := bufio.NewScanner(f)

	timeFormats := []string{
		"2006-01-02 15:04:05.999 MST",
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04:05.999",
		"2006-01-02 15:04:05",
	}

	var current *models.LogEntry

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		matches := stderrLogRe.FindStringSubmatch(line)
		if matches == nil {
			if current != nil {
				current.LogMessage += "\n" + line
			}
			continue
		}

		if current != nil {
			entries = append(entries, *current)
		}

		current = &models.LogEntry{}

		ts := strings.TrimSpace(matches[1])
		for _, fmt := range timeFormats {
			if t, err := time.Parse(fmt, ts); err == nil {
				current.Timestamp = t
				break
			}
		}

		if pid, err := strconv.Atoi(matches[2]); err == nil {
			current.ProcessID = pid
		}

		if matches[3] != "" {
			current.User = matches[3]
		}
		if matches[4] != "" {
			current.Database = matches[4]
		}

		current.Severity = matches[5]
		current.LogMessage = matches[6]
	}

	if current != nil {
		entries = append(entries, *current)
	}

	return entries, scanner.Err()
}

func filterLogEntries(entries []models.LogEntry, q models.LogQuery) []models.LogEntry {
	var result []models.LogEntry
	for _, e := range entries {
		if q.Severity != "" && !strings.EqualFold(e.Severity, q.Severity) {
			continue
		}
		if q.Database != "" && !strings.EqualFold(e.Database, q.Database) {
			continue
		}
		if q.User != "" && !strings.EqualFold(e.User, q.User) {
			continue
		}
		if q.StartTime != nil && e.Timestamp.Before(*q.StartTime) {
			continue
		}
		if q.EndTime != nil && e.Timestamp.After(*q.EndTime) {
			continue
		}
		if q.Search != "" {
			search := strings.ToLower(q.Search)
			matched := strings.Contains(strings.ToLower(e.LogMessage), search) ||
				strings.Contains(strings.ToLower(e.Query), search) ||
				strings.Contains(strings.ToLower(e.Detail), search) ||
				strings.Contains(strings.ToLower(e.Context), search)
			if !matched {
				continue
			}
		}
		result = append(result, e)
	}
	return result
}

func filterConnectionEntries(entries []models.LogEntry) []models.LogEntry {
	var result []models.LogEntry
	for _, e := range entries {
		msg := strings.ToLower(e.LogMessage)
		if strings.Contains(msg, "connection") ||
			strings.Contains(msg, "connect") ||
			strings.Contains(msg, "disconnect") ||
			strings.Contains(msg, "authentication") ||
			e.Severity == "FATAL" ||
			e.Severity == "PANIC" {
			result = append(result, e)
		}
	}
	return result
}

func sortLogEntriesDesc(entries []models.LogEntry) {
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Timestamp.After(entries[i].Timestamp) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}

func (s *LogsService) paginate(entries []models.LogEntry, limit, offset int) *models.LogResponse {
	total := len(entries)
	if offset >= total {
		return &models.LogResponse{
			Entries: []models.LogEntry{},
			Total:   total,
			Limit:   limit,
			Offset:  offset,
		}
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return &models.LogResponse{
		Entries: entries[offset:end],
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}
}

func mapStatusToSeverity(status string) string {
	switch strings.ToLower(status) {
	case "error":
		return "ERROR"
	case "success":
		return "LOG"
	default:
		return "LOG"
	}
}
