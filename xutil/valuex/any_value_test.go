package valuex

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AnyValueTestSuite struct {
	suite.Suite
}

func TestAnyValueSuite(t *testing.T) {
	suite.Run(t, new(AnyValueTestSuite))
}

func (s *AnyValueTestSuite) TestOf() {
	av := Of(42)
	s.Equal(42, av.Val)
	s.NoError(av.Err)
}

func (s *AnyValueTestSuite) TestOf_Nil() {
	av := Of(nil)
	s.Nil(av.Val)
}

func (s *AnyValueTestSuite) TestInt() {
	val, err := Of(42).Int()
	s.NoError(err)
	s.Equal(42, val)

	_, err = Of("42").Int()
	s.ErrorIs(err, ErrTypeMismatch)
}

func (s *AnyValueTestSuite) TestInt8() {
	val, err := Of(int8(8)).Int8()
	s.NoError(err)
	s.Equal(int8(8), val)
}

func (s *AnyValueTestSuite) TestInt16() {
	val, err := Of(int16(16)).Int16()
	s.NoError(err)
	s.Equal(int16(16), val)
}

func (s *AnyValueTestSuite) TestInt32() {
	val, err := Of(int32(32)).Int32()
	s.NoError(err)
	s.Equal(int32(32), val)
}

func (s *AnyValueTestSuite) TestInt64() {
	val, err := Of(int64(64)).Int64()
	s.NoError(err)
	s.Equal(int64(64), val)
}

func (s *AnyValueTestSuite) TestUint() {
	val, err := Of(uint(10)).Uint()
	s.NoError(err)
	s.Equal(uint(10), val)
}

func (s *AnyValueTestSuite) TestUint8() {
	val, err := Of(uint8(8)).Uint8()
	s.NoError(err)
	s.Equal(uint8(8), val)
}

func (s *AnyValueTestSuite) TestUint16() {
	val, err := Of(uint16(16)).Uint16()
	s.NoError(err)
	s.Equal(uint16(16), val)
}

func (s *AnyValueTestSuite) TestUint32() {
	val, err := Of(uint32(32)).Uint32()
	s.NoError(err)
	s.Equal(uint32(32), val)
}

func (s *AnyValueTestSuite) TestUint64() {
	val, err := Of(uint64(64)).Uint64()
	s.NoError(err)
	s.Equal(uint64(64), val)
}

func (s *AnyValueTestSuite) TestFloat32() {
	val, err := Of(float32(3.14)).Float32()
	s.NoError(err)
	s.InDelta(float32(3.14), val, 0.001)
}

func (s *AnyValueTestSuite) TestFloat64() {
	val, err := Of(3.14).Float64()
	s.NoError(err)
	s.InDelta(3.14, val, 0.001)
}

func (s *AnyValueTestSuite) TestString() {
	val, err := Of("hello").String()
	s.NoError(err)
	s.Equal("hello", val)

	_, err = Of(42).String()
	s.ErrorIs(err, ErrTypeMismatch)
}

func (s *AnyValueTestSuite) TestBool() {
	val, err := Of(true).Bool()
	s.NoError(err)
	s.True(val)

	_, err = Of("true").Bool()
	s.ErrorIs(err, ErrTypeMismatch)
}

func (s *AnyValueTestSuite) TestBytes() {
	val, err := Of([]byte("hello")).Bytes()
	s.NoError(err)
	s.Equal([]byte("hello"), val)
}

func (s *AnyValueTestSuite) TestStrict_WithError() {
	av := AnyValue{Err: errors.New("prior error")}

	_, err := av.Int()
	s.Error(err)

	_, err = av.String()
	s.Error(err)

	_, err = av.Bool()
	s.Error(err)
}

func (s *AnyValueTestSuite) TestAsInt() {
	tests := []struct {
		name     string
		input    any
		expected int
	}{
		{"int", 42, 42},
		{"int8", int8(8), 8},
		{"int32", int32(32), 32},
		{"int64", int64(64), 64},
		{"uint", uint(10), 10},
		{"float64", 3.9, 3},
		{"string", "100", 100},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			val, err := Of(tt.input).AsInt()
			s.NoError(err)
			s.Equal(tt.expected, val)
		})
	}
}

func (s *AnyValueTestSuite) TestAsInt_Nil() {
	_, err := Of(nil).AsInt()
	s.ErrorIs(err, ErrNilValue)
}

func (s *AnyValueTestSuite) TestAsInt_Invalid() {
	_, err := Of([]int{1}).AsInt()
	s.ErrorIs(err, ErrConvertFailed)
}

func (s *AnyValueTestSuite) TestAsInt64() {
	val, err := Of(int32(42)).AsInt64()
	s.NoError(err)
	s.Equal(int64(42), val)

	val, err = Of("9999").AsInt64()
	s.NoError(err)
	s.Equal(int64(9999), val)
}

func (s *AnyValueTestSuite) TestAsFloat64() {
	tests := []struct {
		name     string
		input    any
		expected float64
	}{
		{"float64", 3.14, 3.14},
		{"float32", float32(2.5), 2.5},
		{"int", 42, 42.0},
		{"string", "3.14", 3.14},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			val, err := Of(tt.input).AsFloat64()
			s.NoError(err)
			s.InDelta(tt.expected, val, 0.01)
		})
	}
}

func (s *AnyValueTestSuite) TestAsString() {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string", "hello", "hello"},
		{"bytes", []byte("world"), "world"},
		{"int via fmt", 42, "42"},
		{"bool via fmt", true, "true"},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			val, err := Of(tt.input).AsString()
			s.NoError(err)
			s.Equal(tt.expected, val)
		})
	}
}

func (s *AnyValueTestSuite) TestAsBool() {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"int nonzero", 1, true},
		{"int zero", 0, false},
		{"int64 nonzero", int64(1), true},
		{"float64 nonzero", 1.0, true},
		{"string true", "true", true},
		{"string false", "false", false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			val, err := Of(tt.input).AsBool()
			s.NoError(err)
			s.Equal(tt.expected, val)
		})
	}
}

func (s *AnyValueTestSuite) TestAsBool_Nil() {
	_, err := Of(nil).AsBool()
	s.ErrorIs(err, ErrNilValue)
}

func (s *AnyValueTestSuite) TestIntOrDefault() {
	s.Equal(42, Of(42).IntOrDefault(0))
	s.Equal(0, Of("not int").IntOrDefault(0))
	s.Equal(-1, Of(nil).IntOrDefault(-1))
}

func (s *AnyValueTestSuite) TestInt64OrDefault() {
	s.Equal(int64(100), Of(int64(100)).Int64OrDefault(0))
	s.Equal(int64(0), Of(nil).Int64OrDefault(0))
}

func (s *AnyValueTestSuite) TestFloat64OrDefault() {
	s.InDelta(3.14, Of(3.14).Float64OrDefault(0), 0.001)
	s.InDelta(0.0, Of(nil).Float64OrDefault(0), 0.001)
}

func (s *AnyValueTestSuite) TestStringOrDefault() {
	s.Equal("hello", Of("hello").StringOrDefault("default"))
	s.Equal("default", Of(nil).StringOrDefault("default"))
}

func (s *AnyValueTestSuite) TestBoolOrDefault() {
	s.True(Of(true).BoolOrDefault(false))
	s.False(Of(nil).BoolOrDefault(false))
}

func (s *AnyValueTestSuite) TestAsInt_WithExistingError() {
	av := AnyValue{Val: 42, Err: errors.New("existing")}
	_, err := av.AsInt()
	s.Error(err)
	s.Equal("existing", err.Error())
}

func (s *AnyValueTestSuite) TestAsString_WithExistingError() {
	av := AnyValue{Val: "hello", Err: errors.New("existing")}
	_, err := av.AsString()
	s.Error(err)
}
