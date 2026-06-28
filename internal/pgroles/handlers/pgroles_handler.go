package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/pgroles/models"
	pgrolesService "github.com/nexbic/platform/internal/pgroles/service"
	"github.com/nexbic/platform/pkg/response"
)

type PgRolesHandler struct {
	service *pgrolesService.PgRolesService
}

func NewPgRolesHandler(service *pgrolesService.PgRolesService) *PgRolesHandler {
	return &PgRolesHandler{service: service}
}

func (h *PgRolesHandler) ListRoles(c *fiber.Ctx) error {
	roles, err := h.service.ListRoles(c.Context())
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.OK(c, roles)
}

func (h *PgRolesHandler) GetRole(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	role, err := h.service.GetRole(c.Context(), name)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	if role == nil {
		return response.NotFound(c, "role not found")
	}

	return response.OK(c, role)
}

func (h *PgRolesHandler) CreateRole(c *fiber.Ctx) error {
	var req models.CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "role name is required")
	}

	if err := h.service.CreateRole(c.Context(), &req); err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.Created(c, fiber.Map{"role": req.Name})
}

func (h *PgRolesHandler) AlterRole(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	var req models.AlterRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if err := h.service.AlterRole(c.Context(), name, &req); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"role": name, "altered": true})
}

func (h *PgRolesHandler) DropRole(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	if err := h.service.DropRole(c.Context(), name); err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.OK(c, fiber.Map{"role": name, "dropped": true})
}

func (h *PgRolesHandler) ResetPassword(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	var req models.PasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Password == "" {
		return response.BadRequest(c, "password is required")
	}

	if err := h.service.ResetPassword(c.Context(), name, req.Password); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"role": name, "password_reset": true})
}

func (h *PgRolesHandler) GrantDatabase(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	var req struct {
		Database       string `json:"database"`
		Permissions    string `json:"permissions"`
		PermissionType string `json:"permission_type"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Database == "" {
		return response.BadRequest(c, "database is required")
	}

	perms := req.Permissions
	if perms == "" {
		perms = req.PermissionType
	}
	if perms == "" {
		return response.BadRequest(c, "permissions or permission_type is required")
	}

	if err := h.service.GrantDatabase(c.Context(), name, req.Database, perms); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"role": name, "database": req.Database, "granted": true})
}

func (h *PgRolesHandler) GrantSchema(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	var req struct {
		Schema         string `json:"schema"`
		Permissions    string `json:"permissions"`
		PermissionType string `json:"permission_type"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Schema == "" {
		return response.BadRequest(c, "schema is required")
	}

	perms := req.Permissions
	if perms == "" {
		perms = req.PermissionType
	}
	if perms == "" {
		return response.BadRequest(c, "permissions or permission_type is required")
	}

	if err := h.service.GrantSchema(c.Context(), name, req.Schema, perms); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"role": name, "schema": req.Schema, "granted": true})
}

func (h *PgRolesHandler) GrantTable(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	var req struct {
		Schema         string `json:"schema"`
		Table          string `json:"table"`
		Permissions    string `json:"permissions"`
		PermissionType string `json:"permission_type"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	perms := req.Permissions
	if perms == "" {
		perms = req.PermissionType
	}
	if perms == "" {
		return response.BadRequest(c, "permissions or permission_type is required")
	}

	if err := h.service.GrantTable(c.Context(), name, req.Schema, req.Table, perms); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"role": name, "granted": true})
}

func (h *PgRolesHandler) RevokeDatabase(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	var req struct {
		Database       string `json:"database"`
		Permissions    string `json:"permissions"`
		PermissionType string `json:"permission_type"`
		Cascade        bool   `json:"cascade"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Database == "" {
		return response.BadRequest(c, "database is required")
	}

	perms := req.Permissions
	if perms == "" {
		perms = req.PermissionType
	}
	if perms == "" {
		return response.BadRequest(c, "permissions or permission_type is required")
	}

	if err := h.service.RevokeDatabase(c.Context(), name, req.Database, perms, req.Cascade); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"role": name, "database": req.Database, "revoked": true})
}

func (h *PgRolesHandler) RevokeSchema(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	var req struct {
		Schema         string `json:"schema"`
		Permissions    string `json:"permissions"`
		PermissionType string `json:"permission_type"`
		Cascade        bool   `json:"cascade"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Schema == "" {
		return response.BadRequest(c, "schema is required")
	}

	perms := req.Permissions
	if perms == "" {
		perms = req.PermissionType
	}
	if perms == "" {
		return response.BadRequest(c, "permissions or permission_type is required")
	}

	if err := h.service.RevokeSchema(c.Context(), name, req.Schema, perms, req.Cascade); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"role": name, "schema": req.Schema, "revoked": true})
}

func (h *PgRolesHandler) RevokeTable(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	var req struct {
		Schema         string `json:"schema"`
		Table          string `json:"table"`
		Permissions    string `json:"permissions"`
		PermissionType string `json:"permission_type"`
		Cascade        bool   `json:"cascade"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	perms := req.Permissions
	if perms == "" {
		perms = req.PermissionType
	}
	if perms == "" {
		return response.BadRequest(c, "permissions or permission_type is required")
	}

	if err := h.service.RevokeTable(c.Context(), name, req.Schema, req.Table, perms, req.Cascade); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.OK(c, fiber.Map{"role": name, "revoked": true})
}

func (h *PgRolesHandler) AddMember(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	var req models.MembershipRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.MemberOf == "" {
		return response.BadRequest(c, "member_of is required")
	}

	if err := h.service.AddMember(c.Context(), name, req.MemberOf); err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.OK(c, fiber.Map{"role": name, "member_of": req.MemberOf})
}

func (h *PgRolesHandler) RemoveMember(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	var req models.MembershipRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.MemberOf == "" {
		return response.BadRequest(c, "member_of is required")
	}

	if err := h.service.RemoveMember(c.Context(), name, req.MemberOf); err != nil {
		return response.Conflict(c, err.Error())
	}

	return response.OK(c, fiber.Map{"role": name, "removed_from": req.MemberOf})
}

func (h *PgRolesHandler) ListDatabasePrivileges(c *fiber.Ctx) error {
	databaseName := c.Query("database")
	if databaseName == "" {
		return response.BadRequest(c, "database query parameter is required")
	}

	privileges, err := h.service.ListDatabasePrivileges(c.Context(), databaseName)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, privileges)
}

func (h *PgRolesHandler) ListSchemaPrivileges(c *fiber.Ctx) error {
	schemaName := c.Query("schema")
	if schemaName == "" {
		return response.BadRequest(c, "schema query parameter is required")
	}

	privileges, err := h.service.ListSchemaPrivileges(c.Context(), schemaName)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, privileges)
}

func (h *PgRolesHandler) ListTablePrivileges(c *fiber.Ctx) error {
	schemaName := c.Query("schema")
	tableName := c.Query("table")

	privileges, err := h.service.ListTablePrivileges(c.Context(), schemaName, tableName)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, privileges)
}

func (h *PgRolesHandler) GetRoleMembers(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return response.BadRequest(c, "role name is required")
	}

	members, err := h.service.GetRoleMembers(c.Context(), name)
	if err != nil {
		return response.InternalError(c, err.Error())
	}

	return response.OK(c, members)
}
