package db

func ResetDatabase(gormDB *GormDB) error {
	if err := gormDB.db.Migrator().DropTable(DB_Tables...); err != nil {
		return err
	}
	return gormDB.AutoMigrateAll()
}
