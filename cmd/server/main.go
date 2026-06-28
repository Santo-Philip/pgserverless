package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/nexbic/platform/config"
	auditHandlers "github.com/nexbic/platform/internal/audit/handlers"
	auditRoutes "github.com/nexbic/platform/internal/audit/routes"
	auditService "github.com/nexbic/platform/internal/audit/service"
	authHandlers "github.com/nexbic/platform/internal/auth/handlers"
	authRepo "github.com/nexbic/platform/internal/auth/repository"
	authRoutes "github.com/nexbic/platform/internal/auth/routes"
	authService "github.com/nexbic/platform/internal/auth/service"
	backupHandlers "github.com/nexbic/platform/internal/backups/handlers"
	backupRoutes "github.com/nexbic/platform/internal/backups/routes"
	backupService "github.com/nexbic/platform/internal/backups/service"
	dashHandlers "github.com/nexbic/platform/internal/dashboard/handlers"
	dashRoutes "github.com/nexbic/platform/internal/dashboard/routes"
	dashService "github.com/nexbic/platform/internal/dashboard/service"
	explorerHandlers "github.com/nexbic/platform/internal/explorer/handlers"
	explorerRoutes "github.com/nexbic/platform/internal/explorer/routes"
	explorerService "github.com/nexbic/platform/internal/explorer/service"
	extHandlers "github.com/nexbic/platform/internal/extensions/handlers"
	extRoutes "github.com/nexbic/platform/internal/extensions/routes"
	extService "github.com/nexbic/platform/internal/extensions/service"
	logsHandlers "github.com/nexbic/platform/internal/logs/handlers"
	logsRoutes "github.com/nexbic/platform/internal/logs/routes"
	logsService "github.com/nexbic/platform/internal/logs/service"
	"github.com/nexbic/platform/internal/middleware"
	monHandlers "github.com/nexbic/platform/internal/monitoring/handlers"
	monRoutes "github.com/nexbic/platform/internal/monitoring/routes"
	monService "github.com/nexbic/platform/internal/monitoring/service"
	pgroleHandlers "github.com/nexbic/platform/internal/pgroles/handlers"
	pgroleRoutes "github.com/nexbic/platform/internal/pgroles/routes"
	pgroleService "github.com/nexbic/platform/internal/pgroles/service"
	schemaHandlers "github.com/nexbic/platform/internal/schema/handlers"
	schemaRoutes "github.com/nexbic/platform/internal/schema/routes"
	schemaService "github.com/nexbic/platform/internal/schema/service"
	sqlHandlers "github.com/nexbic/platform/internal/sql/handlers"
	sqlRoutes "github.com/nexbic/platform/internal/sql/routes"
	sqlService "github.com/nexbic/platform/internal/sql/service"
	tableHandlers "github.com/nexbic/platform/internal/tables/handlers"
	tableRoutes "github.com/nexbic/platform/internal/tables/routes"
	tableService "github.com/nexbic/platform/internal/tables/service"
	"github.com/nexbic/platform/pkg/database"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	authMW := middleware.NewAuthMiddleware(cfg.JWT)

	userRepo := authRepo.NewUserRepository(db)
	tokenRepo := authRepo.NewRefreshTokenRepo(db)
	authSvc := authService.NewAuthService(userRepo, tokenRepo, cfg.JWT, cfg.SuperAdmin)
	authHandler := authHandlers.NewAuthHandler(authSvc)

	if cfg.SuperAdmin.Email != "" {
		authSvc.SeedSuperAdmin(ctx)
	}

	auditSvc := auditService.NewAuditService(db)
	auditHandler := auditHandlers.NewAuditHandler(auditSvc)

	dashSvc := dashService.NewDashboardService(db)
	dashHandler := dashHandlers.NewDashboardHandler(dashSvc)

	explorerSvc := explorerService.NewExplorerService(db)
	explorerHandler := explorerHandlers.NewExplorerHandler(explorerSvc)

	tableSvc := tableService.NewTablesService(db)
	tableHandler := tableHandlers.NewTablesHandler(tableSvc)

	sqlSvc := sqlService.NewSQLService(db)
	sqlHandler := sqlHandlers.NewSQLHandler(sqlSvc)

	schemaSvc := schemaService.NewSchemaService(db)
	schemaHandler := schemaHandlers.NewSchemaHandler(schemaSvc)

	pgroleSvc := pgroleService.NewPgRolesService(db)
	pgroleHandler := pgroleHandlers.NewPgRolesHandler(pgroleSvc)

	extSvc := extService.NewExtensionsService(db)
	extHandler := extHandlers.NewExtensionsHandler(extSvc)

	monSvc := monService.NewMonitoringService(db)
	monHandler := monHandlers.NewMonitoringHandler(monSvc)

	backupDir := os.Getenv("BACKUP_DIR")
	if backupDir == "" {
		backupDir = "/data/backups"
	}
	backupSvc := backupService.NewBackupService(db, backupDir,
		cfg.Database.Host, strconv.Itoa(cfg.Database.Port),
		cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	backupHandler := backupHandlers.NewBackupHandler(backupSvc)

	logsSvc := logsService.NewLogsService(db)
	logsHandler := logsHandlers.NewLogsHandler(logsSvc)

	f := fiber.New(fiber.Config{
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		AppName:           cfg.AppName,
		EnablePrintRoutes: false,
	})

	f.Use(middleware.RequestID())
	f.Use(middleware.Logger())
	f.Use(middleware.Recover())
	f.Use(middleware.CORS(cfg.Server.CORSOrigins))
	f.Use(middleware.RateLimit(200, 1*time.Minute))

	f.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})

	f.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "nexbic-pg-admin",
		})
	})

	f.Get("/ready", func(c *fiber.Ctx) error {
		if err := db.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "not_ready",
				"reason": "database unavailable",
			})
		}
		return c.JSON(fiber.Map{
			"status": "ready",
		})
	})

	api := f.Group("/api/v1")

	authRoutes.RegisterAuthRoutes(api, authHandler, authMW)
	auditRoutes.RegisterAuditRoutes(api, auditHandler, authMW)
	dashRoutes.RegisterDashboardRoutes(api, dashHandler, authMW)
	explorerRoutes.RegisterExplorerRoutes(api, explorerHandler, authMW)
	tableRoutes.RegisterTablesRoutes(api, tableHandler)
	sqlRoutes.RegisterSQLRoutes(api, sqlHandler, authMW)
	schemaRoutes.RegisterSchemaRoutes(api, schemaHandler)
	pgroleRoutes.RegisterPgRolesRoutes(api, pgroleHandler)
	extRoutes.RegisterExtensionRoutes(api, extHandler, authMW)
	monRoutes.RegisterMonitoringRoutes(api, monHandler, authMW)
	backupRoutes.RegisterBackupRoutes(api, backupHandler, authMW)
	logsRoutes.RegisterLogsRoutes(api, logsHandler, authMW)

	// Serve SvelteKit frontend if build directory exists
	if _, err := os.Stat("../dashboard/build"); err == nil {
		f.Use("/", filesystem.New(filesystem.Config{
			Root:         http.Dir("../dashboard/build"),
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
		addr := cfg.Addr()
		slog.Info("server starting", "address", addr)
		if err := f.Listen(addr); err != nil {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-quit
	slog.Info("shutting down", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := f.ShutdownWithContext(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
