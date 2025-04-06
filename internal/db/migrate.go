package db

import (
	"github.com/vars7899/iots/internal/domain/sensor"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&sensor.Sensor{})
}
