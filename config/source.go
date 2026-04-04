package config

// KeyValue 配置键值对.
type KeyValue struct {
	Key    string
	Value  []byte
	Format string // 如 "json", "yaml", "toml"
}

// Source 配置数据源接口.
type Source interface {
	// Load 从数据源加载配置.
	Load() ([]*KeyValue, error)
	// Watch 创建配置变更监听器.
	Watch() (Watcher, error)
}

// Watcher 配置变更监听器接口.
type Watcher interface {
	// Next 阻塞直到配置变更，返回新的配置数据.
	Next() ([]*KeyValue, error)
	// Stop 停止监听.
	Stop() error
}
