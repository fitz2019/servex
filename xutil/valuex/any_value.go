// Package valuex 提供泛型类型转换工具.
package valuex

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrNilValue      = errors.New("valuex: 值为 nil")
	ErrTypeMismatch  = errors.New("valuex: 类型不匹配")
	ErrConvertFailed = errors.New("valuex: 转换失败")
)

// AnyValue 包装任意值并提供类型安全的访问方法.
type AnyValue struct {
	Val any
	Err error
}

func Of(val any) AnyValue {
	return AnyValue{Val: val}
}

func typeAssert[T any](av AnyValue, typeName string) (T, error) {
	if av.Err != nil {
		var zero T
		return zero, av.Err
	}
	val, ok := av.Val.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("%w: 期望 %s, 实际 %T", ErrTypeMismatch, typeName, av.Val)
	}
	return val, nil
}

func (av AnyValue) Int() (int, error)         { return typeAssert[int](av, "int") }
func (av AnyValue) Int8() (int8, error)       { return typeAssert[int8](av, "int8") }
func (av AnyValue) Int16() (int16, error)     { return typeAssert[int16](av, "int16") }
func (av AnyValue) Int32() (int32, error)     { return typeAssert[int32](av, "int32") }
func (av AnyValue) Int64() (int64, error)     { return typeAssert[int64](av, "int64") }
func (av AnyValue) Uint() (uint, error)       { return typeAssert[uint](av, "uint") }
func (av AnyValue) Uint8() (uint8, error)     { return typeAssert[uint8](av, "uint8") }
func (av AnyValue) Uint16() (uint16, error)   { return typeAssert[uint16](av, "uint16") }
func (av AnyValue) Uint32() (uint32, error)   { return typeAssert[uint32](av, "uint32") }
func (av AnyValue) Uint64() (uint64, error)   { return typeAssert[uint64](av, "uint64") }
func (av AnyValue) Float32() (float32, error) { return typeAssert[float32](av, "float32") }
func (av AnyValue) Float64() (float64, error) { return typeAssert[float64](av, "float64") }
func (av AnyValue) String() (string, error)   { return typeAssert[string](av, "string") }
func (av AnyValue) Bool() (bool, error)       { return typeAssert[bool](av, "bool") }
func (av AnyValue) Bytes() ([]byte, error)    { return typeAssert[[]byte](av, "[]byte") }

func (av AnyValue) AsInt() (int, error) {
	if av.Err != nil {
		return 0, av.Err
	}
	if av.Val == nil {
		return 0, ErrNilValue
	}
	switch v := av.Val.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("%w: 无法将 %T 转换为 int", ErrConvertFailed, av.Val)
	}
}

func (av AnyValue) AsInt64() (int64, error) {
	if av.Err != nil {
		return 0, av.Err
	}
	if av.Val == nil {
		return 0, ErrNilValue
	}
	switch v := av.Val.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("%w: 无法将 %T 转换为 int64", ErrConvertFailed, av.Val)
	}
}

func (av AnyValue) AsFloat64() (float64, error) {
	if av.Err != nil {
		return 0, av.Err
	}
	if av.Val == nil {
		return 0, ErrNilValue
	}
	switch v := av.Val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("%w: 无法将 %T 转换为 float64", ErrConvertFailed, av.Val)
	}
}

func (av AnyValue) AsString() (string, error) {
	if av.Err != nil {
		return "", av.Err
	}
	if av.Val == nil {
		return "", ErrNilValue
	}
	switch v := av.Val.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return fmt.Sprintf("%v", av.Val), nil
	}
}

func (av AnyValue) AsBool() (bool, error) {
	if av.Err != nil {
		return false, av.Err
	}
	if av.Val == nil {
		return false, ErrNilValue
	}
	switch v := av.Val.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("%w: 无法将 %T 转换为 bool", ErrConvertFailed, av.Val)
	}
}

func (av AnyValue) IntOrDefault(def int) int {
	val, err := av.AsInt()
	if err != nil {
		return def
	}
	return val
}

func (av AnyValue) Int64OrDefault(def int64) int64 {
	val, err := av.AsInt64()
	if err != nil {
		return def
	}
	return val
}

func (av AnyValue) Float64OrDefault(def float64) float64 {
	val, err := av.AsFloat64()
	if err != nil {
		return def
	}
	return val
}

func (av AnyValue) StringOrDefault(def string) string {
	val, err := av.AsString()
	if err != nil {
		return def
	}
	return val
}

func (av AnyValue) BoolOrDefault(def bool) bool {
	val, err := av.AsBool()
	if err != nil {
		return def
	}
	return val
}
