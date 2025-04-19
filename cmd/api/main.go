package main

import (
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/db"
	"github.com/vars7899/iots/internal/server"
	"github.com/vars7899/iots/internal/validation"
	"github.com/vars7899/iots/pkg/di"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Init logger
	logger.InitDev()
	defer logger.Sync()

	// Init validator
	validation.Init(logger.L())

	// Load configuration
	if err := config.Load("config.dev", "yaml", "./configs", logger.L()); err != nil {
		logger.L().Fatal("failed to load config, shutting down app", zap.Error(err))
		return
	}

	// Init db
	gormDB, err := db.NewGormDB(logger.L(), config.GlobalConfig.Postgres)
	if err != nil {
		logger.L().Fatal("failed to connect to db", zap.Error(err))
		return
	}
	// auto migration (dev only)
	if !config.InProd() {
		if err := gormDB.AutoMigrateAll(); err != nil {
			logger.L().Fatal("failed to auto-migrate db", zap.Error(err))
		}
	}

	// Load dependencies
	if err := di.NewProvider(gormDB.DB(), logger.L(), config.GlobalConfig); err != nil {
		logger.L().Fatal("failed to load provider dependencies", zap.Error(err))
		return
	}

	// Start http server
	s := server.NewServer(&di.Provider{}, logger.L(), config.GlobalConfig.Server)
	s.Start()
	s.WaitForShutdown()
}

// func main() {
// 	// Initialize custom validator once
// 	validation.InitValidator()

// 	logger.InitLogger(false)

// 	// Fx App - all DI modules and startup logic
// 	app := fx.New(
// 		logger.Module,                 // Logs
// 		config.Module,                 // Load config
// 		db.Module,                     // Connect DB + AutoMigrate
// 		postgres.Module,               // Init all repositories
// 		service.Module,                // Init all services
// 		api.Module,                    // Register API routes
// 		server.Module,                 // Provide echo instance
// 		fx.Invoke(server.StartServer), // Start Echo server on boot
// 	)

// 	app.Run()
// }

// func main() {
// 	validation.InitValidator()
// 	baseLogger := logger.InitLogger(false)

// 	postgresConfig, err := config.Load(".env.dev")
// 	if err != nil {
// 		logger.Lgr.Error(err.Error())
// 		os.Exit(1)
// 	}

// 	postgresDB, err := db.ConnectPostgres(*postgresConfig)
// 	if err != nil {
// 		log.Fatalf("database connection failed: %v", err)
// 	}

// 	if err := db.AutoMigrate(postgresDB); err != nil {
// 		log.Fatalf("database migration failed: %v", err)
// 	}

// 	// Initialize repositories
// 	sensorRepo := postgres.NewSensorRepositoryPostgres(postgresDB, baseLogger)
// 	userRepo := postgres.NewUserRepositoryPostgres(postgresDB)
// 	deviceRepo := postgres.NewDeviceRepositoryPostgres(postgresDB)

// 	// Initialize services
// 	sensorService := service.NewSensorService(sensorRepo)
// 	userService := service.NewUserService(userRepo)
// 	deviceService := service.NewDeviceService(deviceRepo, baseLogger)
// 	tokenService := token.NewJwtTokenService(os.Getenv("JWT_ACCESS_SECRET"), os.Getenv("JWT_REFRESH_SECRET"), 15*time.Hour, 24*7*time.Hour)

// 	// Create dependency injection container
// 	deps := v1.Provider{
// 		SensorService: sensorService,
// 		UserService:   userService,
// 		TokenService:  tokenService,
// 		DeviceService: deviceService,
// 	}

// 	// Initialize Echo
// 	e := echo.New()

// 	// Register routes with DIs
// 	v1.RegisterRoutes(e, deps, baseLogger)

// 	// Start server
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8080"
// 	}
// 	e.Logger.Fatal(e.Start(":" + port))
// }
