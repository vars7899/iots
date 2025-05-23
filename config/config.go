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
	Server    *ServerConfig    `mapstructure:"server"`
	Postgres  *PostgresConfig  `mapstructure:"postgres"`
	Jwt       *JwtConfig       `mapstructure:"jwt"`
	Redis     *RedisConfig     `mapstructure:"redis"`
	Auth      *AuthConfig      `mapstructure:"auth"`
	Frontend  *FrontendConfig  `mapstructure:"frontend"`
	Websocket *WebsocketConfig `mapstructure:"websocket"`
	Nats      *NatsConfig      `mapstructure:"nats"`
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
	AccessSecret             string        `mapstructure:"access_secret"`
	RefreshSecret            string        `mapstructure:"refresh_secret"`
	AccessTokenTTL           time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL          time.Duration `mapstructure:"refresh_token_ttl"`
	DeviceConnectionTokenTTL time.Duration `mapstructure:"device_connection_token_ttl"`
	DeviceRefreshTokenTTL    time.Duration `mapstructure:"device_refresh_token_ttl"`
	Leeway                   time.Duration `mapstructure:"leeway"`
}

type WebsocketConfig struct {
	PingTimeout  time.Duration `mapstructure:"ping_timeout"`
	PongTimeout  time.Duration `mapstructure:"pong_timeout"`
	ReadDeadline int64         `mapstructure:"read_deadline"`
}

type RedisConfig struct {
	Addr     string        `mapstructure:"address"`
	Password string        `mapstructure:"password"`
	DB       int           `mapstructure:"db"`
	Prefix   string        `mapstructure:"prefix"`
	TTL      time.Duration `mapstructure:"ttl"`
}

type AuthConfig struct {
	DefaultNewUserRoleSlug       string        `mapstructure:"default_new_user_role_slug"`
	RequestResetPasswordTokenTTL time.Duration `mapstructure:"request_reset_password_token_ttl"`
}

type FrontendConfig struct {
	BaseUrl string `mapstructure:"base_url"`
}

type NatsConfig struct {
	BaseUrl string `mapstructure:"base_url"`
}

var envBindings = map[string]string{
	"postgres.db_user":     "POSTGRES_DB_USER",
	"postgres.db_password": "POSTGRES_DB_PASSWORD",
	"postgres.db_name":     "POSTGRES_DB_NAME",
	"jwt.access_secret":    "JWT_ACCESS_SECRET",
	"jwt.refresh_secret":   "JWT_REFRESH_SECRET",
	"redis.password":       "REDIS_PASSWORD",
	"frontend.base_url":    "FRONTEND_BASE_URL",
	"nats.base_url":        "NATS_BASE_URL",
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
