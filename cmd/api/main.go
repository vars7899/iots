package main

import (
	"github.com/vars7899/iots/config"
	v1 "github.com/vars7899/iots/internal/api/v1"
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

	// fmt.Println(config.GlobalConfig.Postgres, config.GlobalConfig.Jwt, config.GlobalConfig.Redis, config.GlobalConfig.Server)

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
			return
		}
	}

	// Load dependencies
	deps, err := di.NewProvider(gormDB.DB(), logger.L(), config.GlobalConfig)
	if err != nil {
		logger.L().Fatal("failed to load provider dependencies", zap.Error(err))
		return
	}

	s := server.NewServer(deps, logger.L(), config.GlobalConfig.Server)

	// mount router(s)
	v1.RegisterRoutes(s.E(), s.Provider, logger.L())

	// Start http server
	s.Start()
	s.WaitForShutdown()
}
