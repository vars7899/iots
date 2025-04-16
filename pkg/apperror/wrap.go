package apperror

import "errors"

func WrapAppErrWithContext(err error, contextMessage string, fallbackCode ErrorCode) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		if contextMessage != "" {
			return appErr.WithMessagef("%s: %s", contextMessage, appErr.Message)
		}
		return appErr
	}

	newAppErr := New(fallbackCode)
	if contextMessage != "" {
		newAppErr = newAppErr.WithMessage(contextMessage)
	}
	return newAppErr.Wrap(err)
}
