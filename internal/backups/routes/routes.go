package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/backups/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterBackupRoutes(router fiber.Router, handler *handlers.BackupHandler, authMW *middleware.AuthMiddleware) {
	backups := router.Group("/backups")

	backups.Get("/", authMW.RequireAuth(), handler.ListBackups)
	backups.Post("/", authMW.RequireAuth(), handler.CreateBackup)
	backups.Get("/:id", authMW.RequireAuth(), handler.GetBackup)
	backups.Delete("/:id", authMW.RequireAuth(), handler.DeleteBackup)
	backups.Post("/:id/restore", authMW.RequireAuth(), handler.RestoreBackup)
	backups.Get("/:id/download", authMW.RequireAuth(), handler.DownloadBackup)
	backups.Get("/:id/verify", authMW.RequireAuth(), handler.VerifyBackup)
}
