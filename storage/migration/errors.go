package migration

import "errors"

var (
	// ErrNilDB 数据库连接为空.
	ErrNilDB = errors.New("migration: db is nil")
	// ErrNilRegistry 迁移注册表为空.
	ErrNilRegistry = errors.New("migration: registry is nil")
	// ErrNilLogger 日志记录器为空.
	ErrNilLogger = errors.New("migration: logger is nil")
	// ErrNoMigrations 没有注册任何迁移.
	ErrNoMigrations = errors.New("migration: no migrations registered")
	// ErrAlreadyApplied 迁移已应用.
	ErrAlreadyApplied = errors.New("migration: already applied")
	// ErrNotApplied 迁移未应用.
	ErrNotApplied = errors.New("migration: not applied")
	// ErrVersionNotFound 指定版本不存在.
	ErrVersionNotFound = errors.New("migration: version not found")
)
