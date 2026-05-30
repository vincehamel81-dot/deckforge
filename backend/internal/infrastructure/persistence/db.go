package persistence

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(driver, dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch driver {
	case "postgres":
		dialector = postgres.Open(dsn)
	case "sqlite", "":
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER %q: use 'sqlite' or 'postgres'", driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&UserModel{},
		&GameModel{},
		&ShoeCardModel{},
		&PlayerModel{},
	); err != nil {
		return nil, err
	}
	return db, nil
}
