package sensor

import "errors"

var (
	// CRUD-related errors
	ErrSensorNotFound      = errors.New("sensor not found")
	ErrSensorAlreadyExists = errors.New("sensor with this ID already exists")
	ErrSensorDeleteFailed  = errors.New("failed to delete sensor")
	ErrSensorUpdateFailed  = errors.New("failed to update sensor")

	// Validation-related errors
	ErrInvalidSensorID     = errors.New("invalid sensor ID")
	ErrInvalidSensorName   = errors.New("sensor name is required")
	ErrInvalidSensorType   = errors.New("sensor type is not supported")
	ErrInvalidSensorStatus = errors.New("sensor status is not valid")
	ErrInvalidUnit         = errors.New("unit of measurement is required")
	ErrInvalidPrecision    = errors.New("precision must be non-negative")
	ErrInvalidLocation     = errors.New("location is required")

	// Filter/Search errors
	ErrNoSensorsFound = errors.New("no sensors matched the filter criteria")

	// Ownership/Association errors
	ErrDeviceNotFound      = errors.New("associated device not found")
	ErrSensorNotAssociated = errors.New("sensor is not associated with the specified device")

	// Database/Infrastructure
	ErrSensorStorageFailure = errors.New("sensor storage failed")
	ErrSensorLoadFailure    = errors.New("failed to load sensor from database")
)
