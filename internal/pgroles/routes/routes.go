package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/pgroles/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterPgRolesRoutes(router fiber.Router, handler *handlers.PgRolesHandler, authMW *middleware.AuthMiddleware) {
	roles := router.Group("/roles", authMW.RequireAuth(), authMW.RequireRole("super_admin", "dba"))

	roles.Get("/", handler.ListRoles)
	roles.Post("/", handler.CreateRole)
	roles.Get("/:name", handler.GetRole)
	roles.Patch("/:name", handler.AlterRole)
	roles.Delete("/:name", handler.DropRole)

	roles.Post("/:name/password", handler.ResetPassword)
	roles.Post("/:name/grant-database", handler.GrantDatabase)
	roles.Post("/:name/grant-schema", handler.GrantSchema)
	roles.Post("/:name/grant-table", handler.GrantTable)
	roles.Post("/:name/revoke-database", handler.RevokeDatabase)
	roles.Post("/:name/revoke-schema", handler.RevokeSchema)
	roles.Post("/:name/revoke-table", handler.RevokeTable)
	roles.Post("/:name/add-member", handler.AddMember)
	roles.Post("/:name/remove-member", handler.RemoveMember)

	roles.Get("/privileges/databases", handler.ListDatabasePrivileges)
	roles.Get("/privileges/schemas", handler.ListSchemaPrivileges)
	roles.Get("/privileges/tables", handler.ListTablePrivileges)
	roles.Get("/:name/members", handler.GetRoleMembers)
}
