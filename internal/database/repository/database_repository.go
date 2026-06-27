package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/internal/database/models"
	"github.com/nexbic/platform/pkg/database"
	"github.com/nexbic/platform/pkg/helpers"
)

type DatabaseRepository struct {
	db *database.DB
}

func NewDatabaseRepository(db *database.DB) *DatabaseRepository {
	return &DatabaseRepository{db: db}
}

func (r *DatabaseRepository) Create(ctx context.Context, dbEntry *models.Database) error {
	dbEntry.ID = uuid.New()
	dbEntry.CreatedAt = time.Now()
	dbEntry.UpdatedAt = time.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO databases (id, project_id, name, schema_name, db_user, db_password, status, size_bytes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		dbEntry.ID, dbEntry.ProjectID, dbEntry.Name, dbEntry.SchemaName,
		dbEntry.DBUser, dbEntry.DBPassword, dbEntry.Status, dbEntry.SizeBytes,
		dbEntry.CreatedAt, dbEntry.UpdatedAt,
	)
	return err
}

func (r *DatabaseRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Database, error) {
	d := &models.Database{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, project_id, name, schema_name, db_user, db_password, status, size_bytes, created_at, updated_at
		FROM databases WHERE id = $1`, id).Scan(
		&d.ID, &d.ProjectID, &d.Name, &d.SchemaName, &d.DBUser, &d.DBPassword,
		&d.Status, &d.SizeBytes, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	d.ConnString = fmt.Sprintf("postgres://%s:<password>@localhost:5432/%s?sslmode=disable",
		d.DBUser, d.Name)
	return d, nil
}

func (r *DatabaseRepository) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]models.Database, int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM databases WHERE project_id = $1`, projectID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, project_id, name, schema_name, db_user, db_password, status, size_bytes, created_at, updated_at
		FROM databases WHERE project_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		projectID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var databases []models.Database
	for rows.Next() {
		var d models.Database
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Name, &d.SchemaName, &d.DBUser, &d.DBPassword,
			&d.Status, &d.SizeBytes, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		d.ConnString = fmt.Sprintf("postgres://%s:<password>@localhost:5432/%s?sslmode=disable",
			d.DBUser, d.Name)
		databases = append(databases, d)
	}

	return databases, total, nil
}

func (r *DatabaseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM databases WHERE id = $1`, id)
	return err
}

func (r *DatabaseRepository) ProvisionSchema(ctx context.Context, dbEntry *models.Database) error {
	schemaName := fmt.Sprintf("db_%s", dbEntry.SchemaName)

	cmds := []string{
		fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, quoteIdent(schemaName)),
		fmt.Sprintf(`CREATE ROLE %s WITH LOGIN PASSWORD %s`, quoteIdent(dbEntry.DBUser), quoteLiteral(dbEntry.DBPassword)),
		fmt.Sprintf(`GRANT USAGE ON SCHEMA %s TO %s`, quoteIdent(schemaName), quoteIdent(dbEntry.DBUser)),
		fmt.Sprintf(`GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA %s TO %s`, quoteIdent(schemaName), quoteIdent(dbEntry.DBUser)),
		fmt.Sprintf(`GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA %s TO %s`, quoteIdent(schemaName), quoteIdent(dbEntry.DBUser)),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT ALL ON TABLES TO %s`, quoteIdent(schemaName), quoteIdent(dbEntry.DBUser)),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT ALL ON SEQUENCES TO %s`, quoteIdent(schemaName), quoteIdent(dbEntry.DBUser)),
	}

	for _, cmd := range cmds {
		if _, err := r.db.Pool.Exec(ctx, cmd); err != nil {
			return fmt.Errorf("provision schema: %s: %w", cmd, err)
		}
	}

	return nil
}

func (r *DatabaseRepository) DropSchema(ctx context.Context, dbEntry *models.Database) error {
	schemaName := fmt.Sprintf("db_%s", dbEntry.SchemaName)
	_, err := r.db.Pool.Exec(ctx, fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, quoteIdent(schemaName)))
	if err != nil {
		return err
	}
	_, err = r.db.Pool.Exec(ctx, fmt.Sprintf(`DROP ROLE IF EXISTS %s`, quoteIdent(dbEntry.DBUser)))
	return err
}

func (r *DatabaseRepository) RunSQL(ctx context.Context, dbID uuid.UUID, query string) ([]map[string]any, error) {
	d, err := r.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}

	schemaName := fmt.Sprintf("db_%s", d.SchemaName)
	setPath := fmt.Sprintf(`SET search_path TO %s`, quoteIdent(schemaName))
	if _, err := r.db.Pool.Exec(ctx, setPath); err != nil {
		return nil, fmt.Errorf("set search_path: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	var result []map[string]any

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
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

type UsageStats struct {
	DatabaseID   uuid.UUID `json:"database_id"`
	RequestCount int64     `json:"request_count"`
	StorageBytes int64     `json:"storage_bytes"`
	Period       string    `json:"period"`
}

func (r *DatabaseRepository) GetUsageStats(ctx context.Context, dbID uuid.UUID) (*UsageStats, error) {
	stats := &UsageStats{DatabaseID: dbID}

	r.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(size_bytes), 0) FROM databases WHERE id = $1`, dbID).Scan(&stats.StorageBytes)

	stats.Period = "current"
	return stats, nil
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func quoteLiteral(val string) string {
	return "'" + strings.ReplaceAll(val, "'", "''") + "'"
}

var validIdentReplacer = strings.NewReplacer(`"`, ``)

func validateIdent(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("identifier too long (max 63)")
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_') {
			return fmt.Errorf("invalid character in identifier: %c", r)
		}
	}
	return nil
}

var allowedPgTypes = map[string]bool{
	"bigint": true, "bigserial": true, "boolean": true, "box": true,
	"bytea": true, "character varying": true, "varchar": true, "character": true,
	"char": true, "cidr": true, "circle": true, "date": true,
	"double precision": true, "float8": true, "inet": true, "integer": true,
	"int": true, "interval": true, "json": true, "jsonb": true,
	"line": true, "lseg": true, "macaddr": true, "macaddr8": true,
	"money": true, "numeric": true, "decimal": true, "path": true,
	"pg_lsn": true, "point": true, "polygon": true, "real": true,
	"float4": true, "smallint": true, "int2": true, "serial": true,
	"text": true, "time": true, "time with time zone": true, "timetz": true,
	"timestamp": true, "timestamp with time zone": true, "timestamptz": true,
	"tsquery": true, "tsvector": true, "txid_snapshot": true, "uuid": true,
	"xml": true, "bigint[]": true, "text[]": true, "integer[]": true,
	"boolean[]": true, "uuid[]": true, "jsonb[]": true, "timestamp with time zone[]": true,
}

func validatePgType(t string) error {
	if !allowedPgTypes[t] {
		return fmt.Errorf("unsupported PostgreSQL type: %s", t)
	}
	return nil
}

type RowReader struct {
	rows pgx.Rows
}

func (r *RowReader) Close() {
	r.rows.Close()
}

func (r *RowReader) ReadAll() ([]map[string]any, error) {
	defer r.rows.Close()

	fieldDescs := r.rows.FieldDescriptions()
	var result []map[string]any

	for r.rows.Next() {
		values, err := r.rows.Values()
		if err != nil {
			return nil, err
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

func (r *DatabaseRepository) ListTables(ctx context.Context, dbID uuid.UUID) ([]models.TableInfo, error) {
	d, err := r.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}

	schemaName := fmt.Sprintf("db_%s", d.SchemaName)

	rows, err := r.db.Pool.Query(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = $1 AND table_type = 'BASE TABLE'
		ORDER BY table_name`, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []models.TableInfo
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}

		cols, err := r.listColumns(ctx, schemaName, tableName)
		if err != nil {
			return nil, err
		}

		tables = append(tables, models.TableInfo{
			Name:    tableName,
			Columns: cols,
		})
	}

	if tables == nil {
		tables = []models.TableInfo{}
	}

	return tables, nil
}

func (r *DatabaseRepository) listColumns(ctx context.Context, schema, table string) ([]models.TableColumn, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable,
			COALESCE((
				SELECT TRUE FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu
					ON tc.constraint_name = kcu.constraint_name
					AND tc.table_schema = kcu.table_schema
				WHERE tc.constraint_type = 'PRIMARY KEY'
					AND kcu.column_name = c.column_name
					AND kcu.table_name = c.table_name
					AND kcu.table_schema = c.table_schema
				LIMIT 1
			), FALSE) as is_pk,
			COALESCE(c.column_default, '') as default_value
		FROM information_schema.columns c
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []models.TableColumn
	for rows.Next() {
		var col models.TableColumn
		var nullable string
		if err := rows.Scan(&col.Name, &col.Type, &nullable, &col.IsPK, &col.DefaultValue); err != nil {
			return nil, err
		}
		col.Nullable = nullable == "YES"
		columns = append(columns, col)
	}

	if columns == nil {
		columns = []models.TableColumn{}
	}

	return columns, nil
}

func (r *DatabaseRepository) GetTableData(ctx context.Context, dbID uuid.UUID, table string, limit, offset int) ([]map[string]any, error) {
	d, err := r.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}

	schemaName := fmt.Sprintf("db_%s", d.SchemaName)
	query := fmt.Sprintf(`SELECT * FROM %s.%s LIMIT $1 OFFSET $2`,
		quoteIdent(schemaName), quoteIdent(validIdentReplacer.Replace(table)))

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	var result []map[string]any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
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

func (r *DatabaseRepository) CreateTable(ctx context.Context, dbID uuid.UUID, table string, columns []models.TableColumn) error {
	d, err := r.GetByID(ctx, dbID)
	if err != nil {
		return err
	}

	if err := validateIdent(table); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	schemaName := fmt.Sprintf("db_%s", d.SchemaName)
	var colDefs []string
	for _, col := range columns {
		if err := validateIdent(col.Name); err != nil {
			return fmt.Errorf("invalid column name: %w", err)
		}
		if err := validatePgType(col.Type); err != nil {
			return err
		}
		def := fmt.Sprintf("%s %s", quoteIdent(validIdentReplacer.Replace(col.Name)), col.Type)
		if !col.Nullable {
			def += " NOT NULL"
		}
		if col.DefaultValue != "" {
			def += " DEFAULT " + col.DefaultValue
		}
		colDefs = append(colDefs, def)
	}

	query := fmt.Sprintf(`CREATE TABLE %s.%s (%s)`,
		quoteIdent(schemaName), quoteIdent(validIdentReplacer.Replace(table)),
		strings.Join(colDefs, ", "))

	_, err = r.db.Pool.Exec(ctx, query)
	return err
}

func (r *DatabaseRepository) AddColumn(ctx context.Context, dbID uuid.UUID, table string, col *models.TableColumn) error {
	d, err := r.GetByID(ctx, dbID)
	if err != nil {
		return err
	}

	if err := validateIdent(table); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}
	if err := validateIdent(col.Name); err != nil {
		return fmt.Errorf("invalid column name: %w", err)
	}
	if err := validatePgType(col.Type); err != nil {
		return err
	}

	schemaName := fmt.Sprintf("db_%s", d.SchemaName)
	def := fmt.Sprintf("%s %s", quoteIdent(validIdentReplacer.Replace(col.Name)), col.Type)
	if !col.Nullable {
		def += " NOT NULL"
	}
	if col.DefaultValue != "" {
		def += " DEFAULT " + col.DefaultValue
	}

	query := fmt.Sprintf(`ALTER TABLE %s.%s ADD COLUMN %s`,
		quoteIdent(schemaName), quoteIdent(validIdentReplacer.Replace(table)), def)

	_, err = r.db.Pool.Exec(ctx, query)
	return err
}

func (r *DatabaseRepository) InsertRow(ctx context.Context, dbID uuid.UUID, table string, values map[string]any) (map[string]any, error) {
	d, err := r.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}

	schemaName := fmt.Sprintf("db_%s", d.SchemaName)

	cols := make([]string, 0, len(values))
	vals := make([]any, 0, len(values))
	placeholders := make([]string, 0, len(values))
	i := 1
	for k, v := range values {
		cols = append(cols, quoteIdent(validIdentReplacer.Replace(k)))
		vals = append(vals, v)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}

	query := fmt.Sprintf(`INSERT INTO %s.%s (%s) VALUES (%s) RETURNING *`,
		quoteIdent(schemaName), quoteIdent(validIdentReplacer.Replace(table)),
		strings.Join(cols, ", "), strings.Join(placeholders, ", "))

	rows, err := r.db.Pool.Query(ctx, query, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		fieldDescs := rows.FieldDescriptions()
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := make(map[string]any)
		for i, fd := range fieldDescs {
			row[string(fd.Name)] = values[i]
		}
		return row, nil
	}

	return nil, fmt.Errorf("insert succeeded but no rows returned")
}

func (r *DatabaseRepository) UpdateRow(ctx context.Context, dbID uuid.UUID, table string, values, where map[string]any) ([]map[string]any, error) {
	d, err := r.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}

	schemaName := fmt.Sprintf("db_%s", d.SchemaName)

	setClauses := make([]string, 0, len(values))
	args := make([]any, 0, len(values)+len(where))
	i := 1
	for k, v := range values {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", quoteIdent(validIdentReplacer.Replace(k)), i))
		args = append(args, v)
		i++
	}

	whereClauses := make([]string, 0, len(where))
	for k, v := range where {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", quoteIdent(validIdentReplacer.Replace(k)), i))
		args = append(args, v)
		i++
	}

	query := fmt.Sprintf(`UPDATE %s.%s SET %s WHERE %s RETURNING *`,
		quoteIdent(schemaName), quoteIdent(validIdentReplacer.Replace(table)),
		strings.Join(setClauses, ", "), strings.Join(whereClauses, " AND "))

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	var result []map[string]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := make(map[string]any)
		for i, fd := range fieldDescs {
			row[string(fd.Name)] = vals[i]
		}
		result = append(result, row)
	}

	if result == nil {
		result = []map[string]any{}
	}

	return result, nil
}

func (r *DatabaseRepository) DeleteRow(ctx context.Context, dbID uuid.UUID, table string, where map[string]any) (int64, error) {
	d, err := r.GetByID(ctx, dbID)
	if err != nil {
		return 0, err
	}

	schemaName := fmt.Sprintf("db_%s", d.SchemaName)

	args := make([]any, 0, len(where))
	whereClauses := make([]string, 0, len(where))
	i := 1
	for k, v := range where {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", quoteIdent(validIdentReplacer.Replace(k)), i))
		args = append(args, v)
		i++
	}

	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE %s`,
		quoteIdent(schemaName), quoteIdent(validIdentReplacer.Replace(table)),
		strings.Join(whereClauses, " AND "))

	ct, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return ct.RowsAffected(), nil
}

func (r *DatabaseRepository) ListAvailableExtensions(ctx context.Context) ([]models.Extension, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT
			e.name,
			COALESCE(ev.version, ''),
			COALESCE(e.comment, ''),
			CASE WHEN pg.extname IS NOT NULL THEN true ELSE false END as installed
		FROM pg_available_extensions e
		LEFT JOIN pg_available_extension_versions ev ON e.name = ev.name AND ev.version = e.default_version
		LEFT JOIN pg_extension pg ON e.name = pg.extname
		ORDER BY e.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var extensions []models.Extension
	for rows.Next() {
		var ext models.Extension
		if err := rows.Scan(&ext.Name, &ext.Version, &ext.Description, &ext.Installed); err != nil {
			return nil, err
		}
		extensions = append(extensions, ext)
	}

	if extensions == nil {
		extensions = []models.Extension{}
	}

	return extensions, nil
}

func (r *DatabaseRepository) ToggleExtension(ctx context.Context, name string, install bool) error {
	if install {
		_, err := r.db.Pool.Exec(ctx, fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS %s`, quoteIdent(name)))
		return err
	}
	_, err := r.db.Pool.Exec(ctx, fmt.Sprintf(`DROP EXTENSION IF EXISTS %s`, quoteIdent(name)))
	return err
}

var blockedExtensions = map[string]bool{
	"plpgsql": true, "plpython3u": true, "plperlu": true, "pltclu": true,
	"amcheck": true, "pageinspect": true, "pg_buffercache": true,
	"pgrowlocks": true, "pgstattuple": true, "auto_explain": true,
	"pg_prewarm": true, "old_snapshot": true, "pg_surgery": true,
	"adminpack": true, "pg_freespacemap": true, "pg_visibility": true,
}

func IsExtensionBlocked(name string) bool {
	return blockedExtensions[name]
}
