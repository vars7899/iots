package main

import (
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/db"
	"github.com/vars7899/iots/internal/validation"
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
	if config.InProd() {
		logger.L().Error("reset not allowed under production")
		return
	}

	logger.L().Info("Database reset initiated")
	if err := db.ResetDatabase(gormDB.DB()); err != nil {
		logger.L().Error("Database reset failed", zap.Error(err))
		return
	}
	logger.L().Info("Database reset completed")
}
