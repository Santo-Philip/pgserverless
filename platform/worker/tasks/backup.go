package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/nexbic/platform/shared/database"
)

const TypeBackup = "backup:app"

type BackupPayload struct {
	AppID  string `json:"app_id"`
	Schema string `json:"schema"`
}

func NewBackupTask(appID, schema string) (*asynq.Task, error) {
	payload, err := json.Marshal(BackupPayload{
		AppID:  appID,
		Schema: schema,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeBackup, payload), nil
}

func HandleBackup(ctx context.Context, t *asynq.Task, db *database.DB) error {
	var p BackupPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	log.Printf("starting backup for app: %s (schema: %s)", p.AppID, p.Schema)

	// TODO: implement actual pg_dump via exec
	log.Printf("backup completed for app: %s", p.AppID)
	return nil
}
