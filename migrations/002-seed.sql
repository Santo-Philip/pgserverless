-- Seed default plans
INSERT INTO plans (name, slug, description, max_databases, max_storage_mb, max_connections, max_requests, max_api_keys, price, is_active)
VALUES
    ('Free', 'free', 'Free tier for development and testing', 1, 100, 20, 10000, 5, 0.00, TRUE),
    ('Starter', 'starter', 'Starter tier for small projects', 5, 1024, 100, 100000, 20, 29.00, TRUE),
    ('Pro', 'pro', 'Professional tier for production workloads', 25, 10240, 500, 1000000, 100, 99.00, TRUE),
    ('Enterprise', 'enterprise', 'Enterprise tier with custom limits', 100, 102400, 2000, 10000000, 500, 499.00, TRUE)
ON CONFLICT (slug) DO NOTHING;

-- Platform settings singleton
CREATE TABLE IF NOT EXISTS platform_settings (
    id INTEGER PRIMARY KEY DEFAULT 1,
    settings JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT single_row CHECK (id = 1)
);

INSERT INTO platform_settings (id, settings)
VALUES (1, '{
    "app_name": "Nexbic Database Platform",
    "log_level": "info",
    "default_plan": "free",
    "registration_enabled": true,
    "password_min_length": 8,
    "maintenance_mode": false
}')
ON CONFLICT (id) DO NOTHING;
