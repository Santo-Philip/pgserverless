package main

import (
	"context"
	"fmt"
	"log/slog"
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
	"github.com/redis/go-redis/v9"
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

	var rdb *redis.Client
	if addr := cfg.RedisAddr(); addr != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		if err := rdb.Ping(ctx).Err(); err != nil {
			slog.Warn("redis not available, using in-memory rate limiter", "error", err)
			rdb = nil
		} else {
			defer rdb.Close()
		}
	}

	tp, err := middleware.InitTracing(ctx, cfg.AppName+"-gateway", cfg.Tracing.OTLPEndpoint)
	if err != nil {
		slog.Warn("failed to initialize tracing", "error", err)
	}
	if tp != nil {
		defer func() { _ = tp.Shutdown(ctx) }()
	}

	appResolver := service.NewAppResolver(db)
	postgrestProxy := proxy.NewPostgRESTProxy(cfg.PostgREST.URL, cfg.PostgREST.Timeout)
	authMW := middleware.NewAuthMiddleware(cfg.JWT)

	gatewayHandler := handler.NewGatewayHandler(appResolver, postgrestProxy, authMW)

	f := fiber.New(fiber.Config{
		ReadTimeout:     cfg.Server.ReadTimeout,
		WriteTimeout:    cfg.Server.WriteTimeout,
		AppName:         cfg.AppName + "-gateway",
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

	f.Use(middleware.RateLimit(100, 1*time.Minute, rdb))

	f.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "gateway",
		})
	})

	f.Get("/ready", func(c *fiber.Ctx) error {
		if err := db.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":  "not_ready",
				"service": "gateway",
				"reason":  "database unavailable",
			})
		}
		return c.JSON(fiber.Map{
			"status":  "ready",
			"service": "gateway",
		})
	})

	if cfg.Monitoring.Enabled {
		f.Get(cfg.Monitoring.MetricPath, middleware.MetricsHandler())
	}

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
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		slog.Info("gateway starting", "address", addr)
		if err := f.Listen(addr); err != nil {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-quit
	slog.Info("shutting down gateway", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := f.ShutdownWithContext(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
