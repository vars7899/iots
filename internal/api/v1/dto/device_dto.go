package dto

import (
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/validatorz"
)

type CreateNewDeviceDTO struct {
	Name            string `json:"name" validate:"required,max=255"`
	Description     string `json:"description,omitempty"`
	Manufacturer    string `json:"manufacturer" validate:"required"`
	ModelNumber     string `json:"model_number" validate:"required"`
	SerialNumber    string `json:"serial_number" validate:"required"`
	FirmwareVersion string `json:"firmware_version" validate:"required"`
	IPAddress       string `json:"ip_address" validate:"omitempty,ip"`
	MACAddress      string `json:"mac_address" validate:"omitempty,max=17"`
	ConnectionType  string `json:"connection_type" validate:"omitempty"`
	// Location        DeviceLocationDTO  `json:"location"`
	// TelemetryConfig TelemetryConfigDTO `json:"telemetry_config"`
	// BroadcastConfig BroadcastConfigDTO `json:"broadcast_config"`
	// Capabilities    []string           `json:"capabilities"`
	// Tags            []string           `json:"tags"`
	// Metadata        JSONDTO            `json:"metadata"`
}

func (dto *CreateNewDeviceDTO) Validate() error { return validatorz.Validate.Struct(dto) }

func (dto *CreateNewDeviceDTO) ToDevice() *model.Device {
	return &model.Device{
		Name:            dto.Name,
		Description:     dto.Description,
		Manufacturer:    dto.Manufacturer,
		ModelNumber:     dto.ModelNumber,
		SerialNumber:    dto.SerialNumber,
		FirmwareVersion: dto.FirmwareVersion,
		IPAddress:       dto.IPAddress,
		MACAddress:      dto.MACAddress,
		ConnectionType:  dto.ConnectionType,
	}
}

type UpdateDeviceDTO struct {
	Name            *string `json:"name,omitempty" validate:"omitempty,max=255"`
	Description     *string `json:"description,omitempty" validate:"omitempty"`
	Manufacturer    *string `json:"manufacturer,omitempty" validate:"omitempty"`
	ModelNumber     *string `json:"model_number,omitempty" validate:"omitempty"`
	SerialNumber    *string `json:"serial_number,omitempty" validate:"omitempty"`
	FirmwareVersion *string `json:"firmware_version,omitempty" validate:"omitempty"`
	IPAddress       *string `json:"ip_address,omitempty" validate:"omitempty,ip"`
	MACAddress      *string `json:"mac_address,omitempty" validate:"omitempty,max=17"`
	ConnectionType  *string `json:"connection_type,omitempty" validate:"omitempty"`
}

func (dto *UpdateDeviceDTO) Validate() error { return validatorz.Validate.Struct(dto) }

func (dto *UpdateDeviceDTO) ToDevice() *model.Device {
	device := &model.Device{}

	if dto.Name != nil {
		device.Name = *dto.Name
	}
	if dto.Description != nil {
		device.Description = *dto.Description
	}
	if dto.Manufacturer != nil {
		device.Manufacturer = *dto.Manufacturer
	}
	if dto.ModelNumber != nil {
		device.ModelNumber = *dto.ModelNumber
	}
	if dto.SerialNumber != nil {
		device.SerialNumber = *dto.SerialNumber
	}
	if dto.FirmwareVersion != nil {
		device.FirmwareVersion = *dto.FirmwareVersion
	}
	if dto.IPAddress != nil {
		device.IPAddress = *dto.IPAddress
	}
	if dto.MACAddress != nil {
		device.MACAddress = *dto.MACAddress
	}
	if dto.ConnectionType != nil {
		device.ConnectionType = *dto.ConnectionType
	}

	return device
}

type BroadcastConfigDTO struct {
	BroadcastEnabled bool   `json:"broadcast_enabled"`
	Protocol         string `json:"protocol"`
	BrokerURL        string `json:"broker_url"`
	Topic            string `json:"topic"`
	QoS              int    `json:"qos"`
	RetainMessages   bool   `json:"retain_messages"`
	ClientID         string `json:"client_id"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	CertificatePath  string `json:"certificate_path"`
	PrivateKeyPath   string `json:"private_key_path"`
}

type TelemetryConfigDTO struct {
	Enabled            bool    `json:"enabled"`
	ReportingFrequency int     `json:"reporting_frequency_seconds"`
	BatchSize          int     `json:"batch_size"`
	RetentionPeriod    int     `json:"retention_period_days"`
	StorageQuota       int64   `json:"storage_quota_bytes"`
	CompressionEnabled bool    `json:"compression_enabled"`
	EncryptionEnabled  bool    `json:"encryption_enabled"`
	AlertThresholds    JSONDTO `json:"alert_thresholds"`
}

type DeviceLocationDTO struct {
	Latitude       float64 `json:"latitude" validate:"required,latitude"`
	Longitude      float64 `json:"longitude" validate:"required,longitude"`
	IndoorLocation bool    `json:"indoor_location,omitempty" validate:"boolean"`
	Building       string  `json:"building,omitempty"`
	Floor          int     `json:"floor,omitempty" validate:"number"`
	Room           string  `json:"room,omitempty"`
	Description    string  `json:"location_description,omitempty"`
}

type JSONDTO map[string]interface{}
