package sensor

import "time"

type SensorFilter struct {
	DeviceID  *string
	Name      *string
	Type      *string
	Status    *string
	Unit      *string
	Precision *int
	Location  *string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

var AllowedSortFields = map[string]bool{
	"id":         true,
	"device_id":  true,
	"name":       true,
	"type":       true,
	"status":     true,
	"unit":       true,
	"precision":  true,
	"location":   true,
	"created_at": true,
	"updated_at": true,
}

var AllowedSortOrders = map[string]bool{
	"asc":  true,
	"desc": true,
}
