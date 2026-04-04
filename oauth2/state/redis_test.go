// oauth2/state/redis_test.go
package state

import (
	"testing"

	"github.com/Tsukikage7/servex/oauth2"
)

func TestNewRedisStore_NilClient(t *testing.T) {
	_, err := NewRedisStore(nil)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestRedisStore_ImplementsInterface(t *testing.T) {
	var _ oauth2.StateStore = (*RedisStore)(nil)
}
