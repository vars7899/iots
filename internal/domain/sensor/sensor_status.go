package sensor

type SensorStatus string

const (
	SensorStatusOnline           SensorStatus = "online"            // Sensor is functioning and connected
	SensorStatusOffline          SensorStatus = "offline"           // Sensor is disconnected or unreachable
	SensorStatusError            SensorStatus = "error"             // Sensor has encountered an issue
	SensorStatusIdle             SensorStatus = "idle"              // Sensor is powered on but not actively reporting
	SensorStatusUnderCalibration SensorStatus = "under_calibration" // Sensor is undergoing calibration
	SensorStatusUnderMaintenance SensorStatus = "under_maintenance" // Sensor is under scheduled maintenance
)

func (s SensorStatus) IsValid() bool {
	switch s {
	case SensorStatusOnline, SensorStatusOffline, SensorStatusError, SensorStatusIdle, SensorStatusUnderCalibration, SensorStatusUnderMaintenance:
		return true
	default:
		return false
	}
}
