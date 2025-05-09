package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/db"
	"github.com/vars7899/iots/internal/repository/postgres"
	"github.com/vars7899/iots/internal/service/command"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/pubsub"
	"go.uber.org/zap"
)

func main() {
	// initialize logger
	logger.InitDev()
	defer logger.Sync()

	// load configuration from the path
	if err := config.Load("config.dev", "yaml", "./configs", logger.L()); err != nil {
		logger.L().Fatal("failed to load config, shutting down app", zap.Error(err))
		return
	}

	// load device command schema registry from the path
	registry := config.NewDeviceCommandSchemaRegistry()
	registry.LoadDeviceCommandSchemaRegistry("device.command.schema", "yaml", "./configs", logger.L())

	// initialize connection to database
	gormDB, err := db.NewGormDB(logger.L(), config.GlobalConfig.Postgres)
	if err != nil {
		logger.L().Fatal("failed to connect to db", zap.Error(err))
		return
	}

	deviceRepo := postgres.NewDeviceRepositoryPostgres(gormDB.DB(), logger.L())

	natsPubSub, err := pubsub.NewNatsPubSub(config.GlobalConfig.Nats.BaseUrl, logger.L())
	if err != nil {
		// Use apperror for logging NATS connection errors
		appErr := apperror.WrapAppErrWithContext(err, "Failed to connect to NATS", apperror.ErrCodeInit)
		logger.L().Fatal("Failed to connect to NATS", zap.Error(appErr))
	}
	defer natsPubSub.Close()

	// 7. Initialize CommandProcessorService (Pass schemaRegistry)
	cmdProcessor, err := command.NewCommandProcessorService(registry, natsPubSub, deviceRepo, logger.L())
	if err != nil {
		// NewCommandProcessorService uses fmt.Errorf, wrap it if needed, or just log
		logger.L().Fatal("Failed to create CommandProcessorService", zap.Error(err))
	}

	// 8. Start the CommandProcessorService
	err = cmdProcessor.Start()
	if err != nil {
		// Start method now returns AppError, log it
		logger.L().Fatal("Failed to start CommandProcessorService", zap.Error(err))
	}
	// Stop is deferred below after the signal handler

	logger.L().Info("CommandProcessorService started and listening on NATS", zap.String("topic", pubsub.NatsTopicCommandsInbound))
	logger.L().Info("Press Ctrl+C to stop.")

	// 9. Graceful Shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	<-stopChan // Block until a shutdown signal is received

	logger.L().Info("Shutdown signal received. Stopping CommandProcessorService...")
	cmdProcessor.Stop() // Stop the service

	logger.L().Info("CommandProcessorService stopped. Exiting.")
}
