// Package pagination 提供游标分页支持，适合大数据量和实时数据场景.
package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

// 默认配置.
const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Direction 分页方向.
type Direction string

const (
	// Forward 向前分页（获取更新的数据）.
	Forward Direction = "forward"
	// Backward 向后分页（获取更旧的数据）.
	Backward Direction = "backward"
)

// ErrInvalidCursor 无效的游标.
var ErrInvalidCursor = errors.New("pagination: invalid cursor")

// CursorRequest 游标分页请求.
type CursorRequest struct {
	Cursor    string    `json:"cursor,omitempty"` // 上一页最后一条的游标，为空表示第一页
	Limit     int       `json:"limit"`            // 每页数量，默认 20
	Direction Direction `json:"direction"`        // 方向，默认 Forward
}

// Apply 应用默认值并校验参数.
func (r *CursorRequest) Apply() *CursorRequest {
	if r.Limit <= 0 {
		r.Limit = DefaultLimit
	}
	if r.Limit > MaxLimit {
		r.Limit = MaxLimit
	}
	if r.Direction == "" {
		r.Direction = Forward
	}
	return r
}

// CursorResponse 游标分页响应.
type CursorResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// EncodeCursor 将值编码为游标字符串（base64url + JSON）.
func EncodeCursor(values ...any) string {
	data, err := json.Marshal(values)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeCursor 解码游标字符串.
func DecodeCursor(cursor string) ([]any, error) {
	if cursor == "" {
		return nil, ErrInvalidCursor
	}
	data, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	var values []any
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, ErrInvalidCursor
	}
	return values, nil
}

// GORMPaginate 为 GORM 查询添加游标分页条件.
// orderField 是排序字段名，cursorValue 从 DecodeCursor 获得.
func GORMPaginate(db *gorm.DB, req *CursorRequest, orderField string) *gorm.DB {
	req.Apply()

	query := db

	if req.Cursor != "" {
		values, err := DecodeCursor(req.Cursor)
		if err != nil || len(values) == 0 {
			return query.Where("1 = 0") // 无效游标返回空结果
		}
		cursorValue := values[0]
		if req.Direction == Backward {
			query = query.Where(orderField+" < ?", cursorValue)
		} else {
			query = query.Where(orderField+" > ?", cursorValue)
		}
	}

	if req.Direction == Backward {
		query = query.Order(orderField + " DESC")
	} else {
		query = query.Order(orderField + " ASC")
	}

	// 多取一条用于判断是否有更多数据
	query = query.Limit(req.Limit + 1)

	return query
}
