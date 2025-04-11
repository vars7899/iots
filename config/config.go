package config

import (
	"os"
	"strconv"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/vars7899/iots/pkg/logger"
)

type PostgresConfig struct {
	DBUser     string `env:"POSTGRES_DB_USER"`
	DBPassword string `env:"POSTGRES_DB_PASSWORD"`
	DBName     string `env:"POSTGRES_DB_NAME"`
	DBHost     string `env:"POSTGRES_DB_HOST"`
	DBPort     string `env:"POSTGRES_DB_PORT"`
}

func Load(filenames ...string) (*PostgresConfig, error) {
	if err := godotenv.Load(filenames...); err != nil {
		return nil, err
	}

	var cfg PostgresConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	logger.Lgr.Info("Configuration loaded successfully")
	return &cfg, nil
}

func IsUnderProduction() bool {
	isProductionEnv := os.Getenv("IS_PRODUCTION")
	isProduction, err := strconv.ParseBool(isProductionEnv)
	if err != nil {
		isProduction = false
		logger.Lgr.Error("failed to parse IS_PRODUCTION from .env")
	}
	return isProduction
}
