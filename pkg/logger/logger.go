package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Lgr *zap.Logger
)

func InitLogger(onProductionMode bool) *zap.Logger {
	var cfg zap.Config

	if onProductionMode {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	}
	_logger, err := cfg.Build()
	if err != nil {
		panic("failed to build zap logger: " + err.Error())
	}
	Lgr = _logger
	return Lgr
}
