package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuditAction string

const (
	AuditCreate    AuditAction = "create"
	AuditRead      AuditAction = "read"
	AuditUpdate    AuditAction = "update"
	AuditDelete    AuditAction = "delete"
	AuditLogin     AuditAction = "login"
	AuditSuspend   AuditAction = "suspend"
	AuditActivate  AuditAction = "activate"
	AuditVerify    AuditAction = "verify"
)

func AuditLog(action AuditAction, resource string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		userID, _ := c.Locals("user_id").(string)
		requestID, _ := c.Locals("request_id").(string)

		attrs := []slog.Attr{
			slog.String("event", "audit"),
			slog.String("action", string(action)),
			slog.String("resource", resource),
			slog.String("user_id", userID),
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Time("timestamp", time.Now()),
		}

		if id := c.Params("id"); id != "" {
			if _, err := uuid.Parse(id); err == nil {
				attrs = append(attrs, slog.String("resource_id", id))
			}
		}

		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		slog.LogAttrs(c.Context(), slog.LevelInfo, "audit", attrs...)

		return err
	}
}

func Audit(actorID uuid.UUID, action AuditAction, resource string, resourceID string, metadata map[string]interface{}) {
	attrs := []slog.Attr{
		slog.String("event", "audit"),
		slog.String("action", string(action)),
		slog.String("resource", resource),
		slog.String("actor_id", actorID.String()),
		slog.Time("timestamp", time.Now()),
	}
	if resourceID != "" {
		attrs = append(attrs, slog.String("resource_id", resourceID))
	}
	if metadata != nil {
		attrs = append(attrs, slog.Any("metadata", metadata))
	}
	slog.LogAttrs(nil, slog.LevelInfo, "audit", attrs...)
}
