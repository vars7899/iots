package sensor

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/vars7899/iots/pkg/validatorutils"
)

var (
	nameValidationRules     = []validation.Rule{validation.Required, validation.Length(3, 50)}
	typeValidationRules     = []validation.Rule{validation.Required, validation.By(ValidateSensorType)}
	statusValidationRules   = []validation.Rule{validation.Required, validation.By(ValidateSensorStatus)}
	locationValidationRules = []validation.Rule{validation.Required, validation.Length(3, 100)}

	// Rules for optional fields (used in update DTO)
	optionalDeviceIDValidationRules  = validatorutils.ApplyRulesWhenValueIsPresent(validatorutils.UuidValidationRule[1:])
	optionalNameValidationRules      = validatorutils.ApplyRulesWhenValueIsPresent(nameValidationRules[1:])
	optionalTypeValidationRules      = validatorutils.ApplyRulesWhenValueIsPresent(typeValidationRules[1:])
	optionalStatusValidationRules    = validatorutils.ApplyRulesWhenValueIsPresent(statusValidationRules[1:])
	optionalUnitValidationRules      = validatorutils.ApplyRulesWhenValueIsPresent(validatorutils.UnitValidationRules[1:])
	optionalPrecisionValidationRules = validatorutils.ApplyRulesWhenValueIsPresent(validatorutils.PrecisionValidationRules[1:])
	optionalLocationValidationRules  = validatorutils.ApplyRulesWhenValueIsPresent(locationValidationRules[1:])
)

func ValidateCreateSensorDTO(dto *CreateSensorDTO) error {
	return validation.ValidateStruct(dto,
		validation.Field(&dto.DeviceID, validatorutils.UuidValidationRule...),
		validation.Field(&dto.Name, nameValidationRules...),
		validation.Field(&dto.Type, typeValidationRules...),
		validation.Field(&dto.Status, statusValidationRules...),
		validation.Field(&dto.Unit, validatorutils.UnitValidationRules...),
		validation.Field(&dto.Precision, validatorutils.PrecisionValidationRules...),
		validation.Field(&dto.Location, locationValidationRules...),
	)
}

func ValidateUpdateSensorDTO(dto *UpdateSensorDTO) error {
	return validation.ValidateStruct(dto,
		validation.Field(&dto.DeviceID, optionalDeviceIDValidationRules...),
		validation.Field(&dto.Name, optionalNameValidationRules...),
		validation.Field(&dto.Type, optionalTypeValidationRules...),
		validation.Field(&dto.Status, optionalStatusValidationRules...),
		validation.Field(&dto.Unit, optionalUnitValidationRules...),
		validation.Field(&dto.Precision, optionalPrecisionValidationRules...),
		validation.Field(&dto.Location, optionalLocationValidationRules...),
	)
}

// func ValidateSensorType(value interface{}) error {
// 	if t, ok := value.(string); ok {
// 		sensorType := SensorType(t)
// 		if !sensorType.IsValid() {
// 			return fmt.Errorf("invalid sensor type: %s", t)
// 		}
// 		return nil
// 	}
// 	return fmt.Errorf("invalid type: %T", value)
// }

// func ValidateSensorStatus(value interface{}) error {
// 	if s, ok := value.(string); ok {
// 		sensorStatus := SensorStatus(s)
// 		if !sensorStatus.IsValid() {
// 			return fmt.Errorf("invalid sensor status: %s", s)
// 		}
// 		return nil
// 	}
// 	return fmt.Errorf("invalid status: %T", value)
// }

func ValidateSensorType(value interface{}) error {
	var str string

	switch v := value.(type) {
	case string:
		str = v
	case *string:
		if v == nil || *v == "" {
			return nil // optional and not present
		}
		str = *v
	default:
		return fmt.Errorf("invalid type: %T", value)
	}

	if !SensorType(str).IsValid() {
		return fmt.Errorf("invalid sensor type: %s", str)
	}
	return nil
}

func ValidateSensorStatus(value interface{}) error {
	var str string

	switch v := value.(type) {
	case string:
		str = v
	case *string:
		if v == nil || *v == "" {
			return nil // optional and not present
		}
		str = *v
	default:
		return fmt.Errorf("invalid status: %T", value)
	}

	if !SensorStatus(str).IsValid() {
		return fmt.Errorf("invalid sensor status: %s", str)
	}
	return nil
}
