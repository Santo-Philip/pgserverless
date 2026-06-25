import { browser } from '$app/environment';
import type { AuthResponse, App, APIKey, Domain, User, PlatformSettings } from '$lib/types';

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

class ApiClient {
	private token: string | null = null;

	constructor() {
		if (browser) {
			this.token = localStorage.getItem('nexbic_token');
		}
	}

	setToken(token: string) {
		this.token = token;
		if (browser) {
			localStorage.setItem('nexbic_token', token);
		}
	}

	clearToken() {
		this.token = null;
		if (browser) {
			localStorage.removeItem('nexbic_token');
		}
	}

	get isAuthenticated(): boolean {
		return !!this.token;
	}

	private unwrap<T>(json: unknown): T {
		if (json && typeof json === 'object' && 'message' in json && 'data' in json) {
			return (json as SuccessEnvelope).data as T;
		}
		return json as T;
	}

	private async request<T>(
		method: string,
		path: string,
		body?: unknown
	): Promise<T> {
		const headers: Record<string, string> = {
			'Content-Type': 'application/json',
			Accept: 'application/json'
		};

		if (this.token) {
			headers['Authorization'] = `Bearer ${this.token}`;
		}

		const response = await fetch(`${getBaseUrl()}${path}`, {
			method,
			headers,
			body: body ? JSON.stringify(body) : undefined
		});

		if (response.status === 401) {
			this.clearToken();
			throw new Error('Unauthorized');
		}

		if (!response.ok) {
			const error = await response.json().catch(() => ({ message: response.statusText }));
			throw new Error(error.message || `HTTP ${response.status}`);
		}

		const json = await response.json();
		return this.unwrap<T>(json);
	}

	async get<T>(path: string): Promise<T> {
		return this.request<T>('GET', path);
	}

	async post<T>(path: string, body?: unknown): Promise<T> {
		return this.request<T>('POST', path, body);
	}

	async patch<T>(path: string, body?: unknown): Promise<T> {
		return this.request<T>('PATCH', path, body);
	}

	async del<T>(path: string): Promise<T> {
		return this.request<T>('DELETE', path);
	}

	async login(email: string, password: string) {
		const result = await this.post<AuthResponse>('/api/v1/platform/auth/login', { email, password });
		this.setToken(result.token);
		return result;
	}

	async register(email: string, password: string, name?: string) {
		const result = await this.post<AuthResponse>('/api/v1/platform/auth/register', { email, password, name });
		this.setToken(result.token);
		return result;
	}

	async listApps(): Promise<App[]> {
		type PaginatedApps = { data: App[]; total: number; limit: number; offset: number };
		const result = await this.get<PaginatedApps>('/api/v1/platform/apps');
		return Array.isArray(result) ? result : (result?.data ?? []);
	}

	async getApp(id: string): Promise<App> {
		return this.get<App>(`/api/v1/platform/apps/${id}`);
	}

	async createApp(name: string, slug: string, description?: string): Promise<{
		app: App;
		admin_key: APIKey;
		service_key: APIKey;
		jwt_secret: string;
		connection_uri: string;
	}> {
		return this.post('/api/v1/platform/apps', { name, slug, description });
	}

	async deleteApp(id: string): Promise<void> {
		await this.del(`/api/v1/platform/apps/${id}`);
	}

	async listAPIKeys(appId: string): Promise<APIKey[]> {
		const result = await this.get<unknown>(`/api/v1/platform/apps/${appId}/apikey`);
		return Array.isArray(result) ? result : [];
	}

	async createAPIKey(appId: string, name: string, keyType: string): Promise<APIKey> {
		return this.post(`/api/v1/platform/apps/${appId}/apikey`, { name, key_type: keyType });
	}

	async deactivateAPIKey(appId: string, keyId: string): Promise<void> {
		await this.del(`/api/v1/platform/apps/${appId}/apikey/${keyId}`);
	}

	async listDomains(appId: string): Promise<Domain[]> {
		const result = await this.get<unknown>(`/api/v1/platform/apps/${appId}/domains`);
		return Array.isArray(result) ? result : [];
	}

	async createDomain(appId: string, domain: string): Promise<Domain> {
		return this.post<Domain>(`/api/v1/platform/apps/${appId}/domains`, { domain });
	}

	async deleteDomain(appId: string, domainId: string): Promise<void> {
		await this.del(`/api/v1/platform/apps/${appId}/domains/${domainId}`);
	}

	async verifyDomain(appId: string, domainId: string): Promise<void> {
		await this.post(`/api/v1/platform/apps/${appId}/domains/${domainId}/verify`);
	}

	async listUsers(): Promise<User[]> {
		type PaginatedUsers = { data: User[]; total: number; limit: number; offset: number };
		const result = await this.get<PaginatedUsers>('/api/v1/platform/users');
		return Array.isArray(result) ? result : (result?.data ?? []);
	}

	async getUser(id: string): Promise<User> {
		return this.get<User>(`/api/v1/platform/users/${id}`);
	}

	async suspendUser(id: string): Promise<void> {
		await this.post(`/api/v1/platform/users/${id}/suspend`);
	}

	async activateUser(id: string): Promise<void> {
		await this.post(`/api/v1/platform/users/${id}/activate`);
	}

	async getSettings(): Promise<PlatformSettings> {
		return this.get<PlatformSettings>('/api/v1/platform/settings');
	}

	async updateSettings(settings: Partial<PlatformSettings>): Promise<void> {
		await this.patch('/api/v1/platform/settings', settings);
	}

	async listExtensions(appId: string): Promise<{name: string; version: string; description: string; installed: boolean}[]> {
		const result = await this.get<unknown>(`/api/v1/platform/apps/${appId}/extensions`);
		return Array.isArray(result) ? result : [];
	}

	async toggleExtension(appId: string, name: string, install: boolean): Promise<void> {
		await this.post(`/api/v1/platform/apps/${appId}/extensions/toggle`, { name, install });
	}

	async listTables(appId: string): Promise<{name: string; columns: {name: string; type: string; nullable: boolean; is_pk: boolean; default_value: string}[]}[]> {
		const result = await this.get<unknown>(`/api/v1/platform/apps/${appId}/tables`);
		if (Array.isArray(result)) return result;
		if (result && typeof result === 'object' && 'tables' in result) {
			const tables = (result as {tables: unknown[]}).tables;
			return Array.isArray(tables) ? tables.map((t: unknown) => typeof t === 'string' ? { name: t, columns: [] } : t as {name: string; columns: {name: string; type: string; nullable: boolean; is_pk: boolean; default_value: string}[]}) : [];
		}
		return [];
	}

	async getTableData(appId: string, tableName: string): Promise<Record<string, unknown>[]> {
		const result = await this.get<unknown>(`/api/v1/platform/apps/${appId}/tables/${tableName}`);
		return Array.isArray(result) ? result : [];
	}

	async createTable(appId: string, name: string, columns: {name: string; type: string; nullable: boolean; is_pk: boolean; default_value?: string}[]): Promise<void> {
		await this.post(`/api/v1/platform/apps/${appId}/tables`, { name, columns });
	}

	async addColumn(appId: string, tableName: string, column: {name: string; type: string; nullable: boolean; default_value?: string}): Promise<void> {
		await this.post(`/api/v1/platform/apps/${appId}/tables/${tableName}/columns`, column);
	}

	async insertRow(appId: string, tableName: string, values: Record<string, unknown>): Promise<void> {
		await this.post(`/api/v1/platform/apps/${appId}/tables/${tableName}/rows`, { values });
	}

	async updateRow(appId: string, tableName: string, values: Record<string, unknown>, where: Record<string, unknown>): Promise<void> {
		await this.patch(`/api/v1/platform/apps/${appId}/tables/${tableName}/rows`, { values, where });
	}

	async deleteRow(appId: string, tableName: string, where: Record<string, unknown>): Promise<void> {
		await this.del(`/api/v1/platform/apps/${appId}/tables/${tableName}/rows`);
	}
}

export const api = new ApiClient();
