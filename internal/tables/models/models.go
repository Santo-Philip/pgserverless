package models

type Filter struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    any    `json:"value,omitempty"`
}

type QueryRequest struct {
	Limit         int      `json:"limit"`
	Offset        int      `json:"offset"`
	Sort          string   `json:"sort"`
	Order         string   `json:"order"`
	Filters       []Filter `json:"filters,omitempty"`
	Search        string   `json:"search,omitempty"`
	SearchColumns []string `json:"search_columns,omitempty"`
}

type InsertRowRequest struct {
	Data map[string]any `json:"data"`
}

type UpdateRowRequest struct {
	Data  map[string]any `json:"data"`
	Where map[string]any `json:"where"`
}

type DeleteRowRequest struct {
	Where map[string]any `json:"where"`
}

type BulkInsertRequest struct {
	Rows []map[string]any `json:"rows"`
}

type BulkDeleteRequest struct {
	IDs      []any  `json:"ids"`
	IDColumn string `json:"id_column"`
}

type SearchRequest struct {
	Search  string   `json:"search"`
	Columns []string `json:"columns"`
	Limit   int      `json:"limit"`
}

type CSVExportConfig struct {
	Schema        string   `json:"schema"`
	Table         string   `json:"table"`
	Columns       []string `json:"columns,omitempty"`
	Filters       []Filter `json:"filters,omitempty"`
	Delimiter     string   `json:"delimiter,omitempty"`
	IncludeHeader bool     `json:"include_header"`
}

type CSVImportConfig struct {
	Schema     string   `json:"schema"`
	Table      string   `json:"table"`
	Columns    []string `json:"columns"`
	Delimiter  string   `json:"delimiter,omitempty"`
	SkipHeader bool     `json:"skip_header"`
	OnConflict string   `json:"on_conflict,omitempty"`
}
