package validatorz

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var Validate *validator.Validate

func InitValidator() {
	Validate = validator.New()
	registerCustomValidators()
}

func registerCustomValidators() {
	// Register custom UUID validator
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

	Validate.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		if str, ok := fl.Field().Interface().(string); ok {
			_, err := uuid.Parse(str)
			return err == nil
		}
		return false
	})
}
