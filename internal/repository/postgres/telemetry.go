package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TelemetryRepositoryPostgres struct {
	db *gorm.DB
	l  *zap.Logger
}

func NewTelemetryRepositoryPostgres(db *gorm.DB, baseLogger *zap.Logger) repository.TelemetryRepository {
	return &TelemetryRepositoryPostgres{
		db: db,
		l:  logger.Named(baseLogger, "TelemetryRepositoryPostgres"),
	}
}

func (r *TelemetryRepositoryPostgres) Ingest(ctx context.Context, telemetryData *model.Telemetry) error {
	if err := r.db.WithContext(ctx).Model(&model.Telemetry{}).Create(telemetryData).Error; err != nil {
		return apperror.ErrDBInsert.WithMessagef(apperror.RepoErrorMsg("insert", domain.EntityTelemetry))
	}
	return nil
}

func (r *TelemetryRepositoryPostgres) HardDelete(ctx context.Context, telemetryID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Unscoped().Where("id = ?", telemetryID).Delete(&model.Telemetry{})
	if tx.Error != nil {
		return apperror.ErrDBDelete.WithMessagef(apperror.RepoErrorMsg("hard delete", domain.EntityTelemetry))
	}
	if tx.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessagef("failed to %s %s: no record found", "hard delete", domain.EntityTelemetry)
	}
	return nil
}

func (r *TelemetryRepositoryPostgres) Delete(ctx context.Context, telemetryID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Where("id = ?", telemetryID).Delete(&model.Telemetry{})
	if tx.Error != nil {
		return apperror.ErrDBDelete.WithMessagef(apperror.RepoErrorMsg("delete", domain.EntityTelemetry))
	}
	if tx.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessagef("failed to %s %s: no record found", "delete", domain.EntityTelemetry)
	}
	return nil
}

func (r *TelemetryRepositoryPostgres) FindByID(ctx context.Context, telemetryID uuid.UUID) (*model.Telemetry, error) {
	telemetryData := new(model.Telemetry)
	if err := r.db.WithContext(ctx).Where("id = ?", telemetryID).First(&telemetryData).Error; err != nil {
		return nil, apperror.ErrDBQuery.WithMessagef(apperror.RepoErrorMsg("query", domain.EntityTelemetry))
	}
	return telemetryData, nil
}

// func (r *TelemetryRepositoryPostgres) ListByDeviceID(ctx context.Context, sensorID uuid.UUID) ([]*model.Telemetry, error) {
// 	sensorData := new(model.Device)
// 	tx := r.db.WithContext(ctx).Model(&model.Sensor{}).Where("id = ?", sensorID).First(&sensorData)
// 		if tx.Error != nil {
// 		return sensorData. ,apperror.ErrDBDelete.WithMessagef(apperror.RepoErrorMsg("delete", domain.EntityTelemetry))
// 	}
// 	if tx.RowsAffected == 0 {
// 		return apperror.ErrNotFound.WithMessagef("failed to %s %s: no record found", "delete", domain.EntityTelemetry)
// 	}
// 	return nil
// }
