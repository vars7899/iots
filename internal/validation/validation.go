package validation

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

var Validate *validator.Validate

func Init(baseLogger *zap.Logger) {
	log := logger.Named(baseLogger, "Validator")
	Validate = validator.New()
	registerCustomValidators()
	log.Info("validator initialized successfully")
}

func registerCustomValidators() {
	registerUUIDValidator()
	registerStatusValidator()
	registerConnectionTypeValidator()
}

func registerUUIDValidator() {
	// Register custom UUID converter
	Validate.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if field.Type() == reflect.TypeOf(uuid.UUID{}) {
			uuidVal := field.Interface().(uuid.UUID)
			if uuidVal == uuid.Nil {
				return ""
			}
			return uuidVal.String()
		}
		return nil
	}, uuid.UUID{})

	// Register the UUID validation function
	Validate.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		if str, ok := fl.Field().Interface().(string); ok {
			_, err := uuid.Parse(str)
			return err == nil
		}
		return false
	})
}

func registerStatusValidator() {
	// Register custom Status converter
	Validate.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if field.Type() == reflect.TypeOf(domain.Status("")) {
			return string(field.Interface().(domain.Status))
		}
		return nil
	}, domain.Status(""))

	// Register the Status validation function
	Validate.RegisterValidation("status", func(fl validator.FieldLevel) bool {
		statusStr, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}
		return domain.IsValidStatus(statusStr)
	})
}

func registerConnectionTypeValidator() {
	// Register custom ConnectionType converter
	Validate.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if field.Type() == reflect.TypeOf(domain.ConnectionType("")) {
			return string(field.Interface().(domain.ConnectionType))
		}
		return nil
	}, domain.ConnectionType(""))

	// Register the ConnectionType validation function
	Validate.RegisterValidation("connection_type", func(fl validator.FieldLevel) bool {
		connectionTypeStr, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}
		return domain.IsValidConnectionType(connectionTypeStr)
	})
}
