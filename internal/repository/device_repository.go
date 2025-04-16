package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/pkg/pagination"
)

// DeviceRepository defines all operations for device management
type DeviceRepository interface {
	Create(ctx context.Context, device *model.Device) (*model.Device, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Device, error)
	Update(ctx context.Context, deviceID uuid.UUID, device *model.Device) (*model.Device, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Query operations
	FindAll(ctx context.Context, paginationOpt *pagination.Pagination) ([]*model.Device, int64, error)
	FindByOwnerID(ctx context.Context, ownerID uuid.UUID, paginationOpt *pagination.Pagination) ([]*model.Device, int64, error)
	// FindByType(ctx context.Context, deviceType string, paginationOpt pagination.Pagination) ([]device.Device, int64, error)
	// FindByStatus(ctx context.Context, status string, paginationOpt pagination.Pagination) ([]device.Device, int64, error)
	// FindByTags(ctx context.Context, tags []string, paginationOpt pagination.Pagination) ([]device.Device, int64, error)
	// FindByCapabilities(ctx context.Context, capabilities []string, paginationOpt pagination.Pagination) ([]device.Device, int64, error)
	// Search(ctx context.Context, query string, paginationOpt pagination.Pagination) ([]device.Device, int64, error)

	// // Specialized operations
	// FindNearLocation(ctx context.Context, lat, lon float64, radiusKm float64, page, pageSize int) ([]device.Device, int64, error)
	// FindByLastConnectedRange(ctx context.Context, start, end time.Time, page, pageSize int) ([]device.Device, int64, error)

	// Status management
	UpdateStatus(ctx context.Context, deviceID uuid.UUID, status string) error
	MarkOnline(ctx context.Context, deviceID uuid.UUID) (*model.Device, error)
	MarkOffline(ctx context.Context, deviceID uuid.UUID) (*model.Device, error)

	// // Bulk operations
	BulkCreate(ctx context.Context, devices []*model.Device) ([]*model.Device, error)
	BulkUpdate(ctx context.Context, devices []*model.Device) ([]*model.Device, error)
	BulkDelete(ctx context.Context, ids []uuid.UUID) error
	// UpdateDevicesStatus(ctx context.Context, ids []uuid.UUID, status string) error

	// // Telemetry config management
	// UpdateTelemetryConfig(ctx context.Context, deviceID uuid.UUID, config device.TelemetryConfig) error

	// // Broadcast config management
	// UpdateBroadcastConfig(ctx context.Context, deviceID uuid.UUID, config device.BroadcastConfig) error

	// // Access control
	// AssignOwner(ctx context.Context, deviceID, ownerID uuid.UUID) error
	// AddAccessGroup(ctx context.Context, deviceID, groupID uuid.UUID) error
	// RemoveAccessGroup(ctx context.Context, deviceID, groupID uuid.UUID) error
	// GetAccessGroups(ctx context.Context, deviceID uuid.UUID) ([]device.AccessGroup, error)

	// // Sensor management
	// AddSensor(ctx context.Context, deviceID uuid.UUID, sensor sensor.Sensor) error
	// RemoveSensor(ctx context.Context, deviceID, sensorID uuid.UUID) error
	// GetSensors(ctx context.Context, deviceID uuid.UUID) ([]sensor.Sensor, error)

	// // Event management
	// RecordEvent(ctx context.Context, event device.DeviceEvent) error
	// GetEvents(ctx context.Context, deviceID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]device.DeviceEvent, int64, error)

	// // Firmware management
	// UpdateFirmware(ctx context.Context, deviceID uuid.UUID, version string) error
	// GetFirmwareHistory(ctx context.Context, deviceID uuid.UUID) ([]FirmwareUpdate, error)

	// // Health and diagnostics
	// CountDevicesByStatus(ctx context.Context) (map[string]int64, error)
	// GetAverageUptimeByType(ctx context.Context) (map[string]float64, error)
	// GetDeviceHealthMetrics(ctx context.Context, deviceID uuid.UUID) (*DeviceHealthMetrics, error)

	// // Advanced transaction support
	// WithTransaction(tx interface{}) DeviceRepository

	// Cleanup and maintenance
	// CleanupStaleDevices(ctx context.Context, thresholdDays int) (int64, error)
	// ArchiveInactiveDevices(ctx context.Context, inactiveDays int) (int64, error)

	// // Cache management
	// InvalidateCache(ctx context.Context, deviceID uuid.UUID) error
	// PrefetchDevices(ctx context.Context, deviceIDs []uuid.UUID) error
}

// For completeness, here are the additional models referenced above
type FirmwareUpdate struct {
	ID         uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DeviceID   uuid.UUID `json:"device_id" gorm:"type:uuid"`
	OldVersion string    `json:"old_version"`
	NewVersion string    `json:"new_version"`
	UpdatedBy  uuid.UUID `json:"updated_by" gorm:"type:uuid"`
	UpdatedAt  time.Time `json:"updated_at"`
	Status     string    `json:"status"` // success, failed, in-progress
	Notes      string    `json:"notes"`
}

type DeviceHealthMetrics struct {
	DeviceID           uuid.UUID `json:"device_id"`
	UptimePercentage   float64   `json:"uptime_percentage"`
	LastDowntime       time.Time `json:"last_downtime"`
	ErrorCount         int64     `json:"error_count"`
	AveragePingLatency float64   `json:"average_ping_latency"` // in milliseconds
	BatteryLevel       float64   `json:"battery_level"`        // percentage if applicable
	MemoryUsage        float64   `json:"memory_usage"`         // percentage
	CpuUsage           float64   `json:"cpu_usage"`            // percentage
	StorageUsage       float64   `json:"storage_usage"`        // percentage
	LastHealthCheck    time.Time `json:"last_health_check"`
}
