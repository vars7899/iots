package di

import (
	"context"
	"sync"

	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/cache"
	"github.com/vars7899/iots/internal/cache/redis"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/internal/repository/postgres"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/internal/worker"
	"github.com/vars7899/iots/internal/ws"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AppContainer struct {
	Repositories *RepositoryProvider
	Services     *ServiceProvider
	CoreServices *CoreServiceProvider
	Logger       *zap.Logger
	DB           *gorm.DB
	WsHub        *ws.Hub
	Config       *config.AppConfig
	WaitGroup    *sync.WaitGroup
	Ctx          context.Context
}

type RepositoryProvider struct {
	SensorRepository             repository.SensorRepository
	DeviceRepository             repository.DeviceRepository
	UserRepository               repository.UserRepository
	TelemetryRepository          repository.TelemetryRepository
	RoleRepository               repository.RoleRepository
	ResetPasswordTokenRepository repository.ResetPasswordTokenRepository
}

type ServiceProvider struct {
	SensorService             *service.SensorService
	DeviceService             *service.DeviceService
	UserService               service.UserService
	TelemetryService          *service.TelemetryService
	RoleService               service.RoleService
	ResetPasswordTokenService service.ResetPasswordTokenService
	AuthService               service.AuthService
}

type CoreServiceProvider struct {
	AuthTokenService     auth.AuthTokenService
	AccessControlService auth.AccessControlService
	JWTTokenService      token.TokenService
	JTIStoreService      cache.JTIStore
}

func NewAppContainer(ctx context.Context, wg *sync.WaitGroup, db *gorm.DB, cfg *config.AppConfig, baseLogger *zap.Logger) (*AppContainer, error) {
	logger := logger.Named(baseLogger, "AppContainer")

	repoProvider, err := NewRepositoryProvider(db, logger)
	if err != nil {
		logger.Error("failed to initialize repository provider", zap.Error(err))
		return nil, err
	}

	coreServiceProvider, err := NewCoreServiceProvider(db, cfg, logger)
	if err != nil {
		logger.Error("failed to initialize core service provider", zap.Error(err))
		return nil, err
	}

	serviceProvider, err := NewServiceProvider(repoProvider, coreServiceProvider, cfg, baseLogger)
	if err != nil {
		logger.Error("failed to initialize service provider", zap.Error(err))
		return nil, err
	}

	a := &AppContainer{
		Repositories: repoProvider,
		Services:     serviceProvider,
		CoreServices: coreServiceProvider,
		Config:       cfg,
		Logger:       baseLogger,
		WaitGroup:    wg,
		Ctx:          ctx,
		DB:           db,
	}

	a.initWebsocketHub()
	a.initServiceWorker()

	logger.Info("App container initialized successfully")
	return a, nil
}

func (a *AppContainer) Close() {
	a.Logger.Info("AppContainer waiting for background services to shut down...")
	a.WaitGroup.Wait()
	a.Logger.Info("AppContainer background services shut down complete.")

	// Add cleanup for services/components that need it (Redis, DB if necessary, etc.)
	// Example Redis close:
	if a.CoreServices != nil && a.CoreServices.JTIStoreService != nil {
		if closer, ok := a.CoreServices.JTIStoreService.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				a.Logger.Error("Failed to close JTI store (Redis client)", zap.Error(err))
			} else {
				a.Logger.Info("JTI store (Redis client) closed")
			}
		}
	}

	a.Logger.Info("App container close complete.")
}

func NewRepositoryProvider(db *gorm.DB, baseLogger *zap.Logger) (*RepositoryProvider, error) {
	logger := logger.Named(baseLogger, "RepositoryProvider")

	if db == nil {
		logger.Error("RepositoryProvider initialization failed: missing db")
		return nil, apperror.ErrDBMissing.WithMessage("repository initialization failed: missing db")
	}

	return &RepositoryProvider{
		SensorRepository:             postgres.NewSensorRepositoryPostgres(db, logger),
		DeviceRepository:             postgres.NewDeviceRepositoryPostgres(db, logger),
		UserRepository:               postgres.NewUserRepositoryPostgres(db, logger),
		TelemetryRepository:          postgres.NewTelemetryRepositoryPostgres(db, logger),
		RoleRepository:               postgres.NewRoleRepositoryPostgres(db, logger),
		ResetPasswordTokenRepository: postgres.NewResetPasswordTokenRepositoryPostgres(db, logger),
	}, nil
}

func NewCoreServiceProvider(db *gorm.DB, cfg *config.AppConfig, baseLogger *zap.Logger) (*CoreServiceProvider, error) {
	logger := logger.Named(baseLogger, "CoreServiceProvider")

	if db == nil {
		logger.Error("CoreServiceProvider initialization failed: missing db")
		return nil, apperror.ErrDBMissing.WithMessage("initialization failed: missing db")
	}
	if cfg == nil {
		logger.Error("CoreServiceProvider initialization failed: config is nil")
		return nil, apperror.ErrMissingConfig.WithMessage("application configuration is missing")
	}
	if cfg.Jwt == nil {
		logger.Error("CoreServiceProvider initialization failed: JWT config is missing")
		return nil, apperror.ErrMissingConfig.WithMessage("JWT configuration is missing")
	}
	if cfg.Redis == nil {
		logger.Error("CoreServiceProvider initialization failed: Redis config is missing (required for JTI store)")
		return nil, apperror.ErrMissingConfig.WithMessage("Redis configuration is missing")
	}

	accessControlService, err := auth.NewAccessControlService(db, "internal/casbin/model.conf", logger)
	if err != nil {
		logger.Error("CoreServiceProvider failed to start Access Control service", zap.Error(err))
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeInit, "failed to start access control service")
	}

	jwtTokenService := token.NewJwtTokenService(cfg.Jwt, logger)
	jtiStoreService := redis.NewRedisJTIStore(cfg.Redis, logger)
	authTokenService := auth.NewAuthTokenManger(jwtTokenService, jtiStoreService, logger)

	return &CoreServiceProvider{
		JWTTokenService:      jwtTokenService,
		JTIStoreService:      jtiStoreService,
		AuthTokenService:     authTokenService,
		AccessControlService: accessControlService,
	}, nil
}

func NewServiceProvider(repoProvider *RepositoryProvider, coreProvider *CoreServiceProvider, cfg *config.AppConfig, baseLogger *zap.Logger) (*ServiceProvider, error) {
	logger := logger.Named(baseLogger, "ServiceProvider")

	if repoProvider == nil {
		logger.Error("ServiceProvider initialization failed: missing RepositoryProvider")
		return nil, apperror.ErrMissingDependency.WithMessage("missing repository provider")
	}
	if coreProvider == nil {
		logger.Error("ServiceProvider initialization failed: missing CoreServiceProvider")
		return nil, apperror.ErrMissingDependency.WithMessage("missing core service provider")
	}

	if repoProvider.DeviceRepository == nil ||
		repoProvider.ResetPasswordTokenRepository == nil ||
		repoProvider.RoleRepository == nil ||
		repoProvider.SensorRepository == nil ||
		repoProvider.TelemetryRepository == nil ||
		repoProvider.UserRepository == nil {
		logger.Error("ServiceProvider initialization failed: missing one or more of the required repository")
		return nil, apperror.ErrMissingDependency.WithMessage("missing required one or more repository")
	}

	roleService := service.NewRoleService(repoProvider.RoleRepository, cfg.Auth.DefaultNewUserRoleSlug, logger)
	resetPasswordTokenService := service.NewResetPasswordTokenService(repoProvider.ResetPasswordTokenRepository, repoProvider.UserRepository, logger)
	userService := service.NewUserService(repoProvider.UserRepository, logger)
	sensorService := service.NewSensorService(repoProvider.SensorRepository, logger)
	deviceService := service.NewDeviceService(repoProvider.DeviceRepository, logger)
	telemetryService := service.NewTelemetryService(repoProvider.TelemetryRepository, logger)
	authService := service.NewAuthService(userService, roleService, coreProvider.AccessControlService, coreProvider.AuthTokenService, resetPasswordTokenService, config.GlobalConfig, logger)

	logger.Info("ServiceProvider initialized successfully")

	return &ServiceProvider{
		SensorService:             sensorService,
		DeviceService:             deviceService,
		TelemetryService:          telemetryService,
		UserService:               userService,
		RoleService:               roleService,
		ResetPasswordTokenService: resetPasswordTokenService,
		AuthService:               authService,
	}, nil
}

func (a *AppContainer) initWebsocketHub() {
	a.WsHub = ws.NewHub(a.Logger)

	a.WaitGroup.Add(1)
	go func() {
		a.WsHub.Run(a.Ctx, a.WaitGroup)
	}()
	a.Logger.Info("websocket hub initialized successfully")
}

func (a *AppContainer) initServiceWorker() error {
	if a.WsHub == nil {
		a.Logger.Fatal("websocket hub is nil, cannot start service worker")
		return apperror.ErrMissingDependency.WithMessage("missing websocket hub to start service worker")
	}
	if a.Services == nil {
		a.Logger.Fatal("service provider is nil, cannot start service worker")
		return apperror.ErrMissingDependency.WithMessage("missing services to start service worker")
	}

	l := logger.Named(a.Logger, "ServiceWorker")

	a.WaitGroup.Add(1)
	go worker.TelemetryWorker(a.Ctx, a.WaitGroup, a.WsHub.GetSensorTelemetryMessageChannel(), a.Services.TelemetryService, l)

	l.Info("Telemetry worker started")

	return nil
}
