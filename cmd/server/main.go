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

	f.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "Nexbic PostgreSQL Admin",
			"version": "1.0.0",
			"api_base": "/v1",
			"endpoints": []fiber.Map{
				{"path": "/v1/auth/login", "method": "POST", "description": "Authenticate and receive JWT"},
				{"path": "/v1/auth/refresh", "method": "POST", "description": "Refresh expired JWT"},
				{"path": "/v1/auth/me", "method": "GET", "description": "Get current user profile"},
				{"path": "/v1/auth/password", "method": "PATCH", "description": "Update own password"},
				{"path": "/v1/admin/users", "method": "GET", "description": "List all users"},
				{"path": "/v1/admin/users", "method": "POST", "description": "Create a new user"},
				{"path": "/v1/admin/users/:id", "method": "GET", "description": "Get user by ID"},
				{"path": "/v1/admin/users/:id", "method": "PATCH", "description": "Update user"},
				{"path": "/v1/admin/users/:id", "method": "DELETE", "description": "Delete user"},
				{"path": "/v1/admin/users/:id/password", "method": "PATCH", "description": "Set user password"},
				{"path": "/v1/audit-logs", "method": "GET", "description": "List audit logs"},
				{"path": "/v1/audit-logs/:resource/:resource_id", "method": "GET", "description": "Get audit logs by resource"},
				{"path": "/v1/dashboard/overview", "method": "GET", "description": "Dashboard overview statistics"},
				{"path": "/v1/dashboard/stats", "method": "GET", "description": "Database statistics"},
				{"path": "/v1/dashboard/schemas", "method": "GET", "description": "List schemas"},
				{"path": "/v1/explorer/schemas", "method": "GET", "description": "List all schemas"},
				{"path": "/v1/explorer/schemas/:schema/tables", "method": "GET", "description": "List tables in schema"},
				{"path": "/v1/explorer/schemas/:schema/tables/:table", "method": "GET", "description": "Get table details"},
				{"path": "/v1/explorer/schemas/:schema/views", "method": "GET", "description": "List views in schema"},
				{"path": "/v1/explorer/schemas/:schema/functions", "method": "GET", "description": "List functions in schema"},
				{"path": "/v1/explorer/schemas/:schema/procedures", "method": "GET", "description": "List procedures in schema"},
				{"path": "/v1/explorer/schemas/:schema/triggers", "method": "GET", "description": "List triggers"},
				{"path": "/v1/explorer/schemas/:schema/indexes", "method": "GET", "description": "List indexes"},
				{"path": "/v1/explorer/schemas/:schema/constraints", "method": "GET", "description": "List constraints"},
				{"path": "/v1/explorer/schemas/:schema/sequences", "method": "GET", "description": "List sequences"},
				{"path": "/v1/explorer/schemas/:schema/materialized-views", "method": "GET", "description": "List materialized views"},
				{"path": "/v1/explorer/extensions", "method": "GET", "description": "List available extensions"},
				{"path": "/v1/tables/:schema/:table", "method": "GET", "description": "Browse table rows"},
				{"path": "/v1/tables/:schema/:table/rows", "method": "POST", "description": "Insert row"},
				{"path": "/v1/tables/:schema/:table/rows", "method": "PATCH", "description": "Update rows"},
				{"path": "/v1/tables/:schema/:table/rows", "method": "DELETE", "description": "Delete rows"},
				{"path": "/v1/tables/:schema/:table/rows/bulk", "method": "POST", "description": "Bulk insert rows"},
				{"path": "/v1/tables/:schema/:table/rows/bulk", "method": "DELETE", "description": "Bulk delete rows"},
				{"path": "/v1/tables/:schema/:table/search", "method": "GET", "description": "Search table rows"},
				{"path": "/v1/sql/execute", "method": "POST", "description": "Execute arbitrary SQL"},
				{"path": "/v1/sql/explain", "method": "POST", "description": "Explain query plan"},
				{"path": "/v1/sql/cancel", "method": "POST", "description": "Cancel running query"},
				{"path": "/v1/sql/history", "method": "GET", "description": "Query execution history"},
				{"path": "/v1/sql/saved", "method": "GET", "description": "List saved queries"},
				{"path": "/v1/sql/saved", "method": "POST", "description": "Save a query"},
				{"path": "/v1/sql/saved/:id", "method": "DELETE", "description": "Delete saved query"},
				{"path": "/v1/schemas", "method": "POST", "description": "Create schema"},
				{"path": "/v1/schemas/:name", "method": "DELETE", "description": "Drop schema"},
				{"path": "/v1/schemas/:schema/tables", "method": "POST", "description": "Create table"},
				{"path": "/v1/schemas/:schema/tables/:table", "method": "DELETE", "description": "Drop table"},
				{"path": "/v1/schemas/:schema/tables/:table/columns", "method": "POST", "description": "Add column"},
				{"path": "/v1/schemas/:schema/tables/:table/columns/:column", "method": "DELETE", "description": "Drop column"},
				{"path": "/v1/schemas/:schema/tables/:table/columns/:column", "method": "PATCH", "description": "Alter column"},
				{"path": "/v1/schemas/:schema/tables/:table/constraints", "method": "POST", "description": "Add constraint"},
				{"path": "/v1/schemas/:schema/tables/:table/constraints/:constraint", "method": "DELETE", "description": "Drop constraint"},
				{"path": "/v1/schemas/:schema/tables/:table/indexes", "method": "POST", "description": "Create index"},
				{"path": "/v1/schemas/:schema/tables/:table/indexes/:index", "method": "DELETE", "description": "Drop index"},
				{"path": "/v1/schemas/:schema/sequences", "method": "POST", "description": "Create sequence"},
				{"path": "/v1/schemas/:schema/sequences/:sequence", "method": "PATCH", "description": "Alter sequence"},
				{"path": "/v1/schemas/:schema/sequences/:sequence", "method": "DELETE", "description": "Drop sequence"},
				{"path": "/v1/schemas/:schema/tables/:table/ddl", "method": "GET", "description": "Get table DDL"},
				{"path": "/v1/roles", "method": "GET", "description": "List PostgreSQL roles"},
				{"path": "/v1/roles", "method": "POST", "description": "Create PostgreSQL role"},
				{"path": "/v1/roles/:name", "method": "GET", "description": "Get role details"},
				{"path": "/v1/roles/:name", "method": "PATCH", "description": "Alter role"},
				{"path": "/v1/roles/:name", "method": "DELETE", "description": "Drop role"},
				{"path": "/v1/roles/:name/password", "method": "POST", "description": "Set role password"},
				{"path": "/v1/roles/:role/grant-database", "method": "POST", "description": "Grant database privileges"},
				{"path": "/v1/roles/:role/grant-schema", "method": "POST", "description": "Grant schema privileges"},
				{"path": "/v1/roles/:role/grant-table", "method": "POST", "description": "Grant table privileges"},
				{"path": "/v1/roles/:role/revoke-database", "method": "POST", "description": "Revoke database privileges"},
				{"path": "/v1/roles/:role/revoke-schema", "method": "POST", "description": "Revoke schema privileges"},
				{"path": "/v1/roles/:role/revoke-table", "method": "POST", "description": "Revoke table privileges"},
				{"path": "/v1/roles/:role/add-member", "method": "POST", "description": "Add role member"},
				{"path": "/v1/roles/:role/remove-member", "method": "POST", "description": "Remove role member"},
				{"path": "/v1/roles/privileges/databases", "method": "GET", "description": "Get database privileges"},
				{"path": "/v1/roles/privileges/schemas", "method": "GET", "description": "Get schema privileges"},
				{"path": "/v1/roles/privileges/tables", "method": "GET", "description": "Get table privileges"},
				{"path": "/v1/roles/:name/members", "method": "GET", "description": "Get role memberships"},
				{"path": "/v1/extensions", "method": "GET", "description": "List installed extensions"},
				{"path": "/v1/extensions", "method": "POST", "description": "Install extension"},
				{"path": "/v1/extensions/:name", "method": "DELETE", "description": "Uninstall extension"},
				{"path": "/v1/monitoring/sessions", "method": "GET", "description": "List active sessions"},
				{"path": "/v1/monitoring/slow-queries", "method": "GET", "description": "List slow queries"},
				{"path": "/v1/monitoring/locks", "method": "GET", "description": "List active locks"},
				{"path": "/v1/monitoring/waiting", "method": "GET", "description": "List waiting queries"},
				{"path": "/v1/monitoring/query-stats", "method": "GET", "description": "Query statistics"},
				{"path": "/v1/monitoring/connections", "method": "GET", "description": "Connection statistics"},
				{"path": "/v1/monitoring/cache", "method": "GET", "description": "Cache statistics"},
				{"path": "/v1/monitoring/databases", "method": "GET", "description": "Database statistics"},
				{"path": "/v1/monitoring/table-stats", "method": "GET", "description": "Table statistics"},
				{"path": "/v1/monitoring/index-stats", "method": "GET", "description": "Index statistics"},
				{"path": "/v1/monitoring/sessions/terminate", "method": "POST", "description": "Terminate a session"},
				{"path": "/v1/backups", "method": "GET", "description": "List backups"},
				{"path": "/v1/backups", "method": "POST", "description": "Create backup"},
				{"path": "/v1/backups/:id", "method": "GET", "description": "Get backup details"},
				{"path": "/v1/backups/:id", "method": "DELETE", "description": "Delete backup"},
				{"path": "/v1/backups/:id/restore", "method": "POST", "description": "Restore backup"},
				{"path": "/v1/backups/:id/verify", "method": "POST", "description": "Verify backup"},
				{"path": "/v1/backups/:id/download", "method": "GET", "description": "Download backup"},
				{"path": "/v1/logs", "method": "GET", "description": "List PostgreSQL logs"},
				{"path": "/v1/logs/query", "method": "GET", "description": "Filter query logs"},
				{"path": "/v1/logs/auth", "method": "GET", "description": "Filter auth logs"},
			},
		})
	})

	api := f.Group("/v1")

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
