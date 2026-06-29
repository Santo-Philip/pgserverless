package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/middleware"
	"github.com/nexbic/platform/internal/identity/wallet/handlers"
)

func RegisterWalletRoutes(router fiber.Router, handler *handlers.WalletHandler, authMW *middleware.AuthMiddleware) {
	g := router.Group("/wallet", authMW.RequireAuth())
	g.Get("/balance", handler.GetBalance)
	g.Post("/transactions", handler.CreateTransaction)
	g.Get("/transactions", handler.ListTransactions)
}
