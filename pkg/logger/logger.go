package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

func Init(onProductionMode bool) {
	var cfg zap.Config

	if onProductionMode {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	}
	logger, err := cfg.Build()
	if err != nil {
		panic("failed to initialize zap logger: " + err.Error())
	}
	globalLogger = logger
}

func InitDev() {
	Init(false)
}

func InitProd() {
	Init(true)
}

func Sync() {
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
}

func L() *zap.Logger {
	if globalLogger == nil {
		panic("logger not initialized: nil pointer dereference")
	}
	return globalLogger
}

func Named(baseLogger *zap.Logger, withName string) *zap.Logger {
	if baseLogger == nil {
		baseLogger = zap.NewNop()
	}
	return baseLogger.Named(withName)
}
