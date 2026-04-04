// jobqueue/factory/config_test.go
package factory

import (
	"testing"
)

func TestNewStore_NilConfig(t *testing.T) {
	_, err := NewStore(nil)
	if err == nil {
		t.Fatal("期望 nil config 返回错误，实际为 nil")
	}
}

func TestNewStore_EmptyType(t *testing.T) {
	_, err := NewStore(&StoreConfig{})
	if err == nil {
		t.Fatal("期望空 type 返回错误，实际为 nil")
	}
}

func TestNewStore_UnsupportedType(t *testing.T) {
	_, err := NewStore(&StoreConfig{Type: "unsupported"})
	if err == nil {
		t.Fatal("期望不支持的 type 返回错误，实际为 nil")
	}
}
