import { browser } from '$app/environment';

const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

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

		const response = await fetch(`${BASE_URL}${path}`, {
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

		return response.json();
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

	async listApps(): Promise<PaginatedResponse<App>> {
		return this.get<PaginatedResponse<App>>('/api/v1/platform/apps');
	}

	async getApp(id: string): Promise<App> {
		return this.get<App>(`/api/v1/platform/apps/${id}`);
	}

	async createApp(name: string, slug: string, description?: string): Promise<{ app: App; admin_key: APIKey }> {
		return this.post('/api/v1/platform/apps', { name, slug, description });
	}

	async deleteApp(id: string): Promise<void> {
		return this.del(`/api/v1/platform/apps/${id}`);
	}

	async listAPIKeys(appId: string): Promise<APIKey[]> {
		return this.get<APIKey[]>(`/api/v1/platform/apps/${appId}/apikey`);
	}

	async createAPIKey(appId: string, name: string, keyType: string): Promise<APIKey> {
		return this.post(`/api/v1/platform/apps/${appId}/apikey`, { name, key_type: keyType });
	}

	async deactivateAPIKey(appId: string, keyId: string): Promise<void> {
		return this.del(`/api/v1/platform/apps/${appId}/apikey/${keyId}`);
	}

	async listDomains(appId: string): Promise<Domain[]> {
		return this.get<Domain[]>(`/api/v1/platform/apps/${appId}/domains`);
	}

	async createDomain(appId: string, domain: string): Promise<Domain> {
		return this.post<Domain>(`/api/v1/platform/apps/${appId}/domains`, { domain });
	}

	async deleteDomain(appId: string, domainId: string): Promise<void> {
		return this.del(`/api/v1/platform/apps/${appId}/domains/${domainId}`);
	}

	async verifyDomain(appId: string, domainId: string): Promise<void> {
		return this.post(`/api/v1/platform/apps/${appId}/domains/${domainId}/verify`);
	}

	async listUsers(): Promise<PaginatedResponse<User>> {
		return this.get<PaginatedResponse<User>>('/api/v1/platform/users');
	}

	async getUser(id: string): Promise<User> {
		return this.get<User>(`/api/v1/platform/users/${id}`);
	}

	async listBackups(): Promise<Backup[]> {
		return this.get<Backup[]>('/api/v1/platform/backups');
	}

	async createBackup(): Promise<void> {
		return this.post('/api/v1/platform/backups');
	}

	async getSettings(): Promise<PlatformSettings> {
		return this.get<PlatformSettings>('/api/v1/platform/settings');
	}

	async updateSettings(settings: Partial<PlatformSettings>): Promise<void> {
		return this.patch('/api/v1/platform/settings', settings);
	}
}

export const api = new ApiClient();
