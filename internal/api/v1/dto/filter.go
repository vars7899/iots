package dto

import (
	"time"

	"github.com/google/uuid"
)

type SensorFilter struct {
	ID        *string    `query:"id"`
	DeviceID  *string    `query:"device_id"`
	Name      *string    `query:"name"`
	Type      *string    `query:"type"`
	Status    *string    `query:"status"`
	Unit      *string    `query:"unit"`
	Precision *int       `query:"precision"`
	Location  *string    `query:"location"`
	CreatedAt *time.Time `query:"created_at"`
	UpdatedAt *time.Time `query:"updated_at"`
}

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
