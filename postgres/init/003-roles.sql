INSERT INTO roles (name, description, is_system) VALUES
    ('anon', 'Anonymous unauthenticated user', TRUE),
    ('authenticated', 'Authenticated user', TRUE),
    ('admin', 'Administrator with full access', TRUE)
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (resource, action, description) VALUES
    ('users', 'read', 'Read user profiles'),
    ('users', 'write', 'Create and update users'),
    ('users', 'delete', 'Delete users'),
    ('organizations', 'read', 'Read organization data'),
    ('organizations', 'write', 'Create and update organizations'),
    ('organizations', 'delete', 'Delete organizations'),
    ('roles', 'read', 'Read role definitions'),
    ('roles', 'write', 'Create and update roles'),
    ('roles', 'delete', 'Delete roles'),
    ('permissions', 'read', 'Read permissions'),
    ('permissions', 'write', 'Create and update permissions'),
    ('api_keys', 'read', 'Read API keys'),
    ('api_keys', 'write', 'Create and update API keys'),
    ('api_keys', 'delete', 'Delete API keys'),
    ('logs', 'read', 'Read audit logs'),
    ('sessions', 'read', 'Read sessions'),
    ('sessions', 'delete', 'Delete sessions')
ON CONFLICT (resource, action) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin'
  AND (p.resource, p.action) IN (
      ('users', 'read'),
      ('users', 'write'),
      ('users', 'delete'),
      ('organizations', 'read'),
      ('organizations', 'write'),
      ('organizations', 'delete'),
      ('roles', 'read'),
      ('roles', 'write'),
      ('roles', 'delete'),
      ('permissions', 'read'),
      ('permissions', 'write'),
      ('api_keys', 'read'),
      ('api_keys', 'write'),
      ('api_keys', 'delete'),
      ('logs', 'read'),
      ('sessions', 'read'),
      ('sessions', 'delete')
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'authenticated'
  AND (p.resource, p.action) IN (
      ('users', 'read'),
      ('organizations', 'read')
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;

CREATE OR REPLACE FUNCTION api.is_admin()
RETURNS BOOLEAN
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN current_setting('request.jwt.claim.role', TRUE) = 'admin';
END;
$$;

CREATE OR REPLACE FUNCTION api.current_user_id()
RETURNS UUID
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_sub TEXT;
BEGIN
    v_sub := current_setting('request.jwt.claim.sub', TRUE);
    IF v_sub IS NULL THEN
        RETURN NULL;
    END IF;
    RETURN v_sub::UUID;
END;
$$;

CREATE OR REPLACE FUNCTION api.current_organization_id()
RETURNS UUID
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_org TEXT;
BEGIN
    v_org := current_setting('request.jwt.claim.organization_id', TRUE);
    IF v_org IS NULL THEN
        RETURN NULL;
    END IF;
    RETURN v_org::UUID;
END;
$$;
