package model

import (
	"time"

	"github.com/google/uuid"
)

type DeviceCommandType string

var (
	CommandTypeCommand  DeviceCommandType = "command"
	CommandTypeAck      DeviceCommandType = "ack"
	CommandTypeResponse DeviceCommandType = "response"
	CommandTypeError    DeviceCommandType = "error"
)

type DeviceCommand struct {
	ID          uuid.UUID              `json:"id"`
	DeviceID    uuid.UUID              `json:"device_id"`
	Type        DeviceCommandType      `json:"type"`
	Command     string                 `json:"command"`
	CommandCode string                 `json:"command_code"`
	Payload     map[string]interface{} `json:"payload"`
	Timestamp   time.Time              `json:"timestamp"`
}

const (
	// Device lifecycle events
	EventTypeDeviceConnected    = "device_connected"
	EventTypeDeviceDisconnected = "device_disconnected"
	EventTypeDeviceInitialized  = "device_initialized"

	// Command processing events
	EventTypeCommandReceived  = "command_received"
	EventTypeCommandProcessed = "command_processed"
	EventTypeCommandFailed    = "command_failed"

	// State events
	EventTypeStateChanged = "state_changed"

	// Error events
	EventTypeError = "error"
)

// DeviceEvent represent an event in the device command processor service
type DeviceEvent struct {
	ID            uuid.UUID              `json:"id"`                       // unique identifier for event helpful for tracing
	Type          string                 `json:"type"`                     // type of device event
	DeviceID      uuid.UUID              `json:"device_id"`                // device connected to the event
	CommandCode   string                 `json:"command_code,omitempty"`   // original command code that triggered this event
	Payload       map[string]interface{} `json:"payload"`                  // detailed event data
	Timestamp     time.Time              `json:"timestamp"`                // time when event was triggered
	CorrelationID uuid.UUID              `json:"correlation_id,omitempty"` // for tracing request-response pairs
}

func NewDeviceErrorEvent(commandID, deviceID uuid.UUID, commandCode, message, errorCode string) *DeviceEvent {
	return &DeviceEvent{
		ID:            uuid.New(),
		Type:          EventTypeError,
		DeviceID:      deviceID,
		CommandCode:   commandCode,
		Timestamp:     time.Now(),
		CorrelationID: commandID,
		Payload: map[string]interface{}{
			"error":      message,
			"error_code": errorCode,
		},
	}
}
