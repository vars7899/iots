package db

import (
	"gorm.io/gorm"
)

func ResetDatabase(db *gorm.DB) error {
	// Drop all tables
	err := db.Migrator().DropTable(DB_Tables...)
	if err != nil {
		return err
	}

	// Recreate tables
	return AutoMigrate(db)
}
