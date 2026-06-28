package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/internal/tables/models"
	"github.com/nexbic/platform/pkg/database"
)

type TablesService struct {
	db *database.DB
}

func NewTablesService(db *database.DB) *TablesService {
	return &TablesService{db: db}
}

func quoteIdent(parts ...string) string {
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = `"` + strings.ReplaceAll(p, `"`, `""`) + `"`
	}
	return strings.Join(quoted, ".")
}

func buildFilterClause(f models.Filter, startIdx int) (string, []any, int, error) {
	col := quoteIdent(f.Column)
	idx := startIdx

	switch f.Operator {
	case "eq", "=", "":
		return fmt.Sprintf("%s = $%d", col, idx), []any{f.Value}, idx + 1, nil
	case "neq", "!=", "<>":
		return fmt.Sprintf("%s != $%d", col, idx), []any{f.Value}, idx + 1, nil
	case "gt", ">":
		return fmt.Sprintf("%s > $%d", col, idx), []any{f.Value}, idx + 1, nil
	case "gte", ">=":
		return fmt.Sprintf("%s >= $%d", col, idx), []any{f.Value}, idx + 1, nil
	case "lt", "<":
		return fmt.Sprintf("%s < $%d", col, idx), []any{f.Value}, idx + 1, nil
	case "lte", "<=":
		return fmt.Sprintf("%s <= $%d", col, idx), []any{f.Value}, idx + 1, nil
	case "like":
		return fmt.Sprintf("%s LIKE $%d", col, idx), []any{f.Value}, idx + 1, nil
	case "ilike":
		return fmt.Sprintf("%s ILIKE $%d", col, idx), []any{f.Value}, idx + 1, nil
	case "is_null":
		return fmt.Sprintf("%s IS NULL", col), nil, idx, nil
	case "is_not_null":
		return fmt.Sprintf("%s IS NOT NULL", col), nil, idx, nil
	case "in":
		vals, ok := f.Value.([]any)
		if !ok {
			return "", nil, 0, fmt.Errorf("IN operator requires an array value")
		}
		placeholders := make([]string, len(vals))
		args := make([]any, len(vals))
		for i, v := range vals {
			placeholders[i] = fmt.Sprintf("$%d", idx+i)
			args[i] = v
		}
		return fmt.Sprintf("%s IN (%s)", col, strings.Join(placeholders, ", ")), args, idx + len(vals), nil
	default:
		return "", nil, 0, fmt.Errorf("unsupported filter operator: %s", f.Operator)
	}
}

func (s *TablesService) QueryTable(ctx context.Context, schema, table string, limit, offset int, sort, order string, filters []models.Filter) ([]map[string]any, int, error) {
	whereClauses := make([]string, 0)
	var args []any
	argIdx := 1

	for _, f := range filters {
		clause, fArgs, nextIdx, err := buildFilterClause(f, argIdx)
		if err != nil {
			return nil, 0, err
		}
		whereClauses = append(whereClauses, clause)
		args = append(args, fArgs...)
		argIdx = nextIdx
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", quoteIdent(schema, table), whereSQL)
	var total int
	if err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	orderClause := ""
	if sort != "" {
		dir := "ASC"
		if strings.EqualFold(order, "desc") {
			dir = "DESC"
		}
		orderClause = fmt.Sprintf(" ORDER BY %s %s", quoteIdent(sort), dir)
	}

	dataQuery := fmt.Sprintf("SELECT * FROM %s%s%s LIMIT $%d OFFSET $%d",
		quoteIdent(schema, table), whereSQL, orderClause, argIdx, argIdx+1)

	args = append(args, limit, offset)
	rows, err := s.db.Pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	result, err := rowsToMaps(rows)
	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (s *TablesService) InsertRow(ctx context.Context, schema, table string, data map[string]any) (map[string]any, error) {
	cols := make([]string, 0, len(data))
	vals := make([]any, 0, len(data))
	placeholders := make([]string, 0, len(data))
	i := 1
	for k, v := range data {
		cols = append(cols, quoteIdent(k))
		vals = append(vals, v)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		quoteIdent(schema, table),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "))

	rows, err := s.db.Pool.Query(ctx, query, vals...)
	if err != nil {
		return nil, fmt.Errorf("insert: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		fieldDescs := rows.FieldDescriptions()
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		row := make(map[string]any)
		for i, fd := range fieldDescs {
			row[string(fd.Name)] = values[i]
		}
		return row, nil
	}

	return nil, fmt.Errorf("insert succeeded but no rows returned")
}

func (s *TablesService) UpdateRow(ctx context.Context, schema, table string, data, where map[string]any) ([]map[string]any, error) {
	setClauses := make([]string, 0, len(data))
	args := make([]any, 0, len(data)+len(where))
	i := 1
	for k, v := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", quoteIdent(k), i))
		args = append(args, v)
		i++
	}

	whereClauses := make([]string, 0, len(where))
	for k, v := range where {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", quoteIdent(k), i))
		args = append(args, v)
		i++
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s RETURNING *",
		quoteIdent(schema, table),
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "))

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update: %w", err)
	}
	defer rows.Close()

	result, err := rowsToMaps(rows)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *TablesService) DeleteRow(ctx context.Context, schema, table string, where map[string]any) (int64, error) {
	args := make([]any, 0, len(where))
	whereClauses := make([]string, 0, len(where))
	i := 1
	for k, v := range where {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", quoteIdent(k), i))
		args = append(args, v)
		i++
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s",
		quoteIdent(schema, table),
		strings.Join(whereClauses, " AND "))

	ct, err := s.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("delete: %w", err)
	}

	return ct.RowsAffected(), nil
}

func (s *TablesService) BulkInsert(ctx context.Context, schema, table string, rows []map[string]any) ([]map[string]any, error) {
	if len(rows) == 0 {
		return []map[string]any{}, nil
	}

	allCols := make([]string, 0)
	colSet := make(map[string]bool)
	for _, row := range rows {
		for k := range row {
			if !colSet[k] {
				colSet[k] = true
				allCols = append(allCols, k)
			}
		}
	}

	var results []map[string]any

	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		results = make([]map[string]any, 0, len(rows))

		for _, row := range rows {
			vals := make([]any, 0, len(allCols))
			placeholders := make([]string, 0, len(allCols))
			i := 1
			for _, col := range allCols {
				placeholders = append(placeholders, fmt.Sprintf("$%d", i))
				if v, ok := row[col]; ok {
					vals = append(vals, v)
				} else {
					vals = append(vals, nil)
				}
				i++
			}

			quotedCols := make([]string, len(allCols))
			for j, col := range allCols {
				quotedCols[j] = quoteIdent(col)
			}

			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING *",
				quoteIdent(schema, table),
				strings.Join(quotedCols, ", "),
				strings.Join(placeholders, ", "))

			txRows, err := tx.Query(ctx, query, vals...)
			if err != nil {
				return fmt.Errorf("bulk insert row: %w", err)
			}

			if txRows.Next() {
				values, err := txRows.Values()
				if err != nil {
					txRows.Close()
					return fmt.Errorf("read row: %w", err)
				}
				fieldDescs := txRows.FieldDescriptions()
				rowMap := make(map[string]any)
				for i, fd := range fieldDescs {
					rowMap[string(fd.Name)] = values[i]
				}
				results = append(results, rowMap)
			}
			txRows.Close()
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *TablesService) BulkDelete(ctx context.Context, schema, table string, ids []any, idColumn string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s IN (%s)",
		quoteIdent(schema, table),
		quoteIdent(idColumn),
		strings.Join(placeholders, ", "))

	ct, err := s.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("bulk delete: %w", err)
	}

	return ct.RowsAffected(), nil
}

func (s *TablesService) CountRows(ctx context.Context, schema, table string, filters []models.Filter) (int, error) {
	whereClauses := make([]string, 0)
	var args []any
	argIdx := 1

	for _, f := range filters {
		clause, fArgs, nextIdx, err := buildFilterClause(f, argIdx)
		if err != nil {
			return 0, err
		}
		whereClauses = append(whereClauses, clause)
		args = append(args, fArgs...)
		argIdx = nextIdx
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", quoteIdent(schema, table))
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	if err := s.db.Pool.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count rows: %w", err)
	}

	return total, nil
}

func (s *TablesService) GetPrimaryKeyColumns(ctx context.Context, schema, table string) ([]string, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.constraint_type = 'PRIMARY KEY'
			AND tc.table_schema = $1
			AND tc.table_name = $2
		ORDER BY kcu.ordinal_position`, schema, table)
	if err != nil {
		return nil, fmt.Errorf("get primary keys: %w", err)
	}
	defer rows.Close()

	var pkCols []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, fmt.Errorf("scan pk column: %w", err)
		}
		pkCols = append(pkCols, col)
	}

	if pkCols == nil {
		pkCols = []string{}
	}

	return pkCols, nil
}

func (s *TablesService) SearchTable(ctx context.Context, schema, table, search string, columns []string, limit int) ([]map[string]any, error) {
	if search == "" {
		return []map[string]any{}, nil
	}
	if len(columns) == 0 {
		return []map[string]any{}, nil
	}

	likeClauses := make([]string, len(columns))
	args := make([]any, len(columns)+1)
	for i, col := range columns {
		likeClauses[i] = fmt.Sprintf("%s::text ILIKE $%d", quoteIdent(col), i+1)
		args[i] = "%" + search + "%"
	}
	args[len(columns)] = limit

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT $%d",
		quoteIdent(schema, table),
		strings.Join(likeClauses, " OR "),
		len(columns)+1)

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer rows.Close()

	result, err := rowsToMaps(rows)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func rowsToMaps(rows pgx.Rows) ([]map[string]any, error) {
	fieldDescs := rows.FieldDescriptions()
	var result []map[string]any

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		row := make(map[string]any)
		for i, fd := range fieldDescs {
			row[string(fd.Name)] = values[i]
		}
		result = append(result, row)
	}

	if result == nil {
		result = []map[string]any{}
	}

	return result, nil
}
