package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/vars7899/iots/internal/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Device struct {
	ID              uuid.UUID             `gorm:"type:uuid;primaryKey;default:gen_random_uuid();" json:"id"`
	Name            string                `gorm:"type:varchar(255);not null;index" json:"name"`
	Description     string                `gorm:"type:text;not null" json:"description"`
	Manufacturer    string                `gorm:"type:varchar(255);not null;index" json:"manufacturer"`
	ModelNumber     string                `gorm:"type:varchar(255);not null" json:"model_number"`
	SerialNumber    string                `gorm:"type:varchar(255);not null" json:"serial_number"`
	FirmwareVersion string                `gorm:"type:varchar(20);not null" json:"firmware_version"`
	IPAddress       string                `gorm:"type:varchar(45);index" json:"ip_address"`
	MACAddress      string                `gorm:"type:varchar(17);index" json:"mac_address"`
	IsOnline        bool                  `gorm:"default:false" json:"is_online"`
	ConnectionType  domain.ConnectionType `gorm:"type:varchar(20)" json:"connection_type"`
	Status          domain.Status         `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Location        domain.GeoLocation    `gorm:"embedded" json:"location"`
	CreatedAt       time.Time             `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time             `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt        `gorm:"index" json:"-"`
	LastConnected   *time.Time            `json:"last_connected"`
	Metadata        datatypes.JSON        `gorm:"type:jsonb" json:"metadata"`
	Tags            pq.StringArray        `gorm:"type:text[]" json:"tags"`
	Capabilities    pq.StringArray        `gorm:"type:text[]" json:"capabilities"`
	TelemetryConfig TelemetryConfig       `gorm:"embedded" json:"telemetry_config"`
	BroadcastConfig BroadcastConfig       `gorm:"embedded" json:"broadcast_config"`
	// Sensors         []sensor.Sensor `gorm:"foreignKey:DeviceID" json:"sensors"`
}

type TelemetryConfig struct {
	Enabled            bool           `gorm:"default:true" json:"enabled"`
	ReportingFrequency int            `gorm:"default:60" json:"reporting_frequency_seconds"` // in seconds
	BatchSize          int            `gorm:"default:1" json:"batch_size"`
	RetentionPeriod    int            `gorm:"default:30" json:"retention_period_days"` // in days
	StorageQuota       int64          `json:"storage_quota_bytes"`                     // in bytes
	CompressionEnabled bool           `gorm:"default:false" json:"compression_enabled"`
	EncryptionEnabled  bool           `gorm:"default:false" json:"encryption_enabled"`
	AlertThresholds    datatypes.JSON `gorm:"type:jsonb" json:"alert_thresholds"` // flexible thresholds configuration
}

type BroadcastConfig struct {
	BroadcastEnabled bool   `gorm:"default:false" json:"broadcast_enabled"`
	Protocol         string `json:"protocol"` // MQTT, AMQP, etc.
	BrokerURL        string `json:"broker_url"`
	Topic            string `json:"topic"`
	QoS              int    `gorm:"default:0" json:"qos"` // Quality of Service level
	RetainMessages   bool   `gorm:"default:false" json:"retain_messages"`
	ClientID         string `json:"client_id"`
	Username         string `json:"username"`
	Password         string `json:"-"`
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

type DeviceUpdate struct {
	ID      uuid.UUID
	Updates interface{}
}
