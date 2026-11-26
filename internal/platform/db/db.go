package db

import (
	"fmt"
	"strings"

	"github.com/JINLINBI/pitaya-server/internal/domain/account"
	"github.com/JINLINBI/pitaya-server/internal/domain/mailbox"
	"github.com/JINLINBI/pitaya-server/internal/platform/env"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 按 Settings 打开 GORM 并迁移业务表。
func Open(s env.Settings) (*gorm.DB, error) {
	dialector, err := dialectorFor(s)
	if err != nil {
		return nil, err
	}
	gdb, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}
	if err := gdb.AutoMigrate(&account.Entity{}, &mailbox.Message{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	return gdb, nil
}

func dialectorFor(s env.Settings) (gorm.Dialector, error) {
	driver := strings.ToLower(s.DatabaseDriver)
	dsn := s.DatabaseDSN
	switch driver {
	case "sqlite":
		return sqlite.Open(dsn), nil
	case "postgres", "postgresql":
		return postgres.Open(dsn), nil
	case "mysql":
		return mysql.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported DATABASE_DRIVER %q", s.DatabaseDriver)
	}
}