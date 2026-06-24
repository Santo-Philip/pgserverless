CREATE OR REPLACE FUNCTION api.login(
    p_email TEXT,
    p_password TEXT
)
RETURNS JSON
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_user RECORD;
    v_token TEXT;
    v_payload JSON;
BEGIN
    SELECT * INTO v_user
    FROM users
    WHERE email = p_email
      AND status = 'active'
    LIMIT 1;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'invalid_email_or_password' USING HINT = 'Check your credentials';
    END IF;

    IF v_user.password_hash != crypt(p_password, v_user.password_hash) THEN
        RAISE EXCEPTION 'invalid_email_or_password' USING HINT = 'Check your credentials';
    END IF;

    UPDATE users SET last_login_at = NOW() WHERE id = v_user.id;

    v_token := encode(gen_random_bytes(64), 'hex');

    INSERT INTO sessions (user_id, token, expires_at)
    VALUES (v_user.id, v_token, NOW() + INTERVAL '24 hours');

    v_payload := json_build_object(
        'token', v_token,
        'user_id', v_user.id,
        'email', v_user.email,
        'name', v_user.name,
        'role_id', v_user.role_id,
        'organization_id', v_user.organization_id,
        'expires_at', to_char(NOW() + INTERVAL '24 hours', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
    );

    RETURN v_payload;
END;
$$;

CREATE OR REPLACE FUNCTION api.register(
    p_email TEXT,
    p_password TEXT,
    p_name TEXT DEFAULT NULL
)
RETURNS JSON
LANGUAGE plpgsql
VOLATILE
AS $$
DECLARE
    v_user_id UUID;
    v_default_role UUID;
BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE email = p_email) THEN
        RAISE EXCEPTION 'email_already_exists' USING HINT = 'This email is already registered';
    END IF;

    SELECT id INTO v_default_role FROM roles WHERE name = 'authenticated' AND is_system = TRUE;

    INSERT INTO users (email, password_hash, name, role_id)
    VALUES (p_email, crypt(p_password, gen_salt('bf')), p_name, v_default_role)
    RETURNING id INTO v_user_id;

    RETURN json_build_object(
        'user_id', v_user_id,
        'email', p_email,
        'message', 'Registration successful'
    );
END;
$$;

CREATE OR REPLACE FUNCTION api.change_password(
    p_user_id UUID,
    p_old_password TEXT,
    p_new_password TEXT
)
RETURNS JSON
LANGUAGE plpgsql
VOLATILE
AS $$
DECLARE
    v_user RECORD;
BEGIN
    SELECT * INTO v_user FROM users WHERE id = p_user_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'user_not_found';
    END IF;

    IF v_user.password_hash != crypt(p_old_password, v_user.password_hash) THEN
        RAISE EXCEPTION 'invalid_password';
    END IF;

    UPDATE users
    SET password_hash = crypt(p_new_password, gen_salt('bf'))
    WHERE id = p_user_id;

    DELETE FROM sessions WHERE user_id = p_user_id;

    RETURN json_build_object('message', 'Password changed successfully');
END;
$$;

CREATE OR REPLACE FUNCTION api.validate_api_key(
    p_api_key TEXT
)
RETURNS JSON
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_key RECORD;
BEGIN
    SELECT ak.*, u.email AS user_email
    INTO v_key
    FROM api_keys ak
    LEFT JOIN users u ON u.id = ak.user_id
    WHERE ak.key_hash = crypt(p_api_key, ak.key_hash)
      AND ak.is_active = TRUE
      AND (ak.expires_at IS NULL OR ak.expires_at > NOW());

    IF NOT FOUND THEN
        RAISE EXCEPTION 'invalid_or_expired_api_key';
    END IF;

    UPDATE api_keys SET last_used_at = NOW() WHERE id = v_key.id;

    RETURN json_build_object(
        'valid', TRUE,
        'key_id', v_key.id,
        'user_id', v_key.user_id,
        'user_email', v_key.user_email,
        'organization_id', v_key.organization_id,
        'permissions', v_key.permissions,
        'name', v_key.name
    );
END;
$$;

CREATE OR REPLACE FUNCTION api.generate_api_key(
    p_user_id UUID,
    p_name TEXT,
    p_permissions JSONB DEFAULT '[]'
)
RETURNS JSON
LANGUAGE plpgsql
VOLATILE
AS $$
DECLARE
    v_raw_key TEXT;
    v_key_hash TEXT;
    v_key_prefix TEXT;
    v_key_id UUID;
BEGIN
    v_raw_key := encode(gen_random_bytes(32), 'hex');
    v_key_hash := crypt(v_raw_key, gen_salt('bf'));
    v_key_prefix := LEFT(v_raw_key, 8);

    INSERT INTO api_keys (user_id, name, key_hash, key_prefix, permissions)
    VALUES (p_user_id, p_name, v_key_hash, v_key_prefix, p_permissions)
    RETURNING id INTO v_key_id;

    RETURN json_build_object(
        'key_id', v_key_id,
        'raw_key', v_raw_key,
        'key_prefix', v_key_prefix,
        'name', p_name,
        'warning', 'Store this key securely. It will not be shown again.'
    );
END;
$$;

CREATE OR REPLACE FUNCTION api.get_current_user(
    p_user_id UUID
)
RETURNS JSON
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_user JSON;
BEGIN
    SELECT json_build_object(
        'id', u.id,
        'email', u.email,
        'name', u.name,
        'avatar_url', u.avatar_url,
        'status', u.status,
        'organization_id', u.organization_id,
        'role', json_build_object(
            'id', r.id,
            'name', r.name
        ),
        'permissions', (
            SELECT json_agg(DISTINCT json_build_object(
                'resource', p.resource,
                'action', p.action
            ))
            FROM role_permissions rp
            JOIN permissions p ON p.id = rp.permission_id
            WHERE rp.role_id = u.role_id
        ),
        'last_login_at', u.last_login_at,
        'created_at', u.created_at
    ) INTO v_user
    FROM users u
    LEFT JOIN roles r ON r.id = u.role_id
    WHERE u.id = p_user_id;

    RETURN v_user;
END;
$$;

CREATE OR REPLACE FUNCTION api.healthcheck()
RETURNS JSON
LANGUAGE plpgsql
STABLE
AS $$
BEGIN
    RETURN json_build_object(
        'status', 'healthy',
        'timestamp', NOW(),
        'version', current_setting('server_version')
    );
END;
$$;

CREATE OR REPLACE FUNCTION api.get_stats()
RETURNS JSON
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_result JSON;
BEGIN
    SELECT json_build_object(
        'total_users', (SELECT COUNT(*) FROM users WHERE status != 'deleted'),
        'total_organizations', (SELECT COUNT(*) FROM organizations WHERE status != 'inactive'),
        'total_api_keys', (SELECT COUNT(*) FROM api_keys WHERE is_active = TRUE),
        'active_sessions', (SELECT COUNT(*) FROM sessions WHERE expires_at > NOW()),
        'total_logs', (SELECT COUNT(*) FROM logs)
    ) INTO v_result;

    RETURN v_result;
END;
$$;
