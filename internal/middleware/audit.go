package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nexbic/platform/pkg/database"
)

type AuditAction string

const (
	AuditCreate  AuditAction = "create"
	AuditRead    AuditAction = "read"
	AuditUpdate  AuditAction = "update"
	AuditDelete  AuditAction = "delete"
	AuditLogin   AuditAction = "login"
	AuditRevoke  AuditAction = "revoke"
	AuditRestore AuditAction = "restore"
)

func AuditLog(action AuditAction, resource string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		var actorID uuid.UUID
		if id, ok := c.Locals("user_id").(uuid.UUID); ok {
			actorID = id
		}

		db, ok := c.Locals("db").(*database.DB)
		if !ok || db == nil {
			slog.Warn("no db in context for audit log")
			return err
		}

		resourceID := c.Params("id")
		if resourceID == "" {
			resourceID = c.Params("keyId")
		}

		_, qErr := db.Pool.Exec(c.Context(), `
			INSERT INTO audit_logs (actor_id, action, resource, resource_id, ip_address, user_agent)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			actorID, string(action), resource, resourceID, c.IP(), c.Get("User-Agent"),
		)
		if qErr != nil {
			slog.Error("failed to write audit log", "error", qErr)
		}

		return err
	}
}

func AuditEntry(db *database.DB, actorID uuid.UUID, action AuditAction, resource, resourceID, ip, userAgent string) {
	_, err := db.Pool.Exec(nil, `
		INSERT INTO audit_logs (actor_id, action, resource, resource_id, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		actorID, string(action), resource, resourceID, ip, userAgent,
	)
	if err != nil {
		slog.Error("failed to write audit entry", "error", err)
	}
}

type AuditEntryData struct {
	ActorID    uuid.UUID
	Action     AuditAction
	Resource   string
	ResourceID string
	IP         string
	UserAgent  string
	Metadata   map[string]any
	CreatedAt  time.Time
}
