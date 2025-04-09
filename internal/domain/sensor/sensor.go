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
	ID        string         `gorm:"type:uuid;primaryKey" json:"id"`
	DeviceID  string         `gorm:"type:varchar(100);not null;index" json:"device_id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Type      SensorType     `gorm:"type:varchar(50);not null" json:"type"`
	Status    SensorStatus   `gorm:"type:varchar(50);not null" json:"status"`
	Unit      string         `gorm:"type:varchar(20)" json:"unit"`
	Precision int            `gorm:"type:int" json:"precision"`
	Location  string         `gorm:"type:varchar(255)" json:"location"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
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
