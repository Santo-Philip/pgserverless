package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/database/tables/models"
	"github.com/nexbic/platform/internal/database/tables/service"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type TablesHandler struct {
	svc *service.TablesService
}

func NewTablesHandler(svc *service.TablesService) *TablesHandler {
	return &TablesHandler{svc: svc}
}

func (h *TablesHandler) Query(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return response.BadRequest(c, "schema and table are required")
	}

	p := helpers.ParsePagination(c)
	sort := c.Query("sort")
	order := c.Query("order")

	var filters []models.Filter
	if f := c.Query("filters"); f != "" {
		if err := json.Unmarshal([]byte(f), &filters); err != nil {
			return response.BadRequest(c, "invalid filters JSON")
		}
	}

	search := c.Query("search")
	var searchCols []string
	if sc := c.Query("search_columns"); sc != "" {
		if err := json.Unmarshal([]byte(sc), &searchCols); err != nil {
			searchCols = nil
		}
	}

	if search != "" && len(searchCols) > 0 {
		f := models.Filter{
			Operator: "ilike",
		}
		for _, col := range searchCols {
			f.Column = col
			f.Value = "%" + search + "%"
			filters = append(filters, f)
		}
	}

	rows, total, err := h.svc.QueryTable(c.Context(), schema, table, p.Limit, p.Offset, sort, order, filters)
	if err != nil {
		return response.BadRequest(c, "query failed: "+err.Error())
	}

	return response.Paginated(c, rows, total, p.Limit, p.Offset)
}

func (h *TablesHandler) Insert(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return response.BadRequest(c, "schema and table are required")
	}

	var req models.InsertRowRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if len(req.Data) == 0 {
		return response.BadRequest(c, "data is required")
	}

	row, err := h.svc.InsertRow(c.Context(), schema, table, req.Data)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, row)
}

func (h *TablesHandler) Update(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return response.BadRequest(c, "schema and table are required")
	}

	var req models.UpdateRowRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if len(req.Data) == 0 {
		return response.BadRequest(c, "data is required")
	}
	if len(req.Where) == 0 {
		return response.BadRequest(c, "where condition is required")
	}

	rows, err := h.svc.UpdateRow(c.Context(), schema, table, req.Data, req.Where)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, rows)
}

func (h *TablesHandler) Delete(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return response.BadRequest(c, "schema and table are required")
	}

	var req models.DeleteRowRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if len(req.Where) == 0 {
		return response.BadRequest(c, "where condition is required")
	}

	count, err := h.svc.DeleteRow(c.Context(), schema, table, req.Where)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"deleted": count})
}

func (h *TablesHandler) BulkInsert(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return response.BadRequest(c, "schema and table are required")
	}

	var req models.BulkInsertRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if len(req.Rows) == 0 {
		return response.BadRequest(c, "rows is required")
	}

	results, err := h.svc.BulkInsert(c.Context(), schema, table, req.Rows)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, results)
}

func (h *TablesHandler) BulkDelete(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return response.BadRequest(c, "schema and table are required")
	}

	var req models.BulkDeleteRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if len(req.IDs) == 0 {
		return response.BadRequest(c, "ids is required")
	}
	if req.IDColumn == "" {
		return response.BadRequest(c, "id_column is required")
	}

	count, err := h.svc.BulkDelete(c.Context(), schema, table, req.IDs, req.IDColumn)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"deleted": count})
}

func (h *TablesHandler) Search(c *fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	if schema == "" || table == "" {
		return response.BadRequest(c, "schema and table are required")
	}

	var req models.SearchRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Search == "" {
		return response.BadRequest(c, "search term is required")
	}
	if len(req.Columns) == 0 {
		return response.BadRequest(c, "columns are required")
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}

	rows, err := h.svc.SearchTable(c.Context(), schema, table, req.Search, req.Columns, req.Limit)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, rows)
}
