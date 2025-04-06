package main

import (
	"log"

	"github.com/vars7899/iots/configs"
	"github.com/vars7899/iots/internal/db"
)

func main() {
	postgresConfig, err := configs.Load(".env.dev")
	if err != nil {
		log.Fatal(err)
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
