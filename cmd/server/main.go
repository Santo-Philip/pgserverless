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
	authHandlers "github.com/nexbic/platform/internal/identity/auth/handlers"
	authRepo "github.com/nexbic/platform/internal/identity/auth/repository"
	authRoutes "github.com/nexbic/platform/internal/identity/auth/routes"
	authService "github.com/nexbic/platform/internal/identity/auth/service"
	backupHandlers "github.com/nexbic/platform/internal/database/backups/handlers"
	backupRoutes "github.com/nexbic/platform/internal/database/backups/routes"
	backupService "github.com/nexbic/platform/internal/database/backups/service"
	dashHandlers "github.com/nexbic/platform/internal/dashboard/handlers"
	dashRoutes "github.com/nexbic/platform/internal/dashboard/routes"
	dashService "github.com/nexbic/platform/internal/dashboard/service"
	explorerHandlers "github.com/nexbic/platform/internal/database/explorer/handlers"
	explorerRoutes "github.com/nexbic/platform/internal/database/explorer/routes"
	explorerService "github.com/nexbic/platform/internal/database/explorer/service"
	extHandlers "github.com/nexbic/platform/internal/database/extensions/handlers"
	extRoutes "github.com/nexbic/platform/internal/database/extensions/routes"
	extService "github.com/nexbic/platform/internal/database/extensions/service"
	logsHandlers "github.com/nexbic/platform/internal/database/logs/handlers"
	logsRoutes "github.com/nexbic/platform/internal/database/logs/routes"
	logsService "github.com/nexbic/platform/internal/database/logs/service"
	"github.com/nexbic/platform/internal/middleware"
	projectsHandlers "github.com/nexbic/platform/internal/projects/handlers"
	projectsRoutes "github.com/nexbic/platform/internal/projects/routes"
	monHandlers "github.com/nexbic/platform/internal/database/monitoring/handlers"
	monRoutes "github.com/nexbic/platform/internal/database/monitoring/routes"
	monService "github.com/nexbic/platform/internal/database/monitoring/service"
	pgroleHandlers "github.com/nexbic/platform/internal/database/roles/handlers"
	pgroleRoutes "github.com/nexbic/platform/internal/database/roles/routes"
	pgroleService "github.com/nexbic/platform/internal/database/roles/service"
	schemaHandlers "github.com/nexbic/platform/internal/database/schema/handlers"
	schemaRoutes "github.com/nexbic/platform/internal/database/schema/routes"
	schemaService "github.com/nexbic/platform/internal/database/schema/service"
	storageHandlers "github.com/nexbic/platform/internal/database/storage/handlers"
	storageRepo "github.com/nexbic/platform/internal/database/storage/repository"
	storageRoutes "github.com/nexbic/platform/internal/database/storage/routes"
	storageService "github.com/nexbic/platform/internal/database/storage/service"
	sqlHandlers "github.com/nexbic/platform/internal/database/sql/handlers"
	sqlRoutes "github.com/nexbic/platform/internal/database/sql/routes"
	sqlService "github.com/nexbic/platform/internal/database/sql/service"
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

	authMW := middleware.NewAuthMiddleware(cfg.JWT, db.Pool)

	userRepo := authRepo.NewUserRepository(db)
	tokenRepo := authRepo.NewRefreshTokenRepo(db)
	secRepo := authRepo.NewSecurityRepo(db)
	authSvc := authService.NewAuthService(userRepo, secRepo, tokenRepo, cfg.JWT, cfg.SuperAdmin)
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

	storageRepoInstance := storageRepo.NewStorageRepo(db)
	storageSvc := storageService.NewStorageService(storageRepoInstance)
	storageHandler := storageHandlers.NewStorageHandler(storageSvc)

	projectsHandler := projectsHandlers.NewProjectsHandler(db.Pool)

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

	f.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/", fiber.StatusFound)
	})

	api := f.Group("/v1")

	// ── Independent of projects ──
	authRoutes.RegisterAuthRoutes(api, authHandler, authMW)
	auditRoutes.RegisterAuditRoutes(api, auditHandler, authMW)

	// ── Projects CRUD (meta, at /v1/projects) ──
	projectsRoutes.RegisterProjectsRoutes(api, projectsHandler, authMW)

	// ── Project-scoped resources (under /v1/projects/:projectId) ──
	scope := api.Group("/projects/:projectId", authMW.RequireAuth(), middleware.ProjectGuard(db.Pool))

	dashRoutes.RegisterDashboardRoutes(scope, dashHandler, authMW)
	explorerRoutes.RegisterExplorerRoutes(scope, explorerHandler, authMW)

	sqlRoutes.RegisterSQLRoutes(scope, sqlHandler, authMW)
	schemaRoutes.RegisterSchemaRoutes(scope, schemaHandler, authMW)
	pgroleRoutes.RegisterPgRolesRoutes(scope, pgroleHandler, authMW)
	extRoutes.RegisterExtensionRoutes(scope, extHandler, authMW)
	monRoutes.RegisterMonitoringRoutes(scope, monHandler, authMW)
	backupRoutes.RegisterBackupRoutes(scope, backupHandler, authMW)
	logsRoutes.RegisterLogsRoutes(scope, logsHandler, authMW)
	storageRoutes.RegisterStorageRoutes(scope, storageHandler, authMW)

	// Serve static documentation at /docs
	f.Use("/docs", filesystem.New(filesystem.Config{
		Root:  http.Dir("./docs"),
		Index: "index.html",
	}))

	// Serve SvelteKit frontend if build directory exists
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
