package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexbic/platform/internal/identity/wallet/models"
	"github.com/nexbic/platform/internal/identity/wallet/repository"
)

type WalletService struct {
	repo *repository.WalletRepo
}

func NewWalletService(repo *repository.WalletRepo) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) GetBalance(ctx context.Context, userID uuid.UUID, currency string) (*models.Wallet, error) {
	w, err := s.repo.GetWalletByUser(ctx, userID, currency)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return s.repo.GetOrCreateWallet(ctx, userID, currency)
	}
	return w, nil
}

func (s *WalletService) CreateTransaction(ctx context.Context, userID uuid.UUID, req *models.CreateTransactionRequest) (*models.Transaction, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if req.Type != "credit" && req.Type != "debit" {
		return nil, fmt.Errorf("type must be 'credit' or 'debit'")
	}

	w, err := s.repo.GetOrCreateWallet(ctx, userID, "USD")
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}

	if req.Type == "debit" && w.Balance < req.Amount {
		return nil, fmt.Errorf("insufficient balance")
	}

	balanceBefore := w.Balance
	var balanceAfter int64
	if req.Type == "credit" {
		balanceAfter = balanceBefore + req.Amount
	} else {
		balanceAfter = balanceBefore - req.Amount
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.UpdateBalance(ctx, w.ID, balanceAfter); err != nil {
		return nil, fmt.Errorf("update balance: %w", err)
	}

	t := &models.Transaction{
		WalletID:      w.ID,
		Amount:        req.Amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Type:          req.Type,
		Status:        "completed",
		Description:   req.Description,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
	}
	if t.Description == "" {
		if req.Type == "credit" {
			t.Description = "Deposit"
		} else {
			t.Description = "Withdrawal"
		}
	}

	if err := s.repo.CreateTransaction(ctx, t); err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return t, nil
}

func (s *WalletService) ListTransactions(ctx context.Context, userID uuid.UUID, currency string, limit, offset int) ([]models.Transaction, int, error) {
	w, err := s.repo.GetWalletByUser(ctx, userID, currency)
	if err != nil {
		return nil, 0, err
	}
	if w == nil {
		return []models.Transaction{}, 0, nil
	}
	return s.repo.ListTransactions(ctx, w.ID, limit, offset)
}
