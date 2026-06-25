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
	Name string `json:"name"`
	Type string `json:"type"`
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
	if err := validateIdentifier(schemaName); err != nil {
		return utils.BadRequest(c, "invalid schema name")
	}

	rows, err := h.db.Pool.Query(c.Context(), fmt.Sprintf(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = %s
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`, quoteLiteral(schemaName)))
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
	if err := validateIdentifier(tableName); err != nil {
		return utils.BadRequest(c, "invalid table name")
	}

	app, err := h.appRepo.GetByID(c.Context(), appID)
	if err != nil || app == nil {
		return utils.NotFound(c, "app not found")
	}

	schemaName := app.SchemaName
	if err := validateIdentifier(schemaName); err != nil {
		return utils.BadRequest(c, "invalid schema name")
	}

	limit := 100
	if l := c.QueryInt("limit", 100); l > 0 && l <= 1000 {
		limit = l
	}

	query := fmt.Sprintf(`SELECT * FROM %s.%s LIMIT %d`,
		quoteIdentifier(schemaName),
		quoteIdentifier(tableName),
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

func (h *TableHandler) listColumns(ctx context.Context, schema, table string) ([]TableColumn, error) {
	rows, err := h.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_schema = %s AND table_name = %s
		AND ordinal_position IS NOT NULL
		ORDER BY ordinal_position
	`, quoteLiteral(schema), quoteLiteral(table)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []TableColumn
	for rows.Next() {
		var c TableColumn
		if err := rows.Scan(&c.Name, &c.Type); err != nil {
			continue
		}
		cols = append(cols, c)
	}
	return cols, nil
}

var validIdent = strings.NewReplacer(`"`, ``)

func validateIdentifier(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("identifier cannot be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("identifier too long (max 63 chars)")
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.') {
			return fmt.Errorf("identifier contains invalid character: %c", c)
		}
	}
	return nil
}

func quoteLiteral(val string) string {
	return `'` + strings.ReplaceAll(val, `'`, `''`) + `'`
}

func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
