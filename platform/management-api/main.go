package main

import (
	"context"
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
	appHandler := handler.NewAppHandler(appService)
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

	api.Get("/me", authMW.RequireAuth(), authHandler.Me)

	api.Post("/apps", authMW.RequireAuth(), appHandler.Create)
	api.Get("/apps", authMW.RequireAuth(), appHandler.List)
	api.Get("/apps/:id", authMW.RequireAuth(), appHandler.GetByID)
	api.Delete("/apps/:id", authMW.RequireAuth(), appHandler.Delete)

	api.Post("/apps/:id/apikey", authMW.RequireAuth(), keyHandler.Create)
	api.Get("/apps/:id/apikey", authMW.RequireAuth(), keyHandler.List)
	api.Delete("/apps/:id/apikey/:keyId", authMW.RequireAuth(), keyHandler.Deactivate)

	api.Get("/apps/:id/domains", authMW.RequireAuth(), domainHandler.List)
	api.Post("/apps/:id/domains", authMW.RequireAuth(), domainHandler.Create)
	api.Delete("/apps/:id/domains/:domainId", authMW.RequireAuth(), domainHandler.Delete)
	api.Post("/apps/:id/domains/:domainId/verify", authMW.RequireAuth(), domainHandler.Verify)

	api.Get("/users", authMW.RequireAuth(), authHandler.ListUsers)
	api.Get("/users/:userId", authMW.RequireAuth(), authHandler.GetUser)

	api.Get("/backups", authMW.RequireAuth(), appHandler.ListBackups)
	api.Post("/backups", authMW.RequireAuth(), appHandler.CreateBackup)

	api.Get("/settings", authMW.RequireAuth(), appHandler.GetSettings)
	api.Patch("/settings", authMW.RequireAuth(), appHandler.UpdateSettings)

	f.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "management-api",
		})
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := cfg.Server.Host + ":" + "8081"
		log.Printf("Management API starting on %s", addr)
		if err := f.Listen(addr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")
	f.ShutdownWithTimeout(cfg.Server.ShutdownTimeout)
}
