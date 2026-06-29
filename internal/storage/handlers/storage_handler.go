package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nexbic/platform/internal/storage/models"
	"github.com/nexbic/platform/internal/storage/service"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type StorageHandler struct {
	svc *service.StorageService
}

func NewStorageHandler(svc *service.StorageService) *StorageHandler {
	return &StorageHandler{svc: svc}
}

// ── Providers ──────────────────────────────────────────

func (h *StorageHandler) CreateProvider(c *fiber.Ctx) error {
	var req models.CreateProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" {
		return response.BadRequest(c, "name is required")
	}

	userID := helpers.GetUserID(c)
	p, err := h.svc.CreateProvider(c.Context(), &req, userID)
	if err != nil {
		return response.Conflict(c, err.Error())
	}
	return response.Created(c, p)
}

func (h *StorageHandler) ListProviders(c *fiber.Ctx) error {
	providers, err := h.svc.ListProviders(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to list providers")
	}
	return response.OK(c, providers)
}

func (h *StorageHandler) GetProvider(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "provider")
	if err != nil {
		return err
	}
	p, err := h.svc.GetProvider(c.Context(), id)
	if err != nil || p == nil {
		return response.NotFound(c, "provider not found")
	}
	return response.OK(c, p)
}

func (h *StorageHandler) UpdateProvider(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "provider")
	if err != nil {
		return err
	}
	var req models.UpdateProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if err := h.svc.UpdateProvider(c.Context(), id, &req); err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.OK(c, "provider updated")
}

func (h *StorageHandler) DeleteProvider(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "provider")
	if err != nil {
		return err
	}
	if err := h.svc.DeleteProvider(c.Context(), id); err != nil {
		return response.InternalError(c, "failed to delete provider")
	}
	return response.NoContent(c)
}

// ── Buckets ────────────────────────────────────────────

func (h *StorageHandler) CreateBucket(c *fiber.Ctx) error {
	var req models.CreateBucketRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if req.Name == "" || req.ProviderID == "" {
		return response.BadRequest(c, "name and provider_id are required")
	}

	userID := helpers.GetUserID(c)
	b, err := h.svc.CreateBucket(c.Context(), &req, userID)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}
	return response.Created(c, b)
}

func (h *StorageHandler) ListBuckets(c *fiber.Ctx) error {
	providerID, err := helpers.ParseUUIDParam(c, "provider_id", "provider")
	if err != nil {
		return err
	}
	buckets, err := h.svc.ListBuckets(c.Context(), providerID)
	if err != nil {
		return response.InternalError(c, "failed to list buckets")
	}
	return response.OK(c, buckets)
}

func (h *StorageHandler) GetBucket(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "bucket")
	if err != nil {
		return err
	}
	b, err := h.svc.GetBucket(c.Context(), id)
	if err != nil || b == nil {
		return response.NotFound(c, "bucket not found")
	}
	return response.OK(c, b)
}

func (h *StorageHandler) DeleteBucket(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "bucket")
	if err != nil {
		return err
	}
	if err := h.svc.DeleteBucket(c.Context(), id); err != nil {
		return response.InternalError(c, "failed to delete bucket")
	}
	return response.NoContent(c)
}

// ── Files ──────────────────────────────────────────────

func (h *StorageHandler) UploadFile(c *fiber.Ctx) error {
	bucketID, err := helpers.ParseUUIDParam(c, "bucket_id", "bucket")
	if err != nil {
		return err
	}

	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "file is required")
	}

	fh, err := file.Open()
	if err != nil {
		return response.InternalError(c, "failed to open file")
	}
	defer fh.Close()

	data := make([]byte, file.Size)
	if _, err := fh.Read(data); err != nil {
		return response.InternalError(c, "failed to read file")
	}

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	userID := helpers.GetUserID(c)
	f, err := h.svc.UploadFile(c.Context(), bucketID, file.Filename, mimeType, data, userID)
	if err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.Created(c, f)
}

func (h *StorageHandler) ListFiles(c *fiber.Ctx) error {
	bucketID, err := helpers.ParseUUIDParam(c, "bucket_id", "bucket")
	if err != nil {
		return err
	}
	p := helpers.ParsePagination(c)
	files, total, err := h.svc.ListFiles(c.Context(), bucketID, p.Limit, p.Offset)
	if err != nil {
		return response.InternalError(c, "failed to list files")
	}
	return response.Paginated(c, files, total, p.Limit, p.Offset)
}

func (h *StorageHandler) GetFile(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "file")
	if err != nil {
		return err
	}
	f, err := h.svc.GetFile(c.Context(), id)
	if err != nil || f == nil {
		return response.NotFound(c, "file not found")
	}
	return response.OK(c, f)
}

func (h *StorageHandler) DownloadFile(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "file")
	if err != nil {
		return err
	}
	data, mimeType, err := h.svc.ReadFile(c.Context(), id)
	if err != nil {
		return response.NotFound(c, err.Error())
	}
	c.Set("Content-Type", mimeType)
	c.Set("Content-Disposition", "attachment")
	return c.Send(data)
}

func (h *StorageHandler) DeleteFile(c *fiber.Ctx) error {
	id, err := helpers.ParseUUIDParam(c, "id", "file")
	if err != nil {
		return err
	}
	if err := h.svc.DeleteFile(c.Context(), id); err != nil {
		return response.InternalError(c, err.Error())
	}
	return response.NoContent(c)
}
