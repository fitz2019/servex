package migration

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// AppliedMigration 已应用的迁移记录.
type AppliedMigration struct {
	Version     int64     `gorm:"primaryKey"`
	Description string    `gorm:"type:varchar(255)"`
	AppliedAt   time.Time `gorm:"autoCreateTime"`
}

// TableName 指定表名.
func (AppliedMigration) TableName() string {
	return "schema_migrations"
}

// Store 迁移记录存储接口.
type Store interface {
	// Applied 获取所有已应用的迁移记录.
	Applied(ctx context.Context) ([]AppliedMigration, error)
	// Record 使用指定 db 记录一条迁移已应用.
	Record(ctx context.Context, db *gorm.DB, version int64, description string) error
	// Remove 使用指定 db 移除一条迁移记录.
	Remove(ctx context.Context, db *gorm.DB, version int64) error
	// AutoMigrate 自动创建迁移记录表.
	AutoMigrate(ctx context.Context) error
}

// gormStore 基于 GORM 的迁移记录存储实现.
type gormStore struct {
	db *gorm.DB
}

// newGORMStore 创建 GORM 存储实例.
func newGORMStore(db *gorm.DB) Store {
	return &gormStore{db: db}
}

// Applied 获取所有已应用的迁移记录.
func (s *gormStore) Applied(ctx context.Context) ([]AppliedMigration, error) {
	var records []AppliedMigration
	if err := s.db.WithContext(ctx).Order("version ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// Record 使用指定 db 记录一条迁移已应用.
func (s *gormStore) Record(ctx context.Context, db *gorm.DB, version int64, description string) error {
	record := AppliedMigration{
		Version:     version,
		Description: description,
		AppliedAt:   time.Now(),
	}
	return db.WithContext(ctx).Create(&record).Error
}

// Remove 使用指定 db 移除一条迁移记录.
func (s *gormStore) Remove(ctx context.Context, db *gorm.DB, version int64) error {
	result := db.WithContext(ctx).Delete(&AppliedMigration{}, "version = ?", version)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotApplied
	}
	return nil
}

// AutoMigrate 自动创建迁移记录表.
func (s *gormStore) AutoMigrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&AppliedMigration{})
}
