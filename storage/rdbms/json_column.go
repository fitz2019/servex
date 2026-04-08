package rdbms

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// JsonColumn 数据库 JSON 列类型，不同驱动自动映射为 JSON/JSONB/TEXT.
type JsonColumn[T any] struct {
	Val   T
	Valid bool
}

// NewJsonColumn 创建有效的 JSON 列值.
func NewJsonColumn[T any](val T) JsonColumn[T] {
	return JsonColumn[T]{Val: val, Valid: true}
}

// NullJsonColumn 创建空值的 JSON 列.
func NullJsonColumn[T any]() JsonColumn[T] {
	return JsonColumn[T]{}
}

func (jc JsonColumn[T]) Value() (driver.Value, error) {
	if !jc.Valid {
		return nil, nil
	}
	data, err := json.Marshal(jc.Val)
	if err != nil {
		return nil, fmt.Errorf("database: json 序列化失败: %w", err)
	}
	return string(data), nil
}

func (jc *JsonColumn[T]) Scan(src any) error {
	if src == nil {
		jc.Valid = false
		return nil
	}

	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("database: json 列不支持类型 %T", src)
	}

	if err := json.Unmarshal(data, &jc.Val); err != nil {
		return fmt.Errorf("database: json 反序列化失败: %w", err)
	}
	jc.Valid = true
	return nil
}

func (JsonColumn[T]) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	switch db.Dialector.Name() {
	case DriverMySQL:
		return "JSON"
	case DriverPostgres, DriverPostgreSQL:
		return "JSONB"
	default:
		return "TEXT"
	}
}
