package db

import (
	"fmt"
	"time"

	"github.com/vars7899/iots/configs"
	"github.com/vars7899/iots/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectPostgres(cfg configs.PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres database: %w", err)
	}

	sql_db, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database handle: %w", err)
	}

	// connection pool setting
	sql_db.SetMaxOpenConns(20)
	sql_db.SetMaxIdleConns(20)
	sql_db.SetConnMaxLifetime(time.Minute * 5)

	logger.Lgr.Info("PostgreSQL connected successfully.")

	return db, nil
}
