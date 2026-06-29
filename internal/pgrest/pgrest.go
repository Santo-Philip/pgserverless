package pgrest

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type tableMeta struct {
	Schema    string
	Name      string
	IsView    bool
	Columns   []columnMeta
	PKColumns []string
}

type columnMeta struct {
	Name       string
	DataType   string
	IsNullable bool
}

type filter struct {
	Column   string
	Operator string // eq neq gt gte lt lte like ilike in is isnot
	Values   []string
}

type orderClause struct {
	Column string
	Dir    string
}

type queryParams struct {
	Select  []string
	Filters []filter
	Order   []orderClause
	Limit   int
	Offset  int
}

type PGREST struct {
	pool   *pgxpool.Pool
	tables []*tableMeta
	lookup map[string]*tableMeta
}

func New(pool *pgxpool.Pool) (*PGREST, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tables, err := discoverTables(ctx, pool)
	if err != nil {
		return nil, fmt.Errorf("pgrest init: %w", err)
	}

	lookup := make(map[string]*tableMeta, len(tables))
	for _, t := range tables {
		key := t.Schema + "." + t.Name
		lookup[key] = t
	}

	return &PGREST{pool: pool, tables: tables, lookup: lookup}, nil
}

func (p *PGREST) Tables() []*tableMeta { return p.tables }

func (p *PGREST) Lookup(schema, name string) *tableMeta {
	return p.lookup[schema+"."+name]
}

func (p *PGREST) RegisterRoutes(api fiber.Router, authMW fiber.Handler) {
	g := api.Group("/r", authMW)

	g.Get("/:schema/:table", p.handleList)
	g.Post("/:schema/:table", p.handleInsert)
	g.Patch("/:schema/:table", p.handleUpdate)
	g.Delete("/:schema/:table", p.handleDelete)
	g.Get("/:schema/:table/:pk", p.handleGet)
	g.Patch("/:schema/:table/:pk", p.handleUpdatePK)
	g.Delete("/:schema/:table/:pk", p.handleDeletePK)
}

func (p *PGREST) resolve(c *fiber.Ctx) (*tableMeta, error) {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return nil, response.BadRequest(c, "schema and table are required")
	}
	t := p.Lookup(schema, table)
	if t == nil {
		return nil, response.NotFound(c, fmt.Sprintf("table %s.%s not found", schema, table))
	}
	return t, nil
}

func (p *PGREST) handleList(c *fiber.Ctx) error {
	t, err := p.resolve(c)
	if err != nil {
		return err
	}

	q := parseQueryParams(c)
	rows, total, err := p.queryRows(c.Context(), t, q)
	if err != nil {
		return response.BadRequest(c, "query failed: "+err.Error())
	}
	return response.Paginated(c, rows, total, q.Limit, q.Offset)
}

func (p *PGREST) handleGet(c *fiber.Ctx) error {
	t, err := p.resolve(c)
	if err != nil {
		return err
	}
	if len(t.PKColumns) == 0 {
		return response.NotFound(c, "table has no primary key")
	}

	pkVal := c.Params("pk")
	q := &queryParams{Limit: 1}
	for _, pkCol := range t.PKColumns {
		q.Filters = append(q.Filters, filter{Column: pkCol, Operator: "eq", Values: []string{pkVal}})
	}

	rows, _, err := p.queryRows(c.Context(), t, q)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	if len(rows) == 0 {
		return response.NotFound(c, "row not found")
	}
	return response.OK(c, rows[0])
}

func (p *PGREST) handleInsert(c *fiber.Ctx) error {
	t, err := p.resolve(c)
	if err != nil {
		return err
	}
	if t.IsView {
		return response.BadRequest(c, "cannot insert into a view")
	}

	var body []map[string]any
	if err := c.BodyParser(&body); err != nil {
		var single map[string]any
		if err2 := c.BodyParser(&single); err2 != nil {
			return response.BadRequest(c, "invalid request body: expected object or array")
		}
		body = []map[string]any{single}
	}
	if len(body) == 0 {
		return response.BadRequest(c, "body is required")
	}

	if len(body) == 1 {
		row, err := p.insertRow(c.Context(), t, body[0])
		if err != nil {
			return response.BadRequest(c, err.Error())
		}
		return response.Created(c, row)
	}

	results, err := p.bulkInsert(c.Context(), t, body)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, results)
}

func (p *PGREST) handleUpdate(c *fiber.Ctx) error {
	t, err := p.resolve(c)
	if err != nil {
		return err
	}
	if t.IsView {
		return response.BadRequest(c, "cannot update a view")
	}

	var data map[string]any
	if err := c.BodyParser(&data); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if len(data) == 0 {
		return response.BadRequest(c, "update data is required")
	}

	q := parseQueryParams(c)
	if len(q.Filters) == 0 {
		return response.BadRequest(c, "filters required (use query params like ?col.eq=val)")
	}

	count, err := p.updateRows(c.Context(), t, data, q.Filters)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"updated": count})
}

func (p *PGREST) handleDelete(c *fiber.Ctx) error {
	t, err := p.resolve(c)
	if err != nil {
		return err
	}
	if t.IsView {
		return response.BadRequest(c, "cannot delete from a view")
	}

	var body []string
	if err := c.BodyParser(&body); err == nil && len(body) > 0 && len(t.PKColumns) > 0 {
		count, err := p.bulkDelete(c.Context(), t, body, t.PKColumns[0])
		if err != nil {
			return response.BadRequest(c, err.Error())
		}
		return response.OK(c, fiber.Map{"deleted": count})
	}

	q := parseQueryParams(c)
	if len(q.Filters) == 0 {
		return response.BadRequest(c, "filters or id array required")
	}

	count, err := p.deleteRows(c.Context(), t, q.Filters)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"deleted": count})
}

func (p *PGREST) handleUpdatePK(c *fiber.Ctx) error {
	t, err := p.resolve(c)
	if err != nil {
		return err
	}
	if t.IsView {
		return response.BadRequest(c, "cannot update a view")
	}
	if len(t.PKColumns) == 0 {
		return response.BadRequest(c, "table has no primary key")
	}

	var data map[string]any
	if err := c.BodyParser(&data); err != nil || len(data) == 0 {
		return response.BadRequest(c, "update data is required")
	}

	pkVal := c.Params("pk")
	filters := []filter{{Column: t.PKColumns[0], Operator: "eq", Values: []string{pkVal}}}
	count, err := p.updateRows(c.Context(), t, data, filters)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"updated": count})
}

func (p *PGREST) handleDeletePK(c *fiber.Ctx) error {
	t, err := p.resolve(c)
	if err != nil {
		return err
	}
	if t.IsView {
		return response.BadRequest(c, "cannot delete from a view")
	}
	if len(t.PKColumns) == 0 {
		return response.BadRequest(c, "table has no primary key")
	}
	pkVal := c.Params("pk")
	count, err := p.deleteByPK(c.Context(), t, pkVal)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"deleted": count})
}

func parseQueryParams(c *fiber.Ctx) *queryParams {
	q := &queryParams{Limit: 50, Offset: 0}

	q.Limit = c.QueryInt("limit", 50)
	if q.Limit < 1 {
		q.Limit = 50
	}
	if q.Limit > 1000 {
		q.Limit = 1000
	}
	q.Offset = c.QueryInt("offset", 0)
	if q.Offset < 0 {
		q.Offset = 0
	}

	if pg := c.QueryInt("page", 0); pg > 0 {
		q.Offset = (pg - 1) * q.Limit
	}

	if s := c.Query("select"); s != "" {
		q.Select = strings.Split(s, ",")
	}

	if o := c.Query("order"); o != "" {
		parts := strings.SplitN(o, ".", 2)
		if len(parts) == 2 {
			dir := "asc"
			if strings.ToLower(parts[1]) == "desc" {
				dir = "desc"
			}
			q.Order = append(q.Order, orderClause{Column: parts[0], Dir: dir})
		}
	}

	c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
		k := string(key)
		v := string(value)
		if k == "select" || k == "limit" || k == "offset" || k == "page" || k == "order" {
			return
		}
		parts := strings.SplitN(k, ".", 2)
		if len(parts) != 2 {
			return
		}
		col, op := parts[0], parts[1]
		switch op {
		case "eq", "neq", "gt", "gte", "lt", "lte", "like", "ilike":
			q.Filters = append(q.Filters, filter{Column: col, Operator: op, Values: []string{v}})
		case "in":
			q.Filters = append(q.Filters, filter{Column: col, Operator: "in", Values: strings.Split(v, ",")})
		case "is":
			q.Filters = append(q.Filters, filter{Column: col, Operator: "is", Values: []string{v}})
		case "isnot":
			q.Filters = append(q.Filters, filter{Column: col, Operator: "isnot", Values: []string{v}})
		}
	})

	return q
}

func (p *PGREST) queryRows(ctx context.Context, t *tableMeta, q *queryParams) ([]map[string]any, int, error) {
	countSQL, countArgs := buildCountSQL(t, q)
	var total int
	if err := p.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	dataSQL, dataArgs := buildSelectSQL(t, q)
	rows, err := p.pool.Query(ctx, dataSQL, dataArgs...)
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

func (p *PGREST) insertRow(ctx context.Context, t *tableMeta, data map[string]any) (map[string]any, error) {
	cols := make([]string, 0, len(data))
	vals := make([]any, 0, len(data))
	phs := make([]string, 0, len(data))
	i := 1
	for k, v := range data {
		cols = append(cols, helpers.QuoteIdent(k))
		vals = append(vals, v)
		phs = append(phs, "$"+strconv.Itoa(i))
		i++
	}

	colsStr := strings.Join(cols, ", ")
	phsStr := strings.Join(phs, ", ")
	schemaTable := helpers.QuoteIdent(t.Schema, t.Name)

	rows, err := p.pool.Query(ctx, fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING *", schemaTable, colsStr, phsStr), vals...)
	if err != nil {
		return nil, fmt.Errorf("insert: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		fds := rows.FieldDescriptions()
		row := make(map[string]any, len(fds))
		for i, fd := range fds {
			row[string(fd.Name)] = values[i]
		}
		return row, nil
	}
	return nil, fmt.Errorf("insert succeeded but no rows returned")
}

func (p *PGREST) bulkInsert(ctx context.Context, t *tableMeta, rows []map[string]any) ([]map[string]any, error) {
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

	quotedCols := make([]string, len(allCols))
	for j, col := range allCols {
		quotedCols[j] = helpers.QuoteIdent(col)
	}
	colsStr := strings.Join(quotedCols, ", ")
	schemaTable := helpers.QuoteIdent(t.Schema, t.Name)
	results := make([]map[string]any, 0, len(rows))

	err := p.withTx(ctx, func(tx pgx.Tx) error {
		for _, row := range rows {
			vals := make([]any, 0, len(allCols))
			phs := make([]string, 0, len(allCols))
			for idx, col := range allCols {
				phs = append(phs, "$"+strconv.Itoa(idx+1))
				if v, ok := row[col]; ok {
					vals = append(vals, v)
				} else {
					vals = append(vals, nil)
				}
			}
			phsStr := strings.Join(phs, ", ")

			txRows, err := tx.Query(ctx, fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING *", schemaTable, colsStr, phsStr), vals...)
			if err != nil {
				return fmt.Errorf("bulk insert: %w", err)
			}

			if txRows.Next() {
				values, err := txRows.Values()
				if err != nil {
					txRows.Close()
					return fmt.Errorf("read row: %w", err)
				}
				fds := txRows.FieldDescriptions()
				rowMap := make(map[string]any, len(fds))
				for i, fd := range fds {
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

func (p *PGREST) updateRows(ctx context.Context, t *tableMeta, data map[string]any, filters []filter) (int64, error) {
	schemaTable := helpers.QuoteIdent(t.Schema, t.Name)
	args := make([]any, 0, len(data)+len(filters))
	setClauses := make([]string, 0, len(data))
	idx := 1

	for k, v := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", helpers.QuoteIdent(k), idx))
		args = append(args, v)
		idx++
	}

	whereSQL, whereArgs := buildWhereClause(filters, idx)
	args = append(args, whereArgs...)

	query := fmt.Sprintf("UPDATE %s SET %s%s", schemaTable, strings.Join(setClauses, ", "), whereSQL)
	ct, err := p.pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("update: %w", err)
	}
	return ct.RowsAffected(), nil
}

func (p *PGREST) deleteRows(ctx context.Context, t *tableMeta, filters []filter) (int64, error) {
	schemaTable := helpers.QuoteIdent(t.Schema, t.Name)
	whereSQL, whereArgs := buildWhereClause(filters, 1)

	query := fmt.Sprintf("DELETE FROM %s%s", schemaTable, whereSQL)
	ct, err := p.pool.Exec(ctx, query, whereArgs...)
	if err != nil {
		return 0, fmt.Errorf("delete: %w", err)
	}
	return ct.RowsAffected(), nil
}

func (p *PGREST) deleteByPK(ctx context.Context, t *tableMeta, pkVal string) (int64, error) {
	if len(t.PKColumns) == 0 {
		return 0, fmt.Errorf("no primary key")
	}
	schemaTable := helpers.QuoteIdent(t.Schema, t.Name)
	pkCol := helpers.QuoteIdent(t.PKColumns[0])
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", schemaTable, pkCol)
	ct, err := p.pool.Exec(ctx, query, pkVal)
	if err != nil {
		return 0, fmt.Errorf("delete by pk: %w", err)
	}
	return ct.RowsAffected(), nil
}

func (p *PGREST) bulkDelete(ctx context.Context, t *tableMeta, ids []string, idColumn string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	schemaTable := helpers.QuoteIdent(t.Schema, t.Name)
	idCol := helpers.QuoteIdent(idColumn)
	phs := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		phs[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s IN (%s)", schemaTable, idCol, strings.Join(phs, ","))
	ct, err := p.pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("bulk delete: %w", err)
	}
	return ct.RowsAffected(), nil
}

func (p *PGREST) withTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func buildSelectSQL(t *tableMeta, q *queryParams) (string, []any) {
	selectCols := "*"
	if len(q.Select) > 0 {
		quoted := make([]string, len(q.Select))
		for i, col := range q.Select {
			quoted[i] = helpers.QuoteIdent(col)
		}
		selectCols = strings.Join(quoted, ", ")
	}

	whereSQL, whereArgs := buildWhereClause(q.Filters, 1)
	schemaTable := helpers.QuoteIdent(t.Schema, t.Name)

	orderSQL := ""
	if len(q.Order) > 0 {
		parts := make([]string, len(q.Order))
		for i, o := range q.Order {
			parts[i] = fmt.Sprintf("%s %s", helpers.QuoteIdent(o.Column), strings.ToUpper(o.Dir))
		}
		orderSQL = " ORDER BY " + strings.Join(parts, ", ")
	}

	argIdx := 1 + len(whereArgs)
	sql := fmt.Sprintf("SELECT %s FROM %s%s%s LIMIT $%d OFFSET $%d",
		selectCols, schemaTable, whereSQL, orderSQL, argIdx, argIdx+1)

	args := make([]any, 0, len(whereArgs)+2)
	args = append(args, whereArgs...)
	args = append(args, q.Limit, q.Offset)
	return sql, args
}

func buildCountSQL(t *tableMeta, q *queryParams) (string, []any) {
	whereSQL, whereArgs := buildWhereClause(q.Filters, 1)
	schemaTable := helpers.QuoteIdent(t.Schema, t.Name)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s%s", schemaTable, whereSQL), whereArgs
}

func buildWhereClause(filters []filter, startIdx int) (string, []any) {
	if len(filters) == 0 {
		return "", nil
	}

	clauses := make([]string, 0, len(filters))
	args := make([]any, 0)
	idx := startIdx

	for _, f := range filters {
		col := helpers.QuoteIdent(f.Column)
		switch f.Operator {
		case "eq":
			clauses = append(clauses, fmt.Sprintf("%s = $%d", col, idx))
			args = append(args, f.Values[0])
			idx++
		case "neq":
			clauses = append(clauses, fmt.Sprintf("%s != $%d", col, idx))
			args = append(args, f.Values[0])
			idx++
		case "gt":
			clauses = append(clauses, fmt.Sprintf("%s > $%d", col, idx))
			args = append(args, f.Values[0])
			idx++
		case "gte":
			clauses = append(clauses, fmt.Sprintf("%s >= $%d", col, idx))
			args = append(args, f.Values[0])
			idx++
		case "lt":
			clauses = append(clauses, fmt.Sprintf("%s < $%d", col, idx))
			args = append(args, f.Values[0])
			idx++
		case "lte":
			clauses = append(clauses, fmt.Sprintf("%s <= $%d", col, idx))
			args = append(args, f.Values[0])
			idx++
		case "like":
			clauses = append(clauses, fmt.Sprintf("%s LIKE $%d", col, idx))
			args = append(args, f.Values[0])
			idx++
		case "ilike":
			clauses = append(clauses, fmt.Sprintf("%s ILIKE $%d", col, idx))
			args = append(args, f.Values[0])
			idx++
		case "in":
			phs := make([]string, len(f.Values))
			for i, v := range f.Values {
				phs[i] = "$" + strconv.Itoa(idx)
				args = append(args, v)
				idx++
			}
			clauses = append(clauses, fmt.Sprintf("%s IN (%s)", col, strings.Join(phs, ",")))
		case "is":
			if slices.Contains(f.Values, "null") || len(f.Values) == 0 || strings.ToLower(f.Values[0]) == "null" {
				clauses = append(clauses, fmt.Sprintf("%s IS NULL", col))
			} else {
				clauses = append(clauses, fmt.Sprintf("%s IS NOT NULL", col))
			}
		case "isnot":
			clauses = append(clauses, fmt.Sprintf("%s IS NOT NULL", col))
		}
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

func discoverTables(ctx context.Context, pool *pgxpool.Pool) ([]*tableMeta, error) {
	rows, err := pool.Query(ctx, `
		SELECT table_schema, table_name, table_type
		FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema', '_timescaledb_catalog', '_timescaledb_config')
		  AND table_type IN ('BASE TABLE', 'VIEW')
		ORDER BY table_schema, table_name
	`)
	if err != nil {
		return nil, fmt.Errorf("discover tables: %w", err)
	}
	defer rows.Close()

	var tables []*tableMeta
	for rows.Next() {
		var t tableMeta
		var tt string
		if err := rows.Scan(&t.Schema, &t.Name, &tt); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}
		t.IsView = tt == "VIEW"
		tables = append(tables, &t)
	}

	for _, t := range tables {
		cols, err := pool.Query(ctx, `
			SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_schema = $1 AND table_name = $2
			ORDER BY ordinal_position
		`, t.Schema, t.Name)
		if err != nil {
			return nil, fmt.Errorf("discover columns %s.%s: %w", t.Schema, t.Name, err)
		}
		for cols.Next() {
			var c columnMeta
			var nullable string
			if err := cols.Scan(&c.Name, &c.DataType, &nullable); err != nil {
				cols.Close()
				return nil, fmt.Errorf("scan column: %w", err)
			}
			c.IsNullable = nullable == "YES"
			t.Columns = append(t.Columns, c)
		}
		cols.Close()

		pkRows, err := pool.Query(ctx, `
			SELECT kcu.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
				AND tc.table_name = kcu.table_name
			WHERE tc.table_schema = $1
			  AND tc.table_name = $2
			  AND tc.constraint_type = 'PRIMARY KEY'
			ORDER BY kcu.ordinal_position
		`, t.Schema, t.Name)
		if err != nil {
			return nil, fmt.Errorf("discover PKs %s.%s: %w", t.Schema, t.Name, err)
		}
		for pkRows.Next() {
			var pk string
			if err := pkRows.Scan(&pk); err != nil {
				pkRows.Close()
				return nil, fmt.Errorf("scan PK: %w", err)
			}
			t.PKColumns = append(t.PKColumns, pk)
		}
		pkRows.Close()
	}

	return tables, nil
}

func rowsToMaps(rows pgx.Rows) ([]map[string]any, error) {
	fds := rows.FieldDescriptions()
	var result []map[string]any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		row := make(map[string]any, len(fds))
		for i, fd := range fds {
			row[string(fd.Name)] = values[i]
		}
		result = append(result, row)
	}
	if result == nil {
		result = []map[string]any{}
	}
	return result, nil
}
