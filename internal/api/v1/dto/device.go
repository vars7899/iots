package dto

import (
	"github.com/lib/pq"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/validation"
	"gorm.io/datatypes"
)

type ProvisionDeviceRequest struct {
	DeviceID      string `json:"device_id" validate:"required,uuid"`
	ProvisionCode string `json:"provision_code" validate:"required"`
}

func (dto *ProvisionDeviceRequest) Validate() error { return validation.Validate.Struct(dto) }

type RefreshSessionTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// type CreateNewDeviceDTO struct {
// 	Name            string             `json:"name" validate:"required,max=255"`
// 	Description     string             `json:"description,omitempty" validate:"max=255"`
// 	Manufacturer    string             `json:"manufacturer" validate:"required"`
// 	ModelNumber     string             `json:"model_number" validate:"required"`
// 	SerialNumber    string             `json:"serial_number" validate:"required"`
// 	FirmwareVersion string             `json:"firmware_version" validate:"required"`
// 	IPAddress       string             `json:"ip_address" validate:"omitempty,ip"`
// 	MACAddress      string             `json:"mac_address" validate:"omitempty,max=17"`
// 	ConnectionType  string             `json:"connection_type" validate:"required,connection_type"`
// 	Location        GeoLocationDTO     `json:"location" validate:"required"`
// 	Metadata        *datatypes.JSON    `json:"metadata" validate:"omitempty"` // TODO: meta data can have maximum 100 fields
// 	Tags            pq.StringArray     `json:"tags" validate:"omitempty,dive,min=1"`
// 	Capabilities    pq.StringArray     `json:"capabilities" validate:"omitempty,dive,min=1"`
// 	TelemetryConfig TelemetryConfigDTO `json:"telemetry_config"`
// 	BroadcastConfig BroadcastConfigDTO `json:"broadcast_config"`
// }

// func (dto *CreateNewDeviceDTO) Validate() error { return validation.Validate.Struct(dto) }

// func (dto *CreateNewDeviceDTO) AsModel() *model.Device {
// 	device := &model.Device{
// 		Name:            dto.Name,
// 		Description:     dto.Description,
// 		Manufacturer:    dto.Manufacturer,
// 		ModelNumber:     dto.ModelNumber,
// 		SerialNumber:    dto.SerialNumber,
// 		FirmwareVersion: dto.FirmwareVersion,
// 		IPAddress:       dto.IPAddress,
// 		MACAddress:      dto.MACAddress,
// 		ConnectionType:  domain.ConnectionType(dto.ConnectionType),
// 		Location:        *dto.Location.AsModel(),
// 		Tags:            dto.Tags,
// 		Capabilities:    dto.Capabilities,
// 		TelemetryConfig: *dto.TelemetryConfig.AsModel(),
// 		BroadcastConfig: *dto.BroadcastConfig.AsModel(),
// 	}

// 	if dto.Metadata != nil {
// 		device.Metadata = *dto.Metadata
// 	}

// 	return device
// }

// type UpdateDeviceDTO struct {
// 	Name            *string `json:"name,omitempty" validate:"omitempty,max=255"`
// 	Description     *string `json:"description,omitempty" validate:"omitempty"`
// 	Manufacturer    *string `json:"manufacturer,omitempty" validate:"omitempty"`
// 	ModelNumber     *string `json:"model_number,omitempty" validate:"omitempty"`
// 	SerialNumber    *string `json:"serial_number,omitempty" validate:"omitempty"`
// 	FirmwareVersion *string `json:"firmware_version,omitempty" validate:"omitempty"`
// 	IPAddress       *string `json:"ip_address,omitempty" validate:"omitempty,ip"`
// 	MACAddress      *string `json:"mac_address,omitempty" validate:"omitempty,max=17"`
// 	ConnectionType  *string `json:"connection_type,omitempty" validate:"omitempty"`
// }

// func (dto *UpdateDeviceDTO) Validate() error { return validation.Validate.Struct(dto) }

// func (dto *UpdateDeviceDTO) ToDevice() *model.Device {
// 	device := &model.Device{}

// 	if dto.Name != nil {
// 		device.Name = *dto.Name
// 	}
// 	if dto.Description != nil {
// 		device.Description = *dto.Description
// 	}
// 	if dto.Manufacturer != nil {
// 		device.Manufacturer = *dto.Manufacturer
// 	}
// 	if dto.ModelNumber != nil {
// 		device.ModelNumber = *dto.ModelNumber
// 	}
// 	if dto.SerialNumber != nil {
// 		device.SerialNumber = *dto.SerialNumber
// 	}
// 	if dto.FirmwareVersion != nil {
// 		device.FirmwareVersion = *dto.FirmwareVersion
// 	}
// 	if dto.IPAddress != nil {
// 		device.IPAddress = *dto.IPAddress
// 	}
// 	if dto.MACAddress != nil {
// 		device.MACAddress = *dto.MACAddress
// 	}
// 	if dto.ConnectionType != nil {
// 		device.ConnectionType = domain.ConnectionType(*dto.ConnectionType)
// 	}

// 	return device
// }

// type BulkCreateDevicesDTO struct {
// 	Devices []CreateNewDeviceDTO `json:"devices" validate:"required,dive"`
// }

// func (dto *BulkCreateDevicesDTO) Validate() error {
// 	return validation.Validate.Struct(dto)
// }

// func (dto *BulkCreateDevicesDTO) ToDevices() []*model.Device {
// 	devices := make([]*model.Device, 0, len(dto.Devices))
// 	for _, d := range dto.Devices {
// 		devices = append(devices, d.AsModel())
// 	}
// 	return devices
// }

// type BulkDeleteDeviceDTO struct {
// 	DeviceIDs []string `json:"device_ids" validate:"required,min=1,dive,required,uuid"`
// }

// // After validation, convert strings to UUIDs
// func (dto *BulkDeleteDeviceDTO) ToUUIDs() ([]uuid.UUID, error) {
// 	uuids := make([]uuid.UUID, len(dto.DeviceIDs))
// 	for i, id := range dto.DeviceIDs {
// 		uid, err := uuid.Parse(id)
// 		if err != nil {
// 			return nil, err
// 		}
// 		uuids[i] = uid
// 	}
// 	return uuids, nil
// }

// func (dto *BulkDeleteDeviceDTO) Validate() error {
// 	err := validation.Validate.Struct(dto)
// 	if err != nil {
// 		// Handle validation errors
// 		if validationErrors, ok := err.(validator.ValidationErrors); ok {
// 			// You can format validation errors here
// 			return fmt.Errorf("validation error: %v", validationErrors)
// 		}
// 		return err
// 	}

// 	// Additional custom validation logic
// 	if len(dto.DeviceIDs) == 0 {
// 		return fmt.Errorf("at least one device ID is required")
// 	}

// 	// Check for nil UUIDs
// 	for i, id := range dto.DeviceIDs {
// 		id, _ := uuid.Parse(id)
// 		if id == uuid.Nil {
// 			return fmt.Errorf("device ID at position %d cannot be nil", i)
// 		}
// 	}

// 	return nil
// }

// type BulkUpdateDeviceDTO struct {
// 	Devices []UpdateDeviceDTO `json:"devices" validate:"required,dive"`
// }

// func (dto *BulkUpdateDeviceDTO) Validate() error {
// 	return validation.Validate.Struct(dto)
// }

// func (dto *BulkUpdateDeviceDTO) ToDevices() []*model.Device {
// 	devices := make([]*model.Device, 0, len(dto.Devices))
// 	for _, d := range dto.Devices {
// 		devices = append(devices, d.ToDevice())
// 	}
// 	return devices
// }

// type UpdateDeviceStatusDTO struct {
// 	Status string `json:"status" validate:"required,status"`
// }

// func (dto *UpdateDeviceStatusDTO) Validate() error {
// 	return validation.Validate.Struct(dto)
// }

// type BroadcastConfigDTO struct {
// 	BroadcastEnabled bool   `json:"broadcast_enabled" validate:"boolean"`
// 	Protocol         string `json:"protocol" validate:"required,oneof=MQTT AMQP"`
// 	BrokerURL        string `json:"broker_url" validate:"required,url"`
// 	Topic            string `json:"topic" validate:"required"`
// 	QoS              int    `json:"qos" validate:"oneof=0 1 2"`
// 	RetainMessages   bool   `json:"retain_messages" validate:"boolean"`
// 	ClientID         string `json:"client_id" validate:"required"`
// 	Username         string `json:"username"  validate:"required"`
// 	Password         string `json:"password"  validate:"required"`
// 	CertificatePath  string `json:"certificate_path,omitempty"`
// 	PrivateKeyPath   string `json:"private_key_path,omitempty"`
// }

// func (dto *BroadcastConfigDTO) Validate() error { return validation.Validate.Struct(dto) }

// func (dto *BroadcastConfigDTO) AsModel() *model.BroadcastConfig {
// 	return &model.BroadcastConfig{
// 		BroadcastEnabled: dto.BroadcastEnabled,
// 		Protocol:         dto.Protocol,
// 		BrokerURL:        dto.BrokerURL,
// 		Topic:            dto.Topic,
// 		QoS:              dto.QoS,
// 		RetainMessages:   dto.RetainMessages,
// 		ClientID:         dto.ClientID,
// 		Username:         dto.Username,
// 		Password:         dto.Password,
// 		CertificatePath:  dto.CertificatePath,
// 		PrivateKeyPath:   dto.PrivateKeyPath,
// 	}
// }

// type TelemetryConfigDTO struct {
// 	Enabled            bool           `json:"enabled"`
// 	ReportingFrequency int            `json:"reporting_frequency_seconds"`
// 	BatchSize          int            `json:"batch_size"`
// 	RetentionPeriod    int            `json:"retention_period_days"`
// 	StorageQuota       int64          `json:"storage_quota_bytes"`
// 	CompressionEnabled bool           `json:"compression_enabled"`
// 	EncryptionEnabled  bool           `json:"encryption_enabled"`
// 	AlertThresholds    datatypes.JSON `json:"alert_thresholds"`
// }

// func (dto *TelemetryConfigDTO) Validate() error { return validation.Validate.Struct(dto) }

// func (dto *TelemetryConfigDTO) AsModel() *model.TelemetryConfig {
// 	return &model.TelemetryConfig{
// 		Enabled:            dto.Enabled,
// 		ReportingFrequency: dto.ReportingFrequency,
// 		BatchSize:          dto.BatchSize,
// 		RetentionPeriod:    dto.RetentionPeriod,
// 		StorageQuota:       dto.StorageQuota,
// 		CompressionEnabled: dto.CompressionEnabled,
// 		EncryptionEnabled:  dto.EncryptionEnabled,
// 		AlertThresholds:    dto.AlertThresholds,
// 	}
// }

type LinkDeviceRequestDTO struct {
	DeviceID       string `json:"device_id" validate:"required,uuid"`
	DeviceLinkCode string `json:"link_code" validate:"required"`
}

func (dto *LinkDeviceRequestDTO) Validate() error { return validation.Validate.Struct(dto) }

type RegisterDeviceRequest struct {
	Name            string                     `json:"name" validate:"required"`
	Description     string                     `json:"description"`
	MACAddress      string                     `json:"mac_address" validate:"required,mac"`
	IPAddress       string                     `json:"ip_address" validate:"omitempty,ip"`
	Specification   DeviceSpecificationPayload `json:"specification" validate:"required"`
	TelemetryConfig *TelemetryConfigPayload    `json:"telemetry_config,omitempty"`
	BroadcastConfig *BroadcastConfigPayload    `json:"broadcast_config,omitempty"`
	Metadata        datatypes.JSON             `json:"metadata,omitempty"`
	Tags            pq.StringArray             `json:"tags,omitempty"`
	Capabilities    pq.StringArray             `json:"capabilities,omitempty"`
}

func (dto *RegisterDeviceRequest) Validate() error { return validation.Validate.Struct(dto) }

func (dto *RegisterDeviceRequest) AsModel() *model.Device {
	return &model.Device{
		Name:          dto.Name,
		Description:   dto.Description,
		MACAddress:    dto.MACAddress,
		IPAddress:     dto.IPAddress,
		Specification: *dto.Specification.AsModel(),
		// TelemetryConfig: ,
		Metadata:     dto.Metadata,
		Tags:         dto.Tags,
		Capabilities: dto.Capabilities,
	}
}

type DeviceSpecificationPayload struct {
	Manufacturer    string `json:"manufacturer" validate:"required"`
	ModelNumber     string `json:"model_number" validate:"required"`
	SerialNumber    string `json:"serial_number" validate:"required"`
	FirmwareVersion string `json:"firmware_version,omitempty"`
	HardwareVersion string `json:"hardware_version,omitempty"`
	SoftwareVersion string `json:"software_version,omitempty"`
}

func (dto *DeviceSpecificationPayload) Validate() error { return validation.Validate.Struct(dto) }

func (dto *DeviceSpecificationPayload) AsModel() *model.DeviceSpecification {
	return &model.DeviceSpecification{
		Manufacturer:    dto.Manufacturer,
		ModelNumber:     dto.ModelNumber,
		SerialNumber:    dto.SerialNumber,
		FirmwareVersion: dto.FirmwareVersion,
		HardwareVersion: dto.HardwareVersion,
		SoftwareVersion: dto.SoftwareVersion,
	}
}

type TelemetryConfigPayload struct {
	Enabled            bool                   `json:"enabled"`
	ReportingFrequency int                    `json:"reporting_frequency_seconds"`
	BatchSize          int                    `json:"batch_size"`
	RetentionPeriod    int                    `json:"retention_period_days"`
	StorageQuota       int64                  `json:"storage_quota_bytes"`
	CompressionEnabled bool                   `json:"compression_enabled"`
	EncryptionEnabled  bool                   `json:"encryption_enabled"`
	AlertThresholds    map[string]interface{} `json:"alert_thresholds,omitempty"`
}

func (dto *TelemetryConfigPayload) Validate() error { return validation.Validate.Struct(dto) }

// func (dto *TelemetryConfigPayload) AsModel() *model.TelemetryConfig {
// 	return &model.TelemetryConfig{
// 		Enabled: dto,
// 	}
// }

type BroadcastConfigPayload struct {
	BroadcastEnabled bool   `json:"broadcast_enabled"`
	Protocol         string `json:"protocol,omitempty"` // MQTT, AMQP
	BrokerURL        string `json:"broker_url,omitempty"`
	Topic            string `json:"topic,omitempty"`
	QoS              int    `json:"qos,omitempty"`
	RetainMessages   bool   `json:"retain_messages,omitempty"`
	ClientID         string `json:"client_id,omitempty"`
	Username         string `json:"username,omitempty"`
	Password         string `json:"password,omitempty"` // handle securely!
	CertificatePath  string `json:"certificate_path,omitempty"`
	PrivateKeyPath   string `json:"private_key_path,omitempty"`
}

func (dto *BroadcastConfigPayload) Validate() error { return validation.Validate.Struct(dto) }
