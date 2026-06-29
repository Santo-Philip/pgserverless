package models

type ActiveSession struct {
	PID        int32   `json:"pid"`
	User       string  `json:"user"`
	Database   string  `json:"database"`
	State      string  `json:"state"`
	Query      string  `json:"query"`
	WaitEvent  *string `json:"wait_event"`
	QueryStart *string `json:"query_start"`
	ClientAddr *string `json:"client_addr,omitempty"`
	BackendStart *string `json:"backend_start,omitempty"`
}

type SlowQuery struct {
	PID        int32   `json:"pid"`
	User       string  `json:"user"`
	Database   string  `json:"database"`
	Query      string  `json:"query"`
	Duration   float64 `json:"duration_seconds"`
	State      string  `json:"state"`
	QueryStart *string `json:"query_start"`
	WaitEvent  *string `json:"wait_event,omitempty"`
}

type LockInfo struct {
	PID        int32   `json:"pid"`
	LockType   string  `json:"lock_type"`
	Mode       string  `json:"mode"`
	Granted    bool    `json:"granted"`
	Relation   *string `json:"relation,omitempty"`
	RelationID *int32  `json:"relation_id,omitempty"`
	User       string  `json:"user"`
	Query      string  `json:"query"`
}

type WaitingQuery struct {
	PID        int32   `json:"pid"`
	User       string  `json:"user"`
	Database   string  `json:"database"`
	Query      string  `json:"query"`
	Duration   float64 `json:"duration_seconds"`
	WaitEvent  string  `json:"wait_event"`
	State      string  `json:"state"`
}

type QueryStats struct {
	Query          string   `json:"query"`
	Calls          *int64   `json:"calls,omitempty"`
	TotalTime      *float64 `json:"total_time_ms,omitempty"`
	MeanTime       *float64 `json:"mean_time_ms,omitempty"`
	Rows           *int64   `json:"rows,omitempty"`
	SharedBlksHit  *int64   `json:"shared_blks_hit,omitempty"`
	SharedBlksRead *int64   `json:"shared_blks_read,omitempty"`
}

type ConnectionStats struct {
	Total               int    `json:"total"`
	Active              int    `json:"active"`
	Idle                int    `json:"idle"`
	IdleInTransaction   int    `json:"idle_in_transaction"`
	Waiting             int    `json:"waiting"`
	MaxConnections      int    `json:"max_connections"`
}

type CacheStats struct {
	HitRate   float64 `json:"hit_rate"`
	Reads     int64   `json:"reads"`
	Hits      int64   `json:"hits"`
}

type DatabaseStat struct {
	Datname       string  `json:"datname"`
	NumBackends   int32   `json:"num_backends"`
	XactCommit    int64   `json:"xact_commit"`
	XactRollback  int64   `json:"xact_rollback"`
	BlksRead      int64   `json:"blks_read"`
	BlksHit       int64   `json:"blks_hit"`
	CacheHitRatio float64 `json:"cache_hit_ratio"`
	SizeBytes     *int64  `json:"size_bytes,omitempty"`
}

type TableStat struct {
	Schemaname  string  `json:"schemaname"`
	Tablename   string  `json:"tablename"`
	SeqScan     int64   `json:"seq_scan"`
	SeqTupRead  int64   `json:"seq_tup_read"`
	IdxScan     int64   `json:"idx_scan"`
	IdxTupFetch int64   `json:"idx_tup_fetch"`
	NTupIns     int64   `json:"n_tup_ins"`
	NTupUpd     int64   `json:"n_tup_upd"`
	NTupDel     int64   `json:"n_tup_del"`
	LiveTup     *int64  `json:"live_tup,omitempty"`
	DeadTup     *int64  `json:"dead_tup,omitempty"`
}

type IndexStat struct {
	Schemaname  string `json:"schemaname"`
	Indexname   string `json:"indexname"`
	Tablename   string `json:"tablename"`
	IdxScan     int64  `json:"idx_scan"`
	IdxTupRead  int64  `json:"idx_tup_read"`
	IdxTupFetch int64  `json:"idx_tup_fetch"`
}

type TerminateRequest struct {
	PID int32 `json:"pid"`
}

type CancelRequest struct {
	PID int32 `json:"pid"`
}

type TerminateResponse struct {
	PID     int32  `json:"pid"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}
