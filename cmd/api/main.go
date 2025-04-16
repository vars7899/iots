package main

import (
	"log"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/config"
	api_v1 "github.com/vars7899/iots/internal/api/v1"
	"github.com/vars7899/iots/internal/db"
	"github.com/vars7899/iots/internal/repository/postgres"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/internal/validatorz"
	"github.com/vars7899/iots/pkg/auth/token"
	"github.com/vars7899/iots/pkg/logger"
)

func main() {
	validatorz.InitValidator()
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

	if err := db.AutoMigrate(postgresDB); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	// Initialize repositories
	sensorRepo := postgres.NewSensorRepositoryPostgres(postgresDB)
	userRepo := postgres.NewUserRepositoryPostgres(postgresDB)
	deviceRepo := postgres.NewDeviceRepositoryPostgres(postgresDB)

	// Initialize services
	sensorService := service.NewSensorService(sensorRepo)
	userService := service.NewUserService(userRepo)
	deviceService := service.NewDeviceService(deviceRepo, logger.Lgr)
	tokenService := token.NewJwtTokenService(os.Getenv("JWT_ACCESS_SECRET"), os.Getenv("JWT_REFRESH_SECRET"), 15*time.Hour, 24*7*time.Hour)

	// Create dependency injection container
	deps := api_v1.APIDependencies{
		SensorService: sensorService,
		UserService:   userService,
		TokenService:  tokenService,
		DeviceService: deviceService,
	}

	// Initialize Echo
	e := echo.New()

	// Register routes with DIs
	api_v1.RegisterRoutes(e, deps, logger.Lgr)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
