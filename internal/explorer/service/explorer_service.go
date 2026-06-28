package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/internal/explorer/models"
	"github.com/nexbic/platform/pkg/database"
)

type ExplorerService struct {
	db *database.DB
}

func NewExplorerService(db *database.DB) *ExplorerService {
	return &ExplorerService{db: db}
}

func (s *ExplorerService) ListSchemas(ctx context.Context, schemaFilter string) ([]models.SchemaInfo, error) {
	query := `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		  AND schema_name NOT LIKE 'pg_%'
	`
	var args []any
	if schemaFilter != "" {
		query += " AND schema_name ILIKE $1"
		args = append(args, "%"+schemaFilter+"%")
	}
	query += " ORDER BY schema_name"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list schemas: %w", err)
	}
	defer rows.Close()

	var result []models.SchemaInfo
	for rows.Next() {
		var si models.SchemaInfo
		if err := rows.Scan(&si.SchemaName); err != nil {
			return nil, fmt.Errorf("scan schema: %w", err)
		}
		result = append(result, si)
	}
	return result, nil
}

func (s *ExplorerService) ListTables(ctx context.Context, schema string) ([]models.TableInfo, error) {
	query := `
		SELECT
			t.table_name,
			t.table_type,
			pg_class.reltuples::bigint AS row_estimate
		FROM information_schema.tables t
		LEFT JOIN pg_catalog.pg_class ON pg_class.relname = t.table_name
			AND pg_class.relnamespace = (
				SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = $1
			)
		WHERE t.table_schema = $1
		  AND t.table_type IN ('BASE TABLE', 'VIEW')
		ORDER BY t.table_name
	`

	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var result []models.TableInfo
	for rows.Next() {
		var ti models.TableInfo
		if err := rows.Scan(&ti.TableName, &ti.TableType, &ti.RowEstimate); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}
		result = append(result, ti)
	}
	return result, nil
}

func (s *ExplorerService) GetTableDetails(ctx context.Context, schema, table string) (*models.TableDetails, error) {
	columns, err := s.getColumns(ctx, schema, table)
	if err != nil {
		return nil, err
	}

	constraints, err := s.getConstraintsForTable(ctx, schema, table)
	if err != nil {
		return nil, err
	}

	indexes, err := s.getIndexesForTable(ctx, schema, table)
	if err != nil {
		return nil, err
	}

	triggers, err := s.getTriggersForTable(ctx, schema, table)
	if err != nil {
		return nil, err
	}

	var rowEstimate *int64
	err = s.db.QueryRow(ctx, `
		SELECT c.reltuples::bigint
		FROM pg_catalog.pg_class c
		JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2
	`, schema, table).Scan(&rowEstimate)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("get row estimate: %w", err)
	}

	td := &models.TableDetails{
		Table: models.TableInfo{
			TableName:   table,
			TableType:   "BASE TABLE",
			RowEstimate: rowEstimate,
		},
		Columns:     columns,
		Constraints: constraints,
		Indexes:     indexes,
		Triggers:    triggers,
	}

	if len(td.Columns) == 0 {
		return nil, nil
	}

	return td, nil
}

func (s *ExplorerService) getColumns(ctx context.Context, schema, table string) ([]models.ColumnInfo, error) {
	query := `
		SELECT
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length,
			numeric_precision,
			numeric_scale,
			ordinal_position
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`

	rows, err := s.db.Query(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}
	defer rows.Close()

	var result []models.ColumnInfo
	for rows.Next() {
		var ci models.ColumnInfo
		if err := rows.Scan(
			&ci.ColumnName, &ci.DataType, &ci.IsNullable, &ci.ColumnDefault,
			&ci.CharacterMaximumLength, &ci.NumericPrecision, &ci.NumericScale,
			&ci.OrdinalPosition,
		); err != nil {
			return nil, fmt.Errorf("scan column: %w", err)
		}
		result = append(result, ci)
	}
	return result, nil
}

func (s *ExplorerService) getConstraintsForTable(ctx context.Context, schema, table string) ([]models.ConstraintInfo, error) {
	query := `
		SELECT
			tc.constraint_name,
			tc.constraint_type,
			coalesce(kcu.column_name, '') AS column_name,
			coalesce(ccu.table_name, '') AS ref_table_name,
			coalesce(ccu.column_name, '') AS ref_column_name
		FROM information_schema.table_constraints tc
		LEFT JOIN information_schema.key_column_usage kcu
			ON tc.constraint_catalog = kcu.constraint_catalog
			AND tc.constraint_schema = kcu.constraint_schema
			AND tc.constraint_name = kcu.constraint_name
		LEFT JOIN information_schema.constraint_column_usage ccu
			ON tc.constraint_catalog = ccu.constraint_catalog
			AND tc.constraint_schema = ccu.constraint_schema
			AND tc.constraint_name = ccu.constraint_name
		WHERE tc.table_schema = $1 AND tc.table_name = $2
		ORDER BY tc.constraint_name, kcu.ordinal_position
	`

	rows, err := s.db.Query(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("get constraints: %w", err)
	}
	defer rows.Close()

	var result []models.ConstraintInfo
	for rows.Next() {
		var ci models.ConstraintInfo
		if err := rows.Scan(&ci.ConstraintName, &ci.ConstraintType, &ci.ColumnName, &ci.RefTableName, &ci.RefColumnName); err != nil {
			return nil, fmt.Errorf("scan constraint: %w", err)
		}
		result = append(result, ci)
	}
	return result, nil
}

func (s *ExplorerService) getIndexesForTable(ctx context.Context, schema, table string) ([]models.IndexInfo, error) {
	query := `
		SELECT
			indexname,
			indexdef
		FROM pg_catalog.pg_indexes
		WHERE schemaname = $1 AND tablename = $2
		ORDER BY indexname
	`

	rows, err := s.db.Query(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("get indexes: %w", err)
	}
	defer rows.Close()

	var result []models.IndexInfo
	for rows.Next() {
		var ii models.IndexInfo
		if err := rows.Scan(&ii.IndexName, &ii.IndexDef); err != nil {
			return nil, fmt.Errorf("scan index: %w", err)
		}
		ii.TableName = table
		ii.IsUnique = false
		ii.IsPrimary = false
		ii.IndexType = ""

		if contains(ii.IndexDef, "UNIQUE") {
			ii.IsUnique = true
		}
		if contains(ii.IndexDef, "CREATE UNIQUE INDEX") || contains(ii.IndexDef, "UNIQUE") {
			ii.IsUnique = true
		}
		if contains(ii.IndexDef, "PRIMARY KEY") {
			ii.IsPrimary = true
		}
		if contains(ii.IndexDef, "btree") {
			ii.IndexType = "btree"
		} else if contains(ii.IndexDef, "hash") {
			ii.IndexType = "hash"
		} else if contains(ii.IndexDef, "gin") {
			ii.IndexType = "gin"
		} else if contains(ii.IndexDef, "gist") {
			ii.IndexType = "gist"
		} else if contains(ii.IndexDef, "brin") {
			ii.IndexType = "brin"
		}

		result = append(result, ii)
	}
	return result, nil
}

func (s *ExplorerService) getTriggersForTable(ctx context.Context, schema, table string) ([]models.TriggerInfo, error) {
	query := `
		SELECT
			t.trigger_name,
			t.event_manipulation,
			t.action_timing,
			t.action_statement,
			t.event_object_table AS table_name
		FROM information_schema.triggers t
		WHERE t.event_object_schema = $1 AND t.event_object_table = $2
		ORDER BY t.trigger_name
	`

	rows, err := s.db.Query(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("get triggers: %w", err)
	}
	defer rows.Close()

	var result []models.TriggerInfo
	for rows.Next() {
		var ti models.TriggerInfo
		if err := rows.Scan(&ti.TriggerName, &ti.EventManipulation, &ti.ActionTiming, &ti.ActionStatement, &ti.TableName); err != nil {
			return nil, fmt.Errorf("scan trigger: %w", err)
		}
		result = append(result, ti)
	}
	return result, nil
}

func (s *ExplorerService) ListViews(ctx context.Context, schema string) ([]models.ViewInfo, error) {
	query := `
		SELECT table_name, view_definition
		FROM information_schema.views
		WHERE table_schema = $1
		ORDER BY table_name
	`

	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("list views: %w", err)
	}
	defer rows.Close()

	var result []models.ViewInfo
	for rows.Next() {
		var vi models.ViewInfo
		if err := rows.Scan(&vi.ViewName, &vi.ViewDefinition); err != nil {
			return nil, fmt.Errorf("scan view: %w", err)
		}
		result = append(result, vi)
	}
	return result, nil
}

func (s *ExplorerService) ListFunctions(ctx context.Context, schema string) ([]models.FunctionInfo, error) {
	query := `
		SELECT
			p.proname,
			format_type(p.prorettype, NULL) AS return_type,
			pg_get_function_arguments(p.oid) AS arguments,
			l.lanname AS language
		FROM pg_catalog.pg_proc p
		JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
		JOIN pg_catalog.pg_language l ON l.oid = p.prolang
		WHERE n.nspname = $1
		  AND p.prokind = 'f'
		ORDER BY p.proname
	`

	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("list functions: %w", err)
	}
	defer rows.Close()

	var result []models.FunctionInfo
	for rows.Next() {
		var fi models.FunctionInfo
		if err := rows.Scan(&fi.FunctionName, &fi.ReturnType, &fi.Arguments, &fi.Language); err != nil {
			return nil, fmt.Errorf("scan function: %w", err)
		}
		result = append(result, fi)
	}
	return result, nil
}

func (s *ExplorerService) ListProcedures(ctx context.Context, schema string) ([]models.ProcedureInfo, error) {
	query := `
		SELECT
			p.proname,
			pg_get_function_arguments(p.oid) AS arguments,
			l.lanname AS language
		FROM pg_catalog.pg_proc p
		JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
		JOIN pg_catalog.pg_language l ON l.oid = p.prolang
		WHERE n.nspname = $1
		  AND p.prokind = 'p'
		ORDER BY p.proname
	`

	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("list procedures: %w", err)
	}
	defer rows.Close()

	var result []models.ProcedureInfo
	for rows.Next() {
		var pi models.ProcedureInfo
		if err := rows.Scan(&pi.ProcedureName, &pi.Arguments, &pi.Language); err != nil {
			return nil, fmt.Errorf("scan procedure: %w", err)
		}
		result = append(result, pi)
	}
	return result, nil
}

func (s *ExplorerService) ListTriggers(ctx context.Context, schema string) ([]models.TriggerInfo, error) {
	query := `
		SELECT
			t.trigger_name,
			t.event_manipulation,
			t.action_timing,
			t.action_statement,
			t.event_object_table AS table_name
		FROM information_schema.triggers t
		WHERE t.event_object_schema = $1
		ORDER BY t.event_object_table, t.trigger_name
	`

	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("list triggers: %w", err)
	}
	defer rows.Close()

	var result []models.TriggerInfo
	for rows.Next() {
		var ti models.TriggerInfo
		if err := rows.Scan(&ti.TriggerName, &ti.EventManipulation, &ti.ActionTiming, &ti.ActionStatement, &ti.TableName); err != nil {
			return nil, fmt.Errorf("scan trigger: %w", err)
		}
		result = append(result, ti)
	}
	return result, nil
}

func (s *ExplorerService) ListIndexes(ctx context.Context, schema string) ([]models.IndexInfo, error) {
	query := `
		SELECT
			indexname,
			tablename,
			indexdef
		FROM pg_catalog.pg_indexes
		WHERE schemaname = $1
		ORDER BY tablename, indexname
	`

	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("list indexes: %w", err)
	}
	defer rows.Close()

	var result []models.IndexInfo
	for rows.Next() {
		var ii models.IndexInfo
		if err := rows.Scan(&ii.IndexName, &ii.TableName, &ii.IndexDef); err != nil {
			return nil, fmt.Errorf("scan index: %w", err)
		}
		ii.IsUnique = contains(ii.IndexDef, "UNIQUE")
		ii.IsPrimary = contains(ii.IndexDef, "PRIMARY KEY")
		if contains(ii.IndexDef, "btree") {
			ii.IndexType = "btree"
		} else if contains(ii.IndexDef, "hash") {
			ii.IndexType = "hash"
		} else if contains(ii.IndexDef, "gin") {
			ii.IndexType = "gin"
		} else if contains(ii.IndexDef, "gist") {
			ii.IndexType = "gist"
		} else if contains(ii.IndexDef, "brin") {
			ii.IndexType = "brin"
		}
		result = append(result, ii)
	}
	return result, nil
}

func (s *ExplorerService) ListConstraints(ctx context.Context, schema string) ([]models.ConstraintInfo, error) {
	query := `
		SELECT
			tc.constraint_name,
			tc.constraint_type,
			tc.table_name,
			coalesce(kcu.column_name, '') AS column_name,
			coalesce(ccu.table_name, '') AS ref_table_name,
			coalesce(ccu.column_name, '') AS ref_column_name
		FROM information_schema.table_constraints tc
		LEFT JOIN information_schema.key_column_usage kcu
			ON tc.constraint_catalog = kcu.constraint_catalog
			AND tc.constraint_schema = kcu.constraint_schema
			AND tc.constraint_name = kcu.constraint_name
		LEFT JOIN information_schema.constraint_column_usage ccu
			ON tc.constraint_catalog = ccu.constraint_catalog
			AND tc.constraint_schema = ccu.constraint_schema
			AND tc.constraint_name = ccu.constraint_name
		WHERE tc.table_schema = $1
		ORDER BY tc.table_name, tc.constraint_name, kcu.ordinal_position
	`

	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("list constraints: %w", err)
	}
	defer rows.Close()

	var result []models.ConstraintInfo
	for rows.Next() {
		var ci models.ConstraintInfo
		if err := rows.Scan(&ci.ConstraintName, &ci.ConstraintType, &ci.TableName, &ci.ColumnName, &ci.RefTableName, &ci.RefColumnName); err != nil {
			return nil, fmt.Errorf("scan constraint: %w", err)
		}
		result = append(result, ci)
	}
	return result, nil
}

func (s *ExplorerService) ListExtensions(ctx context.Context) ([]models.ExtensionInfo, error) {
	query := `
		SELECT
			e.name,
			e.default_version,
			coalesce(x.extversion, '') AS installed_version,
			x.extversion IS NOT NULL AS installed,
			coalesce(e.comment, '') AS comment
		FROM pg_catalog.pg_available_extensions e
		LEFT JOIN pg_catalog.pg_extension x ON x.extname = e.name
		ORDER BY e.name
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list extensions: %w", err)
	}
	defer rows.Close()

	var result []models.ExtensionInfo
	for rows.Next() {
		var ei models.ExtensionInfo
		if err := rows.Scan(&ei.Name, &ei.DefaultVersion, &ei.InstalledVersion, &ei.Installed, &ei.Comment); err != nil {
			return nil, fmt.Errorf("scan extension: %w", err)
		}
		result = append(result, ei)
	}
	return result, nil
}

func (s *ExplorerService) ListSequences(ctx context.Context, schema string) ([]models.SequenceInfo, error) {
	query := `
		SELECT
			sequence_name,
			data_type,
			start_value,
			minimum_value,
			maximum_value,
			increment,
			CASE WHEN cycle_option = 'YES' THEN true ELSE false END AS is_cyclic,
			cache_size
		FROM information_schema.sequences
		WHERE sequence_schema = $1
		ORDER BY sequence_name
	`

	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("list sequences: %w", err)
	}
	defer rows.Close()

	var result []models.SequenceInfo
	for rows.Next() {
		var si models.SequenceInfo
		if err := rows.Scan(&si.SequenceName, &si.DataType, &si.StartValue, &si.MinValue, &si.MaxValue, &si.IncrementBy, &si.IsCyclic, &si.CacheSize); err != nil {
			return nil, fmt.Errorf("scan sequence: %w", err)
		}
		result = append(result, si)
	}
	return result, nil
}

func (s *ExplorerService) ListMaterializedViews(ctx context.Context, schema string) ([]models.MaterializedViewInfo, error) {
	query := `
		SELECT
			c.relname,
			COALESCE(u.usename, '') AS owner
		FROM pg_catalog.pg_class c
		JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
		LEFT JOIN pg_catalog.pg_user u ON u.usesysid = c.relowner
		WHERE n.nspname = $1
		  AND c.relkind = 'm'
		ORDER BY c.relname
	`

	rows, err := s.db.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("list materialized views: %w", err)
	}
	defer rows.Close()

	var result []models.MaterializedViewInfo
	for rows.Next() {
		var mvi models.MaterializedViewInfo
		if err := rows.Scan(&mvi.ViewName, &mvi.Owner); err != nil {
			return nil, fmt.Errorf("scan materialized view: %w", err)
		}
		result = append(result, mvi)
	}
	return result, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
