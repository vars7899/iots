package sensor

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SensorType string

const (
	TemperatureSensor SensorType = "temperature_sensor"
	HumiditySensor    SensorType = "humidity_sensor"
	MotionSensor      SensorType = "motion_sensor"
)

func (t SensorType) IsValid() bool {
	switch t {
	case TemperatureSensor, HumiditySensor, MotionSensor:
		return true
	default:
		return false
	}
}

type Sensor struct {
	ID        string         `gorm:"type:uuid;primaryKey"`
	DeviceID  string         `gorm:"type:varchar(100);not null;index"`
	Name      string         `gorm:"type:varchar(255);not null"`
	Type      SensorType     `gorm:"type:varchar(50);not null"`
	Status    SensorStatus   `gorm:"type:varchar(50);not null"`
	Unit      string         `gorm:"type:varchar(20)"`
	Precision int            `gorm:"type:int"`
	Location  string         `gorm:"type:varchar(255)"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	// MetaData  map[string]interface{} `gorm:"type:jsonb"`
}

func (s *Sensor) StampNew() {
	s.ID = uuid.NewString()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
}

func (s *Sensor) StampUpdate() {
	s.UpdatedAt = time.Now()
}
