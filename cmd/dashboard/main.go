package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/nexbic/platform/internal/app"
	auditRoutes "github.com/nexbic/platform/internal/audit/routes"
	authRoutes "github.com/nexbic/platform/internal/auth/routes"
	backupRoutes "github.com/nexbic/platform/internal/backups/routes"
	dashRoutes "github.com/nexbic/platform/internal/dashboard/routes"
	explorerRoutes "github.com/nexbic/platform/internal/explorer/routes"
	extRoutes "github.com/nexbic/platform/internal/extensions/routes"
	logsRoutes "github.com/nexbic/platform/internal/logs/routes"
	"github.com/nexbic/platform/internal/middleware"
	monRoutes "github.com/nexbic/platform/internal/monitoring/routes"
	pgroleRoutes "github.com/nexbic/platform/internal/pgroles/routes"
	schemaRoutes "github.com/nexbic/platform/internal/schema/routes"
	sqlRoutes "github.com/nexbic/platform/internal/sql/routes"
	storageRoutes "github.com/nexbic/platform/internal/storage/routes"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, err := app.New(ctx)
	if err != nil {
		slog.Error("app initialization failed", "error", err)
		os.Exit(1)
	}
	defer a.Close()

	f := fiber.New(fiber.Config{
		ReadTimeout:       a.Config.Server.ReadTimeout,
		WriteTimeout:      a.Config.Server.WriteTimeout,
		AppName:           a.Config.AppName + "-dashboard",
		EnablePrintRoutes: false,
	})

	f.Use(middleware.RequestID())
	f.Use(middleware.Logger())
	f.Use(middleware.Recover())
	f.Use(middleware.CORS(a.Config.Server.CORSOrigins))
	f.Use(middleware.RateLimit(200, 1*time.Minute))

	f.Use(func(c *fiber.Ctx) error {
		c.Locals("db", a.DB)
		return c.Next()
	})

	f.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "service": "nexbic-dashboard"})
	})
	f.Get("/ready", func(c *fiber.Ctx) error {
		if err := a.DB.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "not_ready", "reason": "database unavailable"})
		}
		return c.JSON(fiber.Map{"status": "ready"})
	})

	api := f.Group("/v1")

	authRoutes.RegisterDashboardAuthRoutes(api, a.AuthHandler, a.AuthMW)
	auditRoutes.RegisterAuditRoutes(api, a.AuditHandler, a.AuthMW)

	scope := api.Group("/projects/:projectId", a.AuthMW.RequireAuth(), middleware.ProjectGuard(a.DB.Pool))

	dashRoutes.RegisterDashboardRoutes(scope, a.DashHandler, a.AuthMW)
	explorerRoutes.RegisterExplorerRoutes(scope, a.ExplorerHandler, a.AuthMW)
	sqlRoutes.RegisterSQLRoutes(scope, a.SQLHandler, a.AuthMW)
	schemaRoutes.RegisterSchemaRoutes(scope, a.SchemaHandler, a.AuthMW)
	pgroleRoutes.RegisterPgRolesRoutes(scope, a.PgRoleHandler, a.AuthMW)
	extRoutes.RegisterExtensionRoutes(scope, a.ExtHandler, a.AuthMW)
	monRoutes.RegisterMonitoringRoutes(scope, a.MonHandler, a.AuthMW)
	backupRoutes.RegisterBackupRoutes(scope, a.BackupHandler, a.AuthMW)
	logsRoutes.RegisterLogsRoutes(scope, a.LogsHandler, a.AuthMW)
	storageRoutes.RegisterStorageRoutes(scope, a.StorageHandler, a.AuthMW)

	if _, err := os.Stat("./dashboard/build"); err == nil {
		f.Use("/", filesystem.New(filesystem.Config{
			Root:         http.Dir("./dashboard/build"),
			Index:        "index.html",
			NotFoundFile: "index.html",
		}))
	}

	f.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "not_found",
			"message": "route not found",
		})
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := a.Config.Addr()
		slog.Info("Dashboard server starting", "address", addr)
		if err := f.Listen(addr); err != nil {
			slog.Error("Dashboard server error", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-quit
	slog.Info("shutting down dashboard server", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.Config.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := f.ShutdownWithContext(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
