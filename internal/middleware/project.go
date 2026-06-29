package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type ProjectInfo struct {
	ID        uuid.UUID
	Name      string
	CreatedBy uuid.UUID
}

func ProjectGuard(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectID := c.Params("projectId")
		if projectID == "" {
			return response.BadRequest(c, "project ID required")
		}

		pid, err := uuid.Parse(projectID)
		if err != nil {
			return response.BadRequest(c, "invalid project ID")
		}

		userID := helpers.GetUserID(c)
		if userID == uuid.Nil {
			return response.Unauthorized(c, "authentication required")
		}

		var info ProjectInfo
		err = pool.QueryRow(c.Context(),
			"SELECT id, name, created_by FROM projects WHERE id = $1", pid).
			Scan(&info.ID, &info.Name, &info.CreatedBy)
		if err != nil {
			return response.NotFound(c, "project not found")
		}

		role := helpers.GetUserRole(c)
		if info.CreatedBy != userID && role != "super_admin" {
			return response.Forbidden(c, "access denied to this project")
		}

		c.Locals("project_id", info.ID)
		c.Locals("project_name", info.Name)
		return c.Next()
	}
}
