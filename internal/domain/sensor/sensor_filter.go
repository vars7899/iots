package sensor

import "github.com/vars7899/iots/internal/domain/device"

type SensorFilter struct {
	DeviceID  *device.DeviceID
	Type      *SensorType
	Status    *SensorStatus
	CreatedAt *string
}
