package outbox

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Tsukikage7/servex/storage/rdbms"
)

// txContextKey 事务 context key 类型（避免与其他包冲突）.
type txContextKey struct{}

// InjectTx 将 GORM 事务注入 context，供 Save 读取.
func InjectTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// ExtractTx 从 context 中提取 GORM 事务.
func ExtractTx(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txContextKey{}).(*gorm.DB)
	return tx, ok && tx != nil
}

// TxFunc 在事务 context 中执行的函数.
type TxFunc func(ctx context.Context) error

// Store 发件箱存储接口.
type Store interface {
	// Save 保存消息.
	// 若 ctx 中通过 InjectTx 注入了事务，则在该事务中保存；否则直接保存.
	Save(ctx context.Context, msgs ...*OutboxMessage) error
	// WithTx 在事务中执行 fn，实现原子性语义.
	// fn 收到的 ctx 中已注入事务，可直接调用 Save.
	WithTx(ctx context.Context, fn TxFunc) error
	// FetchPending 拉取待发送消息并标记为 Processing.
	FetchPending(ctx context.Context, limit int) ([]*OutboxMessage, error)
	// MarkSent 批量标记消息为已发送.
	MarkSent(ctx context.Context, ids []uint64) error
	// MarkFailed 标记消息发送失败.
	MarkFailed(ctx context.Context, id uint64, errMsg string) error
	// ResetStale 重置超时的 Processing/Failed 消息为 Pending.
	ResetStale(ctx context.Context, staleDuration time.Duration) (int64, error)
	// Cleanup 清理指定时间之前的已发送消息.
	Cleanup(ctx context.Context, before time.Time) (int64, error)
	// AutoMigrate 自动迁移表结构.
	AutoMigrate() error
}

// GORMStore 基于 GORM 的 Store 实现.
type GORMStore struct {
	db *gorm.DB
}

// 编译期接口合规检查.
var _ Store = (*GORMStore)(nil)

// NewGORMStore 从 rdbms.Database 创建 GORMStore.
func NewGORMStore(db rdbms.Database) *GORMStore {
	return &GORMStore{db: rdbms.AsGORM(db)}
}

// NewGORMStoreFromDB 从 *gorm.DB 创建 GORMStore.
func NewGORMStoreFromDB(db *gorm.DB) *GORMStore {
	return &GORMStore{db: db}
}

// Save 保存消息.
//
// 若 ctx 中注入了事务（通过 InjectTx），则在该事务中保存；否则直接保存.
func (s *GORMStore) Save(ctx context.Context, msgs ...*OutboxMessage) error {
	if len(msgs) == 0 {
		return nil
	}
	db := s.db
	if tx, ok := ExtractTx(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx).Create(msgs).Error
}

// WithTx 在事务中执行 fn.
//
// fn 收到的 ctx 中已通过 InjectTx 注入了事务，可直接调用 Save.
func (s *GORMStore) WithTx(ctx context.Context, fn TxFunc) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := InjectTx(ctx, tx)
		return fn(txCtx)
	})
}

// FetchPending 拉取待发送消息并原子标记为 Processing.
//
// 对支持行锁的数据库（MySQL/PostgreSQL）使用 SELECT FOR UPDATE SKIP LOCKED，
// SQLite 环境自动降级为普通 SELECT.
func (s *GORMStore) FetchPending(ctx context.Context, limit int) ([]*OutboxMessage, error) {
	var msgs []*OutboxMessage

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Where("status = ?", StatusPending).
			Order("id ASC").
			Limit(limit)

		if !s.isSQLite() {
			query = query.Clauses(clause.Locking{
				Strength: "UPDATE",
				Options:  "SKIP LOCKED",
			})
		}

		if err := query.Find(&msgs).Error; err != nil {
			return err
		}

		if len(msgs) == 0 {
			return nil
		}

		ids := make([]uint64, len(msgs))
		for i, m := range msgs {
			ids[i] = m.ID
		}

		return tx.Model(&OutboxMessage{}).
			Where("id IN ?", ids).
			Update("status", StatusProcessing).Error
	})

	return msgs, err
}

// MarkSent 批量标记消息为已发送.
func (s *GORMStore) MarkSent(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	return s.db.WithContext(ctx).
		Model(&OutboxMessage{}).
		Where("id IN ?", ids).
		Updates(map[string]any{
			"status":  StatusSent,
			"sent_at": now,
		}).Error
}

// MarkFailed 标记消息发送失败，递增重试计数.
func (s *GORMStore) MarkFailed(ctx context.Context, id uint64, errMsg string) error {
	return s.db.WithContext(ctx).
		Model(&OutboxMessage{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":      StatusFailed,
			"retry_count": gorm.Expr("retry_count + 1"),
			"last_error":  errMsg,
		}).Error
}

// ResetStale 将超时的 Processing/Failed 消息重置为 Pending.
func (s *GORMStore) ResetStale(ctx context.Context, staleDuration time.Duration) (int64, error) {
	threshold := time.Now().Add(-staleDuration)
	result := s.db.WithContext(ctx).
		Model(&OutboxMessage{}).
		Where("status IN ? AND updated_at < ?", []MessageStatus{StatusProcessing, StatusFailed}, threshold).
		Updates(map[string]any{
			"status": StatusPending,
		})
	return result.RowsAffected, result.Error
}

// Cleanup 删除指定时间之前的已发送消息.
func (s *GORMStore) Cleanup(ctx context.Context, before time.Time) (int64, error) {
	result := s.db.WithContext(ctx).
		Where("status = ? AND sent_at < ?", StatusSent, before).
		Delete(&OutboxMessage{})
	return result.RowsAffected, result.Error
}

// AutoMigrate 自动迁移 outbox_messages 表.
func (s *GORMStore) AutoMigrate() error {
	return s.db.AutoMigrate(&OutboxMessage{})
}

// isSQLite 检测当前是否使用 SQLite.
func (s *GORMStore) isSQLite() bool {
	return s.db.Dialector.Name() == "sqlite"
}
