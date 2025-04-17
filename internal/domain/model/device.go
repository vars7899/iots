package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain"
	"gorm.io/gorm"
)

type DeviceID string

type Device struct {
	ID              uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid();" json:"id"`
	Name            string        `gorm:"type:varchar(255);not null;index" json:"name"`
	Description     string        `gorm:"type:text;not null" json:"description"`
	Manufacturer    string        `gorm:"type:varchar(255);not null;index" json:"manufacturer"`
	ModelNumber     string        `gorm:"type:varchar(255);not null" json:"model_number"`
	SerialNumber    string        `gorm:"type:varchar(255);not null" json:"serial_number"`
	FirmwareVersion string        `gorm:"type:varchar(20);not null" json:"firmware_version"`
	IPAddress       string        `gorm:"type:varchar(45);index" json:"ip_address"`
	MACAddress      string        `gorm:"type:varchar(17);index" json:"mac_address"`
	ConnectionType  string        `gorm:"type:varchar(20)" json:"connection_type"`
	IsOnline        bool          `gorm:"default:false" json:"is_online"`
	Status          domain.Status `gorm:"type:varchar(20);default:'pending'" json:"status"`
	// Location        DeviceLocation `gorm:"embedded" json:"location"`
	// LastConnected   *time.Time      `json:"last_connected"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	// Sensors         []sensor.Sensor `gorm:"foreignKey:DeviceID" json:"sensors"`
	// TelemetryConfig TelemetryConfig `gorm:"embedded" json:"telemetry_config"`
	// BroadcastConfig BroadcastConfig `gorm:"embedded" json:"broadcast_config"`
	// Capabilities    pq.StringArray  `gorm:"type:text[] json:"capabilities""` // array of capability identifiers
	// Tags            pq.StringArray  `gorm:"type:text[] json:"tags""`
	// Metadata        JSON            `gorm:"type:jsonb" json:"metadata"`
}

type DeviceLocation struct {
	Latitude       float64 `gorm:"type:double precision" json:"latitude"`
	Longitude      float64 `gorm:"type:double precision" json:"longitude"`
	Altitude       float64 `gorm:"type:double precision" json:"altitude"`
	IndoorLocation bool    `gorm:"default:false" json:"indoor_location"`
	Building       string  `gorm:"type:varchar(100)" json:"building"`
	Floor          int     `json:"floor"`
	Room           string  `gorm:"type:varchar(50)" json:"room"`
	Description    string  `gorm:"type:text" json:"location_description"`
}

// TelemetryConfig represents configuration for device telemetry
type TelemetryConfig struct {
	Enabled            bool  `json:"enabled" gorm:"default:true"`
	ReportingFrequency int   `json:"reporting_frequency_seconds" gorm:"default:60"` // in seconds
	BatchSize          int   `json:"batch_size" gorm:"default:1"`
	RetentionPeriod    int   `json:"retention_period_days" gorm:"default:30"` // in days
	StorageQuota       int64 `json:"storage_quota_bytes"`                     // in bytes
	CompressionEnabled bool  `json:"compression_enabled" gorm:"default:false"`
	EncryptionEnabled  bool  `json:"encryption_enabled" gorm:"default:false"`
	AlertThresholds    JSON  `json:"alert_thresholds" gorm:"type:jsonb"` // flexible thresholds configuration
}

// BroadcastConfig represents configuration for device broadcasting capabilities
type BroadcastConfig struct {
	BroadcastEnabled bool   `json:"broadcast_enabled" gorm:"default:false"`
	Protocol         string `json:"protocol"` // MQTT, AMQP, etc.
	BrokerURL        string `json:"broker_url"`
	Topic            string `json:"topic"`
	QoS              int    `json:"qos" gorm:"default:0"` // Quality of Service level
	RetainMessages   bool   `json:"retain_messages" gorm:"default:false"`
	ClientID         string `json:"client_id"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	CertificatePath  string `json:"certificate_path"`
	PrivateKeyPath   string `json:"private_key_path"`
}

// AccessGroup represents a group of users with specific access levels
type AccessGroup struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"uniqueIndex"`
	Description string    `json:"description"`
	AccessLevel string    `json:"access_level"` // read, write, admin, etc.
}

// DeviceEvent represents significant events in a device's lifecycle
type DeviceEvent struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DeviceID  uuid.UUID `json:"device_id" gorm:"type:uuid"`
	EventType string    `json:"event_type"` // connection, disconnection, error, update, etc.
	Severity  string    `json:"severity"`   // info, warning, error, critical
	Message   string    `json:"message"`
	Metadata  JSON      `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"created_at"`
}

type JSON map[string]interface{}

type DeviceStatus string

var (
	StatusActive   DeviceStatus = "active"
	StatusInactive DeviceStatus = "inactive"
	StatusPending  DeviceStatus = "pending"
	StatusError    DeviceStatus = "error"
)

func (s DeviceStatus) String() string {
	return string(s)
}

func IsValidDeviceStatus(s string) bool {
	switch DeviceStatus(s) {
	case StatusActive, StatusInactive, StatusPending, StatusError:
		return true
	default:
		return false
	}
}

type DeviceUpdate struct {
	ID      uuid.UUID
	Updates interface{}
}
