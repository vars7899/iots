package model

import (
	"time"

	"github.com/google/uuid"
)

type DeviceStatus string

var (
	DevicePendingProvision DeviceStatus = "pending_provision"
	DeviceProvisioned      DeviceStatus = "provisioned"
	DeviceDecommissioned   DeviceStatus = "decommissioned"
	DeviceOnline           DeviceStatus = "online"
	DeviceOffline          DeviceStatus = "offline"
)

type Device struct {
	ID              uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Manufacturer    string       `json:"manufacturer" gorm:"type:varchar(255);not null"`
	HardwareModel   string       `json:"hardware_model" gorm:"type:varchar(255);not null"`
	FirmwareVersion string       `json:"firmware_version" gorm:"type:varchar(255);not null"`
	CreatedAt       time.Time    `json:"created_at" gorm:"autoCreateTime"`
	Status          DeviceStatus `json:"status" gorm:"type:varchar(30);not null;default:'pending_provision'"`
}

type DeviceManufacturingBatch struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BatchCode          string    `json:"batch_code" gorm:"type:varchar(255);not null;uniqueIndex"`
	FactoryCode        string    `json:"factory_code" gorm:"type:varchar(255);not null"`
	ProductionLineCode string    `json:"production_line_code" gorm:"type:varchar(255)"`
}
