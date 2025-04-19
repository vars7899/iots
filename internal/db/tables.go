package db

import (
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/domain/user"
)

var DB_Tables = []interface{}{
	&user.User{},
	&user.Role{},
	&user.Permission{},
	&user.UserRole{},
	&user.RolePermission{},
	&model.Sensor{},
	&model.Device{},
}
