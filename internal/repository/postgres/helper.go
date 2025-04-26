package postgres

import "github.com/vars7899/iots/pkg/apperror"

func notFoundErr(entity string, operation string) error {
	return apperror.ErrNotFound.WithMessagef("cannot %s %s: no matching record found", operation, entity)
}
