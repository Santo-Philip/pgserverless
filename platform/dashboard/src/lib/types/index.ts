export interface App {
	id: string;
	org_id: string;
	owner_id?: string;
	name: string;
	slug: string;
	schema_name: string;
	description?: string;
	status: 'active' | 'inactive' | 'suspended' | 'deleted';
	region: string;
	visibility: 'public' | 'private';
	settings: Record<string, unknown>;
	created_at: string;
	updated_at: string;
}

export interface APIKey {
	id: string;
	app_id: string;
	name: string;
	key_type: 'publishable' | 'secret' | 'service' | 'admin';
	key_prefix: string;
	raw_key?: string;
	scopes: string[];
	rate_limit: number;
	is_active: boolean;
	expires_at?: string;
	created_at: string;
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
	status: string;
	created_at: string;
}

export interface PaginatedResponse<T> {
	data: T[];
	total: number;
	limit: number;
	offset: number;
}

export interface AuditLog {
	id: string;
	app_id?: string;
	user_id?: string;
	method: string;
	path: string;
	status_code: number;
	response_time_ms: number;
	created_at: string;
}

export interface Domain {
	id: string;
	app_id: string;
	domain: string;
	status: 'pending' | 'active' | 'failed';
	verified: boolean;
	verification_token?: string;
	verified_at?: string;
	created_at: string;
	updated_at: string;
}

export interface Backup {
	id: string;
	app_id: string;
	status: 'running' | 'completed' | 'failed';
	file_size?: number;
	created_at: string;
}

export interface PlatformSettings {
	region: string;
	default_visibility: string;
}
