package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nexbic/platform/pkg/helpers"
	"github.com/nexbic/platform/pkg/response"
)

type ProjectsHandler struct {
	pool *pgxpool.Pool
}

func NewProjectsHandler(pool *pgxpool.Pool) *ProjectsHandler {
	return &ProjectsHandler{pool: pool}
}

func slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

func (h *ProjectsHandler) List(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	role := helpers.GetUserRole(c)

	query := "SELECT id, name, slug, description, created_by, created_at, updated_at FROM projects"
	var args []any

	if role == "super_admin" {
		query += " ORDER BY created_at DESC"
	} else {
		query += " WHERE created_by = $1 ORDER BY created_at DESC"
		args = append(args, userID)
	}

	rows, err := h.pool.Query(c.Context(), query, args...)
	if err != nil {
		return response.BadRequest(c, "query failed: "+err.Error())
	}
	defer rows.Close()

	type Project struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Slug        string    `json:"slug"`
		Description string    `json:"description"`
		CreatedBy   uuid.UUID `json:"created_by"`
		CreatedAt   string    `json:"created_at"`
		UpdatedAt   string    `json:"updated_at"`
	}

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return response.BadRequest(c, "scan failed: "+err.Error())
		}
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []Project{}
	}

	return response.OK(c, projects)
}

func (h *ProjectsHandler) Create(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	var body struct {
		Name        string `json:"name"`
		Slug        string `json:"slug,omitempty"`
		Description string `json:"description,omitempty"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if body.Name == "" {
		return response.BadRequest(c, "name is required")
	}

	slug := body.Slug
	if slug == "" {
		slug = slugify(body.Name)
	}
	desc := body.Description

	var p struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Slug        string    `json:"slug"`
		Description string    `json:"description"`
		CreatedBy   uuid.UUID `json:"created_by"`
		CreatedAt   string    `json:"created_at"`
		UpdatedAt   string    `json:"updated_at"`
	}

	err := h.pool.QueryRow(c.Context(),
		`INSERT INTO projects (name, slug, description, created_by) VALUES ($1, $2, $3, $4)
		 RETURNING id, name, slug, description, created_by, created_at, updated_at`,
		body.Name, slug, desc, userID).
		Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return response.Conflict(c, "a project with this slug already exists")
		}
		return response.BadRequest(c, "create failed: "+err.Error())
	}

	return response.Created(c, p)
}

func (h *ProjectsHandler) Get(c *fiber.Ctx) error {
	pid, err := helpers.ParseUUIDParam(c, "projectId", "project")
	if err != nil {
		return err
	}

	userID := helpers.GetUserID(c)
	role := helpers.GetUserRole(c)

	var p struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Slug        string    `json:"slug"`
		Description string    `json:"description"`
		CreatedBy   uuid.UUID `json:"created_by"`
		CreatedAt   string    `json:"created_at"`
		UpdatedAt   string    `json:"updated_at"`
	}

	err = h.pool.QueryRow(c.Context(),
		`SELECT id, name, slug, description, created_by, created_at, updated_at
		 FROM projects WHERE id = $1`, pid).
		Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return response.NotFound(c, "project not found")
	}

	if p.CreatedBy != userID && role != "super_admin" {
		return response.Forbidden(c, "access denied")
	}

	return response.OK(c, p)
}

func (h *ProjectsHandler) Update(c *fiber.Ctx) error {
	pid, err := helpers.ParseUUIDParam(c, "projectId", "project")
	if err != nil {
		return err
	}
	userID := helpers.GetUserID(c)
	role := helpers.GetUserRole(c)

	var body struct {
		Name        string `json:"name,omitempty"`
		Slug        string `json:"slug,omitempty"`
		Description string `json:"description,omitempty"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	sets := []string{}
	args := []any{}
	idx := 1

	if body.Name != "" {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, body.Name)
		idx++
	}
	if body.Slug != "" {
		sets = append(sets, fmt.Sprintf("slug = $%d", idx))
		args = append(args, body.Slug)
		idx++
	}
	if body.Description != "" {
		sets = append(sets, fmt.Sprintf("description = $%d", idx))
		args = append(args, body.Description)
		idx++
	}

	if len(sets) == 0 {
		return response.BadRequest(c, "nothing to update")
	}

	ownerClause := fmt.Sprintf("id = $%d", idx)
	args = append(args, pid)
	idx++
	if role != "super_admin" {
		ownerClause += fmt.Sprintf(" AND created_by = $%d", idx)
		args = append(args, userID)
	}

	query := fmt.Sprintf("UPDATE projects SET %s WHERE %s RETURNING id, name, slug, description, created_by, created_at, updated_at",
		strings.Join(sets, ", "), ownerClause)

	var p struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Slug        string    `json:"slug"`
		Description string    `json:"description"`
		CreatedBy   uuid.UUID `json:"created_by"`
		CreatedAt   string    `json:"created_at"`
		UpdatedAt   string    `json:"updated_at"`
	}

	err = h.pool.QueryRow(c.Context(), query, args...).
		Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return response.NotFound(c, "project not found or access denied")
	}

	return response.OK(c, p)
}

func (h *ProjectsHandler) Delete(c *fiber.Ctx) error {
	pid, err := helpers.ParseUUIDParam(c, "projectId", "project")
	if err != nil {
		return err
	}
	userID := helpers.GetUserID(c)
	role := helpers.GetUserRole(c)

	query := "DELETE FROM projects WHERE id = $1"
	args := []any{pid}
	if role != "super_admin" {
		query += " AND created_by = $2"
		args = append(args, userID)
	}

	ct, err := h.pool.Exec(c.Context(), query, args...)
	if err != nil {
		return response.BadRequest(c, "delete failed: "+err.Error())
	}
	if ct.RowsAffected() == 0 {
		return response.NotFound(c, "project not found or access denied")
	}

	return response.OK(c, fiber.Map{"deleted": true})
}
