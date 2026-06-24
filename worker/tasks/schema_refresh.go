package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/nexbic/platform/shared/database"
)

const TypeSchemaRefresh = "schema:refresh"

type SchemaRefreshPayload struct {
	AppID      string `json:"app_id"`
	SchemaName string `json:"schema_name"`
	Slug       string `json:"slug"`
}

func NewSchemaRefreshTask(appID, schemaName, slug string) (*asynq.Task, error) {
	payload, err := json.Marshal(SchemaRefreshPayload{
		AppID:      appID,
		SchemaName: schemaName,
		Slug:       slug,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSchemaRefresh, payload), nil
}

func HandleSchemaRefresh(ctx context.Context, t *asynq.Task, db *database.DB) error {
	var p SchemaRefreshPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	log.Printf("refreshing schema cache for app: %s (%s)", p.Slug, p.SchemaName)

	err := db.Exec(ctx, fmt.Sprintf(`NOTIFY pgrst, 'reload schema'`))
	if err != nil {
		return fmt.Errorf("notify postgrest: %w", err)
	}

	log.Printf("schema cache refreshed for %s", p.Slug)
	return nil
}
