-- Enhanced Auth: add columns to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS image TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS recovery_codes JSONB DEFAULT '[]' NOT NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS webauthn_challenge TEXT;

-- Devices
CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_name VARCHAR(255) NOT NULL DEFAULT '',
    device_type VARCHAR(50) NOT NULL DEFAULT 'web',
    ip_address VARCHAR(45) NOT NULL DEFAULT '',
    client_device_id VARCHAR(255),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_devices_user ON devices(user_id);

-- Security Events
CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'low',
    ip_address VARCHAR(45) NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_security_events_user_ts ON security_events(user_id, timestamp DESC);

-- Email Verification Tokens
CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_email_verification_token ON email_verification_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_email_verification_user ON email_verification_tokens(user_id);

-- Password Reset Tokens
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_password_reset_token ON password_reset_tokens(token_hash);

-- WebAuthn Credentials
CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id TEXT NOT NULL,
    public_key TEXT NOT NULL,
    counter INTEGER NOT NULL DEFAULT 0,
    device_name VARCHAR(255) NOT NULL DEFAULT '',
    transports JSONB DEFAULT '[]' NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_webauthn_user ON webauthn_credentials(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_webauthn_credential ON webauthn_credentials(credential_id);

-- API Keys
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    prefix VARCHAR(20) NOT NULL,
    hash TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(hash);

-- Enhance sessions table with new columns
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS device_type VARCHAR(50) NOT NULL DEFAULT 'web';
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS country VARCHAR(100);
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS city VARCHAR(100);
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS client_device_id VARCHAR(255);
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS revoked BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE sessions ALTER COLUMN device_info DROP NOT NULL;
ALTER TABLE sessions ALTER COLUMN device_info SET DEFAULT '';

-- =====================
-- STORAGE FEATURE
-- =====================

-- Storage Providers (S3, GCS, local, etc.)
CREATE TABLE IF NOT EXISTS storage_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    provider_type VARCHAR(50) NOT NULL DEFAULT 'local',
    config JSONB NOT NULL DEFAULT '{}',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_storage_providers_type ON storage_providers(provider_type);

-- Storage Buckets
CREATE TABLE IF NOT EXISTS storage_buckets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    provider_id UUID NOT NULL REFERENCES storage_providers(id) ON DELETE CASCADE,
    path TEXT NOT NULL DEFAULT '',
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_storage_buckets_provider ON storage_buckets(provider_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_storage_buckets_name_provider ON storage_buckets(name, provider_id);

-- Storage Files
CREATE TABLE IF NOT EXISTS storage_files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bucket_id UUID NOT NULL REFERENCES storage_buckets(id) ON DELETE CASCADE,
    name VARCHAR(512) NOT NULL,
    path TEXT NOT NULL,
    mime_type VARCHAR(255) NOT NULL DEFAULT 'application/octet-stream',
    size_bytes BIGINT NOT NULL DEFAULT 0,
    md5_hash VARCHAR(64),
    metadata JSONB DEFAULT '{}',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_storage_files_bucket ON storage_files(bucket_id);
CREATE INDEX IF NOT EXISTS idx_storage_files_path ON storage_files(path);

-- Backup Storage Link (link backup_history to storage_files)
ALTER TABLE backup_history ADD COLUMN IF NOT EXISTS storage_file_id UUID REFERENCES storage_files(id);
ALTER TABLE backup_history ADD COLUMN IF NOT EXISTS storage_provider_id UUID REFERENCES storage_providers(id);

-- Default local storage provider (idempotent)
INSERT INTO storage_providers (name, provider_type, config, is_default, is_enabled)
SELECT 'Local Storage', 'local', '{"base_path": "/data/storage"}'::jsonb, TRUE, TRUE
WHERE NOT EXISTS (SELECT 1 FROM storage_providers WHERE provider_type = 'local' AND is_default = TRUE);
