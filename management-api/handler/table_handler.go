package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/management-api/repository"
	"github.com/nexbic/platform/shared/database"
	"github.com/nexbic/platform/shared/utils"
)

type TableHandler struct {
	db      *database.DB
	appRepo *repository.AppRepository
}

func NewTableHandler(db *database.DB, appRepo *repository.AppRepository) *TableHandler {
	return &TableHandler{db: db, appRepo: appRepo}
}

type TableColumn struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	IsPK     bool   `json:"is_pk"`
	DefaultValue string `json:"default_value,omitempty"`
}

type TableInfo struct {
	Name    string        `json:"name"`
	Columns []TableColumn `json:"columns"`
}

func (h *TableHandler) ListTables(c *fiber.Ctx) error {
	appID, err := utils.ParseUUIDParam(c, "id", "app")
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	app, err := h.appRepo.GetByID(c.Context(), appID)
	if err != nil || app == nil {
		return utils.NotFound(c, "app not found")
	}

	schemaName := app.SchemaName
	if err := validateTblIdent(schemaName); err != nil {
		return utils.BadRequest(c, "invalid schema name")
	}

	rows, err := h.db.Pool.Query(c.Context(), fmt.Sprintf(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = %s
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`, quoteLit(schemaName)))
	if err != nil {
		return utils.InternalError(c, "failed to list tables")
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		cols, err := h.listColumns(c.Context(), schemaName, tableName)
		if err != nil {
			cols = []TableColumn{}
		}

		tables = append(tables, TableInfo{
			Name:    tableName,
			Columns: cols,
		})
	}

	if tables == nil {
		tables = []TableInfo{}
	}

	return utils.OK(c, tables)
}

func (h *TableHandler) GetTableData(c *fiber.Ctx) error {
	appID, err := utils.ParseUUIDParam(c, "id", "app")
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	tableName := c.Params("table")
	if err := validateTblIdent(tableName); err != nil {
		return utils.BadRequest(c, "invalid table name")
	}

	app, err := h.appRepo.GetByID(c.Context(), appID)
	if err != nil || app == nil {
		return utils.NotFound(c, "app not found")
	}

	schemaName := app.SchemaName
	if err := validateTblIdent(schemaName); err != nil {
		return utils.BadRequest(c, "invalid schema name")
	}

	limit := 100
	if l := c.QueryInt("limit", 100); l > 0 && l <= 1000 {
		limit = l
	}

	query := fmt.Sprintf(`SELECT * FROM %s.%s LIMIT %d`,
		quoteIdent(schemaName),
		quoteIdent(tableName),
		limit,
	)

	rows, err := h.db.Pool.Query(context.Background(), query)
	if err != nil {
		return utils.InternalError(c, fmt.Sprintf("failed to query table: %s", err.Error()))
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = fd.Name
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}

	return utils.OK(c, results)
}

type CreateTableRequest struct {
	Name    string        `json:"name"`
	Columns []TableColumn `json:"columns"`
}

func (h *TableHandler) CreateTable(c *fiber.Ctx) error {
	appID, err := utils.ParseUUIDParam(c, "id", "app")
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	app, err := h.appRepo.GetByID(c.Context(), appID)
	if err != nil || app == nil {
		return utils.NotFound(c, "app not found")
	}

	var req CreateTableRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	if err := validateTblIdent(req.Name); err != nil {
		return utils.BadRequest(c, "invalid table name")
	}
	if len(req.Columns) == 0 {
		return utils.BadRequest(c, "at least one column is required")
	}

	schemaName := app.SchemaName
	if err := validateTblIdent(schemaName); err != nil {
		return utils.BadRequest(c, "invalid schema name")
	}

	var colDefs []string
	for _, col := range req.Columns {
		if err := validateTblIdent(col.Name); err != nil {
			return utils.BadRequest(c, fmt.Sprintf("invalid column name: %s", col.Name))
		}
		if err := validatePgType(col.Type); err != nil {
			return utils.BadRequest(c, fmt.Sprintf("invalid type for column %s: %s", col.Name, err.Error()))
		}

		def := fmt.Sprintf("%s %s", quoteIdent(col.Name), col.Type)
		if col.IsPK {
			def += " PRIMARY KEY"
		}
		if !col.Nullable && !col.IsPK {
			def += " NOT NULL"
		}
		if col.DefaultValue != "" {
			def += " DEFAULT " + col.DefaultValue
		}
		colDefs = append(colDefs, def)
	}

	sql := fmt.Sprintf("CREATE TABLE %s.%s (\n  %s\n)",
		quoteIdent(schemaName),
		quoteIdent(req.Name),
		strings.Join(colDefs, ",\n  "),
	)

	if _, err := h.db.Pool.Exec(c.Context(), sql); err != nil {
		return utils.InternalError(c, fmt.Sprintf("failed to create table: %s", err.Error()))
	}

	return utils.OK(c, fiber.Map{
		"message": "Table created successfully",
		"name":    req.Name,
		"schema":  schemaName,
	})
}

type AddColumnRequest struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value,omitempty"`
}

func (h *TableHandler) AddColumn(c *fiber.Ctx) error {
	appID, err := utils.ParseUUIDParam(c, "id", "app")
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	tableName := c.Params("table")
	if err := validateTblIdent(tableName); err != nil {
		return utils.BadRequest(c, "invalid table name")
	}

	app, err := h.appRepo.GetByID(c.Context(), appID)
	if err != nil || app == nil {
		return utils.NotFound(c, "app not found")
	}

	var req AddColumnRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	if err := validateTblIdent(req.Name); err != nil {
		return utils.BadRequest(c, "invalid column name")
	}
	if err := validatePgType(req.Type); err != nil {
		return utils.BadRequest(c, fmt.Sprintf("invalid type: %s", err.Error()))
	}

	schemaName := app.SchemaName
	if err := validateTblIdent(schemaName); err != nil {
		return utils.BadRequest(c, "invalid schema name")
	}

	def := fmt.Sprintf("%s %s", quoteIdent(req.Name), req.Type)
	if !req.Nullable {
		def += " NOT NULL"
	}
	if req.DefaultValue != "" {
		def += " DEFAULT " + req.DefaultValue
	}

	sql := fmt.Sprintf("ALTER TABLE %s.%s ADD COLUMN %s",
		quoteIdent(schemaName),
		quoteIdent(tableName),
		def,
	)

	if _, err := h.db.Pool.Exec(c.Context(), sql); err != nil {
		return utils.InternalError(c, fmt.Sprintf("failed to add column: %s", err.Error()))
	}

	return utils.OK(c, fiber.Map{"message": "Column added successfully"})
}

type InsertRowRequest struct {
	Values map[string]interface{} `json:"values"`
}

func (h *TableHandler) InsertRow(c *fiber.Ctx) error {
	appID, err := utils.ParseUUIDParam(c, "id", "app")
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	tableName := c.Params("table")
	if err := validateTblIdent(tableName); err != nil {
		return utils.BadRequest(c, "invalid table name")
	}

	app, err := h.appRepo.GetByID(c.Context(), appID)
	if err != nil || app == nil {
		return utils.NotFound(c, "app not found")
	}

	var req InsertRowRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	if len(req.Values) == 0 {
		return utils.BadRequest(c, "no values provided")
	}

	schemaName := app.SchemaName
	if err := validateTblIdent(schemaName); err != nil {
		return utils.BadRequest(c, "invalid schema name")
	}

	var cols []string
	var placeholders []string
	var args []interface{}
	i := 1
	for col, val := range req.Values {
		if err := validateTblIdent(col); err != nil {
			return utils.BadRequest(c, fmt.Sprintf("invalid column name: %s", col))
		}
		cols = append(cols, quoteIdent(col))
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, val)
		i++
	}

	sql := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)",
		quoteIdent(schemaName),
		quoteIdent(tableName),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	if _, err := h.db.Pool.Exec(c.Context(), sql, args...); err != nil {
		return utils.InternalError(c, fmt.Sprintf("failed to insert row: %s", err.Error()))
	}

	return utils.OK(c, fiber.Map{"message": "Row inserted successfully"})
}

type UpdateRowRequest struct {
	Values map[string]interface{} `json:"values"`
	Where  map[string]interface{} `json:"where"`
}

func (h *TableHandler) UpdateRow(c *fiber.Ctx) error {
	appID, err := utils.ParseUUIDParam(c, "id", "app")
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	tableName := c.Params("table")
	if err := validateTblIdent(tableName); err != nil {
		return utils.BadRequest(c, "invalid table name")
	}

	app, err := h.appRepo.GetByID(c.Context(), appID)
	if err != nil || app == nil {
		return utils.NotFound(c, "app not found")
	}

	var req UpdateRowRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	if len(req.Values) == 0 || len(req.Where) == 0 {
		return utils.BadRequest(c, "values and where clause are required")
	}

	schemaName := app.SchemaName
	if err := validateTblIdent(schemaName); err != nil {
		return utils.BadRequest(c, "invalid schema name")
	}

	var setClauses []string
	var args []interface{}
	i := 1
	for col, val := range req.Values {
		if err := validateTblIdent(col); err != nil {
			return utils.BadRequest(c, fmt.Sprintf("invalid column name: %s", col))
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", quoteIdent(col), i))
		args = append(args, val)
		i++
	}

	var whereClauses []string
	for col, val := range req.Where {
		if err := validateTblIdent(col); err != nil {
			return utils.BadRequest(c, fmt.Sprintf("invalid column name: %s", col))
		}
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", quoteIdent(col), i))
		args = append(args, val)
		i++
	}

	sql := fmt.Sprintf("UPDATE %s.%s SET %s WHERE %s",
		quoteIdent(schemaName),
		quoteIdent(tableName),
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "),
	)

	if _, err := h.db.Pool.Exec(c.Context(), sql, args...); err != nil {
		return utils.InternalError(c, fmt.Sprintf("failed to update row: %s", err.Error()))
	}

	return utils.OK(c, fiber.Map{"message": "Row updated successfully"})
}

type DeleteRowRequest struct {
	Where map[string]interface{} `json:"where"`
}

func (h *TableHandler) DeleteRow(c *fiber.Ctx) error {
	appID, err := utils.ParseUUIDParam(c, "id", "app")
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	tableName := c.Params("table")
	if err := validateTblIdent(tableName); err != nil {
		return utils.BadRequest(c, "invalid table name")
	}

	app, err := h.appRepo.GetByID(c.Context(), appID)
	if err != nil || app == nil {
		return utils.NotFound(c, "app not found")
	}

	var req DeleteRowRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	if len(req.Where) == 0 {
		return utils.BadRequest(c, "where clause is required")
	}

	schemaName := app.SchemaName
	if err := validateTblIdent(schemaName); err != nil {
		return utils.BadRequest(c, "invalid schema name")
	}

	var whereClauses []string
	var args []interface{}
	i := 1
	for col, val := range req.Where {
		if err := validateTblIdent(col); err != nil {
			return utils.BadRequest(c, fmt.Sprintf("invalid column name: %s", col))
		}
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", quoteIdent(col), i))
		args = append(args, val)
		i++
	}

	sql := fmt.Sprintf("DELETE FROM %s.%s WHERE %s",
		quoteIdent(schemaName),
		quoteIdent(tableName),
		strings.Join(whereClauses, " AND "),
	)

	if _, err := h.db.Pool.Exec(c.Context(), sql, args...); err != nil {
		return utils.InternalError(c, fmt.Sprintf("failed to delete row: %s", err.Error()))
	}

	return utils.OK(c, fiber.Map{"message": "Row deleted successfully"})
}

func (h *TableHandler) listColumns(ctx context.Context, schema, table string) ([]TableColumn, error) {
	rows, err := h.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT
			c.column_name,
			c.data_type,
			CASE WHEN c.is_nullable = 'YES' THEN true ELSE false END AS nullable,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END AS is_pk,
			COALESCE(c.column_default, '') AS default_value
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT kcu.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			WHERE tc.constraint_type = 'PRIMARY KEY'
			AND tc.table_schema = %s
			AND tc.table_name = %s
		) pk ON c.column_name = pk.column_name
		WHERE c.table_schema = %s AND c.table_name = %s
		AND c.ordinal_position IS NOT NULL
		ORDER BY c.ordinal_position
	`, quoteLit(schema), quoteLit(table), quoteLit(schema), quoteLit(table)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []TableColumn
	for rows.Next() {
		var col TableColumn
		if err := rows.Scan(&col.Name, &col.Type, &col.Nullable, &col.IsPK, &col.DefaultValue); err != nil {
			continue
		}
		cols = append(cols, col)
	}
	return cols, nil
}

type RunSQLRequest struct {
	Query string `json:"query"`
}

func (h *TableHandler) RunSQL(c *fiber.Ctx) error {
	appID, err := utils.ParseUUIDParam(c, "id", "app")
	if err != nil {
		return utils.BadRequest(c, "invalid app id")
	}

	app, err := h.appRepo.GetByID(c.Context(), appID)
	if err != nil || app == nil {
		return utils.NotFound(c, "app not found")
	}

	var req RunSQLRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	if strings.TrimSpace(req.Query) == "" {
		return utils.BadRequest(c, "query cannot be empty")
	}

	// Ensure query is scoped to the app's schema
	schemaName := app.SchemaName
	if err := validateTblIdent(schemaName); err != nil {
		return utils.BadRequest(c, "invalid schema name")
	}

	// Set search path to the app schema so unqualified tables resolve correctly
	setPath := fmt.Sprintf("SET search_path TO %s", quoteIdent(schemaName))
	if _, err := h.db.Pool.Exec(c.Context(), setPath); err != nil {
		return utils.InternalError(c, "failed to set schema context")
	}

	rows, err := h.db.Pool.Query(c.Context(), req.Query)
	if err != nil {
		return utils.BadRequest(c, fmt.Sprintf("query error: %s", err.Error()))
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = fd.Name
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}

	return utils.OK(c, results)
}

var validTblIdentReplacer = strings.NewReplacer(`"`, ``)

func validateTblIdent(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("identifier cannot be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("identifier too long (max 63 chars)")
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return fmt.Errorf("identifier contains invalid character: %c", c)
		}
	}
	return nil
}

var allowedPgTypes = map[string]bool{
	"uuid": true, "text": true, "varchar": true, "char": true,
	"integer": true, "int": true, "int2": true, "int4": true, "int8": true,
	"bigint": true, "smallint": true,
	"numeric": true, "decimal": true, "real": true, "double precision": true,
	"float": true, "float4": true, "float8": true,
	"boolean": true, "bool": true,
	"json": true, "jsonb": true,
	"timestamp": true, "timestamptz": true, "timestamp with time zone": true, "timestamp without time zone": true,
	"date": true, "time": true, "timetz": true,
	"bytea": true, "cidr": true, "inet": true, "macaddr": true,
	"money": true, "point": true, "line": true, "lseg": true, "box": true,
	"path": true, "polygon": true, "circle": true,
	"interval": true, "bit": true, "varbit": true,
	"serial": true, "bigserial": true, "smallserial": true,
	"xml": true, "name": true, "oid": true,
}

func validatePgType(t string) error {
	normalized := strings.ToLower(strings.TrimSpace(t))
	if normalized == "" {
		return fmt.Errorf("type cannot be empty")
	}
	// Handle varchar(N), char(N), numeric(P,S) etc
	base := normalized
	if idx := strings.IndexByte(normalized, '('); idx > 0 {
		base = normalized[:idx]
	}
	if !allowedPgTypes[base] {
		return fmt.Errorf("unsupported type: %s", t)
	}
	return nil
}

func quoteLit(val string) string {
	return `'` + strings.ReplaceAll(val, `'`, `''`) + `'`
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
