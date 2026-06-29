package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/database/storage/handlers"
	"github.com/nexbic/platform/internal/middleware"
)

func RegisterStorageRoutes(router fiber.Router, handler *handlers.StorageHandler, authMW *middleware.AuthMiddleware) {
	s := router.Group("/storage")

	// Providers
	s.Get("/providers", authMW.RequireAuth(), handler.ListProviders)
	s.Post("/providers", authMW.RequireAuth(), authMW.RequireRole("super_admin", "dba"), handler.CreateProvider)
	s.Get("/providers/:id", authMW.RequireAuth(), handler.GetProvider)
	s.Patch("/providers/:id", authMW.RequireAuth(), authMW.RequireRole("super_admin", "dba"), handler.UpdateProvider)
	s.Delete("/providers/:id", authMW.RequireAuth(), authMW.RequireRole("super_admin"), handler.DeleteProvider)

	// Buckets
	s.Get("/providers/:provider_id/buckets", authMW.RequireAuth(), handler.ListBuckets)
	s.Post("/buckets", authMW.RequireAuth(), handler.CreateBucket)
	s.Get("/buckets/:id", authMW.RequireAuth(), handler.GetBucket)
	s.Delete("/buckets/:id", authMW.RequireAuth(), handler.DeleteBucket)

	// Files
	s.Post("/buckets/:bucket_id/files", authMW.RequireAuth(), handler.UploadFile)
	s.Get("/buckets/:bucket_id/files", authMW.RequireAuth(), handler.ListFiles)
	s.Get("/files/:id", authMW.RequireAuth(), handler.GetFile)
	s.Get("/files/:id/download", authMW.RequireAuth(), handler.DownloadFile)
	s.Delete("/files/:id", authMW.RequireAuth(), handler.DeleteFile)
}
