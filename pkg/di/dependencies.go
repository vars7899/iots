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
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Provider struct {
	Repositories *RepositoryProvider
	Services     *ServiceProvider
	Helpers      *HelperProvider
	l            *zap.Logger
	db           *gorm.DB
	WsHub        *ws.Hub
	config       *config.AppConfig
	wg           *sync.WaitGroup
	ctx          context.Context
}

type RepositoryProvider struct {
	SensorRepository    repository.SensorRepository
	DeviceRepository    repository.DeviceRepository
	UserRepository      repository.UserRepository
	TelemetryRepository repository.TelemetryRepository
	RoleRepository      repository.RoleRepository
}

type ServiceProvider struct {
	SensorService    *service.SensorService
	DeviceService    *service.DeviceService
	UserService      *service.UserService
	TelemetryService *service.TelemetryService
	RoleService      service.RoleService
	CasbinService    service.CasbinService
}

type HelperProvider struct {
	TokenService token.TokenService
	JTIService   cache.JTIStore
	// In the future:
	// EmailService  email.EmailService
	// Metrics       metrics.MetricsProvider
}

func NewProvider(ctx context.Context, wg *sync.WaitGroup, db *gorm.DB, baseLogger *zap.Logger, cfg *config.AppConfig) (*Provider, error) {
	p := &Provider{
		db:     db,
		l:      logger.Named(baseLogger, "Provider"),
		config: cfg,
		wg:     wg,
		ctx:    ctx,
	}

	if err := p.initRepositoryProvider(); err != nil {
		return nil, err
	}
	if err := p.initServiceProvider(); err != nil {
		return nil, err
	}
	if err := p.initHelperProvider(); err != nil {
		return nil, err
	}

	p.initWebsocketHub()
	p.initServiceWorker()

	return p, nil
}

func (p *Provider) Close() {
	p.l.Info("Provider waiting for background services to shut down...")
	p.wg.Wait()
	p.l.Info("Provider background services shut down complete.")

	// You might add other cleanup here, like closing database connections if GORM
	// requires manual closing (often managed internally) or closing Redis clients.
	// Example (if your Redis client needs closing and is in a provider field):
	// if p.Helpers != nil && p.Helpers.JTIService != nil {
	//     if closer, ok := p.Helpers.JTIService.(interface{ Close() error }); ok {
	//          if err := closer.Close(); err != nil {
	//              p.l.Error("Failed to close JTI service (Redis client)", zap.Error(err))
	//          } else {
	//              p.l.Info("JTI service (Redis client) closed")
	//          }
	//     }
	// }

	p.l.Info("Provider close complete.")
}

func (p *Provider) initRepositoryProvider() error {
	l := logger.Named(p.l, "RepositoryProvider")

	if p.db == nil {
		l.Error("provider repository initialization failed: missing db")
		return apperror.ErrDBMissing.WithMessage("repository initialization failed: missing db")
	}

	p.Repositories = &RepositoryProvider{
		SensorRepository:    postgres.NewSensorRepositoryPostgres(p.db, l),
		DeviceRepository:    postgres.NewDeviceRepositoryPostgres(p.db, l),
		UserRepository:      postgres.NewUserRepositoryPostgres(p.db, l),
		TelemetryRepository: postgres.NewTelemetryRepositoryPostgres(p.db, l),
		RoleRepository:      postgres.NewRoleRepositoryPostgres(p.db, l),
	}
	l.Info("provider repositories initialized successfully")
	return nil
}

func (p *Provider) initServiceProvider() error {
	l := logger.Named(p.l, "ServiceProvider")

	if p.Repositories.SensorRepository == nil {
		l.Error("provider service initialization failed: missing sensor repository")
		return apperror.ErrMissingDependency.WithMessage("missing sensor repository")
	}
	if p.Repositories.DeviceRepository == nil {
		l.Error("provider service initialization failed: missing device repository")
		return apperror.ErrMissingDependency.WithMessage("missing device repository")
	}
	if p.Repositories.UserRepository == nil {
		l.Error("provider service initialization failed: missing user repository")
		return apperror.ErrMissingDependency.WithMessage("missing user repository")
	}
	if p.Repositories.TelemetryRepository == nil {
		l.Error("provider service initialization failed: missing telemetry repository")
		return apperror.ErrMissingDependency.WithMessage("missing telemetry repository")
	}
	if p.Repositories.RoleRepository == nil {
		l.Error("provider service initialization failed: missing role repository")
		return apperror.ErrMissingDependency.WithMessage("missing role repository")
	}

	casbinSrv, err := service.NewCasbinService(p.db, "internal/casbin/model.conf", l)
	if err != nil {
		p.l.Error("failed to start casbin service", zap.Error(err))
		return apperror.ErrorHandler(err, apperror.ErrCodeInit, "failed to start casbin service").WithMessage("missing telemetry repository")
	}

	roleSrv := service.NewRoleService(p.Repositories.RoleRepository, l, p.config.Auth.DefaultNewUserRoleSlug)
	userSrv := service.NewUserService(p.Repositories.UserRepository, roleSrv, casbinSrv, l)

	p.Services = &ServiceProvider{
		SensorService:    service.NewSensorService(p.Repositories.SensorRepository, l),
		DeviceService:    service.NewDeviceService(p.Repositories.DeviceRepository, l),
		TelemetryService: service.NewTelemetryService(p.Repositories.TelemetryRepository, l),
		UserService:      userSrv,
		RoleService:      roleSrv,
		CasbinService:    casbinSrv,
	}
	l.Info("provider services initialized successfully")
	return nil
}

func (p *Provider) initHelperProvider() error {
	l := logger.Named(p.l, "HelperProvider\t")

	if p.config == nil || p.config.Jwt == nil {
		l.Error("provider helper initialization failed: missing sensor repository")
		return apperror.ErrMissingConfig.WithMessage("missing json web token credentials")
	}
	if p.config == nil || p.config.Redis == nil {
		l.Error("provider helper initialization failed: missing sensor repository")
		return apperror.ErrMissingConfig.WithMessage("missing redis credentials")
	}

	p.Helpers = &HelperProvider{
		TokenService: token.NewJwtTokenService(p.config.Jwt, l),
		JTIService:   redis.NewRedisJTIStore(p.config.Redis, l),
	}

	l.Info("provider helpers initialized successfully")
	return nil
}

func (p *Provider) initWebsocketHub() {
	l := logger.Named(p.l, "WebsocketHub\t")
	p.WsHub = ws.NewHub(p.l)

	p.wg.Add(1)
	go func() {
		p.WsHub.Run(p.ctx, p.wg)
	}()
	l.Info("websocket hub initialized successfully")
}

func (p *Provider) initServiceWorker() error {
	if p.WsHub == nil {
		p.l.Fatal("websocket hub is nil, cannot start service worker")
		return apperror.ErrMissingDependency.WithMessage("missing websocket hub to start service worker")
	}
	if p.Services == nil {
		p.l.Fatal("service provider is nil, cannot start service worker")
		return apperror.ErrMissingDependency.WithMessage("missing services to start service worker")
	}

	l := logger.Named(p.l, "ServiceWorker")

	p.wg.Add(1)
	go worker.TelemetryWorker(p.ctx, p.wg, p.WsHub.GetSensorTelemetryMessageChannel(), p.Services.TelemetryService, l)

	l.Info("Telemetry worker started")

	return nil
}
