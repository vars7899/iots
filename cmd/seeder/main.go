package main

import (
	"log"

	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/db"
	"github.com/vars7899/iots/internal/seed"
	"github.com/vars7899/iots/pkg/logger"
)

func main() {
	logger.InitLogger(false)

	postgresConfig, err := config.Load(".env.dev")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	_db, err := db.ConnectPostgres(*postgresConfig)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	if err := db.AutoMigrate(_db); err != nil {
		log.Fatal("migration failed", err)
	}

	err = seed.SeedSensorData(_db, 50)
	if err != nil {
		log.Fatalf("failed to seed sensors: %v", err)
	}

	logger.Lgr.Info("Sensor seeding completed")
}
