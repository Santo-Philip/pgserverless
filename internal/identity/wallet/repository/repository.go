package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/internal/identity/wallet/models"
	"github.com/nexbic/platform/pkg/database"
	"github.com/nexbic/platform/pkg/helpers"
)

type WalletRepo struct {
	db *database.DB
}

func NewWalletRepo(db *database.DB) *WalletRepo {
	return &WalletRepo{db: db}
}

func (r *WalletRepo) GetOrCreateWallet(ctx context.Context, userID uuid.UUID, currency string) (*models.Wallet, error) {
	w := &models.Wallet{}
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO wallets (user_id, balance, currency)
		VALUES ($1, 0, $2)
		ON CONFLICT (user_id, currency)
		DO UPDATE SET updated_at = NOW()
		RETURNING id, user_id, balance, currency, created_at, updated_at`,
		userID, currency).Scan(&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.CreatedAt, &w.UpdatedAt)
	return w, err
}

func (r *WalletRepo) GetWallet(ctx context.Context, walletID uuid.UUID) (*models.Wallet, error) {
	w := &models.Wallet{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, balance, currency, created_at, updated_at
		FROM wallets WHERE id = $1`, walletID).Scan(
		&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return w, nil
}

func (r *WalletRepo) GetWalletByUser(ctx context.Context, userID uuid.UUID, currency string) (*models.Wallet, error) {
	w := &models.Wallet{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, balance, currency, created_at, updated_at
		FROM wallets WHERE user_id = $1 AND currency = $2`, userID, currency).Scan(
		&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, helpers.ErrNoRowsAsNil(err)
	}
	return w, nil
}

func (r *WalletRepo) UpdateBalance(ctx context.Context, walletID uuid.UUID, newBalance int64) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`,
		newBalance, walletID)
	return err
}

func (r *WalletRepo) CreateTransaction(ctx context.Context, t *models.Transaction) error {
	t.ID = uuid.New()
	t.CreatedAt = time.Now()
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO wallet_transactions (id, wallet_id, amount, balance_before, balance_after, type, status, description, reference_type, reference_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		t.ID, t.WalletID, t.Amount, t.BalanceBefore, t.BalanceAfter, t.Type, t.Status, t.Description, t.ReferenceType, t.ReferenceID, t.CreatedAt)
	return err
}

func (r *WalletRepo) ListTransactions(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]models.Transaction, int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM wallet_transactions WHERE wallet_id = $1`, walletID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, wallet_id, amount, balance_before, balance_after, type, status, COALESCE(description, ''), COALESCE(reference_type, ''), reference_id, created_at
		FROM wallet_transactions WHERE wallet_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		walletID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.WalletID, &t.Amount, &t.BalanceBefore, &t.BalanceAfter, &t.Type, &t.Status, &t.Description, &t.ReferenceType, &t.ReferenceID, &t.CreatedAt); err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, t)
	}
	if transactions == nil {
		transactions = []models.Transaction{}
	}
	return transactions, total, nil
}

func (r *WalletRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Pool.Begin(ctx)
}
