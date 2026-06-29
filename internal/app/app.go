package app

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	"github.com/nexbic/platform/config"
	auditHandlers "github.com/nexbic/platform/internal/audit/handlers"
	auditService "github.com/nexbic/platform/internal/audit/service"
	authHandlers "github.com/nexbic/platform/internal/auth/handlers"
	authRepo "github.com/nexbic/platform/internal/auth/repository"
	authService "github.com/nexbic/platform/internal/auth/service"
	backupHandlers "github.com/nexbic/platform/internal/backups/handlers"
	backupService "github.com/nexbic/platform/internal/backups/service"
	dashHandlers "github.com/nexbic/platform/internal/dashboard/handlers"
	dashService "github.com/nexbic/platform/internal/dashboard/service"
	explorerHandlers "github.com/nexbic/platform/internal/explorer/handlers"
	explorerService "github.com/nexbic/platform/internal/explorer/service"
	extHandlers "github.com/nexbic/platform/internal/extensions/handlers"
	extService "github.com/nexbic/platform/internal/extensions/service"
	filesHandlers "github.com/nexbic/platform/internal/files/handlers"
	filesService "github.com/nexbic/platform/internal/files/service"
	logsHandlers "github.com/nexbic/platform/internal/logs/handlers"
	logsService "github.com/nexbic/platform/internal/logs/service"
	"github.com/nexbic/platform/internal/middleware"
	monHandlers "github.com/nexbic/platform/internal/monitoring/handlers"
	monService "github.com/nexbic/platform/internal/monitoring/service"
	pgroleHandlers "github.com/nexbic/platform/internal/pgroles/handlers"
	pgroleService "github.com/nexbic/platform/internal/pgroles/service"
	projectsHandlers "github.com/nexbic/platform/internal/projects/handlers"
	schemaHandlers "github.com/nexbic/platform/internal/schema/handlers"
	schemaService "github.com/nexbic/platform/internal/schema/service"
	sqlHandlers "github.com/nexbic/platform/internal/sql/handlers"
	sqlService "github.com/nexbic/platform/internal/sql/service"
	storageHandlers "github.com/nexbic/platform/internal/storage/handlers"
	storageRepo "github.com/nexbic/platform/internal/storage/repository"
	storageService "github.com/nexbic/platform/internal/storage/service"
	walletHandlers "github.com/nexbic/platform/internal/wallet/handlers"
	walletRepo "github.com/nexbic/platform/internal/wallet/repository"
	walletService "github.com/nexbic/platform/internal/wallet/service"
	"github.com/nexbic/platform/pkg/database"
)

type App struct {
	Config *config.Config
	DB     *database.DB
	AuthMW *middleware.AuthMiddleware

	AuthHandler     *authHandlers.AuthHandler
	AuditHandler    *auditHandlers.AuditHandler
	DashHandler     *dashHandlers.DashboardHandler
	ExplorerHandler *explorerHandlers.ExplorerHandler
	SQLHandler      *sqlHandlers.SQLHandler
	SchemaHandler   *schemaHandlers.SchemaHandler
	PgRoleHandler   *pgroleHandlers.PgRolesHandler
	ExtHandler      *extHandlers.ExtensionsHandler
	MonHandler      *monHandlers.MonitoringHandler
	BackupHandler   *backupHandlers.BackupHandler
	LogsHandler     *logsHandlers.LogsHandler
	StorageHandler  *storageHandlers.StorageHandler
	ProjectsHandler *projectsHandlers.ProjectsHandler
	WalletHandler   *walletHandlers.WalletHandler
	FilesHandler    *filesHandlers.FilesHandler
}

func New(ctx context.Context) (*App, error) {
	cfg := config.Load()

	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	authMW := middleware.NewAuthMiddleware(cfg.JWT, db.Pool)

	userRepo := authRepo.NewUserRepository(db)
	tokenRepo := authRepo.NewRefreshTokenRepo(db)
	secRepo := authRepo.NewSecurityRepo(db)
	authSvc := authService.NewAuthService(userRepo, secRepo, tokenRepo, cfg.JWT, cfg.SuperAdmin)
	authHandler := authHandlers.NewAuthHandler(authSvc)

	if cfg.SuperAdmin.Email != "" {
		authSvc.SeedSuperAdmin(ctx)
	}

	auditSvc := auditService.NewAuditService(db)
	auditHandler := auditHandlers.NewAuditHandler(auditSvc)

	dashSvc := dashService.NewDashboardService(db)
	dashHandler := dashHandlers.NewDashboardHandler(dashSvc)

	explorerSvc := explorerService.NewExplorerService(db)
	explorerHandler := explorerHandlers.NewExplorerHandler(explorerSvc)

	sqlSvc := sqlService.NewSQLService(db)
	sqlHandler := sqlHandlers.NewSQLHandler(sqlSvc)

	schemaSvc := schemaService.NewSchemaService(db)
	schemaHandler := schemaHandlers.NewSchemaHandler(schemaSvc)

	pgroleSvc := pgroleService.NewPgRolesService(db)
	pgroleHandler := pgroleHandlers.NewPgRolesHandler(pgroleSvc)

	extSvc := extService.NewExtensionsService(db)
	extHandler := extHandlers.NewExtensionsHandler(extSvc)

	monSvc := monService.NewMonitoringService(db)
	monHandler := monHandlers.NewMonitoringHandler(monSvc)

	backupDir := os.Getenv("BACKUP_DIR")
	if backupDir == "" {
		backupDir = "/data/backups"
	}
	backupSvc := backupService.NewBackupService(db, backupDir,
		cfg.Database.Host, strconv.Itoa(cfg.Database.Port),
		cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	backupHandler := backupHandlers.NewBackupHandler(backupSvc)

	logsSvc := logsService.NewLogsService(db)
	logsHandler := logsHandlers.NewLogsHandler(logsSvc)

	storageRepoInstance := storageRepo.NewStorageRepo(db)
	storageSvc := storageService.NewStorageService(storageRepoInstance)
	storageHandler := storageHandlers.NewStorageHandler(storageSvc)

	projectsHandler := projectsHandlers.NewProjectsHandler(db.Pool)

	filesDir := os.Getenv("FILES_DIR")
	if filesDir == "" {
		filesDir = "/data/user_files"
	}
	filesSvc := filesService.NewFilesService(filesDir)
	filesHandler := filesHandlers.NewFilesHandler(filesSvc)

	walletRepoInstance := walletRepo.NewWalletRepo(db)
	walletSvc := walletService.NewWalletService(walletRepoInstance)
	walletHandler := walletHandlers.NewWalletHandler(walletSvc)

	slog.Info("app initialized", "env", cfg.AppEnv, "addr", cfg.Addr())

	return &App{
		Config: cfg,
		DB:     db,
		AuthMW: authMW,

		AuthHandler:     authHandler,
		AuditHandler:    auditHandler,
		DashHandler:     dashHandler,
		ExplorerHandler: explorerHandler,
		SQLHandler:      sqlHandler,
		SchemaHandler:   schemaHandler,
		PgRoleHandler:   pgroleHandler,
		ExtHandler:      extHandler,
		MonHandler:      monHandler,
		BackupHandler:   backupHandler,
		LogsHandler:     logsHandler,
		StorageHandler:  storageHandler,
		ProjectsHandler: projectsHandler,
		WalletHandler:   walletHandler,
		FilesHandler:    filesHandler,
	}, nil
}

func (a *App) Close() {
	a.DB.Close()
}
