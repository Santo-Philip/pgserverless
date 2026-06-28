import { browser } from '$app/environment';
import { env } from '$env/dynamic/public';
import type {
  AuthResponse, User, CreateUserRequest, UpdateUserRequest, UpdatePasswordRequest,
  DashboardOverview, SchemaInfo,
  TableInfo, ColumnInfo, ViewInfo, FunctionInfo, ProcedureInfo, TriggerInfo,
  IndexInfo, ConstraintInfo, ExtensionInfo, SequenceInfo, MaterializedViewInfo,
  QueryRequest, ExecuteResponse, ExplainResult, SavedQuery, QueryHistory,
  ActiveSession, SlowQuery, LockInfo, WaitingQuery, QueryStats,
  ConnectionStats, CacheStats, DatabaseStat, TableStat, IndexStat,
  BackupInfo, LogEntry, AuditLog,
  PgRole, PgPrivilege,
  TableRowResponse, PaginatedResponse
} from '$lib/types';

function getBaseUrl(): string {
  if (env.PUBLIC_API_URL) return env.PUBLIC_API_URL;
  if (browser) return window.location.origin;
  return '';
}

interface SuccessEnvelope {
  message: string;
  data: unknown;
}

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

class ApiClient {
  private token: string | null = null;
  private refreshToken: string | null = null;
  private expiresAt: number | null = null;
  private refreshPromise: Promise<void> | null = null;

  constructor() {
    if (browser) {
      this.token = localStorage.getItem('pgadmin_token');
      this.refreshToken = localStorage.getItem('pgadmin_refresh_token');
      const exp = localStorage.getItem('pgadmin_expires_at');
      this.expiresAt = exp ? parseInt(exp, 10) : null;
    }
  }

  setToken(token: string) {
    this.token = token;
    if (browser) localStorage.setItem('pgadmin_token', token);
  }

  setRefreshToken(token: string) {
    this.refreshToken = token;
    if (browser) localStorage.setItem('pgadmin_refresh_token', token);
  }

  setExpiresAt(iso: string) {
    this.expiresAt = new Date(iso).getTime();
    if (browser) localStorage.setItem('pgadmin_expires_at', String(this.expiresAt));
  }

  clearToken() {
    this.token = null;
    this.refreshToken = null;
    this.expiresAt = null;
    if (browser) {
      localStorage.removeItem('pgadmin_token');
      localStorage.removeItem('pgadmin_refresh_token');
      localStorage.removeItem('pgadmin_expires_at');
    }
  }

  get isAuthenticated(): boolean {
    return !!this.token;
  }

  private isTokenExpired(): boolean {
    if (!this.expiresAt) return false;
    return Date.now() >= this.expiresAt;
  }

  private async tryRefreshToken(): Promise<void> {
    if (!this.refreshToken) throw new Error('No refresh token');
    const response = await fetch(`${getBaseUrl()}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: this.refreshToken }),
    });
    if (!response.ok) {
      this.clearToken();
      throw new Error('Session expired');
    }
    const json = await response.json();
    const result = json.data || json;
    this.setToken(result.token);
    this.setRefreshToken(result.refresh_token || this.refreshToken);
    if (result.expires_at) this.setExpiresAt(result.expires_at);
  }

  private async ensureAuth(): Promise<void> {
    if (!this.token) return;
    if (!this.isTokenExpired()) return;
    if (this.refreshPromise) return this.refreshPromise;
    this.refreshPromise = this.tryRefreshToken().finally(() => { this.refreshPromise = null; });
    return this.refreshPromise;
  }

  private unwrap<T>(json: unknown): T {
    if (json && typeof json === 'object' && 'message' in json && 'data' in json) {
      return (json as SuccessEnvelope).data as T;
    }
    return json as T;
  }

  private async request<T>(method: string, path: string, body?: unknown, retried = false): Promise<T> {
    await this.ensureAuth();
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    };
    if (this.token) headers['Authorization'] = `Bearer ${this.token}`;
    const response = await fetch(`${getBaseUrl()}${path}`, { method, headers, body: body ? JSON.stringify(body) : undefined });
    if (response.status === 401 && !retried && this.refreshToken) {
      try {
        await this.tryRefreshToken();
        return this.request<T>(method, path, body, true);
      } catch {
        this.clearToken();
        throw new ApiError('Session expired. Please log in again.', 401);
      }
    }
    if (response.status === 401) {
      this.clearToken();
      throw new ApiError('Unauthorized', 401);
    }
    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: response.statusText }));
      throw new ApiError(error.message || `HTTP ${response.status}`, response.status);
    }
    const json = await response.json();
    return this.unwrap<T>(json);
  }

  async get<T>(path: string): Promise<T> { return this.request<T>('GET', path); }
  async post<T>(path: string, body?: unknown): Promise<T> { return this.request<T>('POST', path, body); }
  async patch<T>(path: string, body?: unknown): Promise<T> { return this.request<T>('PATCH', path, body); }
  async del<T>(path: string, body?: unknown): Promise<T> { return this.request<T>('DELETE', path, body); }

  // ── Auth ─────────────────────────────────────────────
  async login(email: string, password: string) {
    const result = await this.post<AuthResponse>('/api/v1/auth/login', { email, password });
    this.setToken(result.token);
    if (result.refresh_token) this.setRefreshToken(result.refresh_token);
    if (result.expires_at) this.setExpiresAt(result.expires_at);
    return result;
  }

  async refreshTokenEndpoint() {
    return this.post<AuthResponse>('/api/v1/auth/refresh');
  }

  async getMe() {
    return this.get<User>('/api/v1/auth/me');
  }

  async updatePassword(data: UpdatePasswordRequest) {
    return this.patch<void>('/api/v1/auth/password', data);
  }

  // ── Admin Users ──────────────────────────────────────
  async listUsers() {
    const result = await this.get<PaginatedResponse<User>>('/api/v1/admin/users');
    return result?.data ?? [];
  }

  async getUser(id: string) {
    return this.get<User>(`/api/v1/admin/users/${id}`);
  }

  async createUser(data: CreateUserRequest) {
    return this.post<User>('/api/v1/admin/users', data);
  }

  async updateUser(id: string, data: UpdateUserRequest) {
    return this.patch<User>(`/api/v1/admin/users/${id}`, data);
  }

  async deleteUser(id: string) {
    return this.del<void>(`/api/v1/admin/users/${id}`);
  }

  async updateUserPassword(id: string, password: string) {
    return this.patch<void>(`/api/v1/admin/users/${id}/password`, { password });
  }

  // ── Dashboard ────────────────────────────────────────
  async getOverview() {
    return this.get<DashboardOverview>('/api/v1/dashboard/overview');
  }

  async getStats() {
    return this.get<Record<string, unknown>>('/api/v1/dashboard/stats');
  }

  async getSchemas() {
    const result = await this.get<SchemaInfo[]>('/api/v1/dashboard/schemas');
    return Array.isArray(result) ? result : [];
  }

  // ── Explorer ─────────────────────────────────────────
  async listSchemas() {
    const result = await this.get<SchemaInfo[]>('/api/v1/explorer/schemas');
    return Array.isArray(result) ? result : [];
  }

  async listTables(schema: string) {
    const result = await this.get<TableInfo[]>(`/api/v1/explorer/schemas/${schema}/tables`);
    return Array.isArray(result) ? result : [];
  }

  async getTableDetails(schema: string, table: string) {
    return this.get<{ columns: ColumnInfo[]; info: TableInfo }>(`/api/v1/explorer/schemas/${schema}/tables/${table}`);
  }

  async listViews(schema: string) {
    const result = await this.get<ViewInfo[]>(`/api/v1/explorer/schemas/${schema}/views`);
    return Array.isArray(result) ? result : [];
  }

  async listFunctions(schema: string) {
    const result = await this.get<FunctionInfo[]>(`/api/v1/explorer/schemas/${schema}/functions`);
    return Array.isArray(result) ? result : [];
  }

  async listProcedures(schema: string) {
    const result = await this.get<ProcedureInfo[]>(`/api/v1/explorer/schemas/${schema}/procedures`);
    return Array.isArray(result) ? result : [];
  }

  async listTriggers(schema: string, table?: string) {
    const path = table
      ? `/api/v1/explorer/schemas/${schema}/tables/${table}/triggers`
      : `/api/v1/explorer/schemas/${schema}/triggers`;
    const result = await this.get<TriggerInfo[]>(path);
    return Array.isArray(result) ? result : [];
  }

  async listIndexes(schema: string, table?: string) {
    const path = table
      ? `/api/v1/explorer/schemas/${schema}/tables/${table}/indexes`
      : `/api/v1/explorer/schemas/${schema}/indexes`;
    const result = await this.get<IndexInfo[]>(path);
    return Array.isArray(result) ? result : [];
  }

  async listConstraints(schema: string, table?: string) {
    const path = table
      ? `/api/v1/explorer/schemas/${schema}/tables/${table}/constraints`
      : `/api/v1/explorer/schemas/${schema}/constraints`;
    const result = await this.get<ConstraintInfo[]>(path);
    return Array.isArray(result) ? result : [];
  }

  async listSequences(schema: string) {
    const result = await this.get<SequenceInfo[]>(`/api/v1/explorer/schemas/${schema}/sequences`);
    return Array.isArray(result) ? result : [];
  }

  async listMaterializedViews(schema: string) {
    const result = await this.get<MaterializedViewInfo[]>(`/api/v1/explorer/schemas/${schema}/materialized-views`);
    return Array.isArray(result) ? result : [];
  }

  async listExtensions(schema?: string) {
    const path = schema ? `/api/v1/explorer/schemas/${schema}/extensions` : '/api/v1/explorer/extensions';
    const result = await this.get<ExtensionInfo[]>(path);
    return Array.isArray(result) ? result : [];
  }

  // ── Table Data ───────────────────────────────────────
  async queryTable(schema: string, table: string, limit = 100, offset = 0, sort?: string, order?: 'asc' | 'desc') {
    let path = `/api/v1/tables/${schema}/${table}?limit=${limit}&offset=${offset}`;
    if (sort) path += `&sort=${sort}&order=${order || 'asc'}`;
    return this.get<TableRowResponse>(path);
  }

  async insertRow(schema: string, table: string, values: Record<string, unknown>) {
    return this.post<Record<string, unknown>>(`/api/v1/tables/${schema}/${table}/rows`, { values });
  }

  async updateRow(schema: string, table: string, values: Record<string, unknown>, where: Record<string, unknown>) {
    return this.patch<Record<string, unknown>>(`/api/v1/tables/${schema}/${table}/rows`, { values, where });
  }

  async deleteRow(schema: string, table: string, where: Record<string, unknown>) {
    return this.del<void>(`/api/v1/tables/${schema}/${table}/rows`, { where });
  }

  async bulkInsert(schema: string, table: string, rows: Record<string, unknown>[]) {
    return this.post<{ inserted: number }>(`/api/v1/tables/${schema}/${table}/rows/bulk`, { rows });
  }

  async bulkDelete(schema: string, table: string, ids: unknown[]) {
    return this.del<{ deleted: number }>(`/api/v1/tables/${schema}/${table}/rows/bulk`, { ids });
  }

  async searchTable(schema: string, table: string, query: string, limit = 50) {
    return this.get<TableRowResponse>(`/api/v1/tables/${schema}/${table}/search?q=${encodeURIComponent(query)}&limit=${limit}`);
  }

  // ── SQL ──────────────────────────────────────────────
  async executeSQL(data: QueryRequest) {
    return this.post<ExecuteResponse>('/api/v1/sql/execute', data);
  }

  async explainQuery(data: QueryRequest) {
    return this.post<ExplainResult>('/api/v1/sql/explain', data);
  }

  async cancelQuery(pid: number) {
    return this.post<void>('/api/v1/sql/cancel', { pid });
  }

  async getQueryHistory(limit = 50, offset = 0) {
    const result = await this.get<PaginatedResponse<QueryHistory>>(`/api/v1/sql/history?limit=${limit}&offset=${offset}`);
    return result?.data ?? [];
  }

  async getSavedQueries() {
    const result = await this.get<SavedQuery[]>('/api/v1/sql/saved');
    return Array.isArray(result) ? result : [];
  }

  async saveQuery(data: { name: string; query: string; database: string }) {
    return this.post<SavedQuery>('/api/v1/sql/saved', data);
  }

  async deleteSavedQuery(id: string) {
    return this.del<void>(`/api/v1/sql/saved/${id}`);
  }

  // ── Schema ───────────────────────────────────────────
  async createSchema(name: string) {
    return this.post<void>('/api/v1/schemas', { name });
  }

  async dropSchema(name: string, cascade = false) {
    return this.del<void>(`/api/v1/schemas/${name}?cascade=${cascade}`);
  }

  async createTable(schema: string, name: string, columns: { name: string; type: string; nullable: boolean; is_pk: boolean; default?: string }[]) {
    return this.post<void>(`/api/v1/schemas/${schema}/tables`, { name, columns });
  }

  async dropTable(schema: string, table: string) {
    return this.del<void>(`/api/v1/schemas/${schema}/tables/${table}`);
  }

  async addColumn(schema: string, table: string, column: { name: string; type: string; nullable: boolean; default?: string }) {
    return this.post<void>(`/api/v1/schemas/${schema}/tables/${table}/columns`, column);
  }

  async dropColumn(schema: string, table: string, column: string) {
    return this.del<void>(`/api/v1/schemas/${schema}/tables/${table}/columns/${column}`);
  }

  async alterColumn(schema: string, table: string, column: string, changes: { new_name?: string; new_type?: string; nullable?: boolean; default?: string | null }) {
    return this.patch<void>(`/api/v1/schemas/${schema}/tables/${table}/columns/${column}`, changes);
  }

  async addConstraint(schema: string, table: string, constraint: { name: string; type: string; columns: string[]; ref_table?: string; ref_columns?: string[]; definition?: string }) {
    return this.post<void>(`/api/v1/schemas/${schema}/tables/${table}/constraints`, constraint);
  }

  async dropConstraint(schema: string, table: string, constraint: string) {
    return this.del<void>(`/api/v1/schemas/${schema}/tables/${table}/constraints/${constraint}`);
  }

  async createIndex(schema: string, table: string, index: { name: string; columns: string[]; unique?: boolean; method?: string }) {
    return this.post<void>(`/api/v1/schemas/${schema}/tables/${table}/indexes`, index);
  }

  async dropIndex(schema: string, table: string, index: string) {
    return this.del<void>(`/api/v1/schemas/${schema}/tables/${table}/indexes/${index}`);
  }

  async createSequence(schema: string, sequence: { name: string; data_type?: string; start?: number; increment?: number; min?: number; max?: number; cache?: number; cycle?: boolean }) {
    return this.post<void>(`/api/v1/schemas/${schema}/sequences`, sequence);
  }

  async alterSequence(schema: string, sequence: string, changes: { restart?: number; increment?: number; min?: number; max?: number; cache?: number; cycle?: boolean }) {
    return this.patch<void>(`/api/v1/schemas/${schema}/sequences/${sequence}`, changes);
  }

  async dropSequence(schema: string, sequence: string) {
    return this.del<void>(`/api/v1/schemas/${schema}/sequences/${sequence}`);
  }

  async getTableDDL(schema: string, table: string) {
    const result = await this.get<{ ddl: string }>(`/api/v1/schemas/${schema}/tables/${table}/ddl`);
    return result.ddl;
  }

  // ── Roles ────────────────────────────────────────────
  async listRoles() {
    const result = await this.get<PgRole[]>('/api/v1/roles');
    return Array.isArray(result) ? result : [];
  }

  async createRole(data: { name: string; login?: boolean; superuser?: boolean; createdb?: boolean; createrole?: boolean; replication?: boolean; password?: string; connection_limit?: number; valid_until?: string }) {
    return this.post<PgRole>('/api/v1/roles', data);
  }

  async getRole(name: string) {
    return this.get<PgRole>(`/api/v1/roles/${name}`);
  }

  async alterRole(name: string, changes: Record<string, unknown>) {
    return this.patch<PgRole>(`/api/v1/roles/${name}`, changes);
  }

  async dropRole(name: string) {
    return this.del<void>(`/api/v1/roles/${name}`);
  }

  async resetPassword(name: string, password: string) {
    return this.post<void>(`/api/v1/roles/${name}/password`, { password });
  }

  async grantDatabase(role: string, database: string, permissions: string) {
    return this.post<void>(`/api/v1/roles/${role}/grant-database`, { database, permissions });
  }

  async grantSchema(role: string, schema: string, permissions: string) {
    return this.post<void>(`/api/v1/roles/${role}/grant-schema`, { schema, permissions });
  }

  async grantTable(role: string, schema: string, table: string, permissions: string) {
    return this.post<void>(`/api/v1/roles/${role}/grant-table`, { schema, table, permissions });
  }

  async revokeDatabase(role: string, database: string, permissions: string, cascade = false) {
    return this.post<void>(`/api/v1/roles/${role}/revoke-database`, { database, permissions, cascade });
  }

  async revokeSchema(role: string, schema: string, permissions: string, cascade = false) {
    return this.post<void>(`/api/v1/roles/${role}/revoke-schema`, { schema, permissions, cascade });
  }

  async revokeTable(role: string, schema: string, table: string, permissions: string, cascade = false) {
    return this.post<void>(`/api/v1/roles/${role}/revoke-table`, { schema, table, permissions, cascade });
  }

  async addMember(role: string, memberOf: string) {
    return this.post<void>(`/api/v1/roles/${role}/add-member`, { member_of: memberOf });
  }

  async removeMember(role: string, memberOf: string) {
    return this.post<void>(`/api/v1/roles/${role}/remove-member`, { member_of: memberOf });
  }

  async listDatabasePrivileges(database: string) {
    const result = await this.get<PgPrivilege[]>(`/api/v1/roles/privileges/databases?database=${encodeURIComponent(database)}`);
    return Array.isArray(result) ? result : [];
  }

  async listSchemaPrivileges(schema: string) {
    const result = await this.get<PgPrivilege[]>(`/api/v1/roles/privileges/schemas?schema=${encodeURIComponent(schema)}`);
    return Array.isArray(result) ? result : [];
  }

  async listTablePrivileges(schema: string, table: string) {
    const result = await this.get<PgPrivilege[]>(`/api/v1/roles/privileges/tables?schema=${encodeURIComponent(schema)}&table=${encodeURIComponent(table)}`);
    return Array.isArray(result) ? result : [];
  }

  async getRoleMembers(name: string) {
    const result = await this.get<{ direct: string[]; indirect: string[] }>(`/api/v1/roles/${name}/members`);
    return result;
  }

  // ── Extensions ───────────────────────────────────────
  async installExtension(name: string, schema?: string) {
    return this.post<void>('/api/v1/extensions', { name, schema });
  }

  async uninstallExtension(name: string) {
    return this.del<void>(`/api/v1/extensions/${name}`);
  }

  // ── Monitoring ───────────────────────────────────────
  async getActiveSessions() {
    const result = await this.get<ActiveSession[]>('/api/v1/monitoring/sessions');
    return Array.isArray(result) ? result : [];
  }

  async getSlowQueries(minDurationMs = 1000) {
    const result = await this.get<SlowQuery[]>(`/api/v1/monitoring/slow-queries?min_duration_ms=${minDurationMs}`);
    return Array.isArray(result) ? result : [];
  }

  async getLocks() {
    const result = await this.get<LockInfo[]>('/api/v1/monitoring/locks');
    return Array.isArray(result) ? result : [];
  }

  async getWaitingQueries() {
    const result = await this.get<WaitingQuery[]>('/api/v1/monitoring/waiting');
    return Array.isArray(result) ? result : [];
  }

  async getQueryStats(limit = 50) {
    const result = await this.get<QueryStats[]>(`/api/v1/monitoring/query-stats?limit=${limit}`);
    return Array.isArray(result) ? result : [];
  }

  async getConnectionStats() {
    return this.get<ConnectionStats>('/api/v1/monitoring/connections');
  }

  async getCacheStats() {
    return this.get<CacheStats>('/api/v1/monitoring/cache');
  }

  async getDatabaseStats() {
    const result = await this.get<DatabaseStat[]>('/api/v1/monitoring/databases');
    return Array.isArray(result) ? result : [];
  }

  async getTableStats(schema?: string) {
    const path = schema ? `/api/v1/monitoring/table-stats?schema=${schema}` : '/api/v1/monitoring/table-stats';
    const result = await this.get<TableStat[]>(path);
    return Array.isArray(result) ? result : [];
  }

  async getIndexStats(schema?: string) {
    const path = schema ? `/api/v1/monitoring/index-stats?schema=${schema}` : '/api/v1/monitoring/index-stats';
    const result = await this.get<IndexStat[]>(path);
    return Array.isArray(result) ? result : [];
  }

  async terminateSession(pid: number) {
    return this.post<void>('/api/v1/monitoring/sessions/terminate', { pid });
  }

  // ── Backups ──────────────────────────────────────────
  async listBackups() {
    const result = await this.get<BackupInfo[]>('/api/v1/backups');
    return Array.isArray(result) ? result : [];
  }

  async createBackup(database = 'postgres', type = 'full') {
    return this.post<BackupInfo>('/api/v1/backups', { database, type });
  }

  async getBackup(id: string) {
    return this.get<BackupInfo>(`/api/v1/backups/${id}`);
  }

  async deleteBackup(id: string) {
    return this.del<void>(`/api/v1/backups/${id}`);
  }

  async restoreBackup(id: string) {
    return this.post<void>(`/api/v1/backups/${id}/restore`);
  }

  async verifyBackup(id: string) {
    return this.post<{ verified: boolean }>(`/api/v1/backups/${id}/verify`);
  }

  async downloadBackup(id: string): Promise<string> {
    const result = await this.get<{ url: string }>(`/api/v1/backups/${id}/download`);
    return result.url;
  }

  // ── Logs ─────────────────────────────────────────────
  async getLogs(limit = 100, offset = 0, level?: string) {
    let path = `/api/v1/logs?limit=${limit}&offset=${offset}`;
    if (level) path += `&level=${level}`;
    const result = await this.get<PaginatedResponse<LogEntry>>(path);
    return result?.data ?? [];
  }

  async getErrorLogs(limit = 100) {
    return this.getLogs(limit, 0, 'error');
  }

  async getQueryLogs(limit = 100, offset = 0) {
    const result = await this.get<PaginatedResponse<LogEntry>>(`/api/v1/logs/query?limit=${limit}&offset=${offset}`);
    return result?.data ?? [];
  }

  async getAuthLogs(limit = 100, offset = 0) {
    const result = await this.get<PaginatedResponse<LogEntry>>(`/api/v1/logs/auth?limit=${limit}&offset=${offset}`);
    return result?.data ?? [];
  }

  async getConnectionLogs(limit = 100, offset = 0) {
    const result = await this.get<PaginatedResponse<LogEntry>>(`/api/v1/logs/connections?limit=${limit}&offset=${offset}`);
    return result?.data ?? [];
  }

  // ── Audit ────────────────────────────────────────────
  async listAuditLogs(limit = 50, offset = 0) {
    const result = await this.get<PaginatedResponse<AuditLog>>(`/api/v1/audit-logs?limit=${limit}&offset=${offset}`);
    return result?.data ?? [];
  }

  // ── Profile ──────────────────────────────────────────
  async getSessions() {
    const result = await this.get<{ id: string; ip: string; user_agent: string; created_at: string; current: boolean }[]>('/api/v1/auth/sessions');
    return Array.isArray(result) ? result : [];
  }

  async getLoginHistory() {
    const result = await this.get<{ id: string; ip: string; user_agent: string; success: boolean; created_at: string }[]>('/api/v1/auth/login-history');
    return Array.isArray(result) ? result : [];
  }
}

export const api = new ApiClient();
