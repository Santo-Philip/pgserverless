import { browser } from '$app/environment';
import type {
	AuthResponse, Project, Database, APIKey,
	Plan, AuditLog, QuotaUsage, QuotaLimit,
	TableInfo, Extension
} from '$lib/types';

function getBaseUrl(): string {
	if (browser) {
		return window.location.origin;
	}
	return '';
}

interface SuccessEnvelope {
	message: string;
	data: unknown;
}

class ApiError extends Error {
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
			this.token = localStorage.getItem('nexbic_token');
			this.refreshToken = localStorage.getItem('nexbic_refresh_token');
			const exp = localStorage.getItem('nexbic_expires_at');
			this.expiresAt = exp ? parseInt(exp, 10) : null;
		}
	}

	setToken(token: string) {
		this.token = token;
		if (browser) localStorage.setItem('nexbic_token', token);
	}

	setRefreshToken(token: string) {
		this.refreshToken = token;
		if (browser) localStorage.setItem('nexbic_refresh_token', token);
	}

	setExpiresAt(iso: string) {
		this.expiresAt = new Date(iso).getTime();
		if (browser) localStorage.setItem('nexbic_expires_at', String(this.expiresAt));
	}

	clearToken() {
		this.token = null;
		this.refreshToken = null;
		this.expiresAt = null;
		if (browser) {
			localStorage.removeItem('nexbic_token');
			localStorage.removeItem('nexbic_refresh_token');
			localStorage.removeItem('nexbic_expires_at');
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
	async del<T>(path: string): Promise<T> { return this.request<T>('DELETE', path); }

	// Auth
	async login(email: string, password: string) {
		const result = await this.post<AuthResponse>('/api/v1/auth/login', { email, password });
		this.setToken(result.token);
		if (result.refresh_token) this.setRefreshToken(result.refresh_token);
		if (result.expires_at) this.setExpiresAt(result.expires_at);
		return result;
	}

	async register(email: string, password: string, name?: string) {
		const result = await this.post<AuthResponse>('/api/v1/auth/register', { email, password, name });
		this.setToken(result.token);
		if (result.refresh_token) this.setRefreshToken(result.refresh_token);
		if (result.expires_at) this.setExpiresAt(result.expires_at);
		return result;
	}

	async getMe() {
		return this.get<User>('/api/v1/auth/me');
	}

	// Projects
	async listProjects(): Promise<Project[]> {
		const result = await this.get<PaginatedResponse<Project>>('/api/v1/projects');
		return result?.data ?? [];
	}

	async getProject(id: string): Promise<Project> {
		return this.get<Project>(`/api/v1/projects/${id}`);
	}

	async createProject(name: string, slug: string, description?: string): Promise<Project> {
		return this.post<Project>('/api/v1/projects', { name, slug, description });
	}

	async updateProject(id: string, data: Partial<Project>): Promise<Project> {
		return this.patch<Project>(`/api/v1/projects/${id}`, data);
	}

	async deleteProject(id: string): Promise<void> {
		await this.del(`/api/v1/projects/${id}`);
	}

	// Databases
	async listDatabases(projectId: string): Promise<Database[]> {
		const result = await this.get<PaginatedResponse<Database>>(`/api/v1/projects/${projectId}/databases`);
		return result?.data ?? [];
	}

	async getDatabase(id: string): Promise<Database> {
		return this.get<Database>(`/api/v1/databases/${id}`);
	}

	async createDatabase(projectId: string, name: string): Promise<Database> {
		return this.post<Database>('/api/v1/databases', { project_id: projectId, name });
	}

	async deleteDatabase(id: string): Promise<void> {
		await this.del(`/api/v1/databases/${id}`);
	}

	// SQL
	async runSQL(databaseId: string, query: string): Promise<Record<string, unknown>[]> {
		const result = await this.post<{ data: Record<string, unknown>[] } | Record<string, unknown>[]>(`/api/v1/databases/${databaseId}/sql`, { query });
		return Array.isArray(result) ? result : [];
	}

	// Tables
	async listTables(databaseId: string): Promise<TableInfo[]> {
		const result = await this.get<TableInfo[]>(`/api/v1/databases/${databaseId}/tables`);
		return Array.isArray(result) ? result : [];
	}

	async getTableData(databaseId: string, table: string, limit = 100, offset = 0): Promise<Record<string, unknown>[]> {
		const result = await this.get<Record<string, unknown>[]>(`/api/v1/databases/${databaseId}/tables/${table}?limit=${limit}&offset=${offset}`);
		return Array.isArray(result) ? result : [];
	}

	async createTable(databaseId: string, name: string, columns: { name: string; type: string; nullable: boolean; is_pk: boolean; default_value?: string }[]): Promise<void> {
		await this.post(`/api/v1/databases/${databaseId}/tables`, { name, columns });
	}

	async addColumn(databaseId: string, table: string, column: { name: string; type: string; nullable: boolean; default_value?: string }): Promise<void> {
		await this.post(`/api/v1/databases/${databaseId}/tables/${table}/columns`, column);
	}

	async insertRow(databaseId: string, table: string, values: Record<string, unknown>): Promise<void> {
		await this.post(`/api/v1/databases/${databaseId}/tables/${table}/rows`, { values });
	}

	async updateRow(databaseId: string, table: string, values: Record<string, unknown>, where: Record<string, unknown>): Promise<void> {
		await this.patch(`/api/v1/databases/${databaseId}/tables/${table}/rows`, { values, where });
	}

	async deleteRow(databaseId: string, table: string, where: Record<string, unknown>): Promise<void> {
		await this.del(`/api/v1/databases/${databaseId}/tables/${table}/rows`, { where });
	}

	// API Keys
	async listAPIKeys(): Promise<APIKey[]> {
		const result = await this.get<PaginatedResponse<APIKey>>('/api/v1/api-keys');
		return result?.data ?? [];
	}

	async listProjectAPIKeys(projectId: string): Promise<APIKey[]> {
		const result = await this.get<PaginatedResponse<APIKey>>(`/api/v1/projects/${projectId}/api-keys`);
		return result?.data ?? [];
	}

	async createAPIKey(name: string, keyType: string, projectId?: string): Promise<APIKey> {
		return this.post<APIKey>('/api/v1/api-keys', { name, key_type: keyType, project_id: projectId });
	}

	async revokeAPIKey(id: string): Promise<void> {
		await this.del(`/api/v1/api-keys/${id}`);
	}

	// Plans
	async listPlans(): Promise<Plan[]> {
		const result = await this.get<Plan[]>('/api/v1/plans');
		return Array.isArray(result) ? result : [];
	}

	async createPlan(data: Partial<Plan>): Promise<Plan> {
		return this.post<Plan>('/api/v1/plans', data);
	}

	async updatePlan(id: string, data: Partial<Plan>): Promise<Plan> {
		return this.patch<Plan>(`/api/v1/plans/${id}`, data);
	}

	// Quota
	async getProjectQuota(projectId: string): Promise<{ usage: QuotaUsage; limits: QuotaLimit }> {
		return this.get(`/api/v1/projects/${projectId}/quota`);
	}

	// Audit Logs
	async listAuditLogs(): Promise<AuditLog[]> {
		const result = await this.get<PaginatedResponse<AuditLog>>('/api/v1/audit-logs');
		return result?.data ?? [];
	}

	// Extensions
	async listExtensions(): Promise<Extension[]> {
		const result = await this.get<Extension[]>('/api/v1/extensions');
		return Array.isArray(result) ? result : [];
	}

	async toggleExtension(name: string, install: boolean): Promise<void> {
		await this.post('/api/v1/extensions/toggle', { name, install });
	}
}

export const api = new ApiClient();
export { ApiError };

interface PaginatedResponse<T> {
	data: T[];
	total: number;
	limit: number;
	offset: number;
}
