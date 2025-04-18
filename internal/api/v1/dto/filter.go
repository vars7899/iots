package dto

import "time"

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
