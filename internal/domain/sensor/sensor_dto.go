package sensor

import (
	"strconv"
	"time"
)

type CreateSensorDTO struct {
	DeviceID  string `json:"device_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Unit      string `json:"unit"`
	Precision int    `json:"precision"`
	Location  string `json:"location"`
}

func (dto *CreateSensorDTO) Validate() error {
	return ValidateCreateSensorDTO(dto)
}

func (dto CreateSensorDTO) ToSensorModel() Sensor {
	return Sensor{
		DeviceID:  dto.DeviceID,
		Name:      dto.Name,
		Type:      SensorType(dto.Type),
		Status:    SensorStatus(dto.Status),
		Unit:      dto.Unit,
		Precision: dto.Precision,
		Location:  dto.Location,
	}
}

type UpdateSensorDTO struct {
	DeviceID  *string `json:"device_id,omitempty"`
	Name      *string `json:"name,omitempty"`
	Type      *string `json:"type,omitempty"`
	Status    *string `json:"status,omitempty"`
	Unit      *string `json:"unit,omitempty"`
	Precision *int    `json:"precision,omitempty"`
	Location  *string `json:"location,omitempty"`
}

func (dto *UpdateSensorDTO) Validate() error {
	return ValidateUpdateSensorDTO(dto)
}

// ApplyUpdates applies the non-nil fields from UpdateSensorDTO to an existing sensor model
func (dto UpdateSensorDTO) ApplyUpdates(sensor *Sensor) {
	if dto.DeviceID != nil {
		sensor.DeviceID = *dto.DeviceID
	}
	if dto.Name != nil {
		sensor.Name = *dto.Name
	}
	if dto.Type != nil {
		sensor.Type = SensorType(*dto.Type)
	}
	if dto.Status != nil {
		sensor.Status = SensorStatus(*dto.Status)
	}
	if dto.Unit != nil {
		sensor.Unit = *dto.Unit
	}
	if dto.Precision != nil {
		sensor.Precision = *dto.Precision
	}
	if dto.Location != nil {
		sensor.Location = *dto.Location
	}

	// Always update the UpdatedAt field when changes are applied
	sensor.UpdatedAt = time.Now()
}

type SensorQueryParamsDTO struct {
	DeviceID  string `query:"device_id"`
	Name      string `query:"name"`
	Type      string `query:"type"`
	Status    string `query:"status"`
	Unit      string `query:"unit"`
	Precision string `query:"precision"`
	Location  string `query:"location"`
	CreatedAt string `query:"created_at"`
	UpdatedAt string `query:"updated_at"`
	Limit     int    `query:"limit"`
	Offset    int    `query:"offset"`
	SortBy    string `query:"sort_by"`
	SortOrder string `query:"sort_order"`
}

func (dto *SensorQueryParamsDTO) Validate() error {
	return ValidateSensorQueryParamsDTO(dto)
}

func (dto *SensorQueryParamsDTO) ToFilter() (SensorFilter, error) {
	filter := SensorFilter{
		Limit:     dto.Limit,
		Offset:    dto.Offset,
		SortBy:    dto.SortBy,
		SortOrder: dto.SortOrder,
	}

	if dto.DeviceID != "" {
		filter.DeviceID = &dto.DeviceID
	}
	if dto.Name != "" {
		filter.Name = &dto.Name
	}
	if dto.Type != "" {
		filter.Type = &dto.Type
	}
	if dto.Status != "" {
		filter.Status = &dto.Status
	}
	if dto.Unit != "" {
		filter.Unit = &dto.Unit
	}
	if dto.Precision != "" {
		if p, err := strconv.Atoi(dto.Precision); err == nil {
			filter.Precision = &p
		}
	}
	if dto.Location != "" {
		filter.Location = &dto.Location
	}
	if dto.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, dto.CreatedAt); err == nil {
			filter.CreatedAt = &t
		}
	}
	if dto.UpdatedAt != "" {
		if t, err := time.Parse(time.RFC3339, dto.UpdatedAt); err == nil {
			filter.UpdatedAt = &t
		}
	}
	return filter, nil
}
