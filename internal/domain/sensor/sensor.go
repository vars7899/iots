package sensor

import (
	"time"

	"gorm.io/gorm"
)

type SensorID string

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
	ID        SensorID
	DeviceID  string
	Name      string
	Type      SensorType
	Status    SensorStatus
	Unit      string
	Precision int
	Location  string
	MetaData  map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}
