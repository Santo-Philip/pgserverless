package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nexbic/platform/config"
	"github.com/nexbic/platform/pkg/response"
)

type AuthMiddleware struct {
	cfg config.JWTConfig
}

func NewAuthMiddleware(cfg config.JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg}
}

func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString, err := extractToken(c)
		if err != nil {
			return response.Unauthorized(c, "missing or malformed authorization header")
		}

		claims, err := m.verifyToken(tokenString)
		if err != nil {
			return response.Unauthorized(c, "invalid or expired token")
		}

		sub, _ := claims.GetSubject()
		uid, err := uuid.Parse(sub)
		if err != nil {
			return response.Unauthorized(c, "invalid token payload")
		}

		email, _ := claims["email"].(string)
		role, _ := claims["role"].(string)

		c.Locals("user_id", uid)
		c.Locals("email", email)
		c.Locals("role", role)

		return c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok {
			return response.Forbidden(c, "access denied")
		}
		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}
		return response.Forbidden(c, "insufficient permissions")
	}
}

func (m *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString, err := extractToken(c)
		if err != nil {
			return c.Next()
		}

		claims, err := m.verifyToken(tokenString)
		if err != nil {
			return c.Next()
		}

		sub, _ := claims.GetSubject()
		if uid, err := uuid.Parse(sub); err == nil {
			c.Locals("user_id", uid)
		}
		if email, _ := claims["email"].(string); email != "" {
			c.Locals("email", email)
		}
		if role, _ := claims["role"].(string); role != "" {
			c.Locals("role", role)
		}

		return c.Next()
	}
}

func (m *AuthMiddleware) verifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		aud, _ := claims.GetAudience()
		if len(aud) > 0 && aud[0] != m.cfg.Audience {
			return nil, jwt.ErrTokenInvalidAudience
		}
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

func extractToken(c *fiber.Ctx) (string, error) {
	auth := c.Get("Authorization")
	if auth == "" {
		return "", fiber.ErrUnauthorized
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fiber.ErrUnauthorized
	}

	return parts[1], nil
}
