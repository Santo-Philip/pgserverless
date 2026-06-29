package handlers

import (
	"github.com/gofiber/fiber/v2"
	explorerservice "github.com/nexbic/platform/internal/database/explorer/service"
	"github.com/nexbic/platform/pkg/response"
)

type ExplorerHandler struct {
	service *explorerservice.ExplorerService
}

func NewExplorerHandler(service *explorerservice.ExplorerService) *ExplorerHandler {
	return &ExplorerHandler{service: service}
}

func (h *ExplorerHandler) ListSchemas(c *fiber.Ctx) error {
	schemas, err := h.service.ListSchemas(c.Context(), c.Query("filter"))
	if err != nil {
		return response.InternalError(c, "failed to list schemas")
	}
	return response.OK(c, schemas)
}

func (h *ExplorerHandler) ListResource(c *fiber.Ctx) error {
	schema := c.Params("schema")
	rtype := c.Params("resource")
	if schema == "" || rtype == "" {
		return response.BadRequest(c, "schema and resource type are required")
	}

	switch rtype {
	case "tables":
		v, err := h.service.ListTables(c.Context(), schema)
		if err != nil { return response.InternalError(c, err.Error()) }
		return response.OK(c, v)
	case "views":
		v, err := h.service.ListViews(c.Context(), schema)
		if err != nil { return response.InternalError(c, err.Error()) }
		return response.OK(c, v)
	case "functions":
		v, err := h.service.ListFunctions(c.Context(), schema)
		if err != nil { return response.InternalError(c, err.Error()) }
		return response.OK(c, v)
	case "procedures":
		v, err := h.service.ListProcedures(c.Context(), schema)
		if err != nil { return response.InternalError(c, err.Error()) }
		return response.OK(c, v)
	case "triggers":
		v, err := h.service.ListTriggers(c.Context(), schema)
		if err != nil { return response.InternalError(c, err.Error()) }
		return response.OK(c, v)
	case "indexes":
		v, err := h.service.ListIndexes(c.Context(), schema)
		if err != nil { return response.InternalError(c, err.Error()) }
		return response.OK(c, v)
	case "constraints":
		v, err := h.service.ListConstraints(c.Context(), schema)
		if err != nil { return response.InternalError(c, err.Error()) }
		return response.OK(c, v)
	case "sequences":
		v, err := h.service.ListSequences(c.Context(), schema)
		if err != nil { return response.InternalError(c, err.Error()) }
		return response.OK(c, v)
	case "materialized-views":
		v, err := h.service.ListMaterializedViews(c.Context(), schema)
		if err != nil { return response.InternalError(c, err.Error()) }
		return response.OK(c, v)
	case "extensions":
		v, err := h.service.ListExtensions(c.Context())
		if err != nil { return response.InternalError(c, err.Error()) }
		return response.OK(c, v)
	default:
		return response.BadRequest(c, "unknown resource type: "+rtype)
	}
}

func (h *ExplorerHandler) GetTableDetails(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return response.BadRequest(c, "schema and table are required")
	}

	details, err := h.service.GetTableDetails(c.Context(), schema, table)
	if err != nil {
		return response.InternalError(c, "failed to get table details")
	}
	if details == nil {
		return response.NotFound(c, "table not found")
	}
	return response.OK(c, details)
}

func (h *ExplorerHandler) ListExtensions(c *fiber.Ctx) error {
	extensions, err := h.service.ListExtensions(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to list extensions")
	}
	return response.OK(c, extensions)
}
