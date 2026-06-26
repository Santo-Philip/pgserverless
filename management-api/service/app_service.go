package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/management-api/repository"
	"github.com/nexbic/platform/shared/database"
	"github.com/nexbic/platform/shared/models"
	"github.com/nexbic/platform/shared/utils"
)

type AppService struct {
	appRepo        *repository.AppRepository
	keyRepo        *repository.APIKeyRepository
	userRepo       *repository.UserRepository
	db             *database.DB
	jwtSecret      string
	postgrestAdmin string
}

func NewAppService(
	appRepo *repository.AppRepository,
	keyRepo *repository.APIKeyRepository,
	userRepo *repository.UserRepository,
	db *database.DB,
	jwtSecret string,
	postgrestAdmin string,
) *AppService {
	return &AppService{
		appRepo:        appRepo,
		keyRepo:        keyRepo,
		userRepo:       userRepo,
		db:             db,
		jwtSecret:      jwtSecret,
		postgrestAdmin: postgrestAdmin,
	}
}

type CreateAppRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	OrgID       string `json:"org_id,omitempty"`
}

type CreateAppResponse struct {
	App            models.App              `json:"app"`
	AdminKey       models.APIKeyResponse   `json:"admin_key"`
	ServiceKey     models.APIKeyResponse   `json:"service_key"`
	JWTSecret      string                  `json:"jwt_secret"`
	ConnectionURI  string                  `json:"connection_uri"`
}

func (s *AppService) CreateApp(ctx context.Context, req CreateAppRequest, userID uuid.UUID) (*CreateAppResponse, error) {
	v := utils.NewValidator()
	v.Required("name", req.Name)
	v.Required("slug", req.Slug)
	v.MinLength("name", req.Name, 2)
	v.MaxLength("name", req.Name, 255)
	v.MinLength("slug", req.Slug, 2)
	v.MaxLength("slug", req.Slug, 100)
	v.Slug("slug", req.Slug)
	if v.HasErrors() {
		return nil, fmt.Errorf("validation: %s", v.Error())
	}

	schemaName := "app_" + req.Slug
	appRole := "app_" + req.Slug
	appJWTSecret := generateJWTSecret()

	if err := validateIdentifier(schemaName); err != nil {
		return nil, fmt.Errorf("invalid schema name: %w", err)
	}
	if err := validateIdentifier(appRole); err != nil {
		return nil, fmt.Errorf("invalid role name: %w", err)
	}

	var orgID *uuid.UUID
	if req.OrgID != "" {
		parsed, err := uuid.Parse(req.OrgID)
		if err != nil {
			return nil, fmt.Errorf("invalid org_id: %w", err)
		}
		orgID = &parsed
	}

	ownerID := userID
	app := &models.App{
		OrgID:       orgID,
		OwnerID:     &ownerID,
		Name:        req.Name,
		Slug:        req.Slug,
		SchemaName:  schemaName,
		Description: req.Description,
		Status:      models.AppStatusActive,
		Region:      "us-east",
		Visibility:  models.VisibilityPublic,
		Settings:    models.JSON{},
	}

	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		if err := s.appRepo.Create(ctx, app); err != nil {
			return fmt.Errorf("create app: %w", err)
		}

		if _, err := tx.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, quoteIdentifier(schemaName))); err != nil {
			return fmt.Errorf("create schema: %w", err)
		}

		if _, err := tx.Exec(ctx, fmt.Sprintf(`CREATE ROLE %s WITH LOGIN PASSWORD %s NOINHERIT`, quoteIdentifier(appRole), quoteLiteral(appJWTSecret))); err != nil {
			return fmt.Errorf("create role: %w", err)
		}

		if _, err := tx.Exec(ctx, fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT ALL ON TABLES TO %s`, quoteIdentifier(schemaName), quoteIdentifier(appRole))); err != nil {
			return fmt.Errorf("grant schema privileges: %w", err)
		}

		if _, err := tx.Exec(ctx, fmt.Sprintf(`GRANT USAGE ON SCHEMA %s TO %s`, quoteIdentifier(schemaName), quoteIdentifier(appRole))); err != nil {
			return fmt.Errorf("grant schema usage: %w", err)
		}

		if _, err := tx.Exec(ctx, fmt.Sprintf(`GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA %s TO %s`, quoteIdentifier(schemaName), quoteIdentifier(appRole))); err != nil {
			return fmt.Errorf("grant table privileges: %w", err)
		}

		if _, err := tx.Exec(ctx, fmt.Sprintf(`GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA %s TO %s`, quoteIdentifier(schemaName), quoteIdentifier(appRole))); err != nil {
			return fmt.Errorf("grant sequence privileges: %w", err)
		}

		if _, err := tx.Exec(ctx, `INSERT INTO jwt_secrets (app_id, secret) VALUES ($1, $2)`, app.ID, appJWTSecret); err != nil {
			return fmt.Errorf("insert jwt secret: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	go s.reloadPostgREST()

	adminRawKey, adminPrefix, _ := utils.GenerateAPIKey()
	serviceRawKey, servicePrefix, _ := utils.GenerateAPIKey()

	adminKey := &models.APIKey{
		AppID:     app.ID,
		UserID:    &userID,
		Name:      "admin-key",
		KeyType:   models.KeyTypeAdmin,
		KeyHash:   s.keyRepo.HashKey(adminRawKey),
		KeyPrefix: adminPrefix,
		Scopes:    []string{"*"},
		RateLimit: 10000,
		IsActive:  true,
	}

	serviceKey := &models.APIKey{
		AppID:     app.ID,
		UserID:    &userID,
		Name:      "service-key",
		KeyType:   models.KeyTypeService,
		KeyHash:   s.keyRepo.HashKey(serviceRawKey),
		KeyPrefix: servicePrefix,
		Scopes:    []string{"read", "write"},
		RateLimit: 1000,
		IsActive:  true,
	}

	if err := s.keyRepo.Create(ctx, adminKey); err != nil {
		slog.Warn("failed to create admin key", "error", err)
	}

	if err := s.keyRepo.Create(ctx, serviceKey); err != nil {
		slog.Warn("failed to create service key", "error", err)
	}

	return &CreateAppResponse{
		App: *app,
		AdminKey: models.APIKeyResponse{
			ID:        adminKey.ID,
			Name:      adminKey.Name,
			KeyType:   adminKey.KeyType,
			KeyPrefix: adminPrefix,
			RawKey:    adminRawKey,
			Scopes:    adminKey.Scopes,
			CreatedAt: adminKey.CreatedAt,
		},
		ServiceKey: models.APIKeyResponse{
			ID:        serviceKey.ID,
			Name:      serviceKey.Name,
			KeyType:   serviceKey.KeyType,
			KeyPrefix: servicePrefix,
			RawKey:    serviceRawKey,
			Scopes:    serviceKey.Scopes,
			CreatedAt: serviceKey.CreatedAt,
		},
		JWTSecret:     appJWTSecret,
		ConnectionURI: fmt.Sprintf("https://api.nexbic.com/v1/%s", req.Slug),
	}, nil
}

func generateJWTSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validateIdentifier(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("identifier cannot be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("identifier too long (max 63 chars)")
	}
	if !validIdentifier.MatchString(name) {
		return fmt.Errorf("identifier contains invalid characters: %s", name)
	}
	return nil
}

func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func quoteLiteral(val string) string {
	return `'` + strings.ReplaceAll(val, `'`, `''`) + `'`
}

func (s *AppService) reloadPostgREST() {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic in postgrest reload", "panic", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reloadURL := strings.TrimRight(s.postgrestAdmin, "/") + "/r/reload"
	req, err := http.NewRequestWithContext(ctx, "POST", reloadURL, nil)
	if err != nil {
		slog.Warn("failed to create postgrest reload request", "error", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Warn("failed to reload postgrest schema cache", "error", err)
		return
	}
	resp.Body.Close()
}

func (s *AppService) UpdateApp(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	v := utils.NewValidator()
	if name, ok := updates["name"]; ok {
		nameStr, _ := name.(string)
		v.Required("name", nameStr)
		v.MinLength("name", nameStr, 2)
		v.MaxLength("name", nameStr, 255)
	}
	if desc, ok := updates["description"]; ok {
		descStr, _ := desc.(string)
		v.MaxLength("description", descStr, 1000)
	}
	if v.HasErrors() {
		return fmt.Errorf("validation: %s", v.Error())
	}
	return s.appRepo.Update(ctx, id, updates)
}

func (s *AppService) GetApp(ctx context.Context, id uuid.UUID) (*models.App, error) {
	return s.appRepo.GetByID(ctx, id)
}

func (s *AppService) ListApps(ctx context.Context, orgID *uuid.UUID, limit, offset int) ([]models.App, int, error) {
	return s.appRepo.List(ctx, orgID, limit, offset)
}

func (s *AppService) DeleteApp(ctx context.Context, id uuid.UUID) error {
	app, err := s.appRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if app == nil {
		return fmt.Errorf("app not found")
	}

	return s.appRepo.Delete(ctx, id)
}
