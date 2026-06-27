package response

import (
	"github.com/gofiber/fiber/v2"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type SuccessBody struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type PaginatedBody struct {
	Data   any `json:"data"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(SuccessBody{
		Message: "success",
		Data:    data,
	})
}

func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(SuccessBody{
		Message: "created",
		Data:    data,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Error(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(ErrorBody{
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

func TooManyRequests(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusTooManyRequests, "rate_limited", message)
}

func InternalError(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusInternalServerError, "internal_error", message)
}

func Paginated(c *fiber.Ctx, data any, total, limit, offset int) error {
	c.Set("X-Total-Count", itoa(total))
	return c.JSON(PaginatedBody{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
