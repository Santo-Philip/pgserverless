package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func init() {
	level := slog.LevelInfo
	if l := os.Getenv("LOG_LEVEL"); l != "" {
		switch l {
		case "debug":
			level = slog.LevelDebug
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
}

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

type LogCtx string

const LogCtxKey LogCtx = "logger"

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()
		requestID := c.Locals("request_id")

		attrs := []slog.Attr{
			slog.String("request_id", toString(requestID)),
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("duration", duration),
			slog.String("ip", c.IP()),
		}

		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		slog.LogAttrs(c.Context(), level, "request", attrs...)

		return err
	}
}

func Recover() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered",
					"panic", r,
					"request_id", c.Locals("request_id"),
				)
				err = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    "internal_error",
					"message": "an unexpected error occurred",
				})
			}
		}()
		return c.Next()
	}
}

func ErrorLogMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			slog.Warn("handler error",
				"request_id", c.Locals("request_id"),
				"error", err.Error(),
			)
		}
		return err
	}
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
