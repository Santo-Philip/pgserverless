package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/config"
	"github.com/nexbic/platform/internal/apikey/handlers"
	apikeyRepo "github.com/nexbic/platform/internal/apikey/repository"
	apikeyRoutes "github.com/nexbic/platform/internal/apikey/routes"
	apikeyService "github.com/nexbic/platform/internal/apikey/service"
	auditHandlers "github.com/nexbic/platform/internal/audit/handlers"
	auditRoutes "github.com/nexbic/platform/internal/audit/routes"
	auditService "github.com/nexbic/platform/internal/audit/service"
	authHandlers "github.com/nexbic/platform/internal/auth/handlers"
	authRepo "github.com/nexbic/platform/internal/auth/repository"
	authRoutes "github.com/nexbic/platform/internal/auth/routes"
	authService "github.com/nexbic/platform/internal/auth/service"
	dbHandlers "github.com/nexbic/platform/internal/database/handlers"
	dbRepo "github.com/nexbic/platform/internal/database/repository"
	dbRoutes "github.com/nexbic/platform/internal/database/routes"
	dbService "github.com/nexbic/platform/internal/database/service"
	"github.com/nexbic/platform/internal/middleware"
	planHandlers "github.com/nexbic/platform/internal/plan/handlers"
	planRepo "github.com/nexbic/platform/internal/plan/repository"
	planRoutes "github.com/nexbic/platform/internal/plan/routes"
	planService "github.com/nexbic/platform/internal/plan/service"
	projectHandlers "github.com/nexbic/platform/internal/project/handlers"
	projectRepo "github.com/nexbic/platform/internal/project/repository"
	projectRoutes "github.com/nexbic/platform/internal/project/routes"
	projectService "github.com/nexbic/platform/internal/project/service"
	quotaHandlers "github.com/nexbic/platform/internal/quota/handlers"
	quotaRoutes "github.com/nexbic/platform/internal/quota/routes"
	quotaService "github.com/nexbic/platform/internal/quota/service"
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
	authSvc := authService.NewAuthService(userRepo, tokenRepo, cfg.JWT)
	authHandler := authHandlers.NewAuthHandler(authSvc)

	projectRepo := projectRepo.NewProjectRepository(db)
	projectSvc := projectService.NewProjectService(projectRepo)
	projectHandler := projectHandlers.NewProjectHandler(projectSvc)

	databaseRepo := dbRepo.NewDatabaseRepository(db)
	databaseSvc := dbService.NewDatabaseService(databaseRepo)
	databaseHandler := dbHandlers.NewDatabaseHandler(databaseSvc)

	apikeyRepo := apikeyRepo.NewAPIKeyRepository(db)
	apikeySvc := apikeyService.NewAPIKeyService(apikeyRepo)
	apikeyHandler := handlers.NewAPIKeyHandler(apikeySvc)

	planRepo := planRepo.NewPlanRepository(db)
	planSvc := planService.NewPlanService(planRepo)
	planHandler := planHandlers.NewPlanHandler(planSvc)

	quotaSvc := quotaService.NewQuotaService(db, projectRepo)
	quotaHandler := quotaHandlers.NewQuotaHandler(quotaSvc)

	auditSvc := auditService.NewAuditService(db)
	auditHandler := auditHandlers.NewAuditHandler(auditSvc)

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
			"service": "nexbic-db-platform",
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
	projectRoutes.RegisterProjectRoutes(api, projectHandler, authMW)
	dbRoutes.RegisterDatabaseRoutes(api, databaseHandler, authMW)
	apikeyRoutes.RegisterAPIKeyRoutes(api, apikeyHandler, authMW)
	planRoutes.RegisterPlanRoutes(api, planHandler, authMW)
	quotaRoutes.RegisterQuotaRoutes(api, quotaHandler, authMW)
	auditRoutes.RegisterAuditRoutes(api, auditHandler, authMW)

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
