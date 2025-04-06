package configs

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type PostgresConfig struct {
	DBUser     string `env:"POSTGRES_DB_USER"`
	DBPassword string `env:"POSTGRES_DB_PASSWORD"`
	DBName     string `env:"POSTGRES_DB_NAME"`
	DBHost     string `env:"POSTGRES_DB_HOST"`
	DBPort     string `env:"POSTGRES_DB_PORT"`
}

func Load(filenames ...string) (*PostgresConfig, error) {
	_ = godotenv.Load(filenames...)

	var cfg PostgresConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
