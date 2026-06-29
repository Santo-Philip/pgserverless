package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/internal/identity/auth/models"
	"github.com/nexbic/platform/pkg/database"
	"github.com/nexbic/platform/pkg/helpers"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, name, role, is_active, image, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		user.ID, user.Email, user.PasswordHash, user.Name, user.Role,
		user.IsActive, user.Image, user.EmailVerified, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	var image, totpSecret *string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, name, COALESCE(image, ''), role, is_active,
		       email_verified, totp_secret, totp_enabled, COALESCE(recovery_codes::text, '[]'),
		       last_login_at, created_at, updated_at
		FROM users WHERE id = $1`, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &image,
		&user.Role, &user.IsActive, &user.EmailVerified, &totpSecret, &user.TOTPEnabled,
		&user.RecoveryCodes, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	if image != nil {
		user.Image = *image
	}
	if totpSecret != nil {
		user.TOTPSecret = *totpSecret
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	var image, totpSecret *string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, name, COALESCE(image, ''), role, is_active,
		       email_verified, totp_secret, totp_enabled, COALESCE(recovery_codes::text, '[]'),
		       last_login_at, created_at, updated_at
		FROM users WHERE email = $1`, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &image,
		&user.Role, &user.IsActive, &user.EmailVerified, &totpSecret, &user.TOTPEnabled,
		&user.RecoveryCodes, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	if image != nil {
		user.Image = *image
	}
	if totpSecret != nil {
		user.TOTPSecret = *totpSecret
	}
	return user, nil
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE is_active = true`).Scan(&count)
	return count, err
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, email, password_hash, name, COALESCE(image, ''), role, is_active,
		       email_verified, totp_secret, totp_enabled, COALESCE(recovery_codes::text, '[]'),
		       last_login_at, created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var image, totpSecret *string
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &image,
			&u.Role, &u.IsActive, &u.EmailVerified, &totpSecret, &u.TOTPEnabled,
			&u.RecoveryCodes, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if image != nil {
			u.Image = *image
		}
		if totpSecret != nil {
			u.TOTPSecret = *totpSecret
		}
		users = append(users, u)
	}

	return users, total, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE users SET last_login_at = NOW(), updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE users SET name = $1, role = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4`,
		user.Name, user.Role, user.IsActive, user.ID)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, hash, id)
	return err
}

func (r *UserRepository) UpdateEmailVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE users SET email_verified = TRUE, updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *UserRepository) UpdateTOTP(ctx context.Context, id uuid.UUID, secret string) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE users SET totp_secret = $1, totp_enabled = TRUE, updated_at = NOW() WHERE id = $2`, secret, id)
	return err
}

func (r *UserRepository) DisableTOTP(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE users SET totp_secret = NULL, totp_enabled = FALSE, recovery_codes = '[]'::jsonb, updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *UserRepository) SetRecoveryCodes(ctx context.Context, id uuid.UUID, codes []string) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE users SET recovery_codes = $1, updated_at = NOW() WHERE id = $2`, codes, id)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

// ── Refresh Tokens ──────────────────────────────────────

type RefreshTokenRepo struct {
	db *database.DB
}

func NewRefreshTokenRepo(db *database.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`, userID, token, expiresAt)
	return err
}

func (r *RefreshTokenRepo) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

func (r *RefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	return err
}

func (r *RefreshTokenRepo) FindByToken(ctx context.Context, tokenHash string) (*uuid.UUID, error) {
	var userID uuid.UUID
	err := r.db.Pool.QueryRow(ctx, `
		SELECT user_id FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > NOW()`, tokenHash).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &userID, nil
}
