package db

import (
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/internal/domain/model"
)

var DB_Tables = []interface{}{
	&model.User{},
	&model.Role{},
	&model.Permission{},
	&model.UserRole{},
	&model.RolePermission{},
	&model.Device{},
	&model.Sensor{},
	&model.Telemetry{},
	&model.AccessGroup{},
	&model.DeviceEvent{},
	&domain.GeoLocation{},
}
