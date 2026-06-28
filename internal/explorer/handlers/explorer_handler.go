package handlers

import (
	"github.com/gofiber/fiber/v2"
	explorerservice "github.com/nexbic/platform/internal/explorer/service"
	"github.com/nexbic/platform/pkg/response"
)

type ExplorerHandler struct {
	service *explorerservice.ExplorerService
}

func NewExplorerHandler(service *explorerservice.ExplorerService) *ExplorerHandler {
	return &ExplorerHandler{service: service}
}

func (h *ExplorerHandler) ListSchemas(c *fiber.Ctx) error {
	schemaFilter := c.Query("filter")
	schemas, err := h.service.ListSchemas(c.Context(), schemaFilter)
	if err != nil {
		return response.InternalError(c, "failed to list schemas")
	}
	return response.OK(c, schemas)
}

func (h *ExplorerHandler) ListTables(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	tables, err := h.service.ListTables(c.Context(), schema)
	if err != nil {
		return response.InternalError(c, "failed to list tables")
	}
	return response.OK(c, tables)
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

func (h *ExplorerHandler) ListViews(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	views, err := h.service.ListViews(c.Context(), schema)
	if err != nil {
		return response.InternalError(c, "failed to list views")
	}
	return response.OK(c, views)
}

func (h *ExplorerHandler) ListFunctions(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	functions, err := h.service.ListFunctions(c.Context(), schema)
	if err != nil {
		return response.InternalError(c, "failed to list functions")
	}
	return response.OK(c, functions)
}

func (h *ExplorerHandler) ListProcedures(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	procedures, err := h.service.ListProcedures(c.Context(), schema)
	if err != nil {
		return response.InternalError(c, "failed to list procedures")
	}
	return response.OK(c, procedures)
}

func (h *ExplorerHandler) ListTriggers(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	triggers, err := h.service.ListTriggers(c.Context(), schema)
	if err != nil {
		return response.InternalError(c, "failed to list triggers")
	}
	return response.OK(c, triggers)
}

func (h *ExplorerHandler) ListIndexes(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	indexes, err := h.service.ListIndexes(c.Context(), schema)
	if err != nil {
		return response.InternalError(c, "failed to list indexes")
	}
	return response.OK(c, indexes)
}

func (h *ExplorerHandler) ListConstraints(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	constraints, err := h.service.ListConstraints(c.Context(), schema)
	if err != nil {
		return response.InternalError(c, "failed to list constraints")
	}
	return response.OK(c, constraints)
}

func (h *ExplorerHandler) ListExtensions(c *fiber.Ctx) error {
	extensions, err := h.service.ListExtensions(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to list extensions")
	}
	return response.OK(c, extensions)
}

func (h *ExplorerHandler) ListSequences(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	sequences, err := h.service.ListSequences(c.Context(), schema)
	if err != nil {
		return response.InternalError(c, "failed to list sequences")
	}
	return response.OK(c, sequences)
}

func (h *ExplorerHandler) ListMaterializedViews(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	views, err := h.service.ListMaterializedViews(c.Context(), schema)
	if err != nil {
		return response.InternalError(c, "failed to list materialized views")
	}
	return response.OK(c, views)
}
