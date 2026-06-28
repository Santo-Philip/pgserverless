package models

import "fmt"

type ColumnDef struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Nullable     bool    `json:"nullable"`
	IsPK         bool    `json:"is_pk"`
	DefaultValue *string `json:"default_value,omitempty"`
}

type CreateTableRequest struct {
	Name        string       `json:"name"`
	Columns     []ColumnDef  `json:"columns"`
	IfNotExists bool         `json:"if_not_exists"`
}

type AlterTableOperation struct {
	Operation    string  `json:"operation"`
	ColumnName   string  `json:"column_name"`
	DataType     string  `json:"data_type,omitempty"`
	Nullable     *bool   `json:"nullable,omitempty"`
	DefaultValue *string `json:"default_value,omitempty"`
}

type AlterTableRequest struct {
	Operations []AlterTableOperation `json:"operations"`
}

type DropTableRequest struct {
	Name     string `json:"name"`
	Cascade  bool   `json:"cascade"`
}

type AddColumnRequest struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Nullable     bool    `json:"nullable"`
	DefaultValue *string `json:"default_value,omitempty"`
}

type AlterColumnRequest struct {
	DataType     string  `json:"data_type,omitempty"`
	Nullable     *bool   `json:"nullable,omitempty"`
	DefaultValue *string `json:"default_value,omitempty"`
}

type DropColumnRequest struct {
	Name string `json:"name"`
}

type AddConstraintRequest struct {
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Columns     []string `json:"columns,omitempty"`
	RefTable    string   `json:"ref_table,omitempty"`
	RefColumn   string   `json:"ref_column,omitempty"`
	CheckExpr   string   `json:"check_expr,omitempty"`
}

type DropConstraintRequest struct {
	Name string `json:"name"`
}

type CreateIndexRequest struct {
	Name    string   `json:"name"`
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Method  string   `json:"method"`
	Where   string   `json:"where,omitempty"`
}

type DropIndexRequest struct {
	Name string `json:"name"`
}

type SequenceOptions struct {
	Start     *int64  `json:"start,omitempty"`
	Increment *int64  `json:"increment,omitempty"`
	MinValue  *int64  `json:"min_value,omitempty"`
	MaxValue  *int64  `json:"max_value,omitempty"`
	Cycle     *bool   `json:"cycle,omitempty"`
	Cache     *int64  `json:"cache,omitempty"`
	Owner     *string `json:"owner,omitempty"`
}

type CreateSequenceRequest struct {
	Name    string          `json:"name"`
	Options *SequenceOptions `json:"options,omitempty"`
}

type AlterSequenceRequest struct {
	Options SequenceOptions `json:"options"`
}

type DropSequenceRequest struct {
	Name string `json:"name"`
}

type CreateSchemaRequest struct {
	Name string `json:"name"`
}

type DropSchemaRequest struct {
	Name    string `json:"name"`
	Cascade bool   `json:"cascade"`
}

type DDLResponse struct {
	DDL string `json:"ddl"`
}

func (r *CreateTableRequest) Validate() string {
	if r.Name == "" {
		return "name is required"
	}
	if len(r.Columns) == 0 {
		return "at least one column is required"
	}
	for i, c := range r.Columns {
		if c.Name == "" {
			return jsonError("columns[%d].name is required", i)
		}
		if c.Type == "" {
			return jsonError("columns[%d].type is required", i)
		}
	}
	return ""
}

func jsonError(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
