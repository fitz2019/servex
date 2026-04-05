package idgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnowflake_Unique(t *testing.T) {
	gen, err := NewSnowflake(&SnowflakeConfig{WorkerID: 1, DatacenterID: 1})
	require.NoError(t, err)

	seen := make(map[string]bool)
	count := 10000
	for i := 0; i < count; i++ {
		id, err := gen.NextID()
		require.NoError(t, err)
		assert.NotEmpty(t, id)
		assert.False(t, seen[id], "重复 ID: %s (第 %d 次)", id, i)
		seen[id] = true
	}
	assert.Len(t, seen, count)
}

func TestSnowflake_Monotonic(t *testing.T) {
	gen, err := NewSnowflake(&SnowflakeConfig{})
	require.NoError(t, err)

	prev := ""
	for i := 0; i < 100; i++ {
		id, err := gen.NextID()
		require.NoError(t, err)
		if prev != "" {
			assert.Greater(t, id, prev, "ID 应单调递增")
		}
		prev = id
	}
}

func TestSnowflake_Validation(t *testing.T) {
	// 无效 WorkerID
	_, err := NewSnowflake(&SnowflakeConfig{WorkerID: -1})
	assert.ErrorIs(t, err, ErrInvalidWorkerID)

	_, err = NewSnowflake(&SnowflakeConfig{WorkerID: 1024})
	assert.ErrorIs(t, err, ErrInvalidWorkerID)

	// 无效 DatacenterID
	_, err = NewSnowflake(&SnowflakeConfig{DatacenterID: -1})
	assert.ErrorIs(t, err, ErrInvalidDatacenterID)

	_, err = NewSnowflake(&SnowflakeConfig{DatacenterID: 32})
	assert.ErrorIs(t, err, ErrInvalidDatacenterID)

	// 有效配置
	gen, err := NewSnowflake(&SnowflakeConfig{WorkerID: 0, DatacenterID: 0})
	assert.NoError(t, err)
	assert.NotNil(t, gen)

	gen, err = NewSnowflake(&SnowflakeConfig{WorkerID: 1023, DatacenterID: 31})
	assert.NoError(t, err)
	assert.NotNil(t, gen)

	// nil 配置
	gen, err = NewSnowflake(nil)
	assert.NoError(t, err)
	assert.NotNil(t, gen)
}

func TestULID_Format(t *testing.T) {
	gen := NewULID()
	for i := 0; i < 100; i++ {
		id, err := gen.NextID()
		require.NoError(t, err)
		assert.Len(t, id, 26, "ULID 应为 26 个字符")
	}
}

func TestULID_Unique(t *testing.T) {
	gen := NewULID()
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id, err := gen.NextID()
		require.NoError(t, err)
		assert.False(t, seen[id], "重复 ULID: %s", id)
		seen[id] = true
	}
}

func TestNanoID_Length(t *testing.T) {
	gen := NewNanoID()
	for i := 0; i < 100; i++ {
		id, err := gen.NextID()
		require.NoError(t, err)
		assert.Len(t, id, 21, "默认 NanoID 应为 21 个字符")
	}
}

func TestNanoID_CustomAlphabet(t *testing.T) {
	gen := NewNanoID(WithAlphabet("0123456789"), WithSize(10))
	for i := 0; i < 100; i++ {
		id, err := gen.NextID()
		require.NoError(t, err)
		assert.Len(t, id, 10)
		for _, c := range id {
			assert.True(t, c >= '0' && c <= '9', "字符应为数字: %c", c)
		}
	}
}

func TestNanoID_CustomSize(t *testing.T) {
	gen := NewNanoID(WithSize(32))
	id, err := gen.NextID()
	require.NoError(t, err)
	assert.Len(t, id, 32)
}

func TestConvenienceFunctions(t *testing.T) {
	// Snowflake
	id := Snowflake()
	assert.NotEmpty(t, id)

	// ULID
	id = ULID()
	assert.Len(t, id, 26)

	// NanoID
	id = NanoID()
	assert.Len(t, id, 21)

	// UUID
	id = UUID()
	assert.Len(t, id, 36) // UUID 格式 xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	assert.Contains(t, id, "-")
}

func BenchmarkSnowflake(b *testing.B) {
	gen, _ := NewSnowflake(&SnowflakeConfig{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.NextID()
	}
}

func BenchmarkULID(b *testing.B) {
	gen := NewULID()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.NextID()
	}
}

func BenchmarkNanoID(b *testing.B) {
	gen := NewNanoID()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.NextID()
	}
}
