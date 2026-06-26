package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/management-api/handler"
	"github.com/nexbic/platform/management-api/repository"
	"github.com/nexbic/platform/management-api/service"
	"github.com/nexbic/platform/shared/config"
	"github.com/nexbic/platform/shared/database"
	"github.com/nexbic/platform/shared/middleware"
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

	tp, err := middleware.InitTracing(ctx, cfg.AppName+"-mgmt-api", cfg.Tracing.OTLPEndpoint)
	if err != nil {
		slog.Warn("failed to initialize tracing", "error", err)
	}
	if tp != nil {
		defer func() { _ = tp.Shutdown(ctx) }()
	}

	appRepo := repository.NewAppRepository(db)
	keyRepo := repository.NewAPIKeyRepository(db)
	userRepo := repository.NewUserRepository(db)
	domainRepo := repository.NewDomainRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWT)
	appService := service.NewAppService(appRepo, keyRepo, userRepo, db, cfg.JWT.Secret, cfg.PostgREST.AdminURL)
	keyService := service.NewAPIKeyService(keyRepo)
	domainService := service.NewDomainService(domainRepo)
	settingsService := service.NewSettingsService(db)
	extensionService := service.NewExtensionService(db)

	authHandler := handler.NewAuthHandler(authService)
	appHandler := handler.NewAppHandler(appService, settingsService)
	keyHandler := handler.NewAPIKeyHandler(keyService)
	domainHandler := handler.NewDomainHandler(domainService)
	extensionHandler := handler.NewExtensionHandler(extensionService)
	tableHandler := handler.NewTableHandler(db, appRepo)
	logHandler := handler.NewLogHandler(db, appRepo)

	authMW := middleware.NewAuthMiddleware(cfg.JWT)

	f := fiber.New(fiber.Config{
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		AppName:           cfg.AppName + "-mgmt-api",
		EnablePrintRoutes: false,
	})

	f.Use(middleware.RequestID())
	f.Use(middleware.Logger())
	f.Use(middleware.Recover())
	f.Use(middleware.CORS(cfg.Server.CORSOrigins))
	f.Use(middleware.MetricsMiddleware())

	if cfg.Tracing.Enabled {
		f.Use(middleware.TracingMiddleware())
	}

	f.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "management-api",
		})
	})

	f.Get("/ready", func(c *fiber.Ctx) error {
		if err := db.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":  "not_ready",
				"service": "management-api",
				"reason":  "database unavailable",
			})
		}
		return c.JSON(fiber.Map{
			"status":  "ready",
			"service": "management-api",
		})
	})

	if cfg.Monitoring.Enabled {
		f.Get(cfg.Monitoring.MetricPath, middleware.MetricsHandler())
	}

	api := f.Group("/api/v1/platform")

	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/refresh", authHandler.RefreshToken)

	api.Get("/me", authMW.RequireAuth(), authMW.RequireAdmin(), authHandler.Me)

	api.Post("/apps", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditCreate, "app"), appHandler.Create)
	api.Get("/apps", authMW.RequireAuth(), authMW.RequireAdmin(), appHandler.List)
	api.Get("/apps/:id", authMW.RequireAuth(), authMW.RequireAdmin(), appHandler.GetByID)
	api.Patch("/apps/:id", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditUpdate, "app"), appHandler.Update)
	api.Delete("/apps/:id", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditDelete, "app"), appHandler.Delete)

	api.Post("/apps/:id/apikey", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditCreate, "api_key"), keyHandler.Create)
	api.Get("/apps/:id/apikey", authMW.RequireAuth(), authMW.RequireAdmin(), keyHandler.List)
	api.Delete("/apps/:id/apikey/:keyId", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditDelete, "api_key"), keyHandler.Deactivate)

	api.Get("/apps/:id/domains", authMW.RequireAuth(), authMW.RequireAdmin(), domainHandler.List)
	api.Post("/apps/:id/domains", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditCreate, "domain"), domainHandler.Create)
	api.Delete("/apps/:id/domains/:domainId", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditDelete, "domain"), domainHandler.Delete)
	api.Post("/apps/:id/domains/:domainId/verify", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditVerify, "domain"), domainHandler.Verify)

	api.Get("/users", authMW.RequireAuth(), authMW.RequireAdmin(), authHandler.ListUsers)
	api.Get("/users/:userId", authMW.RequireAuth(), authMW.RequireAdmin(), authHandler.GetUser)
	api.Post("/users/:userId/suspend", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditSuspend, "user"), authHandler.SuspendUser)
	api.Post("/users/:userId/activate", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditActivate, "user"), authHandler.ActivateUser)

	api.Get("/apps/:id/extensions", authMW.RequireAuth(), authMW.RequireAdmin(), extensionHandler.List)
	api.Post("/apps/:id/extensions/toggle", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditUpdate, "extension"), extensionHandler.Toggle)

	api.Get("/apps/:id/tables", authMW.RequireAuth(), authMW.RequireAdmin(), tableHandler.ListTables)
	api.Post("/apps/:id/tables", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditCreate, "table"), tableHandler.CreateTable)
	api.Get("/apps/:id/tables/:table", authMW.RequireAuth(), authMW.RequireAdmin(), tableHandler.GetTableData)
	api.Post("/apps/:id/tables/:table/rows", authMW.RequireAuth(), authMW.RequireAdmin(), tableHandler.InsertRow)
	api.Patch("/apps/:id/tables/:table/rows", authMW.RequireAuth(), authMW.RequireAdmin(), tableHandler.UpdateRow)
	api.Delete("/apps/:id/tables/:table/rows", authMW.RequireAuth(), authMW.RequireAdmin(), tableHandler.DeleteRow)
	api.Post("/apps/:id/tables/:table/columns", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditUpdate, "column"), tableHandler.AddColumn)

	api.Post("/apps/:id/sql", authMW.RequireAuth(), authMW.RequireAdmin(), tableHandler.RunSQL)

	api.Get("/apps/:id/logs", authMW.RequireAuth(), authMW.RequireAdmin(), logHandler.ListAppLogs)
	api.Get("/logs", authMW.RequireAuth(), authMW.RequireAdmin(), logHandler.ListGlobalLogs)

	api.Get("/settings", authMW.RequireAuth(), authMW.RequireAdmin(), appHandler.GetSettings)
	api.Patch("/settings", authMW.RequireAuth(), authMW.RequireAdmin(), middleware.AuditLog(middleware.AuditUpdate, "settings"), appHandler.UpdateSettings)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		port := os.Getenv("MGMT_PORT")
		if port == "" {
			port = "8081"
		}
		addr := fmt.Sprintf("%s:%s", cfg.Server.Host, port)
		slog.Info("management API starting", "address", addr)
		if err := f.Listen(addr); err != nil {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-quit
	slog.Info("shutting down management API", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := f.ShutdownWithContext(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
