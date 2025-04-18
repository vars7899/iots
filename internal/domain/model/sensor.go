package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const SensorName = "sensor"

type Sensor struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DeviceID  string         `gorm:"type:varchar(100);not null;index" json:"device_id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Type      SensorType     `gorm:"type:varchar(50);not null" json:"type"`
	Status    domain.Status  `gorm:"type:varchar(50);not null" json:"status"`
	Unit      string         `gorm:"type:varchar(20)" json:"unit"`
	Precision int            `gorm:"type:int" json:"precision"`
	Location  string         `gorm:"type:varchar(255)" json:"location"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	MetaData  datatypes.JSON `gorm:"type:jsonb"`
}

type SensorType string

const (
	TemperatureSensor SensorType = "temperature"
	HumiditySensor    SensorType = "humidity"
	MotionSensor      SensorType = "motion"
)

func (t SensorType) IsValid() bool {
	switch t {
	case TemperatureSensor, HumiditySensor, MotionSensor:
		return true
	default:
		return false
	}
}
