package validator

import (
	"fmt"

	"github.com/vars7899/iots/internal/domain/sensor"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func ValidateSensor(s *sensor.Sensor) error {
	return validation.ValidateStruct(s,
		validation.Field(&s.ID, validation.Required, validation.Length(36, 36), is.UUID),
		validation.Field(&s.DeviceID, validation.Required, validation.Length(36, 36), is.UUID),
		validation.Field(&s.Name, validation.Required, validation.Length(3, 50)),
		validation.Field(&s.Type, validation.Required, validation.By(ValidateSensorType)),
		validation.Field(&s.Status, validation.Required, validation.By(ValidateSensorStatus)),
		validation.Field(&s.Unit, validation.Required, validation.Length(1, 10)),
		validation.Field(&s.Precision, validation.Required, validation.Min(0), validation.Max(10)),
		validation.Field(&s.Location, validation.Required, validation.Length(3, 100)),
		validation.Field(&s.MetaData, validation.NotNil),
	)
}

func ValidateSensorType(value interface{}) error {
	if t, ok := value.(sensor.SensorType); ok {
		if !t.IsValid() {
			return fmt.Errorf("invalid sensor type: %s", t)
		}
		return nil
	}
	return fmt.Errorf("invalid type: %s", value)
}

func ValidateSensorStatus(value interface{}) error {
	if s, ok := value.(sensor.SensorStatus); ok {
		if !s.IsValid() {
			return fmt.Errorf("invalid sensor status: %s", s)
		}
		return nil
	}
	return fmt.Errorf("invalid status: %s", value)
}

func ValidateSensorID(sid string) error {
	if sid == "" {
		return sensor.ErrInvalidSensorID
	}
	return nil
}
