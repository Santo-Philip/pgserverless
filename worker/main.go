package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/nexbic/platform/shared/config"
	"github.com/nexbic/platform/shared/database"
	"github.com/nexbic/platform/shared/middleware"
	"github.com/nexbic/platform/worker/tasks"
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

	tp, err := middleware.InitTracing(ctx, cfg.AppName+"-worker", cfg.Tracing.OTLPEndpoint)
	if err != nil {
		slog.Warn("failed to initialize tracing", "error", err)
	}
	if tp != nil {
		defer func() { _ = tp.Shutdown(ctx) }()
	}

	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.AsynqAddr(),
			Password: cfg.Asynq.Password,
			DB:       cfg.Asynq.DB,
		},
		asynq.Config{
			Concurrency: cfg.Asynq.Concurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			Logger: asynqLogger{},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeSchemaRefresh, func(ctx context.Context, t *asynq.Task) error {
		return tasks.HandleSchemaRefresh(ctx, t, db)
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("worker starting", "concurrency", cfg.Asynq.Concurrency)
		if err := asynqServer.Run(mux); err != nil {
			slog.Error("worker error", "error", err)
			os.Exit(1)
		}
	}()

	sig := <-quit
	slog.Info("shutting down worker", "signal", sig)

	asynqServer.Shutdown()
	slog.Info("worker stopped")
}

type asynqLogger struct{}

func (l asynqLogger) Debug(args ...interface{}) {
	slog.Debug("asynq", "args", fmt.Sprint(args...))
}

func (l asynqLogger) Info(args ...interface{}) {
	slog.Info("asynq", "args", fmt.Sprint(args...))
}

func (l asynqLogger) Warn(args ...interface{}) {
	slog.Warn("asynq", "args", fmt.Sprint(args...))
}

func (l asynqLogger) Error(args ...interface{}) {
	slog.Error("asynq", "args", fmt.Sprint(args...))
}

func (l asynqLogger) Fatal(args ...interface{}) {
	slog.Error("asynq fatal", "args", fmt.Sprint(args...))
	os.Exit(1)
}
