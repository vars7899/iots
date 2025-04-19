package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

var GlobalConfig *AppConfig

type AppConfig struct {
	Server   *ServerConfig   `mapstructure:"server"`
	Postgres *PostgresConfig `mapstructure:"postgres"`
	Jwt      *JwtConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Port            string        `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	HideBanner      bool          `mapstructure:"hide_banner"`
	HidePort        bool          `mapstructure:"hide_port"`
}

type PostgresConfig struct {
	DBUser     string `mapstructure:"db_user"`
	DBPassword string `mapstructure:"db_password"`
	DBName     string `mapstructure:"db_name"`
	DBHost     string `mapstructure:"db_host"`
	DBPort     string `mapstructure:"db_port"`
}

type JwtConfig struct {
	AccessSecret    string        `mapstructure:"access_secret"`
	RefreshSecret   string        `mapstructure:"refresh_secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

var envBindings = map[string]string{
	"postgres.db_user":     "POSTGRES_DB_USER",
	"postgres.db_password": "POSTGRES_DB_PASSWORD",
	"postgres.db_name":     "POSTGRES_DB_NAME",
	"postgres.db_host":     "POSTGRES_DB_HOST",
	"postgres.db_port":     "POSTGRES_DB_PORT",
	"jwt.access_secret":    "JWT_ACCESS_SECRET",
	"jwt.refresh_secret":   "JWT_REFRESH_SECRET",
}

func Load(filename string, filetype string, path string, baseLogger *zap.Logger) error {
	logger := logger.Named(baseLogger, "Config")

	if err := godotenv.Load(); err != nil {
		logger.Warn("failed to load .env file, continuing with system env", zap.Error(err))
	}

	viper.SetConfigName(filename)
	viper.SetConfigType(filetype)
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		logger.Error("error reading config file", zap.Error(err))
		return err
	}

	// Bind environment variables
	viper.AutomaticEnv()
	for key, env := range envBindings {
		if err := viper.BindEnv(key, env); err != nil {
			logger.Warn("failed to bind environment variable", zap.String("key", key), zap.String("env", env), zap.Error(err))
		}
	}
	logger.Info(".env binding loaded successfully")

	var cfg AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		logger.Error("failed to unmarshal app config", zap.Error(err))
		return err
	}
	GlobalConfig = &cfg

	logger.Info("configuration loaded successfully")
	return nil
}

func InProd() bool {
	isProductionEnv := os.Getenv("IS_PRODUCTION")
	isProduction, err := strconv.ParseBool(isProductionEnv)
	if err != nil {
		isProduction = false
		logger.L().Error("failed to parse IS_PRODUCTION from .env")
	}
	return isProduction
}
