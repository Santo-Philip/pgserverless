package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/gateway/proxy"
	"github.com/nexbic/platform/gateway/service"
	mw "github.com/nexbic/platform/shared/middleware"
	"github.com/nexbic/platform/shared/utils"
)

type GatewayHandler struct {
	resolver *service.AppResolver
	proxy    *proxy.PostgRESTProxy
	authMW   *mw.AuthMiddleware
}

func NewGatewayHandler(resolver *service.AppResolver, proxy *proxy.PostgRESTProxy, authMW *mw.AuthMiddleware) *GatewayHandler {
	return &GatewayHandler{
		resolver: resolver,
		proxy:    proxy,
		authMW:   authMW,
	}
}

func (h *GatewayHandler) HandleDeveloperAPI(c *fiber.Ctx) error {
	appSlug := c.Params("app_slug")
	if appSlug == "" {
		return utils.NotFound(c, "app not found")
	}

	app, err := h.resolver.ResolveBySlug(c.Context(), appSlug)
	if err != nil {
		return utils.NotFound(c, "app not found or inactive")
	}

	c.Locals("app_id", app.ID.String())
	c.Locals("app_slug", appSlug)
	c.Locals("schema_name", app.SchemaName)

	return h.proxy.Forward(c, appSlug)
}

func (h *GatewayHandler) HandleHostBasedAPI(c *fiber.Ctx) error {
	host := c.Hostname()
	if host == "" {
		return c.Next()
	}

	if !strings.Contains(host, ".api.") && !strings.HasPrefix(host, "localhost") {
		return c.Next()
	}

	appSlug := extractSlugFromHost(host)
	if appSlug == "" {
		return c.Next()
	}

	app, err := h.resolver.ResolveBySlug(c.Context(), appSlug)
	if err != nil {
		return utils.NotFound(c, "app not found or inactive")
	}

	c.Locals("app_id", app.ID.String())
	c.Locals("app_slug", appSlug)
	c.Locals("schema_name", app.SchemaName)

	return h.proxy.Forward(c, appSlug)
}

func extractSlugFromHost(host string) string {
	host = strings.Split(host, ":")[0]
	parts := strings.SplitN(host, ".", 3)
	if len(parts) >= 3 {
		return parts[0]
	}
	return ""
}
