package service

import (
	"context"
	"errors"

	"github.com/vars7899/iots/pkg/apperror"
)

func ServiceError(err error, args ...any) error {
	if err == nil {
		return nil
	}

	var contextMessage string
	var fallback apperror.ErrorCode = apperror.ErrCodeInternal

	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			contextMessage = v
		case apperror.ErrorCode:
			fallback = v
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		if contextMessage == "" {
			return apperror.ErrTimeout.Wrap(err)
		}
		return apperror.ErrTimeout.Wrap(err).WithMessage("operation timed out")
	}

	return apperror.WrapAppErrWithContext(err, contextMessage, fallback)
}
