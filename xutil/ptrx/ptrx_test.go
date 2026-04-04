package ptrx

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PtrxTestSuite struct {
	suite.Suite
}

func TestPtrxSuite(t *testing.T) {
	suite.Run(t, new(PtrxTestSuite))
}

func (s *PtrxTestSuite) TestToPtr() {
	tests := []struct {
		name string
		val  int
	}{
		{name: "正数", val: 42},
		{name: "零值", val: 0},
		{name: "负数", val: -1},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			p := ToPtr(tt.val)
			s.Require().NotNil(p)
			s.Equal(tt.val, *p)
		})
	}
}

func (s *PtrxTestSuite) TestToPtr_String() {
	p := ToPtr("hello")
	s.Require().NotNil(p)
	s.Equal("hello", *p)
}

func (s *PtrxTestSuite) TestToPtrSlice() {
	src := []int{1, 2, 3}
	dst := ToPtrSlice(src)

	s.Len(dst, 3)
	for i, v := range dst {
		s.Require().NotNil(v)
		s.Equal(src[i], *v)
	}
}

func (s *PtrxTestSuite) TestToPtrSlice_Empty() {
	dst := ToPtrSlice([]string{})
	s.Empty(dst)
}

func (s *PtrxTestSuite) TestValue() {
	tests := []struct {
		name     string
		ptr      *int
		expected int
	}{
		{name: "非 nil 指针", ptr: ToPtr(42), expected: 42},
		{name: "nil 指针返回零值", ptr: nil, expected: 0},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, Value(tt.ptr))
		})
	}
}

func (s *PtrxTestSuite) TestValue_StringNil() {
	s.Equal("", Value[string](nil))
}

func (s *PtrxTestSuite) TestEqual() {
	tests := []struct {
		name     string
		a        *int
		b        *int
		expected bool
	}{
		{name: "两个 nil", a: nil, b: nil, expected: true},
		{name: "a 为 nil", a: nil, b: ToPtr(1), expected: false},
		{name: "b 为 nil", a: ToPtr(1), b: nil, expected: false},
		{name: "值相等", a: ToPtr(42), b: ToPtr(42), expected: true},
		{name: "值不等", a: ToPtr(1), b: ToPtr(2), expected: false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, Equal(tt.a, tt.b))
		})
	}
}
