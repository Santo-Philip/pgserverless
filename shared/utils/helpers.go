package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ParseUUIDParam extracts and validates a UUID from a URL parameter.
// Returns the UUID and an error response if invalid.
func ParseUUIDParam(c *fiber.Ctx, param, entity string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(param))
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// GetUserID extracts the user ID from fiber context locals.
// Returns the UUID and true if found and valid, or uuid.Nil and false otherwise.
func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	val, ok := c.Locals("user_id").(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// Pagination holds parsed limit/offset from query parameters.
type Pagination struct {
	Limit  int
	Offset int
}

// ParsePagination extracts limit and offset from query parameters with defaults.
func ParsePagination(c *fiber.Ctx) Pagination {
	return Pagination{
		Limit:  c.QueryInt("limit", 20),
		Offset: c.QueryInt("offset", 0),
	}
}

// ErrNoRowsAsNil converts pgx.ErrNoRows to (nil, nil) for convenience.
// Use: if err := ErrNoRowsAsNil(err); err != nil { return nil, err }
func ErrNoRowsAsNil(err error) error {
	if err == pgx.ErrNoRows {
		return nil
	}
	return err
}
