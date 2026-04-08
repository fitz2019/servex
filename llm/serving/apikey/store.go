package apikey

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

// Store API Key 持久化接口.
type Store interface {
	// Save 保存新的 API Key.
	Save(ctx context.Context, key *Key) error
	// GetByHash 根据哈希值查找 API Key.
	GetByHash(ctx context.Context, hashedKey string) (*Key, error)
	// GetByID 根据 ID 查找 API Key.
	GetByID(ctx context.Context, id string) (*Key, error)
	// List 列出指定 Owner 的所有 API Key.
	List(ctx context.Context, ownerID string) ([]*Key, error)
	// Update 更新 API Key.
	Update(ctx context.Context, key *Key) error
	// Delete 删除 API Key.
	Delete(ctx context.Context, id string) error
	// AutoMigrate 自动迁移数据库表结构.
	AutoMigrate(ctx context.Context) error
}

// GORMStore 基于 GORM 的 Store 实现.
type GORMStore struct {
	db *gorm.DB
}

// NewGORMStore 创建基于 GORM 的 Store.
func NewGORMStore(db *gorm.DB) *GORMStore {
	return &GORMStore{db: db}
}

func (s *GORMStore) Save(ctx context.Context, key *Key) error {
	return s.db.WithContext(ctx).Create(key).Error
}

func (s *GORMStore) GetByHash(ctx context.Context, hashedKey string) (*Key, error) {
	var key Key
	if err := s.db.WithContext(ctx).Where("hashed_key = ?", hashedKey).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (s *GORMStore) GetByID(ctx context.Context, id string) (*Key, error) {
	var key Key
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (s *GORMStore) List(ctx context.Context, ownerID string) ([]*Key, error) {
	var keys []*Key
	if err := s.db.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *GORMStore) Update(ctx context.Context, key *Key) error {
	return s.db.WithContext(ctx).Save(key).Error
}

func (s *GORMStore) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&Key{}).Error
}

func (s *GORMStore) AutoMigrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&Key{})
}

// MemoryStore 基于内存的 Store 实现，用于测试.
type MemoryStore struct {
	mu     sync.RWMutex
	byID   map[string]*Key
	byHash map[string]*Key
}

// NewMemoryStore 创建基于内存的 Store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		byID:   make(map[string]*Key),
		byHash: make(map[string]*Key),
	}
}

func (s *MemoryStore) Save(_ context.Context, key *Key) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 深拷贝避免外部修改
	copied := *key
	s.byID[copied.ID] = &copied
	s.byHash[copied.HashedKey] = &copied
	return nil
}

func (s *MemoryStore) GetByHash(_ context.Context, hashedKey string) (*Key, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, ok := s.byHash[hashedKey]
	if !ok {
		return nil, ErrKeyNotFound
	}
	copied := *key
	return &copied, nil
}

func (s *MemoryStore) GetByID(_ context.Context, id string) (*Key, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, ok := s.byID[id]
	if !ok {
		return nil, ErrKeyNotFound
	}
	copied := *key
	return &copied, nil
}

func (s *MemoryStore) List(_ context.Context, ownerID string) ([]*Key, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var keys []*Key
	for _, key := range s.byID {
		if key.OwnerID == ownerID {
			copied := *key
			keys = append(keys, &copied)
		}
	}
	return keys, nil
}

func (s *MemoryStore) Update(_ context.Context, key *Key) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.byID[key.ID]
	if !ok {
		return ErrKeyNotFound
	}

	// 删除旧的哈希索引
	delete(s.byHash, existing.HashedKey)

	copied := *key
	s.byID[copied.ID] = &copied
	s.byHash[copied.HashedKey] = &copied
	return nil
}

func (s *MemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, ok := s.byID[id]
	if !ok {
		return ErrKeyNotFound
	}
	delete(s.byHash, key.HashedKey)
	delete(s.byID, id)
	return nil
}

func (s *MemoryStore) AutoMigrate(_ context.Context) error {
	return nil
}
