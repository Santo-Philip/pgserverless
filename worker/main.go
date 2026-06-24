package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/nexbic/platform/shared/config"
	"github.com/nexbic/platform/shared/database"
	"github.com/nexbic/platform/worker/tasks"
)

func main() {
	cfg := config.Load()

	db, err := database.New(context.Background(), cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

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
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeSchemaRefresh, func(ctx context.Context, t *asynq.Task) error {
		return tasks.HandleSchemaRefresh(ctx, t, db)
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Worker starting with concurrency %d", cfg.Asynq.Concurrency)
		if err := asynqServer.Run(mux); err != nil {
			log.Fatalf("worker error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down worker...")
	asynqServer.Shutdown()
}
