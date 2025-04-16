package validatorutils

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	UuidValidationRule       = []validation.Rule{validation.Required, validation.Length(36, 36), is.UUID}
	UnitValidationRules      = []validation.Rule{validation.Required, validation.Length(1, 10)}
	PrecisionValidationRules = []validation.Rule{validation.Required, validation.Min(0), validation.Max(10)}
)

func ApplyRulesWhenValueIsPresent(rules []validation.Rule) []validation.Rule {
	return []validation.Rule{
		validation.By(func(value interface{}) error {
			// if value not present skip
			if value == nil {
				return nil
			}
			if str, ok := value.(string); ok && str == "" {
				return nil
			}
			// if value present check value for rules
			for _, rule := range rules {
				if err := rule.Validate(value); err != nil {
					return err
				}
			}
			return nil
		}),
	}
}

func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func IsPgDuplicateKeyError(err error) *pgconn.PgError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return pgErr // Unique violation code
		}
	}
	return nil
}
