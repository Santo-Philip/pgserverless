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
	backupRoutes "github.com/nexbic/platform/internal/database/backups/routes"
	explorerRoutes "github.com/nexbic/platform/internal/database/explorer/routes"
	extRoutes "github.com/nexbic/platform/internal/database/extensions/routes"
	logsRoutes "github.com/nexbic/platform/internal/database/logs/routes"
	monRoutes "github.com/nexbic/platform/internal/database/monitoring/routes"
	pgroleRoutes "github.com/nexbic/platform/internal/database/roles/routes"
	schemaRoutes "github.com/nexbic/platform/internal/database/schema/routes"
	sqlRoutes "github.com/nexbic/platform/internal/database/sql/routes"
	storageRoutes "github.com/nexbic/platform/internal/database/storage/routes"
	dashRoutes "github.com/nexbic/platform/internal/dashboard/routes"
	authRoutes "github.com/nexbic/platform/internal/identity/auth/routes"
	walletRoutes "github.com/nexbic/platform/internal/identity/wallet/routes"
	"github.com/nexbic/platform/internal/middleware"
	projectsRoutes "github.com/nexbic/platform/internal/projects/routes"
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

	v1 := f.Group("/v1")

	// Identity endpoints (dashboard-appropriate subset)
	authRoutes.RegisterDashboardAuthRoutes(v1, a.AuthHandler, a.AuthMW)

	// Project-scoped internal database services
	scope := v1.Group("/projects/:projectId", a.AuthMW.RequireAuth(), middleware.ProjectGuard(a.DB.Pool))

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

	// Dashboard also manages projects
	projectsRoutes.RegisterProjectsRoutes(v1, a.ProjectsHandler, a.AuthMW)
	walletRoutes.RegisterWalletRoutes(v1, a.WalletHandler, a.AuthMW)

	// Audit (global, not scoped to projects)
	auditRoutes.RegisterAuditRoutes(v1, a.AuditHandler, a.AuthMW)

	// Svelte frontend
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
		slog.Info("Nexbic Dashboard starting", "address", addr)
		if err := f.Listen(addr); err != nil {
			slog.Error("Dashboard server error", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-quit
	slog.Info("shutting down Nexbic Dashboard", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.Config.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := f.ShutdownWithContext(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
