package models

type DBStats struct {
	Databases    int               `json:"databases"`
	Version      string            `json:"version"`
	TotalSize    string            `json:"total_size"`
	ActiveConns  int               `json:"active_connections"`
	Uptime       string            `json:"uptime"`
	CacheHitRate float64           `json:"cache_hit_ratio"`
	Transactions int64             `json:"transactions"`
	Replication  []ReplicationInfo `json:"replication,omitempty"`
}

type ReplicationInfo struct {
	AppName    string `json:"app_name"`
	State      string `json:"state"`
	SyncState  string `json:"sync_state"`
	ClientAddr string `json:"client_addr"`
	LagBytes   string `json:"lag_bytes"`
}

type SchemaInfo struct {
	SchemaName string `json:"schema_name"`
	Owner      string `json:"owner"`
	SizeBytes  int64  `json:"size_bytes"`
	TableCount int    `json:"table_count"`
}

type TableCountResult struct {
	Schema string `json:"schema"`
	Count  int    `json:"count"`
}

type DashboardOverview struct {
	Stats       DBStats      `json:"stats"`
	Schemas     []SchemaInfo `json:"schemas"`
	TableCounts []TableCountResult `json:"table_counts"`
}
