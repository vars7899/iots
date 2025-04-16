package utils

import (
	"context"

	"github.com/vars7899/iots/pkg/apperror"
)

func ConvertVectorToPointerVector[T any](i []T) []*T {
	o := make([]*T, len(i))
	for idx := range i {
		o[idx] = &i[idx]
	}
	return o
}

// TODO: might remove this later
func CheckContextForError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return apperror.ErrContextCancelled.WithDetails(err)
	}
	return nil
}
