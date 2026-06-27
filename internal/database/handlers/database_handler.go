package handlers

import (
	"github.com/gofiber/fiber/v2"
	databaseDto "github.com/nexbic/platform/internal/database/dto"
	databaseService "github.com/nexbic/platform/internal/database/service"
	databaseValidation "github.com/nexbic/platform/internal/database/validation"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type DatabaseHandler struct {
	service *databaseService.DatabaseService
}

func NewDatabaseHandler(service *databaseService.DatabaseService) *DatabaseHandler {
	return &DatabaseHandler{service: service}
}

func (h *DatabaseHandler) Create(c *fiber.Ctx) error {
	var req databaseDto.CreateDatabaseRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := databaseValidation.ValidateCreate(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	dbEntry, err := h.service.Create(c.Context(), &req)
	if err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.Created(c, dbEntry)
}

func (h *DatabaseHandler) GetByID(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "database")
	if err != nil {
		return err
	}

	dbEntry, err := h.service.GetByID(c.Context(), id)
	if err != nil || dbEntry == nil {
		return response.NotFound(c, "database not found")
	}

	return response.OK(c, dbEntry)
}

func (h *DatabaseHandler) ListByProject(c *fiber.Ctx) error {
	projectID, err := helpers.ParseUUIDParam(c, "project_id", "project")
	if err != nil {
		return err
	}

	p := helpers.ParsePagination(c)
	databases, total, err := h.service.ListByProject(c.Context(), projectID, p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to list databases")
	}

	return response.Paginated(c, databases, total, p.Limit, p.Offset)
}

func (h *DatabaseHandler) Delete(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "database")
	if err != nil {
		return err
	}

	if err := h.service.Delete(c.Context(), id); err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.NoContent(c)
}

func (h *DatabaseHandler) RunSQL(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "database")
	if err != nil {
		return err
	}

	var req databaseDto.RunSQLRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := databaseValidation.ValidateRunSQL(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	results, err := h.service.RunSQL(c.Context(), id, req.Query)
	if err != nil {
		return response.BadRequest(c, "query failed: "+err.Error())
	}

	return response.OK(c, results)
}

func (h *DatabaseHandler) ListTables(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "database")
	if err != nil {
		return err
	}

	tables, err := h.service.ListTables(c.Context(), id)
	if err != nil {
		return response.InternalError(c, "failed to list tables")
	}

	return response.OK(c, tables)
}

func (h *DatabaseHandler) GetTableData(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "database")
	if err != nil {
		return err
	}

	table := c.Params("table")
	if table == "" {
		return response.BadRequest(c, "table name is required")
	}

	p := helpers.ParsePagination(c)
	rows, err := h.service.GetTableData(c.Context(), id, table, p.Limit, p.Offset)
	if err != nil {
		return response.BadRequest(c, "failed to query table: "+err.Error())
	}

	return response.OK(c, rows)
}

func (h *DatabaseHandler) CreateTable(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "database")
	if err != nil {
		return err
	}

	var req databaseDto.CreateTableRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := databaseValidation.ValidateCreateTable(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	if err := h.service.CreateTable(c.Context(), id, &req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, fiber.Map{"table": req.Name})
}

func (h *DatabaseHandler) AddColumn(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "database")
	if err != nil {
		return err
	}

	table := c.Params("table")
	if table == "" {
		return response.BadRequest(c, "table name is required")
	}

	var req databaseDto.AddColumnRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if v := databaseValidation.ValidateAddColumn(&req); v.HasErrors() {
		return response.BadRequest(c, v.Error())
	}

	if err := h.service.AddColumn(c.Context(), id, table, &req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, fiber.Map{"column": req.Name})
}

func (h *DatabaseHandler) InsertRow(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "database")
	if err != nil {
		return err
	}

	table := c.Params("table")
	if table == "" {
		return response.BadRequest(c, "table name is required")
	}

	var req databaseDto.InsertRowRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	row, err := h.service.InsertRow(c.Context(), id, table, &req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, row)
}

func (h *DatabaseHandler) UpdateRow(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "database")
	if err != nil {
		return err
	}

	table := c.Params("table")
	if table == "" {
		return response.BadRequest(c, "table name is required")
	}

	var req databaseDto.UpdateRowRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	rows, err := h.service.UpdateRow(c.Context(), id, table, &req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, rows)
}

func (h *DatabaseHandler) DeleteRow(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "database")
	if err != nil {
		return err
	}

	table := c.Params("table")
	if table == "" {
		return response.BadRequest(c, "table name is required")
	}

	var req databaseDto.DeleteRowRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	count, err := h.service.DeleteRow(c.Context(), id, table, &req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"deleted": count})
}

func (h *DatabaseHandler) ListExtensions(c *fiber.Ctx) error {
	extensions, err := h.service.ListExtensions(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to list extensions")
	}

	return response.OK(c, extensions)
}

func (h *DatabaseHandler) ToggleExtension(c *fiber.Ctx) error {
	var req databaseDto.ToggleExtensionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if err := h.service.ToggleExtension(c.Context(), &req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"extension": req.Name, "installed": req.Install})
}
