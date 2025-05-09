package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/vars7899/iots/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Device struct {
	ID              uuid.UUID           `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid();"`
	OwnerID         *uuid.UUID          `json:"owner_id" gorm:"type:uuid;index"`
	ProvisionCode   string              `json:"provision_code" gorm:"type:text;not null"`
	ConnectionToken string              `json:"-" gorm:"type:text;not null"`
	Name            string              `json:"name" gorm:"type:varchar(255);not null;index"`
	Description     string              `json:"description" gorm:"type:text"`
	IPAddress       string              `json:"ip_address" gorm:"type:varchar(45);index"`
	MACAddress      string              `json:"mac_address" gorm:"type:varchar(17);index"`
	Status          DeviceStatus        `json:"status" gorm:"type:varchar(20);default:'pending_provision'"`
	Specification   DeviceSpecification `json:"specification" gorm:"embedded"`
	TelemetryConfig TelemetryConfig     `json:"telemetry_config" gorm:"embedded"`
	BroadcastConfig BroadcastConfig     `json:"broadcast_config" gorm:"embedded"`
	Location        domain.GeoLocation  `json:"location" gorm:"embedded"`
	Metadata        datatypes.JSON      `json:"metadata" gorm:"type:jsonb"`
	Tags            pq.StringArray      `json:"tags" gorm:"type:text[]"`
	Capabilities    pq.StringArray      `json:"capabilities" gorm:"type:text[]"`
	LastConnected   *time.Time          `json:"last_connected"`
	CreatedAt       time.Time           `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time           `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt      `json:"-" gorm:"index"`
	Sensors         []Sensor            `json:"sensors" gorm:"foreignKey:DeviceID;constraints:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

func (d *Device) PublicView() *Device {
	return d
}

func (d *Device) StoreProvisionCode(codeStr string) error {
	hashedCode, err := bcrypt.GenerateFromPassword([]byte(codeStr), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	d.ProvisionCode = string(hashedCode)
	return nil
}

func (d *Device) CompareProvisionCode(raw string) error {
	return bcrypt.CompareHashAndPassword([]byte(d.ProvisionCode), []byte(raw))
}

func (d *Device) HashConnectionToken(tokenStr string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(tokenStr), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	d.ConnectionToken = string(hashed)
	return nil
}

func (d *Device) CompareConnectionToken(rawStr string) error {
	return bcrypt.CompareHashAndPassword([]byte(d.ConnectionToken), []byte(rawStr))
}

type DeviceSpecification struct {
	Manufacturer    string `json:"manufacturer" gorm:"type:varchar(255);not null;index"`
	ModelNumber     string `json:"model_number" gorm:"type:varchar(255);not null"`
	SerialNumber    string `json:"serial_number" gorm:"type:varchar(255);not null"`
	FirmwareVersion string `json:"firmware_version" gorm:"type:varchar(20)"`
	HardwareVersion string `json:"hardware_version" gorm:"type:varchar(20)"`
	SoftwareVersion string `json:"software_version" gorm:"type:varchar(20)"`
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
// type DeviceEvent struct {
// 	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
// 	DeviceID  uuid.UUID `json:"device_id" gorm:"type:uuid"`
// 	EventType string    `json:"event_type"` // connection, disconnection, error, update, etc.
// 	Severity  string    `json:"severity"`   // info, warning, error, critical
// 	Message   string    `json:"message"`
// 	Metadata  JSON      `json:"metadata" gorm:"type:jsonb"`
// 	CreatedAt time.Time `json:"created_at"`
// }

type JSON map[string]interface{}

type DeviceUpdate struct {
	ID      uuid.UUID
	Updates interface{}
}

// DeviceStatus represent specific lifecycle status of the digital device
type DeviceStatus string

var (
	DeviceStatusPendingProvision DeviceStatus = "pending_provision"
	DeviceStatusProvisioned      DeviceStatus = "provisioned"
	DeviceStatusDecommissioned   DeviceStatus = "decommissioned"
	DeviceStatusOnline           DeviceStatus = "online"
	DeviceStatusOffline          DeviceStatus = "offline"
	DeviceStatusFaulty           DeviceStatus = "faulty"
	DeviceStatusUnderMaintenance DeviceStatus = "under_maintenance"
	DeviceStatusSuspended        DeviceStatus = "suspended"
)

func IsValidDeviceStatus(inputStr string) bool {
	switch DeviceStatus(inputStr) {
	case
		DeviceStatusPendingProvision,
		DeviceStatusProvisioned,
		DeviceStatusDecommissioned,
		DeviceStatusOnline,
		DeviceStatusOffline,
		DeviceStatusFaulty,
		DeviceStatusUnderMaintenance,
		DeviceStatusSuspended:
		return true
	default:
		return false
	}
}
