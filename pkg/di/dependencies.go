package di

import (
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/internal/repository/postgres"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Provider struct {
	TokenService token.TokenService
	Repositories *RepositoryProvider
	Services     *ServiceProvider
	Helpers      *HelperProvider
	logger       *zap.Logger
	db           *gorm.DB
	config       *config.AppConfig
}

type RepositoryProvider struct {
	SensorRepository repository.SensorRepository
	DeviceRepository repository.DeviceRepository
	UserRepository   repository.UserRepository
}

type ServiceProvider struct {
	SensorService *service.SensorService
	DeviceService *service.DeviceService
	UserService   *service.UserService
}

type HelperProvider struct {
	TokenService token.TokenService
	// In the future:
	// CacheService  cache.CacheService
	// EmailService  email.EmailService
	// Metrics       metrics.MetricsProvider
}

func NewProvider(db *gorm.DB, baseLogger *zap.Logger, cfg *config.AppConfig) error {
	p := &Provider{
		db:     db,
		logger: logger.Named(baseLogger, "Provider"),
		config: cfg,
	}

	if err := p.initRepositoryProvider(); err != nil {
		return err
	}
	p.logger.Info("repositories initialized successfully")
	if err := p.initServiceProvider(); err != nil {
		return err
	}
	p.logger.Info("services initialized successfully")
	if err := p.initHelperProvider(); err != nil {
		return err
	}
	p.logger.Info("helpers initialized successfully")

	return nil
}

func (p *Provider) initRepositoryProvider() error {
	repoLogger := logger.Named(p.logger, "RepositoryProvider")

	if p.db == nil {
		return apperror.ErrDBMissing.WithMessage("repository initialization failed: missing db")
	}

	p.Repositories = &RepositoryProvider{
		SensorRepository: postgres.NewSensorRepositoryPostgres(p.db, repoLogger),
		DeviceRepository: postgres.NewDeviceRepositoryPostgres(p.db, repoLogger),
		UserRepository:   postgres.NewUserRepositoryPostgres(p.db, repoLogger),
	}
	return nil
}

func (p *Provider) initServiceProvider() error {
	serviceLogger := logger.Named(p.logger, "ServiceProvider")

	if p.Repositories.SensorRepository == nil {
		return apperror.ErrMissingDependency.WithMessage("missing sensor repository")
	}
	if p.Repositories.DeviceRepository == nil {
		return apperror.ErrMissingDependency.WithMessage("missing device repository")
	}
	if p.Repositories.UserRepository == nil {
		return apperror.ErrMissingDependency.WithMessage("missing user repository")
	}

	p.Services = &ServiceProvider{
		SensorService: service.NewSensorService(p.Repositories.SensorRepository, serviceLogger),
		DeviceService: service.NewDeviceService(p.Repositories.DeviceRepository, serviceLogger),
		UserService:   service.NewUserService(p.Repositories.UserRepository, serviceLogger),
	}
	return nil
}

func (p *Provider) initHelperProvider() error {
	helperLogger := logger.Named(p.logger, "HelperProvider")

	if p.config == nil || p.config.Jwt == nil {
		return apperror.ErrMissingConfig.WithMessage("missing json web token credentials")
	}

	p.Helpers = &HelperProvider{
		TokenService: token.NewJwtTokenService(p.config.Jwt, helperLogger),
	}

	return nil
}
