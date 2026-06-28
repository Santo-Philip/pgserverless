// Auth
export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  refresh_token: string;
  user: User;
  expires_at: string;
}

export interface User {
  id: string;
  email: string;
  name: string;
  role: 'admin' | 'superadmin' | 'viewer';
  is_active: boolean;
  last_login_at?: string;
  created_at: string;
}

export interface CreateUserRequest {
  email: string;
  password: string;
  name: string;
  role: 'admin' | 'superadmin' | 'viewer';
}

export interface UpdateUserRequest {
  email?: string;
  name?: string;
  role?: 'admin' | 'superadmin' | 'viewer';
  is_active?: boolean;
}

export interface UpdatePasswordRequest {
  current_password: string;
  new_password: string;
}

// Dashboard
export interface DashboardOverview {
  pg_version: string;
  db_size: number;
  active_connections: number;
  uptime: string;
  cache_hit_ratio: number;
  tps: number;
  replication_status: 'healthy' | 'degraded' | 'down';
  databases: string[];
  schemas: string[];
  table_count: number;
}

export interface SchemaInfo {
  schema_name: string;
  owner: string;
  size_bytes: number;
  table_count: number;
}

export interface TableInfo {
  schema_name: string;
  table_name: string;
  row_estimate: number;
  has_pk: boolean;
  column_count: number;
  size_bytes: number;
}

export interface ColumnInfo {
  name: string;
  data_type: string;
  is_nullable: boolean;
  is_pk: boolean;
  default_value: string | null;
  is_generated: boolean;
}

export interface ViewInfo {
  schema_name: string;
  view_name: string;
  definition: string;
  is_materialized: boolean;
}

export interface FunctionInfo {
  schema_name: string;
  function_name: string;
  arguments: string;
  return_type: string;
  language: string;
  is_aggregate: boolean;
  is_window: boolean;
  security_definer: boolean;
}

export interface ProcedureInfo {
  schema_name: string;
  procedure_name: string;
  arguments: string;
  language: string;
  security_definer: boolean;
}

export interface TriggerInfo {
  schema_name: string;
  table_name: string;
  trigger_name: string;
  event: string;
  timing: string;
  procedure: string;
  enabled: boolean;
}

export interface IndexInfo {
  schema_name: string;
  table_name: string;
  index_name: string;
  index_type: string;
  columns: string[];
  is_unique: boolean;
  is_primary: boolean;
  is_clustered: boolean;
  size_bytes: number;
  scan_type: string;
}

export interface ConstraintInfo {
  schema_name: string;
  table_name: string;
  constraint_name: string;
  constraint_type: 'PRIMARY KEY' | 'FOREIGN KEY' | 'UNIQUE' | 'CHECK' | 'EXCLUDE';
  columns: string[];
  ref_table?: string;
  ref_columns?: string[];
  definition: string;
}

export interface ExtensionInfo {
  name: string;
  version: string;
  installed_version: string | null;
  comment: string;
  installed: boolean;
}

export interface SequenceInfo {
  schema_name: string;
  sequence_name: string;
  data_type: string;
  start_value: number;
  min_value: number;
  max_value: number;
  increment_by: number;
  cache_size: number;
  is_cycling: boolean;
  last_value: number | null;
}

export interface MaterializedViewInfo {
  schema_name: string;
  view_name: string;
  row_count: number;
  size_bytes: number;
  definition: string;
  last_refresh: string | null;
}

// SQL
export interface QueryRequest {
  query: string;
  params?: unknown[];
}

export interface ExecuteResponse {
  columns: { name: string; data_type: string }[];
  rows: Record<string, unknown>[];
  row_count: number;
  duration_ms: number;
}

export interface ExplainResult {
  plan: Record<string, unknown>;
}

export interface SavedQuery {
  id: string;
  name: string;
  query: string;
  database: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface QueryHistory {
  id: string;
  query: string;
  database: string;
  duration_ms: number;
  row_count: number;
  status: string;
  error?: string;
  executed_by: string;
  executed_at: string;
}

// Monitoring
export interface ActiveSession {
  pid: number;
  database: string;
  user: string;
  application_name: string;
  client_addr: string;
  state: string;
  query: string;
  query_start: string;
  wait_event: string | null;
  wait_event_type: string | null;
  backend_type: string;
}

export interface SlowQuery {
  pid: number;
  database: string;
  user: string;
  query: string;
  duration_ms: number;
  state: string;
  query_start: string;
  wait_event: string | null;
}

export interface LockInfo {
  pid: number;
  database: string;
  relation: string;
  lock_type: string;
  lock_mode: string;
  granted: boolean;
  blocked_by: number[];
  blocked_query: string;
  blocked_pid: number;
  waiting_pid: number;
  waiting_query: string;
}

export interface WaitingQuery {
  pid: number;
  database: string;
  user: string;
  query: string;
  waiting_for: string;
  blocked_by_pid: number;
  wait_duration_ms: number;
  state: string;
}

export interface QueryStats {
  queryid: number;
  query: string;
  calls: number;
  total_time_ms: number;
  mean_time_ms: number;
  min_time_ms: number;
  max_time_ms: number;
  rows_returned: number;
  shared_blks_hit: number;
  shared_blks_read: number;
  hit_ratio: number;
}

export interface ConnectionStats {
  total: number;
  active: number;
  idle: number;
  idle_in_transaction: number;
  waiting: number;
  max_connections: number;
  by_database: { database: string; count: number }[];
}

export interface CacheStats {
  hit_ratio: number;
  shared_hit: number;
  shared_read: number;
  local_hit: number;
  local_read: number;
  temp_read: number;
  hit_ratio_wal: number;
}

export interface DatabaseStat {
  database: string;
  size_bytes: number;
  connections: number;
  transactions_committed: number;
  transactions_rolled_back: number;
  tuples_returned: number;
  tuples_fetched: number;
  tuples_inserted: number;
  tuples_updated: number;
  tuples_deleted: number;
  hit_ratio: number;
}

export interface TableStat {
  schema_name: string;
  table_name: string;
  seq_scan: number;
  seq_tup_read: number;
  idx_scan: number;
  idx_tup_fetch: number;
  n_tup_ins: number;
  n_tup_upd: number;
  n_tup_del: number;
  n_tup_hot_upd: number;
  n_live_tup: number;
  n_dead_tup: number;
  vacuum_count: number;
  analyze_count: number;
  last_vacuum: string | null;
  last_analyze: string | null;
}

export interface IndexStat {
  schema_name: string;
  table_name: string;
  index_name: string;
  idx_scan: number;
  idx_tup_read: number;
  idx_tup_fetch: number;
  size_bytes: number;
  unique: boolean;
}

// Backups
export interface BackupInfo {
  id: string;
  database: string;
  size_bytes: number;
  status: string;
  started_at: string;
  finished_at: string | null;
  type: string;
  created_by: string;
  verified: boolean;
  download_url?: string;
}

// Logs
export interface LogEntry {
  id: string;
  timestamp: string;
  level: string;
  message: string;
  source: string;
  pid?: number;
  database?: string;
  user?: string;
}

export interface AuditLog {
  id: string;
  actor_id: string;
  actor_name: string;
  action: string;
  resource: string;
  resource_id: string;
  metadata?: Record<string, unknown>;
  ip_address: string;
  user_agent: string;
  created_at: string;
}

// Roles & Privileges
export interface PgRole {
  rolname: string;
  rolsuper: boolean;
  rolinherit: boolean;
  rolcreaterole: boolean;
  rolcreatedb: boolean;
  rolcanlogin: boolean;
  rolreplication: boolean;
  rolconnlimit: number;
  rolvaliduntil: string | null;
  member_of: string[];
  members: string[];
}

export interface PgPrivilege {
  grantor: string;
  grantee: string;
  privilege_type: string;
  is_grantable: boolean;
}

// Table data response
export interface TableRowResponse {
  columns: { name: string; data_type: string }[];
  rows: Record<string, unknown>[];
  total: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}
