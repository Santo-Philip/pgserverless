package helpers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/pkg/response"
)

func ParseUUIDParam(c *fiber.Ctx, param, entity string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(param))
	if err != nil {
		return uuid.Nil, response.BadRequest(c, "invalid "+entity+" id")
	}
	return id, nil
}

func GetUserID(c *fiber.Ctx) uuid.UUID {
	if id, ok := c.Locals("user_id").(uuid.UUID); ok {
		return id
	}
	if id, ok := c.Locals("user_id").(string); ok {
		if uid, err := uuid.Parse(id); err == nil {
			return uid
		}
	}
	return uuid.Nil
}

func GetUserRole(c *fiber.Ctx) string {
	if role, ok := c.Locals("role").(string); ok {
		return role
	}
	return ""
}

type Pagination struct {
	Limit  int
	Offset int
}

func ParsePagination(c *fiber.Ctx) Pagination {
	limit := c.QueryInt("limit", 50)
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}
	return Pagination{Limit: limit, Offset: offset}
}

func ErrNoRowsAsNil(err error) error {
	if err == pgx.ErrNoRows {
		return nil
	}
	return err
}
