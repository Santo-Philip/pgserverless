package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/shared/database"
	"github.com/nexbic/platform/shared/models"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = models.NewID()
	user.CreatedAt = models.Now()
	user.UpdatedAt = models.Now()

	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, name, avatar_url, status, organization_id, role_id, metadata, last_login_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, user.ID, user.Email, user.PasswordHash, user.Name, user.AvatarURL, user.Status, user.OrganizationID, user.RoleID, user.Metadata, user.LastLoginAt, user.CreatedAt, user.UpdatedAt)

	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, name, avatar_url, status, organization_id, role_id, metadata, last_login_at, created_at, updated_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL, &user.Status, &user.OrganizationID, &user.RoleID, &user.Metadata, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, name, avatar_url, status, organization_id, role_id, metadata, last_login_at, created_at, updated_at, deleted_at
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL, &user.Status, &user.OrganizationID, &user.RoleID, &user.Metadata, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, email, password_hash, name, avatar_url, status, organization_id, role_id, metadata, last_login_at, created_at, updated_at, deleted_at
		FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.AvatarURL, &u.Status, &u.OrganizationID, &u.RoleID, &u.Metadata, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.UserStatus) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`, status, id)
	return err
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE users SET last_login_at = NOW() WHERE id = $1`, id)
	return err
}
