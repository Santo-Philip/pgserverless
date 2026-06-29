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
	authRoutes "github.com/nexbic/platform/internal/auth/routes"
	filesRoutes "github.com/nexbic/platform/internal/files/routes"
	"github.com/nexbic/platform/internal/middleware"
	projectsRoutes "github.com/nexbic/platform/internal/projects/routes"
	walletRoutes "github.com/nexbic/platform/internal/wallet/routes"
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
		AppName:           a.Config.AppName + "-api",
		EnablePrintRoutes: false,
	})

	f.Use(middleware.RequestID())
	f.Use(middleware.Logger())
	f.Use(middleware.Recover())
	f.Use(middleware.CORS(a.Config.Server.CORSOrigins))
	f.Use(middleware.RateLimit(200, 1*time.Minute))

	f.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "service": "nexbic-api"})
	})
	f.Get("/ready", func(c *fiber.Ctx) error {
		if err := a.DB.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "not_ready", "reason": "database unavailable"})
		}
		return c.JSON(fiber.Map{"status": "ready"})
	})
	f.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/", fiber.StatusFound)
	})

	api := f.Group("/v1")

	authRoutes.RegisterAuthRoutes(api, a.AuthHandler, a.AuthMW)
	projectsRoutes.RegisterProjectsRoutes(api, a.ProjectsHandler, a.AuthMW)

	walletRoutes.RegisterWalletRoutes(api, a.WalletHandler, a.AuthMW)
	filesRoutes.RegisterFilesRoutes(api, a.FilesHandler, a.AuthMW)

	f.Use("/docs", filesystem.New(filesystem.Config{
		Root:  http.Dir("./docs"),
		Index: "index.html",
	}))

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
		slog.Info("API server starting", "address", addr)
		if err := f.Listen(addr); err != nil {
			slog.Error("API server error", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-quit
	slog.Info("shutting down API server", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.Config.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := f.ShutdownWithContext(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
