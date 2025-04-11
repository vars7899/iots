package db

import (
	"github.com/vars7899/iots/internal/domain/sensor"
	"github.com/vars7899/iots/internal/domain/user"
	"gorm.io/gorm"
)

var DB_Tables = []interface{}{
	&user.User{},
	&user.Role{},
	&user.Permission{},
	&user.UserRole{},
	&user.RolePermission{},
	&sensor.Sensor{},
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(DB_Tables...)
}
