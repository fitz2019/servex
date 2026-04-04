package eventsourcing

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EventStore 事件存储接口.
type EventStore interface {
	// Save 批量保存事件.
	Save(ctx context.Context, events []Event) error
	// Load 从指定版本加载事件.
	Load(ctx context.Context, aggregateID string, fromVersion int64) ([]Event, error)
	// LoadAll 加载聚合的全部事件.
	LoadAll(ctx context.Context, aggregateID string) ([]Event, error)
}

// SnapshotStore 快照存储接口.
type SnapshotStore interface {
	// Save 保存快照（upsert 语义）.
	Save(ctx context.Context, snapshot Snapshot) error
	// Load 加载最新快照.
	Load(ctx context.Context, aggregateID string) (*Snapshot, error)
}

// GORMEventStore 基于 GORM 的事件存储实现.
type GORMEventStore struct {
	db *gorm.DB
}

// 编译期接口合规检查.
var _ EventStore = (*GORMEventStore)(nil)

// NewGORMEventStore 创建 GORM 事件存储.
func NewGORMEventStore(db *gorm.DB) *GORMEventStore {
	return &GORMEventStore{db: db}
}

// AutoMigrate 自动迁移 events 表.
func (s *GORMEventStore) AutoMigrate() error {
	return s.db.AutoMigrate(&Event{})
}

// Save 批量保存事件.
//
// 利用 (aggregate_id, aggregate_type, version) 唯一索引实现乐观并发控制，
// 当版本冲突时返回 ErrConcurrencyConflict.
func (s *GORMEventStore) Save(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return ErrNoEvents
	}

	err := s.db.WithContext(ctx).Create(&events).Error
	if err != nil && isConcurrencyError(err) {
		return ErrConcurrencyConflict
	}
	return err
}

// Load 从指定版本之后加载事件.
func (s *GORMEventStore) Load(ctx context.Context, aggregateID string, fromVersion int64) ([]Event, error) {
	var events []Event
	err := s.db.WithContext(ctx).
		Where("aggregate_id = ? AND version > ?", aggregateID, fromVersion).
		Order("version ASC").
		Find(&events).Error
	return events, err
}

// LoadAll 加载聚合的全部事件.
func (s *GORMEventStore) LoadAll(ctx context.Context, aggregateID string) ([]Event, error) {
	var events []Event
	err := s.db.WithContext(ctx).
		Where("aggregate_id = ?", aggregateID).
		Order("version ASC").
		Find(&events).Error
	return events, err
}

// isConcurrencyError 检测唯一约束冲突错误.
func isConcurrencyError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	// SQLite / MySQL / PostgreSQL 唯一约束冲突关键词
	return strings.Contains(msg, "unique") ||
		strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "constraint")
}

// GORMSnapshotStore 基于 GORM 的快照存储实现.
type GORMSnapshotStore struct {
	db *gorm.DB
}

// 编译期接口合规检查.
var _ SnapshotStore = (*GORMSnapshotStore)(nil)

// NewGORMSnapshotStore 创建 GORM 快照存储.
func NewGORMSnapshotStore(db *gorm.DB) *GORMSnapshotStore {
	return &GORMSnapshotStore{db: db}
}

// AutoMigrate 自动迁移 snapshots 表.
func (s *GORMSnapshotStore) AutoMigrate() error {
	return s.db.AutoMigrate(&Snapshot{})
}

// Save 保存快照（upsert 语义）.
//
// 使用 ON CONFLICT UPDATE 实现 upsert.
func (s *GORMSnapshotStore) Save(ctx context.Context, snapshot Snapshot) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "aggregate_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"version", "data", "created_at"}),
		}).
		Create(&snapshot).Error
}

// Load 加载聚合的最新快照.
func (s *GORMSnapshotStore) Load(ctx context.Context, aggregateID string) (*Snapshot, error) {
	var snapshot Snapshot
	err := s.db.WithContext(ctx).
		Where("aggregate_id = ?", aggregateID).
		First(&snapshot).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &snapshot, nil
}
