package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Telemetry struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid();" json:"id"`
	DeviceID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"device_id"`
	Device    Device         `gorm:"foreignKey:DeviceID;references:ID;constraints:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	SensorID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"sensor_id"`
	Sensor    Sensor         `gorm:"foreignKey:SensorID;references:ID;constraints:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Timestamp time.Time      `gorm:"not null;index" json:"timestamp"`
	Data      datatypes.JSON `gorm:"type:jsonb" json:"data"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}
