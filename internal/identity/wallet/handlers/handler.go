package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/identity/wallet/models"
	"github.com/nexbic/platform/internal/identity/wallet/service"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type WalletHandler struct {
	svc *service.WalletService
}

func NewWalletHandler(svc *service.WalletService) *WalletHandler {
	return &WalletHandler{svc: svc}
}

func (h *WalletHandler) GetBalance(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	currency := c.Query("currency", "USD")
	w, err := h.svc.GetBalance(c.Context(), userID, currency)
	if err != nil {
		return response.InternalError(c, "failed to get balance")
	}
	return response.OK(c, w)
}

func (h *WalletHandler) CreateTransaction(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	var req models.CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	t, err := h.svc.CreateTransaction(c.Context(), userID, &req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, t)
}

func (h *WalletHandler) ListTransactions(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	currency := c.Query("currency", "USD")
	p := helpers.ParsePagination(c)
	transactions, total, err := h.svc.ListTransactions(c.Context(), userID, currency, p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to list transactions")
	}
	return response.Paginated(c, transactions, total, p.Limit, p.Offset)
}
