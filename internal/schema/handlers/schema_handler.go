package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/schema/models"
	"github.com/nexbic/platform/internal/schema/service"
	"github.com/nexbic/platform/pkg/response"
)

type SchemaHandler struct {
	svc *service.SchemaService
}

func NewSchemaHandler(svc *service.SchemaService) *SchemaHandler {
	return &SchemaHandler{svc: svc}
}

func (h *SchemaHandler) CreateSchema(c *fiber.Ctx) error {
	var req models.CreateSchemaRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	if err := h.svc.CreateSchema(c.Context(), req.Name); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, fiber.Map{"name": req.Name})
}

func (h *SchemaHandler) DropSchema(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	cascade := c.Query("cascade") == "true"
	if err := h.svc.DropSchema(c.Context(), schema, cascade); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "schema dropped"})
}

func (h *SchemaHandler) CreateTable(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	var req models.CreateTableRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if msg := req.Validate(); msg != "" {
		return response.BadRequest(c, msg)
	}
	if err := h.svc.CreateTable(c.Context(), schema, req); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, fiber.Map{"table": req.Name})
}

func (h *SchemaHandler) DropTable(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return response.BadRequest(c, "schema and table are required")
	}
	cascade := c.Query("cascade") == "true"
	if err := h.svc.DropTable(c.Context(), schema, table, cascade); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "table dropped"})
}

func (h *SchemaHandler) AddColumn(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return response.BadRequest(c, "schema and table are required")
	}
	var req models.AddColumnRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	if req.Type == "" {
		return response.BadRequest(c, "type is required")
	}
	if err := h.svc.AddColumn(c.Context(), schema, table, req); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, fiber.Map{"column": req.Name})
}

func (h *SchemaHandler) DropColumn(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	column := c.Params("column")
	if schema == "" || table == "" || column == "" {
		return response.BadRequest(c, "schema, table, and column are required")
	}
	if err := h.svc.DropColumn(c.Context(), schema, table, column); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "column dropped"})
}

func (h *SchemaHandler) AlterColumn(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	column := c.Params("column")
	if schema == "" || table == "" || column == "" {
		return response.BadRequest(c, "schema, table, and column are required")
	}
	var req models.AlterColumnRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.AlterColumn(c.Context(), schema, table, column, req); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "column altered"})
}

func (h *SchemaHandler) AddConstraint(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return response.BadRequest(c, "schema and table are required")
	}
	var req models.AddConstraintRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	if req.Type == "" {
		return response.BadRequest(c, "type is required")
	}
	if err := h.svc.AddConstraint(c.Context(), schema, table, req); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, fiber.Map{"constraint": req.Name})
}

func (h *SchemaHandler) DropConstraint(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	constraint := c.Params("constraint")
	if schema == "" || table == "" || constraint == "" {
		return response.BadRequest(c, "schema, table, and constraint are required")
	}
	if err := h.svc.DropConstraint(c.Context(), schema, table, constraint); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "constraint dropped"})
}

func (h *SchemaHandler) CreateIndex(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	var req models.CreateIndexRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	if req.Table == "" {
		return response.BadRequest(c, "table is required")
	}
	if len(req.Columns) == 0 {
		return response.BadRequest(c, "at least one column is required")
	}
	if err := h.svc.CreateIndex(c.Context(), schema, req); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, fiber.Map{"index": req.Name})
}

func (h *SchemaHandler) DropIndex(c *fiber.Ctx) error {
	schema := c.Params("schema")
	name := c.Params("name")
	if schema == "" || name == "" {
		return response.BadRequest(c, "schema and index name are required")
	}
	if err := h.svc.DropIndex(c.Context(), schema, name); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "index dropped"})
}

func (h *SchemaHandler) CreateSequence(c *fiber.Ctx) error {
	schema := c.Params("schema")
	if schema == "" {
		return response.BadRequest(c, "schema is required")
	}
	var req models.CreateSequenceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}
	if err := h.svc.CreateSequence(c.Context(), schema, req); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, fiber.Map{"sequence": req.Name})
}

func (h *SchemaHandler) DropSequence(c *fiber.Ctx) error {
	schema := c.Params("schema")
	name := c.Params("name")
	if schema == "" || name == "" {
		return response.BadRequest(c, "schema and sequence name are required")
	}
	if err := h.svc.DropSequence(c.Context(), schema, name); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "sequence dropped"})
}

func (h *SchemaHandler) AlterSequence(c *fiber.Ctx) error {
	schema := c.Params("schema")
	name := c.Params("name")
	if schema == "" || name == "" {
		return response.BadRequest(c, "schema and sequence name are required")
	}
	var req models.AlterSequenceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.AlterSequence(c.Context(), schema, name, req); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, fiber.Map{"message": "sequence altered"})
}

func (h *SchemaHandler) GetTableDDL(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return response.BadRequest(c, "schema and table are required")
	}
	ddl, err := h.svc.GetTableDDL(c.Context(), schema, table)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, models.DDLResponse{DDL: ddl})
}
