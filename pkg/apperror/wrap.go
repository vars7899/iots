package apperror

func WrapAppErrWithContext(err error, contextMessage string, fallbackCode ErrorCode) error {
	if err == nil {
		return nil
	}
	if appError, ok := err.(*AppError); ok {
		return appError.WithMessage(contextMessage)
	}

	return New(fallbackCode).WithMessage(contextMessage).Wrap(err)
}
