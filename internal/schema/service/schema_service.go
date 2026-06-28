package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/internal/schema/models"
	"github.com/nexbic/platform/pkg/database"
)

type SchemaService struct {
	db *database.DB
}

func NewSchemaService(db *database.DB) *SchemaService {
	return &SchemaService{db: db}
}

func quoteIdent(parts ...string) string {
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = `"` + strings.ReplaceAll(p, `"`, `""`) + `"`
	}
	return strings.Join(quoted, ".")
}

func quoteIdents(parts ...string) []string {
	res := make([]string, len(parts))
	for i, p := range parts {
		res[i] = quoteIdent(p)
	}
	return res
}

func (s *SchemaService) CreateSchema(ctx context.Context, name string) error {
	sql := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", quoteIdent(name))
	return s.db.Exec(ctx, sql)
}

func (s *SchemaService) DropSchema(ctx context.Context, name string, cascade bool) error {
	sql := fmt.Sprintf("DROP SCHEMA %s", quoteIdent(name))
	if cascade {
		sql += " CASCADE"
	}
	return s.db.Exec(ctx, sql)
}

func (s *SchemaService) CreateTable(ctx context.Context, schema string, req models.CreateTableRequest) error {
	var stmt strings.Builder

	stmt.WriteString("CREATE TABLE ")
	if req.IfNotExists {
		stmt.WriteString("IF NOT EXISTS ")
	}
	stmt.WriteString(quoteIdent(schema, req.Name))
	stmt.WriteString(" (\n")

	var pkCols []string
	colDefs := make([]string, 0, len(req.Columns))
	for _, col := range req.Columns {
		def := fmt.Sprintf("  %s %s", quoteIdent(col.Name), col.Type)
		if !col.Nullable {
			def += " NOT NULL"
		}
		if col.DefaultValue != nil {
			def += " DEFAULT " + *col.DefaultValue
		}
		colDefs = append(colDefs, def)
		if col.IsPK {
			pkCols = append(pkCols, quoteIdent(col.Name))
		}
	}

	for i, def := range colDefs {
		if i > 0 {
			stmt.WriteString(",\n")
		}
		stmt.WriteString(def)
	}

	if len(pkCols) > 0 {
		stmt.WriteString(",\n  PRIMARY KEY (")
		stmt.WriteString(strings.Join(pkCols, ", "))
		stmt.WriteString(")")
	}

	stmt.WriteString("\n)")

	return s.db.Exec(ctx, stmt.String())
}

func (s *SchemaService) DropTable(ctx context.Context, schema, name string, cascade bool) error {
	sql := fmt.Sprintf("DROP TABLE %s", quoteIdent(schema, name))
	if cascade {
		sql += " CASCADE"
	}
	return s.db.Exec(ctx, sql)
}

func (s *SchemaService) AddColumn(ctx context.Context, schema, table string, req models.AddColumnRequest) error {
	var stmt strings.Builder
	stmt.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
		quoteIdent(schema, table), quoteIdent(req.Name), req.Type))
	if !req.Nullable {
		stmt.WriteString(" NOT NULL")
	}
	if req.DefaultValue != nil {
		stmt.WriteString(" DEFAULT " + *req.DefaultValue)
	}
	return s.db.Exec(ctx, stmt.String())
}

func (s *SchemaService) DropColumn(ctx context.Context, schema, table, column string) error {
	sql := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s",
		quoteIdent(schema, table), quoteIdent(column))
	return s.db.Exec(ctx, sql)
}

func (s *SchemaService) AlterColumn(ctx context.Context, schema, table, column string, req models.AlterColumnRequest) error {
	if req.DataType != "" {
		sql := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s",
			quoteIdent(schema, table), quoteIdent(column), req.DataType)
		if err := s.db.Exec(ctx, sql); err != nil {
			return err
		}
	}
	if req.DefaultValue != nil {
		sql := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s",
			quoteIdent(schema, table), quoteIdent(column), *req.DefaultValue)
		if err := s.db.Exec(ctx, sql); err != nil {
			return err
		}
	}
	if req.Nullable != nil {
		if *req.Nullable {
			sql := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL",
				quoteIdent(schema, table), quoteIdent(column))
			if err := s.db.Exec(ctx, sql); err != nil {
				return err
			}
		} else {
			sql := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL",
				quoteIdent(schema, table), quoteIdent(column))
			if err := s.db.Exec(ctx, sql); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SchemaService) AddConstraint(ctx context.Context, schema, table string, req models.AddConstraintRequest) error {
	var stmt strings.Builder
	stmt.WriteString(fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s ",
		quoteIdent(schema, table), quoteIdent(req.Name)))

	switch strings.ToUpper(req.Type) {
	case "PK", "PRIMARY KEY":
		if len(req.Columns) == 0 {
			return fmt.Errorf("columns required for primary key constraint")
		}
		stmt.WriteString(fmt.Sprintf("PRIMARY KEY (%s)",
			strings.Join(quoteIdents(req.Columns...), ", ")))
	case "FK", "FOREIGN KEY":
		if len(req.Columns) == 0 || req.RefTable == "" || req.RefColumn == "" {
			return fmt.Errorf("columns, ref_table and ref_column required for foreign key")
		}
		stmt.WriteString(fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)",
			strings.Join(quoteIdents(req.Columns...), ", "),
			quoteIdent(schema, req.RefTable),
			quoteIdent(req.RefColumn)))
	case "UNIQUE":
		if len(req.Columns) == 0 {
			return fmt.Errorf("columns required for unique constraint")
		}
		stmt.WriteString(fmt.Sprintf("UNIQUE (%s)",
			strings.Join(quoteIdents(req.Columns...), ", ")))
	case "CHECK":
		if req.CheckExpr == "" {
			return fmt.Errorf("check_expr required for check constraint")
		}
		stmt.WriteString("CHECK (" + req.CheckExpr + ")")
	default:
		return fmt.Errorf("unsupported constraint type: %s", req.Type)
	}

	return s.db.Exec(ctx, stmt.String())
}

func (s *SchemaService) DropConstraint(ctx context.Context, schema, table, constraint string) error {
	sql := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s",
		quoteIdent(schema, table), quoteIdent(constraint))
	return s.db.Exec(ctx, sql)
}

func (s *SchemaService) CreateIndex(ctx context.Context, schema string, req models.CreateIndexRequest) error {
	var stmt strings.Builder

	stmt.WriteString("CREATE ")
	if req.Unique {
		stmt.WriteString("UNIQUE ")
	}
	stmt.WriteString("INDEX ")
	stmt.WriteString(quoteIdent(req.Name))
	stmt.WriteString(" ON ")
	stmt.WriteString(quoteIdent(schema, req.Table))

	if req.Method != "" {
		stmt.WriteString(" USING ")
		stmt.WriteString(req.Method)
	}

	stmt.WriteString(" (")
	stmt.WriteString(strings.Join(quoteIdents(req.Columns...), ", "))
	stmt.WriteString(")")

	if req.Where != "" {
		stmt.WriteString(" WHERE ")
		stmt.WriteString(req.Where)
	}

	return s.db.Exec(ctx, stmt.String())
}

func (s *SchemaService) DropIndex(ctx context.Context, schema, name string) error {
	sql := fmt.Sprintf("DROP INDEX %s", quoteIdent(schema, name))
	return s.db.Exec(ctx, sql)
}

func (s *SchemaService) CreateSequence(ctx context.Context, schema string, req models.CreateSequenceRequest) error {
	var stmt strings.Builder
	stmt.WriteString("CREATE SEQUENCE ")
	stmt.WriteString(quoteIdent(schema, req.Name))

	if req.Options != nil {
		if req.Options.Increment != nil {
			stmt.WriteString(" INCREMENT BY ")
			stmt.WriteString(strconv.FormatInt(*req.Options.Increment, 10))
		}
		if req.Options.MinValue != nil {
			stmt.WriteString(" MINVALUE ")
			stmt.WriteString(strconv.FormatInt(*req.Options.MinValue, 10))
		}
		if req.Options.MaxValue != nil {
			stmt.WriteString(" MAXVALUE ")
			stmt.WriteString(strconv.FormatInt(*req.Options.MaxValue, 10))
		}
		if req.Options.Start != nil {
			stmt.WriteString(" START WITH ")
			stmt.WriteString(strconv.FormatInt(*req.Options.Start, 10))
		}
		if req.Options.Cache != nil {
			stmt.WriteString(" CACHE ")
			stmt.WriteString(strconv.FormatInt(*req.Options.Cache, 10))
		}
		if req.Options.Cycle != nil && *req.Options.Cycle {
			stmt.WriteString(" CYCLE")
		}
	}

	return s.db.Exec(ctx, stmt.String())
}

func (s *SchemaService) DropSequence(ctx context.Context, schema, name string) error {
	sql := fmt.Sprintf("DROP SEQUENCE %s", quoteIdent(schema, name))
	return s.db.Exec(ctx, sql)
}

func (s *SchemaService) AlterSequence(ctx context.Context, schema, name string, req models.AlterSequenceRequest) error {
	var stmt strings.Builder
	stmt.WriteString("ALTER SEQUENCE ")
	stmt.WriteString(quoteIdent(schema, name))

	if req.Options.Increment != nil {
		stmt.WriteString(" INCREMENT BY ")
		stmt.WriteString(strconv.FormatInt(*req.Options.Increment, 10))
	}
	if req.Options.MinValue != nil {
		stmt.WriteString(" MINVALUE ")
		stmt.WriteString(strconv.FormatInt(*req.Options.MinValue, 10))
	}
	if req.Options.MaxValue != nil {
		stmt.WriteString(" MAXVALUE ")
		stmt.WriteString(strconv.FormatInt(*req.Options.MaxValue, 10))
	}
	if req.Options.Start != nil {
		stmt.WriteString(" START WITH ")
		stmt.WriteString(strconv.FormatInt(*req.Options.Start, 10))
	}
	if req.Options.Cache != nil {
		stmt.WriteString(" CACHE ")
		stmt.WriteString(strconv.FormatInt(*req.Options.Cache, 10))
	}
	if req.Options.Cycle != nil {
		if *req.Options.Cycle {
			stmt.WriteString(" CYCLE")
		} else {
			stmt.WriteString(" NO CYCLE")
		}
	}

	return s.db.Exec(ctx, stmt.String())
}

type columnInfo struct {
	name      string
	dataType  string
	notNull   bool
	defaultFn string
	identity  string
	generated string
}

type constraintInfo struct {
	name     string
	typ      string
	def      string
	columns  []int16
	fwdRelID uint32
	fwdKeys  []int16
}

func (s *SchemaService) GetTableDDL(ctx context.Context, schema, table string) (string, error) {
	var oid uint32
	err := s.db.QueryRow(ctx, `
		SELECT c.oid
		FROM pg_catalog.pg_class c
		JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2 AND c.relkind = 'r'
	`, schema, table).Scan(&oid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("table %s.%s not found", schema, table)
		}
		return "", fmt.Errorf("lookup table: %w", err)
	}

	cols, err := s.loadColumns(ctx, oid)
	if err != nil {
		return "", err
	}

	constraints, err := s.loadConstraints(ctx, oid)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("CREATE TABLE ")
	b.WriteString(quoteIdent(schema, table))
	b.WriteString(" (\n")

	colLines := make([]string, 0, len(cols))
	for _, col := range cols {
		line := fmt.Sprintf("    %s %s", quoteIdent(col.name), col.dataType)
		if col.identity != "" {
			line += " GENERATED " + col.identity + " AS IDENTITY"
		}
		if !col.notNull && col.identity == "" {
			line += " NULL"
		}
		if col.notNull {
			line += " NOT NULL"
		}
		if col.defaultFn != "" {
			line += " DEFAULT " + col.defaultFn
		}
		if col.generated != "" {
			line += " GENERATED ALWAYS AS (" + col.generated + ") STORED"
		}
		colLines = append(colLines, line)
	}

	for i, line := range colLines {
		if i > 0 {
			b.WriteString(",\n")
		}
		b.WriteString(line)
	}

	for _, con := range constraints {
		b.WriteString(",\n")
		switch con.typ {
		case "p":
			b.WriteString("    CONSTRAINT ")
			b.WriteString(quoteIdent(con.name))
			b.WriteString(" PRIMARY KEY (")
			colNames := s.resolveColumnNames(cols, con.columns)
			b.WriteString(strings.Join(colNames, ", "))
			b.WriteString(")")
		case "f":
			b.WriteString("    CONSTRAINT ")
			b.WriteString(quoteIdent(con.name))
			b.WriteString(" FOREIGN KEY (")
			colNames := s.resolveColumnNames(cols, con.columns)
			b.WriteString(strings.Join(colNames, ", "))
			b.WriteString(") REFERENCES ")
			refSchema, refTable, refCols := s.resolveFKRef(ctx, con.fwdRelID, con.fwdKeys)
			if refSchema != "" {
				b.WriteString(quoteIdent(refSchema, refTable))
			} else {
				b.WriteString(quoteIdent(refTable))
			}
			b.WriteString(" (")
			b.WriteString(strings.Join(refCols, ", "))
			b.WriteString(")")
		case "u":
			b.WriteString("    CONSTRAINT ")
			b.WriteString(quoteIdent(con.name))
			b.WriteString(" UNIQUE (")
			colNames := s.resolveColumnNames(cols, con.columns)
			b.WriteString(strings.Join(colNames, ", "))
			b.WriteString(")")
		case "c":
			b.WriteString("    CONSTRAINT ")
			b.WriteString(quoteIdent(con.name))
			b.WriteString(" CHECK (")
			b.WriteString(con.def)
			b.WriteString(")")
		}
	}

	b.WriteString("\n)")

	return b.String(), nil
}

func (s *SchemaService) loadColumns(ctx context.Context, tableOID uint32) ([]columnInfo, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			a.attname,
			pg_catalog.format_type(a.atttypid, a.atttypmod),
			a.attnotnull,
			COALESCE(pg_catalog.pg_get_expr(d.adbin, d.adrelid), '') AS default_expr,
			a.attidentity,
			COALESCE(pg_catalog.pg_get_expr(ad.adbin, ad.adrelid), '') AS generated_expr
		FROM pg_catalog.pg_attribute a
		LEFT JOIN pg_catalog.pg_attrdef d
			ON d.adrelid = a.attrelid AND d.adnum = a.attnum AND a.attgenerated = ''
		LEFT JOIN pg_catalog.pg_attrdef ad
			ON ad.adrelid = a.attrelid AND ad.adnum = a.attnum AND a.attgenerated != ''
		WHERE a.attrelid = $1 AND a.attnum > 0 AND NOT a.attisdropped
		ORDER BY a.attnum
	`, tableOID)
	if err != nil {
		return nil, fmt.Errorf("load columns: %w", err)
	}
	defer rows.Close()

	var cols []columnInfo
	for rows.Next() {
		var c columnInfo
		if err := rows.Scan(&c.name, &c.dataType, &c.notNull, &c.defaultFn, &c.identity, &c.generated); err != nil {
			return nil, fmt.Errorf("scan column: %w", err)
		}
		cols = append(cols, c)
	}
	return cols, nil
}

func (s *SchemaService) loadConstraints(ctx context.Context, tableOID uint32) ([]constraintInfo, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			conname,
			contype,
			COALESCE(pg_catalog.pg_get_constraintdef(con.oid), '') AS consrc,
			COALESCE(conkey, '{}')::int2[] AS conkeys,
			COALESCE(confrelid, 0) AS confrelid,
			COALESCE(confkey, '{}')::int2[] AS confkeys
		FROM pg_catalog.pg_constraint
		WHERE conrelid = $1 AND contype IN ('p', 'f', 'u', 'c')
		ORDER BY
			CASE contype
				WHEN 'p' THEN 1
				WHEN 'u' THEN 2
				WHEN 'f' THEN 3
				WHEN 'c' THEN 4
			END, conname
	`, tableOID)
	if err != nil {
		return nil, fmt.Errorf("load constraints: %w", err)
	}
	defer rows.Close()

	var cons []constraintInfo
	for rows.Next() {
		var c constraintInfo
		if err := rows.Scan(&c.name, &c.typ, &c.def, &c.columns, &c.fwdRelID, &c.fwdKeys); err != nil {
			return nil, fmt.Errorf("scan constraint: %w", err)
		}
		cons = append(cons, c)
	}
	return cons, nil
}

func (s *SchemaService) resolveColumnNames(cols []columnInfo, attnums []int16) []string {
	lookup := make(map[int16]string)
	for i, col := range cols {
		lookup[int16(i+1)] = col.name
	}
	names := make([]string, len(attnums))
	for i, an := range attnums {
		if name, ok := lookup[an]; ok {
			names[i] = quoteIdent(name)
		} else {
			names[i] = quoteIdent(fmt.Sprintf("?%d", an))
		}
	}
	return names
}

func (s *SchemaService) resolveFKRef(ctx context.Context, relID uint32, fwdKeys []int16) (schema, table string, colNames []string) {
	if relID == 0 {
		return "", "", nil
	}

	err := s.db.QueryRow(ctx, `
		SELECT n.nspname, c.relname
		FROM pg_catalog.pg_class c
		JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
		WHERE c.oid = $1
	`, relID).Scan(&schema, &table)
	if err != nil {
		return "", "", nil
	}

	colNames = make([]string, len(fwdKeys))
	for i, k := range fwdKeys {
		var colName string
		err := s.db.QueryRow(ctx, `
			SELECT attname FROM pg_catalog.pg_attribute
			WHERE attrelid = $1 AND attnum = $2 AND NOT attisdropped
		`, relID, k).Scan(&colName)
		if err != nil {
			colNames[i] = "?"
		} else {
			colNames[i] = quoteIdent(colName)
		}
	}

	return schema, table, colNames
}
