package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/identity/auth/models"
	"github.com/nexbic/platform/pkg/database"
	"github.com/nexbic/platform/pkg/helpers"
)

type SecurityRepo struct {
	db *database.DB
}

func NewSecurityRepo(db *database.DB) *SecurityRepo {
	return &SecurityRepo{db: db}
}

// ── Email Verification Tokens ──────────────────────────

func (r *SecurityRepo) CreateEmailVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`, userID, tokenHash, expiresAt)
	return err
}

func (r *SecurityRepo) FindEmailVerificationToken(ctx context.Context, tokenHash string) (*uuid.UUID, error) {
	var userID uuid.UUID
	err := r.db.Pool.QueryRow(ctx, `
		SELECT user_id FROM email_verification_tokens
		WHERE token_hash = $1 AND expires_at > NOW()`, tokenHash).Scan(&userID)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return &userID, nil
}

func (r *SecurityRepo) DeleteEmailVerificationTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM email_verification_tokens WHERE user_id = $1`, userID)
	return err
}

// ── Password Reset Tokens ──────────────────────────────

func (r *SecurityRepo) CreatePasswordResetToken(ctx context.Context, email, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (email, token_hash, expires_at)
		VALUES ($1, $2, $3)`, email, tokenHash, expiresAt)
	return err
}

func (r *SecurityRepo) FindPasswordResetToken(ctx context.Context, tokenHash string) (*string, error) {
	var email string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT email FROM password_reset_tokens
		WHERE token_hash = $1 AND expires_at > NOW()`, tokenHash).Scan(&email)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return &email, nil
}

func (r *SecurityRepo) DeletePasswordResetTokens(ctx context.Context, email string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM password_reset_tokens WHERE email = $1`, email)
	return err
}

// ── Devices ─────────────────────────────────────────────

func (r *SecurityRepo) UpsertDevice(ctx context.Context, userID uuid.UUID, deviceName, deviceType, ipAddress, clientDeviceID string) (*models.Device, error) {
	// Try find existing device by client_device_id
	if clientDeviceID != "" {
		existing := &models.Device{}
		err := r.db.Pool.QueryRow(ctx, `
			SELECT id, user_id, device_name, device_type, ip_address, client_device_id, last_used_at, created_at
			FROM devices WHERE user_id = $1 AND client_device_id = $2`, userID, clientDeviceID).Scan(
			&existing.ID, &existing.UserID, &existing.DeviceName, &existing.DeviceType,
			&existing.IPAddress, &existing.ClientDeviceID, &existing.LastUsedAt, &existing.CreatedAt)
		if err == nil {
			existing.DeviceName = deviceName
			existing.DeviceType = deviceType
			existing.IPAddress = ipAddress
			existing.LastUsedAt = time.Now()
			_, err := r.db.Pool.Exec(ctx, `
				UPDATE devices SET device_name = $1, device_type = $2, ip_address = $3, last_used_at = $4
				WHERE id = $5`, deviceName, deviceType, ipAddress, existing.LastUsedAt, existing.ID)
			if err != nil {
				return nil, err
			}
			return existing, nil
		}
	}

	device := &models.Device{
		ID:             uuid.New(),
		UserID:         userID,
		DeviceName:     deviceName,
		DeviceType:     deviceType,
		IPAddress:      ipAddress,
		ClientDeviceID: clientDeviceID,
		LastUsedAt:     time.Now(),
		CreatedAt:      time.Now(),
	}
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO devices (id, user_id, device_name, device_type, ip_address, client_device_id, last_used_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		device.ID, device.UserID, device.DeviceName, device.DeviceType,
		device.IPAddress, device.ClientDeviceID, device.LastUsedAt, device.CreatedAt)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (r *SecurityRepo) ListDevices(ctx context.Context, userID uuid.UUID) ([]models.Device, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, user_id, device_name, device_type, ip_address, COALESCE(client_device_id, ''), last_used_at, created_at
		FROM devices WHERE user_id = $1 ORDER BY last_used_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.UserID, &d.DeviceName, &d.DeviceType,
			&d.IPAddress, &d.ClientDeviceID, &d.LastUsedAt, &d.CreatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	if devices == nil {
		devices = []models.Device{}
	}
	return devices, nil
}

func (r *SecurityRepo) DeleteDevice(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM devices WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

// ── Security Events ─────────────────────────────────────

func (r *SecurityRepo) LogEvent(ctx context.Context, userID uuid.UUID, message, severity, ipAddress, userAgent string) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO security_events (user_id, message, severity, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)`, userID, message, severity, ipAddress, userAgent)
	return err
}

func (r *SecurityRepo) ListEvents(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.SecurityEvent, int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM security_events WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, user_id, message, severity, ip_address, user_agent, timestamp
		FROM security_events WHERE user_id = $1 ORDER BY timestamp DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []models.SecurityEvent
	for rows.Next() {
		var e models.SecurityEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.Message, &e.Severity,
			&e.IPAddress, &e.UserAgent, &e.Timestamp); err != nil {
			return nil, 0, err
		}
		events = append(events, e)
	}
	if events == nil {
		events = []models.SecurityEvent{}
	}
	return events, total, nil
}

// ── API Keys ────────────────────────────────────────────

func (r *SecurityRepo) CreateAPIKey(ctx context.Context, key *models.APIKey) error {
	key.ID = uuid.New()
	key.CreatedAt = time.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO api_keys (id, user_id, name, prefix, hash, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		key.ID, key.UserID, key.Name, key.Prefix, key.Hash, key.Status, key.ExpiresAt, key.CreatedAt)
	return err
}

func (r *SecurityRepo) ListAPIKeys(ctx context.Context, userID uuid.UUID) ([]models.APIKey, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, user_id, name, prefix, hash, status, revoked_at, last_used_at, expires_at, created_at
		FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.Prefix, &k.Hash,
			&k.Status, &k.RevokedAt, &k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	if keys == nil {
		keys = []models.APIKey{}
	}
	return keys, nil
}

func (r *SecurityRepo) GetAPIKeyByHash(ctx context.Context, hash string) (*models.APIKey, error) {
	k := &models.APIKey{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, name, prefix, hash, status, revoked_at, last_used_at, expires_at, created_at
		FROM api_keys WHERE hash = $1 AND status = 'active' AND (expires_at IS NULL OR expires_at > NOW())`, hash).Scan(
		&k.ID, &k.UserID, &k.Name, &k.Prefix, &k.Hash,
		&k.Status, &k.RevokedAt, &k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return k, nil
}

func (r *SecurityRepo) RevokeAPIKey(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE api_keys SET status = 'revoked', revoked_at = NOW() WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}
