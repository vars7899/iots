package repository

import (
	"context"

	"github.com/vars7899/iots/internal/domain/model"
)

type TelemetryRepository interface {
	Ingest(ctx context.Context, telemetryData *model.Telemetry) error
}