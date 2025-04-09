package user

import (
	"time"

	"github.com/google/uuid"
)

type UserFilter struct {
	ID           *uuid.UUID
	Username     *string
	Email        *string
	PhoneNumber  *string
	IsActive     *bool
	CreatedBy    *uuid.UUID
	CreatedAt    *time.Time
	CreatedAtGTE *time.Time
	CreatedAtLTE *time.Time
	UpdatedAt    *time.Time
	UpdatedAtGTE *time.Time
	UpdatedAtLTE *time.Time
	Limit        int
	Offset       int
	SortBy       string
	SortOrder    string // "ASC" or "DESC"
}
