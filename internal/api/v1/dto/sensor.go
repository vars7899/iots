package dto

import (
	"strconv"
	"strings"
	"time"

	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/validation"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/pagination"
	"gorm.io/datatypes"
)

type CreateSensorDTO struct {
	DeviceID  string          `json:"device_id" validate:"required,uuid"`
	Name      string          `json:"name" validate:"required,max=255"`
	Type      string          `json:"type" validate:"required,oneof=temperature humidity pressure level flow"` // TODO: update when adding more sensor type
	Status    string          `json:"status,omitempty" validate:"omitempty"`
	Unit      string          `json:"unit" validate:"omitempty,max=20"`
	Precision int             `json:"precision" validate:"omitempty,min=0,max=10"`
	Location  string          `json:"location" validate:"omitempty,max=255"`
	Metadata  *datatypes.JSON `json:"metadata" validate:"omitempty"`
}

func (dto *CreateSensorDTO) Validate() error {
	return validation.Validate.Struct(dto)
}

func (dto *CreateSensorDTO) AsModel() *model.Sensor {
	sensor := &model.Sensor{
		DeviceID:  dto.DeviceID,
		Name:      dto.Name,
		Type:      model.SensorType(dto.Type),
		Status:    domain.Pending, // enforced for default create
		Unit:      dto.Unit,
		Precision: dto.Precision,
		Location:  dto.Location,
	}
	if dto.Metadata != nil {
		sensor.MetaData = *dto.Metadata
	}
	return sensor
}

type UpdateSensorDTO struct {
	DeviceID  *string         `json:"device_id,omitempty" validate:"omitempty,uuid"`
	Name      *string         `json:"name,omitempty" validate:"omitempty,max=255"`
	Type      *string         `json:"type,omitempty" validate:"omitempty,oneof=temperature humidity pressure level flow"` // TODO: update when adding more sensor type
	Status    *string         `json:"status,omitempty" validate:"omitempty,status"`
	Unit      *string         `json:"unit,omitempty" validate:"omitempty,max=20"`
	Precision *int            `json:"precision,omitempty" validate:"omitempty,max=20"`
	Location  *string         `json:"location,omitempty" validate:"omitempty,max=255"`
	Metadata  *datatypes.JSON `json:"metadata" validate:"omitempty"`
}

func (dto *UpdateSensorDTO) Validate() error {
	return validation.Validate.Struct(dto)
}

func (dto UpdateSensorDTO) AsModel() *model.Sensor {
	sensor := model.Sensor{}
	if dto.DeviceID != nil {
		sensor.DeviceID = *dto.DeviceID
	}
	if dto.Name != nil {
		sensor.Name = *dto.Name
	}
	if dto.Type != nil {
		sensor.Type = model.SensorType(*dto.Type)
	}
	if dto.Status != nil {
		sensor.Status = domain.Status(*dto.Status)
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
	if dto.Metadata != nil {
		sensor.MetaData = *dto.Metadata
	}

	return &sensor
}

type SensorQueryParamsDTO struct {
	ID        *string `query:"id"`
	DeviceID  *string `query:"device_id"`
	Name      *string `query:"name"`
	Type      *string `query:"type"`
	Status    *string `query:"status"`
	Unit      *string `query:"unit"`
	Precision *string `query:"precision"`
	Location  *string `query:"location"`
	CreatedAt *string `query:"created_at"`
	UpdatedAt *string `query:"updated_at"`
	Limit     int     `query:"limit"`
	Offset    int     `query:"offset"`
	SortBy    string  `query:"sort_by"`
	SortOrder string  `query:"sort_order"`
}

func (dto *SensorQueryParamsDTO) Validate() error {
	return validation.Validate.Struct(dto)
}

func (dto *SensorQueryParamsDTO) AsModel() (*pagination.Pagination, *SensorFilter, error) {
	const defaultLimit = 20
	const maxLimit = 100

	var paginationConfig *pagination.Pagination
	hasPagination := dto.Limit > 0 || dto.Offset > 0

	if hasPagination {
		// Sanitize Limit
		limit := dto.Limit
		if limit <= 0 {
			limit = defaultLimit
		} else if limit > maxLimit {
			return nil, nil, apperror.ErrBadRequest.WithMessagef("limit exceeds maximum allowed %d", maxLimit)
		}

		// Sanitize Offset
		if dto.Offset < 0 {
			return nil, nil, apperror.ErrBadRequest.WithMessagef("offset cannot be negative %d", dto.Offset)
		}

		// Calculate Page from Offset
		page := 1
		if dto.Offset > 0 {
			page = (dto.Offset / limit) + 1
		}

		// Sanitize Sort
		sortBy := dto.SortBy
		if sortBy == "" {
			sortBy = "created_at"
		}
		sortOrder := strings.ToUpper(dto.SortOrder)
		if sortOrder != "ASC" && sortOrder != "DESC" {
			sortOrder = "DESC"
		}

		paginationConfig = &pagination.Pagination{
			Page:      page,
			PageSize:  limit,
			SortBy:    sortBy,
			SortOrder: sortOrder,
		}
	}

	filter := &SensorFilter{}

	// Apply filters only if set and non-empty
	if dto.ID != nil && *dto.ID != "" {
		filter.ID = dto.ID
	}
	if dto.DeviceID != nil && *dto.DeviceID != "" {
		filter.DeviceID = dto.DeviceID
	}
	if dto.Name != nil && *dto.Name != "" {
		filter.Name = dto.Name
	}
	if dto.Type != nil && *dto.Type != "" {
		filter.Type = dto.Type
	}
	if dto.Status != nil && *dto.Status != "" {
		filter.Status = dto.Status
	}
	if dto.Unit != nil && *dto.Unit != "" {
		filter.Unit = dto.Unit
	}
	if dto.Precision != nil && *dto.Precision != "" {
		if p, err := strconv.Atoi(*dto.Precision); err == nil {
			filter.Precision = &p
		} else {
			return nil, nil, apperror.ErrBadRequest.WithMessage("invalid precision value")
		}
	}
	if dto.CreatedAt != nil && *dto.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, *dto.CreatedAt); err == nil {
			filter.CreatedAt = &t
		} else {
			return nil, nil, apperror.ErrBadRequest.WithMessage("invalid created_at format (expected RFC3339)")
		}
	}
	if dto.UpdatedAt != nil && *dto.UpdatedAt != "" {
		if t, err := time.Parse(time.RFC3339, *dto.UpdatedAt); err == nil {
			filter.UpdatedAt = &t
		} else {
			return nil, nil, apperror.ErrBadRequest.WithMessage("invalid updated_at format (expected RFC3339)")
		}
	}

	return paginationConfig, filter, nil
}
