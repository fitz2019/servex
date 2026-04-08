// oauth2/state/memory_test.go
package state

import (
	"testing"

	"github.com/Tsukikage7/servex/oauth2"
)

func TestMemoryStore_GenerateAndValidate(t *testing.T) {
	s := NewMemoryStore()
	ctx := t.Context()

	state, err := s.Generate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state == "" {
		t.Fatal("state should not be empty")
	}

	ok, err := s.Validate(ctx, state)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("valid state should return true")
	}

	// 第二次验证应该失败（一次性使用）
	ok, _ = s.Validate(ctx, state)
	if ok {
		t.Error("state should be consumed after first validation")
	}
}

func TestMemoryStore_InvalidState(t *testing.T) {
	s := NewMemoryStore()
	ok, err := s.Validate(t.Context(), "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("nonexistent state should return false")
	}
}

func TestMemoryStore_ImplementsInterface(t *testing.T) {
	var _ oauth2.StateStore = (*MemoryStore)(nil)
}

func TestMemoryStore_ExpiredState(t *testing.T) {
	s := NewMemoryStore()
	s.ttl = 0 // states expire immediately
	ctx := t.Context()

	state, err := s.Generate(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Validate should fail because state expired.
	ok, err := s.Validate(ctx, state)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expired state should return false")
	}
}

func TestMemoryStore_DoubleValidate(t *testing.T) {
	s := NewMemoryStore()
	ctx := t.Context()

	state, _ := s.Generate(ctx)
	ok1, _ := s.Validate(ctx, state)
	ok2, _ := s.Validate(ctx, state)

	if !ok1 {
		t.Error("first validation should succeed")
	}
	if ok2 {
		t.Error("second validation should fail (consumed)")
	}
}

func TestMemoryStore_MultipleStates(t *testing.T) {
	s := NewMemoryStore()
	ctx := t.Context()

	s1, _ := s.Generate(ctx)
	s2, _ := s.Generate(ctx)

	if s1 == s2 {
		t.Error("generated states should be unique")
	}

	ok1, _ := s.Validate(ctx, s1)
	ok2, _ := s.Validate(ctx, s2)

	if !ok1 || !ok2 {
		t.Error("both states should be valid")
	}
}
