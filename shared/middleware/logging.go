package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set("X-Request-ID", id)
		c.Locals("request_id", id)
		return c.Next()
	}
}

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()
		requestID := c.Locals("request_id")

		log.Printf("[%s] %s %s %d %v %s",
			requestID, method, path, status, duration, c.IP())

		return err
	}
}

func Recover() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC: %v", r)
				err = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    "internal_error",
					"message": "an unexpected error occurred",
				})
			}
		}()
		return c.Next()
	}
}
