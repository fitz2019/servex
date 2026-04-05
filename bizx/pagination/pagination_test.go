package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name   string
		values []any
	}{
		{"单个字符串", []any{"abc123"}},
		{"单个数字", []any{float64(42)}},
		{"多个值", []any{"id_100", float64(1680000000)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := EncodeCursor(tt.values...)
			assert.NotEmpty(t, cursor)

			decoded, err := DecodeCursor(cursor)
			require.NoError(t, err)
			assert.Equal(t, tt.values, decoded)
		})
	}
}

func TestDecodeCursor_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		cursor string
	}{
		{"空字符串", ""},
		{"非 base64", "!!!invalid!!!"},
		{"非 JSON", base64Encode("not json")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeCursor(tt.cursor)
			assert.ErrorIs(t, err, ErrInvalidCursor)
		})
	}
}

func base64Encode(s string) string {
	return "bm90IGpzb24=" // "not json" 的 base64
}

func TestCursorResponse(t *testing.T) {
	resp := CursorResponse[string]{
		Items:      []string{"a", "b", "c"},
		NextCursor: EncodeCursor("c"),
		HasMore:    true,
	}

	assert.Len(t, resp.Items, 3)
	assert.True(t, resp.HasMore)
	assert.NotEmpty(t, resp.NextCursor)
	assert.Empty(t, resp.PrevCursor)
}

func TestApplyDefaults(t *testing.T) {
	t.Run("默认值", func(t *testing.T) {
		req := &CursorRequest{}
		req.Apply()
		assert.Equal(t, DefaultLimit, req.Limit)
		assert.Equal(t, Forward, req.Direction)
	})

	t.Run("超过最大值", func(t *testing.T) {
		req := &CursorRequest{Limit: 999}
		req.Apply()
		assert.Equal(t, MaxLimit, req.Limit)
	})

	t.Run("保留有效值", func(t *testing.T) {
		req := &CursorRequest{Limit: 50, Direction: Backward}
		req.Apply()
		assert.Equal(t, 50, req.Limit)
		assert.Equal(t, Backward, req.Direction)
	})
}

// testItem 测试用数据模型.
type testItem struct {
	ID   int    `gorm:"primaryKey"`
	Name string `gorm:"size:100"`
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&testItem{}))

	// 插入测试数据
	for i := 1; i <= 10; i++ {
		require.NoError(t, db.Create(&testItem{ID: i, Name: "item"}).Error)
	}
	return db
}

func TestGORMPaginate(t *testing.T) {
	db := setupTestDB(t)

	t.Run("第一页", func(t *testing.T) {
		req := &CursorRequest{Limit: 3}
		query := GORMPaginate(db.Model(&testItem{}), req, "id")

		var items []testItem
		require.NoError(t, query.Find(&items).Error)
		// limit+1 = 4，所以取到4条
		assert.Len(t, items, 4)
		assert.Equal(t, 1, items[0].ID)
	})

	t.Run("向前翻页", func(t *testing.T) {
		cursor := EncodeCursor(float64(3))
		req := &CursorRequest{Cursor: cursor, Limit: 3, Direction: Forward}
		query := GORMPaginate(db.Model(&testItem{}), req, "id")

		var items []testItem
		require.NoError(t, query.Find(&items).Error)
		assert.True(t, len(items) > 0)
		assert.Equal(t, 4, items[0].ID)
	})

	t.Run("向后翻页", func(t *testing.T) {
		cursor := EncodeCursor(float64(5))
		req := &CursorRequest{Cursor: cursor, Limit: 3, Direction: Backward}
		query := GORMPaginate(db.Model(&testItem{}), req, "id")

		var items []testItem
		require.NoError(t, query.Find(&items).Error)
		assert.True(t, len(items) > 0)
		// Backward 按 DESC 排序，第一条应该是 4
		assert.Equal(t, 4, items[0].ID)
	})

	t.Run("无效游标", func(t *testing.T) {
		req := &CursorRequest{Cursor: "invalid", Limit: 3}
		query := GORMPaginate(db.Model(&testItem{}), req, "id")

		var items []testItem
		require.NoError(t, query.Find(&items).Error)
		assert.Empty(t, items)
	})
}
