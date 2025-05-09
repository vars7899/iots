package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/pubsub"
	"go.uber.org/zap"
)

// Would you like me to include persistence (PostgreSQL or Redis), a WebSocket state broadcast hook, or integration with telemetry ingestion?

type DeviceRecord struct {
	Metadata *DeviceMetadata
	State    *DeviceState
}

func (d *DeviceRecord) Clone() *DeviceRecord {
	clone := &DeviceRecord{}

	if d.Metadata != nil {
		clone.Metadata = d.Metadata.Clone()
	}
	if d.State != nil {
		clone.State = d.State.Clone()
	}

	return clone
}

type DeviceMetadata struct {
	DeviceID        uuid.UUID `json:"device_id"`
	FirmwareVersion string    `json:"firmware_version"`
	HardwareModel   string    `json:"hardware_model"`
	Manufacturer    string    `json:"manufacturer"`
	RegisteredAt    time.Time `json:"registered_at"`
}

func (d *DeviceMetadata) Clone() *DeviceMetadata {
	return &DeviceMetadata{
		DeviceID:        d.DeviceID,
		FirmwareVersion: d.FirmwareVersion,
		HardwareModel:   d.HardwareModel,
		Manufacturer:    d.Manufacturer,
		RegisteredAt:    d.RegisteredAt,
	}
}

type DeviceState struct {
	Status       domain.Status `json:"status"`
	Timestamp    time.Time     `json:"timestamp"`
	IPAddress    string        `json:"ip_address"`
	ErrorCode    string        `json:"error_code,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
	LastSeen     time.Time     `json:"last_seen"`
}

func (d *DeviceState) Clone() *DeviceState {
	return &DeviceState{
		Status:       d.Status,
		Timestamp:    d.Timestamp,
		IPAddress:    d.IPAddress,
		ErrorCode:    d.ErrorCode,
		ErrorMessage: d.ErrorMessage,
		LastSeen:     d.LastSeen,
	}
}

type PartialDeviceState struct {
	Status       *domain.Status `json:"status,omitempty"`
	Timestamp    *time.Time     `json:"timestamp,omitempty"`
	IPAddress    *string        `json:"ip_address,omitempty"`
	ErrorCode    *string        `json:"error_code,omitempty"`
	ErrorMessage *string        `json:"error_message,omitempty"`
}

type DeviceStateManager struct {
	records   map[uuid.UUID]*DeviceRecord
	mu        sync.RWMutex
	publisher pubsub.PubSubPublisher
	logger    *zap.Logger
}

func NewDeviceStateManager(publisher pubsub.PubSubPublisher, baseLogger *zap.Logger) *DeviceStateManager {
	return &DeviceStateManager{
		records:   map[uuid.UUID]*DeviceRecord{},
		publisher: publisher,
		logger:    logger.Named(baseLogger, "DeviceStateManager"),
	}
}

func (m *DeviceStateManager) RegisterMetadata(ctx context.Context, metadata *DeviceMetadata, purgeStaleState bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentTime := time.Now().UTC()
	existingRecord, ok := m.records[metadata.DeviceID]
	if !ok {
		// new device, add record
		metadata.RegisteredAt = currentTime

		m.records[metadata.DeviceID] = &DeviceRecord{
			Metadata: metadata.Clone(),
			State: &DeviceState{
				Status:    domain.Offline,
				Timestamp: currentTime,
				LastSeen:  currentTime,
			},
		}
		m.logger.Info("Device registered", zap.String("device_id", metadata.DeviceID.String()))

		// Publish a "device_registered" event
		// m.publishAsync(context.Background(), formatTopic(TopicDeviceRegistered, metadata.DeviceID.String()), m.records[metadata.DeviceID].Metadata)

		return nil
	}

	m.logger.Info("Device re-registering (updating metadata)...", zap.String("device_id", metadata.DeviceID.String()))

	metadata.RegisteredAt = currentTime
	existingRecord.Metadata = metadata.Clone()

	if purgeStaleState {
		// TODO: delete persisted data
		existingRecord.State = &DeviceState{Status: domain.Offline, Timestamp: currentTime, LastSeen: currentTime}
		m.logger.Info("Device metadata updated and state reset", zap.String("device_id", metadata.DeviceID.String()))
	} else {
		if existingRecord.State == nil { // Ensure state is not nil if not purging
			existingRecord.State = &DeviceState{
				Status:    domain.Offline,
				Timestamp: currentTime,
			}
		}
		existingRecord.State.LastSeen = currentTime
		m.logger.Info("Device metadata updated and state preserved", zap.String("device_id", metadata.DeviceID.String()))
	}

	// m.publishAsync(context.Background(), formatTopic(TopicDeviceMetadata, metadata.DeviceID.String()), existingRecord.Metadata)

	return nil
}

func (m *DeviceStateManager) UpdateState(ctx context.Context, deviceID uuid.UUID, updates *PartialDeviceState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentTime := time.Now().UTC()
	device, ok := m.records[deviceID]
	if !ok {
		m.logger.Error("State update received for unknown device", zap.String("device_id", deviceID.String()))
		return fmt.Errorf("device %s not found for state update", deviceID)
	}

	stateChanged := false

	if device.State == nil {
		device.State = &DeviceState{
			Timestamp: currentTime,
			Status:    domain.Offline,
		}
		m.logger.Info("Initialized state for device during first state update", zap.String("device_id", deviceID.String()))
	}
	if updates.Status != nil {
		if !domain.IsValidStatus(string(*updates.Status)) {
			m.logger.Warn(fmt.Sprintf("Received invalid status '%s' from device %s. Ignoring status update for status.", *updates.Status, deviceID.String()))
			// Don't update status if invalid
		} else {
			if device.State.Status != *updates.Status {
				device.State.Status = *updates.Status
				stateChanged = true
			}
		}
	}

	if updates.Timestamp != nil {
		if device.State.Timestamp.Equal(*updates.Timestamp) {
			device.State.Timestamp = *updates.Timestamp
			stateChanged = true
		}
	} else {
		device.State.Timestamp = currentTime
		m.logger.Warn("Device state update missing timestamp, using server time", zap.String("device_id", deviceID.String()))
		stateChanged = true
	}

	if updates.IPAddress != nil {
		if device.State.IPAddress != *updates.IPAddress {
			device.State.IPAddress = *updates.IPAddress
			stateChanged = true
		}
	}
	if updates.ErrorCode != nil {
		if device.State.ErrorCode != *updates.ErrorCode {
			device.State.ErrorCode = *updates.ErrorCode
			stateChanged = true
		}
	}
	if updates.ErrorMessage != nil {
		if device.State.ErrorMessage != *updates.ErrorMessage {
			device.State.ErrorMessage = *updates.ErrorMessage
			stateChanged = true
		}
	}

	device.State.LastSeen = currentTime
	stateChanged = true

	m.logger.Info("Device state updated", zap.String("device_id", deviceID.String()))

	if stateChanged {
		// m.publishAsync(ctx, formatTopic(TopicDeviceState, deviceID.String()), device.State)
	}
	return nil
}
func (m *DeviceStateManager) GetDeviceRecord(ctx context.Context, deviceID uuid.UUID) (*DeviceRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	current, ok := m.records[deviceID]
	if !ok {
		return nil, false
	}
	// return deep copy not the real memory
	return current.Clone(), true
}
func (m *DeviceStateManager) GetDeviceMetadata(ctx context.Context, deviceID uuid.UUID) (*DeviceMetadata, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	current, ok := m.records[deviceID]
	if !ok || current.Metadata == nil {
		return nil, false
	}
	// return deep copy not the real memory
	return current.Metadata.Clone(), true
}

func (m *DeviceStateManager) GetDeviceState(ctx context.Context, deviceID uuid.UUID) (*DeviceState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	current, ok := m.records[deviceID]
	if !ok || current.State == nil {
		return nil, false
	}
	// return deep copy not the real memory
	return current.State.Clone(), true
}

func (m *DeviceStateManager) publishAsync(ctx context.Context, topic string, message interface{}) {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		m.logger.Error("Failed to marshal message for Pub/Sub publish", zap.String("topic", topic), zap.Error(err))
		return
	}

	go func() {
		pubErr := m.publisher.Publish(ctx, topic, msgBytes)
		if pubErr != nil {
			m.logger.Error("Failed to publish message to Pub/Sub", zap.String("topic", topic), zap.Error(pubErr))
		}
	}()
}
