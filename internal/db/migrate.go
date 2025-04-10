package db

import (
	"github.com/vars7899/iots/internal/domain/sensor"
	"github.com/vars7899/iots/internal/domain/user"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&sensor.Sensor{}, &user.User{})
}
