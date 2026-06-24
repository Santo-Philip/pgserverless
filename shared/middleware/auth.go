package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nexbic/platform/shared/config"
	"github.com/nexbic/platform/shared/utils"
)

type AuthMiddleware struct {
	cfg config.JWTConfig
}

func NewAuthMiddleware(cfg config.JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg}
}

func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := extractToken(c)
		if err != nil {
			return utils.Unauthorized(c, "missing or invalid authorization header")
		}

		claims, err := m.verifyToken(token)
		if err != nil {
			return utils.Unauthorized(c, "invalid or expired token")
		}

		c.Locals("user_id", claims["sub"])
		c.Locals("email", claims["email"])
		c.Locals("role", claims["role"])
		c.Locals("is_super_admin", claims["is_super_admin"])
		c.Locals("token_claims", claims)

		return c.Next()
	}
}

func (m *AuthMiddleware) RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		isAdmin, ok := c.Locals("is_super_admin").(bool)
		if !ok || !isAdmin {
			return utils.Forbidden(c, "admin access required")
		}
		return c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok {
			return utils.Forbidden(c, "no role found")
		}

		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}

		return utils.Forbidden(c, "insufficient permissions")
	}
}

func (m *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := extractToken(c)
		if err != nil {
			return c.Next()
		}

		claims, err := m.verifyToken(token)
		if err != nil {
			return c.Next()
		}

		c.Locals("user_id", claims["sub"])
		c.Locals("email", claims["email"])
		c.Locals("role", claims["role"])
		c.Locals("token_claims", claims)

		return c.Next()
	}
}

func (m *AuthMiddleware) verifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

func extractToken(c *fiber.Ctx) (string, error) {
	auth := c.Get("Authorization")
	if auth == "" {
		auth = c.Get("X-API-Key")
		if auth != "" {
			return auth, nil
		}
		return "", fiber.ErrUnauthorized
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", fiber.ErrUnauthorized
	}

	return parts[1], nil
}
