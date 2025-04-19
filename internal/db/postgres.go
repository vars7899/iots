package db

import (
	"fmt"
	"time"

	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GormDB struct {
	db     *gorm.DB
	logger *zap.Logger
	config *config.PostgresConfig
}

func NewGormDB(baseLogger *zap.Logger, cfg *config.PostgresConfig) (*GormDB, error) {
	gormDB := &GormDB{
		logger: logger.Named(baseLogger, "GormDB"),
		config: cfg,
	}

	dbConnect, err := connect(gormDB.logger, gormDB.config)
	if err != nil {
		return nil, apperror.ErrorHandler(err, apperror.ErrCodeDBConnect, "failed to initiate gorm")
	}
	gormDB.db = dbConnect
	return gormDB, nil
}

func connect(logger *zap.Logger, cfg *config.PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, apperror.ErrDBConnect.WithMessage("failed to connect to postgres database").Wrap(err)
	}

	sql_db, err := db.DB()
	if err != nil {
		return nil, apperror.ErrDBConnect.WithMessage("failed to extract database handler").Wrap(err)
	}

	// connection pool setting
	sql_db.SetMaxOpenConns(20)
	sql_db.SetMaxIdleConns(20)
	sql_db.SetConnMaxLifetime(time.Minute * 5)

	logger.Info("connection established successfully to postgres")
	return db, nil
}

func (g *GormDB) DB() *gorm.DB {
	return g.db
}

func (g *GormDB) Ping() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return apperror.ErrDBPing.Wrap(err)
	}
	return sqlDB.Ping()
}

func (g *GormDB) AutoMigrate(models ...interface{}) error {
	return g.db.AutoMigrate(models...)
}

func (g *GormDB) AutoMigrateAll() error {
	if err := g.db.AutoMigrate(DB_Tables...); err != nil {
		g.logger.Error("failed to auto-migrate the database", zap.Error(err))
		return err
	}
	g.logger.Info("database migration completed")
	return nil
}
