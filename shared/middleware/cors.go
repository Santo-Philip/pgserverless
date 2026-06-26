package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CORS(origins []string) fiber.Handler {
	allowCredentials := true
	allowOrigins := joinOrigins(origins)

	if allowOrigins == "*" || allowOrigins == "" {
		allowOrigins = "http://localhost:5173"
		allowCredentials = true
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowHeaders:     "Authorization, X-API-Key, Content-Type, Accept, Accept-Profile, Content-Profile, Prefer, X-Request-ID",
		ExposeHeaders:    "Content-Range, Range, Preference-Applied, X-Request-ID, X-Total-Count, X-RateLimit-Limit, X-RateLimit-Remaining",
		AllowCredentials: allowCredentials,
		MaxAge:           86400,
	})
}

func joinOrigins(origins []string) string {
	if len(origins) == 0 {
		return ""
	}
	result := origins[0]
	for _, o := range origins[1:] {
		result += ", " + o
	}
	return result
}
