package dto

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/validation"
	"gorm.io/datatypes"
)

type TelemetryPayloadDTO struct {
	SensorID      string                 `json:"sensor_id" validate:"required,uuid"`
	SensorCode    string                 `json:"sensor_code" validate:"required"`
	SchemaVersion int                    `json:"schema_version" validate:"required"`
	Data          map[string]interface{} `json:"data" validate:"required"`
}

func (dto *TelemetryPayloadDTO) ValidateBasicStructure() error {
	return validation.Validate.Struct(dto)
}

func (dto *TelemetryPayloadDTO) ValidateAgainstSchema() error {
	var matchingSchema *config.SensorSchema
	for _, schema := range config.SensorSchemaRepository.Schema {
		if schema.SensorCode == dto.SensorCode && schema.SchemaVersion == dto.SchemaVersion {
			matchingSchema = &schema
			break
		}
	}
	if matchingSchema == nil {
		return fmt.Errorf("no schema found for sensor %d with %d", dto.SensorCode, dto.SchemaVersion)
	}

	fieldMap := make(map[int64]config.SensorField)
	for _, field := range matchingSchema.Fields {
		fieldMap[field.Code] = field
	}

	// validate data
	for key, value := range dto.Data {
		fieldCode, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid field code format: %s", key)
		}

		field, exists := fieldMap[fieldCode]
		if !exists {
			return fmt.Errorf("field with code %d is not defined in the schema", fieldCode)
		}

		// Validate data type
		if err := validateDataType(value, field.DataType); err != nil {
			return fmt.Errorf("field %s (code %d): %w", field.Name, fieldCode, err)
		}
	}
	return nil
}

func (dto *TelemetryPayloadDTO) AsModel() (*model.Telemetry, error) {
	sensorUUID, err := uuid.Parse(dto.SensorID)
	if err != nil {
		// Should ideally not happen if "uuid" validate tag worked, but defensive
		return nil, fmt.Errorf("failed to parse sensor_id as UUID: %w", err)
	}

	// Marshal the Data map into JSON for the datatypes.JSON field
	jsonData, err := json.Marshal(dto.Data)
	if err != nil {
		// This could happen if the map contains types that cannot be marshalled
		return nil, fmt.Errorf("failed to marshal Data map to JSON: %w", err)
	}

	return &model.Telemetry{
		SensorID:  sensorUUID,
		Timestamp: time.Now(),
		Data:      datatypes.JSON(jsonData),
	}, nil
}

func validateDataType(value interface{}, expectedType string) error {
	switch expectedType {
	case "float":
		switch v := value.(type) {
		case float32, float64:
			return nil
		case int, int8, int16, int32, int64:
			return nil
		default:
			return fmt.Errorf("expected float, got %T", v)
		}
	case "int":
		switch v := value.(type) {
		case int, int8, int16, int32, int64:
			return nil
		default:
			return fmt.Errorf("expected integer, got %T", v)
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	default:
		return fmt.Errorf("unsupported data types %s", expectedType)
	}
	return nil
}
