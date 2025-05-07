package domain

type DeviceFilter struct {
	Name   *string `query:"name"`
	Status *string `query:"status"`
}
