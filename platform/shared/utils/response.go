package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/shared/models"
)

func OK(c *fiber.Ctx, data interface{}) error {
	return c.JSON(models.SuccessResponse{
		Message: "success",
		Data:    data,
	})
}

func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(models.SuccessResponse{
		Message: "created",
		Data:    data,
	})
}

func Error(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(models.ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func BadRequest(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusBadRequest, "bad_request", message)
}

func Unauthorized(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusUnauthorized, "unauthorized", message)
}

func Forbidden(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusForbidden, "forbidden", message)
}

func NotFound(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusNotFound, "not_found", message)
}

func Conflict(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusConflict, "conflict", message)
}

func InternalError(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusInternalServerError, "internal_error", message)
}

func Paginated(c *fiber.Ctx, data interface{}, total, limit, offset int) error {
	c.Set("X-Total-Count", itoa(total))
	c.Set("X-Page-Limit", itoa(limit))
	c.Set("X-Page-Offset", itoa(offset))
	return c.JSON(models.PaginatedResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}
