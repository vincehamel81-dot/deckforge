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

	if driver == "sqlite" || driver == "" {
		// SQLite does not support concurrent writers. A single connection serialises
		// all access and prevents "database is locked" errors under concurrent HTTP requests.
		sqlDB, err := db.DB()
		if err != nil {
			return nil, err
		}
		sqlDB.SetMaxOpenConns(1)
		// WAL mode allows concurrent readers alongside the single writer.
		db.Exec("PRAGMA journal_mode=WAL;")
		// Wait up to 5 s instead of failing immediately on lock contention.
		db.Exec("PRAGMA busy_timeout=5000;")
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
