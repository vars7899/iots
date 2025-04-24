package service

import (
	"context"

	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type TelemetryService struct {
	telemetryRepo repository.TelemetryRepository
	l             *zap.Logger
}

func NewTelemetryService(telemetryRepo repository.TelemetryRepository, baseLogger *zap.Logger) *TelemetryService {
	return &TelemetryService{
		telemetryRepo: telemetryRepo,
		l:             logger.Named(baseLogger, "TelemetryService"),
	}
}

func (s *TelemetryService) IngestSensorTelemetry(ctx context.Context, data *model.Telemetry) error {
	return s.telemetryRepo.Ingest(ctx, data)
}
