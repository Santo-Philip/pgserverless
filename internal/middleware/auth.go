package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nexbic/platform/config"
	"github.com/nexbic/platform/pkg/response"
)

type AuthMiddleware struct {
	cfg config.JWTConfig
	pool *pgxpool.Pool
}

func NewAuthMiddleware(cfg config.JWTConfig, pool *pgxpool.Pool) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg, pool: pool}
}

func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try JWT first
		tokenString, err := extractToken(c)
		if err == nil {
			claims, err := m.verifyToken(tokenString)
			if err == nil {
				sub, _ := claims.GetSubject()
				uid, err := uuid.Parse(sub)
				if err == nil {
					email, _ := claims["email"].(string)
					role, _ := claims["role"].(string)
					c.Locals("user_id", uid)
					c.Locals("email", email)
					c.Locals("role", role)
					return c.Next()
				}
			}
		}

		// Fallback: try API key via X-API-Key header
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			// Also check Authorization: Bearer nxc_*
			if tokenString != "" && strings.HasPrefix(tokenString, "nxc_") {
				apiKey = tokenString
			}
		}

		if apiKey != "" && m.pool != nil {
			hash := sha256.Sum256([]byte(apiKey))
			hashStr := hex.EncodeToString(hash[:])

			var userID uuid.UUID
			err := m.pool.QueryRow(context.Background(), `
				SELECT user_id FROM api_keys
				WHERE hash = $1 AND status = 'active'
				AND (expires_at IS NULL OR expires_at > NOW())`, hashStr).Scan(&userID)
			if err == nil {
				var role string
				m.pool.QueryRow(context.Background(), `SELECT role FROM users WHERE id = $1`, userID).Scan(&role)
				c.Locals("user_id", userID)
				c.Locals("role", role)
				return c.Next()
			}
		}

		return response.Unauthorized(c, "missing or invalid credentials")
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
