package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/gateway/handler"
	"github.com/nexbic/platform/gateway/proxy"
	"github.com/nexbic/platform/gateway/service"
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

	appResolver := service.NewAppResolver(db)
	postgrestProxy := proxy.NewPostgRESTProxy(cfg.PostgREST.URL, cfg.PostgREST.Timeout)
	authMW := middleware.NewAuthMiddleware(cfg.JWT)

	gatewayHandler := handler.NewGatewayHandler(appResolver, postgrestProxy, authMW)

	f := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		AppName:      cfg.AppName + "-gateway",
	})

	f.Use(middleware.RequestID())
	f.Use(middleware.Logger())
	f.Use(middleware.Recover())
	f.Use(middleware.CORS(cfg.Server.CORSOrigins))
	f.Use(middleware.RateLimit(100, 1*time.Minute))

	f.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "gateway",
		})
	})

	api := f.Group("/api")

	api.All("/v1/:app_slug/*", gatewayHandler.HandleDeveloperAPI)
	api.All("/v1/:app_slug", gatewayHandler.HandleDeveloperAPI)

	f.Use(gatewayHandler.HandleHostBasedAPI)

	f.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "not_found",
			"message": "route not found",
		})
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := cfg.Server.Host + ":" + "8080"
		log.Printf("Gateway starting on %s", addr)
		if err := f.Listen(addr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down gateway...")
	f.ShutdownWithTimeout(cfg.Server.ShutdownTimeout)
}
