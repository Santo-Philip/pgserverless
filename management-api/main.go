package main

import (
	"context"
	"fmt"
	"log"
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

	db, err := database.New(context.Background(), cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	appRepo := repository.NewAppRepository(db)
	keyRepo := repository.NewAPIKeyRepository(db)
	userRepo := repository.NewUserRepository(db)
	domainRepo := repository.NewDomainRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWT)
	appService := service.NewAppService(appRepo, keyRepo, userRepo, db, cfg.JWT.Secret, cfg.PostgREST.AdminURL)
	keyService := service.NewAPIKeyService(keyRepo)
	domainService := service.NewDomainService(domainRepo)

	authHandler := handler.NewAuthHandler(authService)
	appHandler := handler.NewAppHandler(appService, db)
	keyHandler := handler.NewAPIKeyHandler(keyService, keyRepo)
	domainHandler := handler.NewDomainHandler(domainService)

	authMW := middleware.NewAuthMiddleware(cfg.JWT)

	f := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		AppName:      cfg.AppName,
	})

	f.Use(middleware.RequestID())
	f.Use(middleware.Logger())
	f.Use(middleware.Recover())
	f.Use(middleware.CORS(cfg.Server.CORSOrigins))

	api := f.Group("/api/v1/platform")

	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/refresh", authHandler.RefreshToken)

	api.Get("/me", authMW.RequireAuth(), authMW.RequireAdmin(), authHandler.Me)

	api.Post("/apps", authMW.RequireAuth(), authMW.RequireAdmin(), appHandler.Create)
	api.Get("/apps", authMW.RequireAuth(), authMW.RequireAdmin(), appHandler.List)
	api.Get("/apps/:id", authMW.RequireAuth(), authMW.RequireAdmin(), appHandler.GetByID)
	api.Delete("/apps/:id", authMW.RequireAuth(), authMW.RequireAdmin(), appHandler.Delete)

	api.Post("/apps/:id/apikey", authMW.RequireAuth(), authMW.RequireAdmin(), keyHandler.Create)
	api.Get("/apps/:id/apikey", authMW.RequireAuth(), authMW.RequireAdmin(), keyHandler.List)
	api.Delete("/apps/:id/apikey/:keyId", authMW.RequireAuth(), authMW.RequireAdmin(), keyHandler.Deactivate)

	api.Get("/apps/:id/domains", authMW.RequireAuth(), authMW.RequireAdmin(), domainHandler.List)
	api.Post("/apps/:id/domains", authMW.RequireAuth(), authMW.RequireAdmin(), domainHandler.Create)
	api.Delete("/apps/:id/domains/:domainId", authMW.RequireAuth(), authMW.RequireAdmin(), domainHandler.Delete)
	api.Post("/apps/:id/domains/:domainId/verify", authMW.RequireAuth(), authMW.RequireAdmin(), domainHandler.Verify)

	api.Get("/users", authMW.RequireAuth(), authMW.RequireAdmin(), authHandler.ListUsers)
	api.Get("/users/:userId", authMW.RequireAuth(), authMW.RequireAdmin(), authHandler.GetUser)
	api.Post("/users/:userId/suspend", authMW.RequireAuth(), authMW.RequireAdmin(), authHandler.SuspendUser)
	api.Post("/users/:userId/activate", authMW.RequireAuth(), authMW.RequireAdmin(), authHandler.ActivateUser)

	api.Get("/backups", authMW.RequireAuth(), authMW.RequireAdmin(), appHandler.ListBackups)
	api.Post("/backups", authMW.RequireAuth(), authMW.RequireAdmin(), appHandler.CreateBackup)

	api.Get("/settings", authMW.RequireAuth(), authMW.RequireAdmin(), appHandler.GetSettings)
	api.Patch("/settings", authMW.RequireAuth(), authMW.RequireAdmin(), appHandler.UpdateSettings)

	f.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "management-api",
		})
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf("0.0.0.0:%s", os.Getenv("MGMT_PORT"))
		if os.Getenv("MGMT_PORT") == "" {
			addr = "0.0.0.0:8081"
		}
		log.Printf("Management API starting on %s", addr)
		if err := f.Listen(addr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")
	f.ShutdownWithTimeout(cfg.Server.ShutdownTimeout)
}
