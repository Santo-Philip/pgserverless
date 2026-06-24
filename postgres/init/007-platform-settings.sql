-- Platform settings (singleton row)
CREATE TABLE IF NOT EXISTS platform_settings (
    id INTEGER PRIMARY KEY DEFAULT 1,
    settings JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT single_row CHECK (id = 1)
);

-- Insert default settings row
INSERT INTO platform_settings (id, settings)
VALUES (1, '{
    "app_name": "Nexbic Platform",
    "log_level": "info",
    "region": "us-east",
    "default_visibility": "public",
    "jwt_access_ttl": "15m",
    "jwt_refresh_ttl": "168h",
    "otp_expiry": "5m",
    "default_user_role": "authenticated",
    "registration_enabled": true,
    "email_verification_required": false,
    "cors_origins": "*",
    "password_min_length": 8,
    "password_require_special": false,
    "password_require_numbers": false,
    "max_db_connections": 20,
    "min_db_connections": 2,
    "health_check_period": "30s",
    "api_rate_limits": "1000/h",
    "storage_limit_mb": 100,
    "monitoring_enabled": true,
    "maintenance_mode": false,
    "feature_flags": "{}"
}')
ON CONFLICT (id) DO NOTHING;
