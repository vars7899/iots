package main

import (
	"log"
	"os"

	"github.com/vars7899/iots/configs"
	"github.com/vars7899/iots/internal/db"
	"github.com/vars7899/iots/pkg/logger"
)

func main() {
	logger.InitLogger(false)

	postgresConfig, err := configs.Load(".env.dev")
	if err != nil {
		logger.Lgr.Error(err.Error())
		os.Exit(1)
	}

	// // Load config
	// cfg, err := configs.Load()
	// if err != nil {
	// 	log.Fatalf("failed to load config: %v", err)
	// }

	// // Initialize DB
	// db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	// if err != nil {
	// 	log.Fatalf("failed to connect to database: %v", err)
	// }

	// // Initialize repositories
	// sensorRepo := postgres.NewSensorRepositoryPostgres(db)
	// deviceRepo := postgres.NewDeviceRepositoryPostgres(db)

	// // Initialize services
	// sensorService := service.NewSensorService(sensorRepo)
	// deviceService := service.NewDeviceService(deviceRepo)

	// // Create dependency injection container
	// deps := handler.HandlerDeps{
	// 	SensorService: sensorService,
	// 	DeviceService: deviceService,
	// }

	// // Initialize Echo
	// e := echo.New()

	// // Register routes with DI
	// api.RegisterRoutes(e, deps)

	// // Start server
	// port := os.Getenv("PORT")
	// if port == "" {
	// 	port = "8080"
	// }
	// e.Logger.Fatal(e.Start(":" + port))

	_, err = db.ConnectPostgres(*postgresConfig)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
}
