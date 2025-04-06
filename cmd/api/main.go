package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/config"
	api_v1 "github.com/vars7899/iots/internal/api/v1"
	"github.com/vars7899/iots/internal/db"
	"github.com/vars7899/iots/internal/repository/postgres"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/logger"
)

func main() {
	logger.InitLogger(false)

	postgresConfig, err := config.Load(".env.dev")
	if err != nil {
		logger.Lgr.Error(err.Error())
		os.Exit(1)
	}

	postgresDB, err := db.ConnectPostgres(*postgresConfig)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	// Initialize repositories
	sensorRepo := postgres.NewSensorRepositoryPostgres(postgresDB)

	// Initialize services
	sensorService := service.NewSensorService(sensorRepo)

	// Create dependency injection container
	deps := api_v1.APIDependencies{
		SensorService: sensorService,
	}

	// Initialize Echo
	e := echo.New()

	// Register routes with DIs
	api_v1.RegisterRoutes(e, deps)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
