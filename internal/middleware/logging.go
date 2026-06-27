package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LogCtx string

const LogCtxKey LogCtx = "logger"

func init() {
	level := slog.LevelInfo
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	if os.Getenv("APP_ENV") == "development" {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)))
	} else {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))
	}
}

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rid := c.Get("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set("X-Request-ID", rid)
		c.Locals("request_id", rid)
		return c.Next()
	}
}

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()

		attrs := []slog.Attr{
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", c.IP()),
			slog.String("request_id", toString(c.Locals("request_id"))),
		}

		if userID := c.Locals("user_id"); userID != nil {
			attrs = append(attrs, slog.String("user_id", toString(userID)))
		}

		if status >= 500 {
			slog.LogAttrs(nil, slog.LevelError, "request failed", attrs...)
		} else if status >= 400 {
			slog.LogAttrs(nil, slog.LevelWarn, "request warning", attrs...)
		} else {
			slog.LogAttrs(nil, slog.LevelInfo, "request", attrs...)
		}

		return err
	}
}

func Recover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered", "error", toString(r))
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    "internal_error",
					"message": "an unexpected error occurred",
				})
			}
		}()
		return c.Next()
	}
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
