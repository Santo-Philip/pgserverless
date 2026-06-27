package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CORS(origins []string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     joinOrigins(origins),
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Authorization,Content-Type,X-API-Key",
		ExposeHeaders:    "X-Total-Count,X-RateLimit-Limit,X-RateLimit-Remaining",
		AllowCredentials: true,
		MaxAge:           300,
	})
}

func joinOrigins(origins []string) string {
	result := ""
	for i, o := range origins {
		if i > 0 {
			result += ", "
		}
		result += o
	}
	return result
}
