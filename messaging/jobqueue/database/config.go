package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewStoreFromConfig 根据驱动名、DSN 和表名创建 Database Store.
// driver 支持 "mysql"、"postgres"、"sqlite".
func NewStoreFromConfig(driver, dsn, table string) (*Store, error) {
	var dialector gorm.Dialector
	switch driver {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("jobqueue/database: 不支持的数据库驱动 %q", driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	var opts []Option
	if table != "" {
		opts = append(opts, WithTableName(table))
	}
	return NewStore(db, opts...)
}
