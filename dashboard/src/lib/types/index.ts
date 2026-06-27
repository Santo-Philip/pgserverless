export interface Project {
	id: string;
	name: string;
	slug: string;
	description?: string;
	plan_id?: string;
	status: string;
	created_at: string;
	updated_at: string;
}

export interface Database {
	id: string;
	project_id: string;
	name: string;
	schema_name: string;
	db_user: string;
	connection_string?: string;
	status: string;
	size_bytes: number;
	created_at: string;
	updated_at: string;
}

export interface APIKey {
	id: string;
	name: string;
	key_type: 'system' | 'service' | 'project';
	key_prefix: string;
	raw_key?: string;
	scopes: string[];
	project_id?: string;
	rate_limit: number;
	allowed_ips?: string[];
	origins?: string[];
	expires_at?: string;
	is_active: boolean;
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
	role: string;
	is_active: boolean;
	last_login_at?: string;
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
	actor_id: string;
	action: string;
	resource: string;
	resource_id: string;
	metadata?: Record<string, unknown>;
	ip_address: string;
	user_agent: string;
	created_at: string;
}

export interface Plan {
	id: string;
	name: string;
	slug: string;
	description?: string;
	max_databases: number;
	max_storage_mb: number;
	max_connections: number;
	max_requests: number;
	max_api_keys: number;
	price: number;
	is_active: boolean;
}

export interface QuotaUsage {
	databases_used: number;
	storage_bytes: number;
	requests_used: number;
	api_keys_used: number;
	period_start: string;
	period_end: string;
}

export interface QuotaLimit {
	max_databases: number;
	max_storage_mb: number;
	max_connections: number;
	max_requests: number;
	max_api_keys: number;
}

export interface TableInfo {
	name: string;
	columns: TableColumn[];
}

export interface TableColumn {
	name: string;
	type: string;
	nullable: boolean;
	is_pk: boolean;
	default_value?: string;
}

export interface Extension {
	name: string;
	version: string;
	description: string;
	installed: boolean;
}
