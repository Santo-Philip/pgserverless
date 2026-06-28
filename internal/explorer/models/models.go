package models

type SchemaInfo struct {
	SchemaName string `json:"schema_name"`
	Owner      string `json:"owner,omitempty"`
}

type TableInfo struct {
	TableName  string  `json:"table_name"`
	TableType  string  `json:"table_type"`
	RowEstimate *int64 `json:"row_estimate,omitempty"`
}

type ColumnInfo struct {
	ColumnName            string  `json:"column_name"`
	DataType              string  `json:"data_type"`
	IsNullable            string  `json:"is_nullable"`
	ColumnDefault         *string `json:"column_default,omitempty"`
	CharacterMaximumLength *int   `json:"character_maximum_length,omitempty"`
	NumericPrecision      *int    `json:"numeric_precision,omitempty"`
	NumericScale          *int    `json:"numeric_scale,omitempty"`
	OrdinalPosition       int     `json:"ordinal_position"`
}

type ConstraintInfo struct {
	ConstraintName string `json:"constraint_name"`
	ConstraintType string `json:"constraint_type"`
	TableName      string `json:"table_name"`
	ColumnName     string `json:"column_name,omitempty"`
	RefTableName   string `json:"ref_table_name,omitempty"`
	RefColumnName  string `json:"ref_column_name,omitempty"`
}

type IndexInfo struct {
	IndexName  string `json:"index_name"`
	IndexType  string `json:"index_type"`
	IndexDef   string `json:"index_def"`
	IsUnique   bool   `json:"is_unique"`
	IsPrimary  bool   `json:"is_primary"`
	TableName  string `json:"table_name"`
}

type TriggerInfo struct {
	TriggerName    string `json:"trigger_name"`
	EventManipulation string `json:"event_manipulation"`
	ActionTiming   string `json:"action_timing"`
	ActionStatement string `json:"action_statement"`
	TableName      string `json:"table_name"`
}

type ViewInfo struct {
	ViewName       string `json:"view_name"`
	ViewDefinition string `json:"view_definition"`
}

type MaterializedViewInfo struct {
	ViewName string `json:"view_name"`
	Owner    string `json:"owner,omitempty"`
}

type FunctionInfo struct {
	FunctionName string `json:"function_name"`
	ReturnType   string `json:"return_type"`
	Arguments    string `json:"arguments"`
	Language     string `json:"language"`
}

type ProcedureInfo struct {
	ProcedureName string `json:"procedure_name"`
	Arguments     string `json:"arguments"`
	Language      string `json:"language"`
}

type SequenceInfo struct {
	SequenceName string `json:"sequence_name"`
	DataType     string `json:"data_type"`
	StartValue   int64  `json:"start_value"`
	MinValue     int64  `json:"min_value"`
	MaxValue     int64  `json:"max_value"`
	IncrementBy  int64  `json:"increment_by"`
	IsCyclic     bool   `json:"is_cyclic"`
	CacheSize    int64  `json:"cache_size"`
}

type ExtensionInfo struct {
	Name           string `json:"name"`
	DefaultVersion string `json:"default_version"`
	InstalledVersion string `json:"installed_version,omitempty"`
	Installed      bool   `json:"installed"`
	Comment        string `json:"comment,omitempty"`
}

type TableDetails struct {
	Table       TableInfo       `json:"table"`
	Columns     []ColumnInfo    `json:"columns"`
	Constraints []ConstraintInfo `json:"constraints"`
	Indexes     []IndexInfo     `json:"indexes"`
	Triggers    []TriggerInfo   `json:"triggers"`
}
