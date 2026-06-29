package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/database/sql/models"
	"github.com/nexbic/platform/pkg/database"
)

type SQLService struct {
	db           *database.DB
	queryTimeout time.Duration
}

func NewSQLService(db *database.DB) *SQLService {
	return &SQLService{
		db:           db,
		queryTimeout: 30 * time.Second,
	}
}

func isQueryStmt(query string) bool {
	trimmed := strings.TrimSpace(strings.ToUpper(query))
	if strings.HasPrefix(trimmed, "SELECT") ||
		strings.HasPrefix(trimmed, "WITH") ||
		strings.HasPrefix(trimmed, "EXPLAIN") ||
		strings.HasPrefix(trimmed, "SHOW") ||
		strings.HasPrefix(trimmed, "VALUES") ||
		strings.HasPrefix(trimmed, "TABLE") {
		return true
	}
	upper := strings.ToUpper(query)
	if (strings.HasPrefix(trimmed, "INSERT") ||
		strings.HasPrefix(trimmed, "UPDATE") ||
		strings.HasPrefix(trimmed, "DELETE") ||
		strings.HasPrefix(trimmed, "MERGE")) &&
		strings.Contains(upper, "RETURNING") {
		return true
	}
	return false
}

func (s *SQLService) ExecuteSQL(ctx context.Context, query string, params []any) (*models.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	if isQueryStmt(query) {
		rows, err := s.db.Pool.Query(ctx, query, params...)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		defer rows.Close()

		fields := rows.FieldDescriptions()
		columns := make([]string, len(fields))
		for i, f := range fields {
			columns[i] = string(f.Name)
		}

		var resultRows []map[string]any
		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				return nil, fmt.Errorf("row scan failed: %w", err)
			}
			row := make(map[string]any)
			for i, fd := range fields {
				row[string(fd.Name)] = values[i]
			}
			resultRows = append(resultRows, row)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("rows iteration failed: %w", err)
		}

		if resultRows == nil {
			resultRows = []map[string]any{}
		}

		return &models.QueryResult{
			Columns: columns,
			Rows:    resultRows,
		}, nil
	}

	tag, err := s.db.Pool.Exec(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("exec failed: %w", err)
	}

	return &models.QueryResult{
		RowsAffected: tag.RowsAffected(),
	}, nil
}

func (s *SQLService) ExplainQuery(ctx context.Context, query string) (*models.QueryResult, error) {
	explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", query)
	ctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	rows, err := s.db.Pool.Query(ctx, explainQuery)
	if err != nil {
		return nil, fmt.Errorf("explain failed: %w", err)
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, f := range fields {
		columns[i] = string(f.Name)
	}

	var resultRows []map[string]any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		row := make(map[string]any)
		for i, fd := range fields {
			row[string(fd.Name)] = values[i]
		}
		resultRows = append(resultRows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	if resultRows == nil {
		resultRows = []map[string]any{}
	}

	return &models.QueryResult{
		Columns: columns,
		Rows:    resultRows,
	}, nil
}

func (s *SQLService) CancelQuery(ctx context.Context, pid int) error {
	ctx, cancel := context.WithTimeout(ctx, s.queryTimeout)
	defer cancel()

	_, err := s.db.Pool.Exec(ctx, "SELECT pg_cancel_backend($1)", pid)
	if err != nil {
		return fmt.Errorf("cancel query failed: %w", err)
	}
	return nil
}

func (s *SQLService) GetQueryHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.QueryHistory, int, error) {
	var total int
	err := s.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM query_history WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count query history: %w", err)
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, user_id, query_text, duration_ms, rows_affected, status, error_message, created_at
		FROM query_history
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	var history []models.QueryHistory
	for rows.Next() {
		var h models.QueryHistory
		if err := rows.Scan(&h.ID, &h.UserID, &h.QueryText, &h.DurationMs,
			&h.RowsAffected, &h.Status, &h.ErrorMessage, &h.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan history: %w", err)
		}
		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}

	if history == nil {
		history = []models.QueryHistory{}
	}

	return history, total, nil
}

func (s *SQLService) SaveQuery(ctx context.Context, userID uuid.UUID, name, query, description string, isShared bool) (*models.SavedQuery, error) {
	now := time.Now()
	q := &models.SavedQuery{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		QueryText:   query,
		Description: description,
		IsShared:    isShared,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO saved_queries (id, user_id, name, query_text, description, is_shared, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		q.ID, q.UserID, q.Name, q.QueryText, q.Description, q.IsShared, q.CreatedAt, q.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("save query: %w", err)
	}

	return q, nil
}

func (s *SQLService) GetSavedQueries(ctx context.Context, userID uuid.UUID) ([]models.SavedQuery, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, user_id, name, query_text, description, is_shared, created_at, updated_at
		FROM saved_queries
		WHERE user_id = $1 OR is_shared = true
		ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("get saved queries: %w", err)
	}
	defer rows.Close()

	var queries []models.SavedQuery
	for rows.Next() {
		var q models.SavedQuery
		if err := rows.Scan(&q.ID, &q.UserID, &q.Name, &q.QueryText,
			&q.Description, &q.IsShared, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan saved query: %w", err)
		}
		queries = append(queries, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	if queries == nil {
		queries = []models.SavedQuery{}
	}

	return queries, nil
}

func (s *SQLService) DeleteSavedQuery(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := s.db.Pool.Exec(ctx,
		`DELETE FROM saved_queries WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete saved query: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("saved query not found")
	}
	return nil
}

func (s *SQLService) LogQuery(ctx context.Context, userID uuid.UUID, query string, durationMs int64, rowsAffected int, status, errorMsg string) error {
	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO query_history (id, user_id, query_text, duration_ms, rows_affected, status, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.New(), userID, query, durationMs, rowsAffected, status, errorMsg, time.Now())
	if err != nil {
		return fmt.Errorf("log query: %w", err)
	}
	return nil
}
