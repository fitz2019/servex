package strx

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type StrxTestSuite struct {
	suite.Suite
}

func TestStrxSuite(t *testing.T) {
	suite.Run(t, new(StrxTestSuite))
}

func (s *StrxTestSuite) TestSplitName() {
	tests := []struct {
		name      string
		fullName  string
		wantFirst string
		wantLast  string
	}{
		{name: "正常全名", fullName: "John Doe", wantFirst: "John", wantLast: "Doe"},
		{name: "仅名", fullName: "John", wantFirst: "John", wantLast: ""},
		{name: "含多个空格的姓", fullName: "John van Doe", wantFirst: "John", wantLast: "van Doe"},
		{name: "空字符串", fullName: "", wantFirst: "", wantLast: ""},
		{name: "仅空格", fullName: "   ", wantFirst: "", wantLast: ""},
		{name: "前后有空格", fullName: "  Alice Bob  ", wantFirst: "Alice", wantLast: "Bob"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			first, last := SplitName(tt.fullName)
			s.Equal(tt.wantFirst, first)
			s.Equal(tt.wantLast, last)
		})
	}
}

func (s *StrxTestSuite) TestJoinName() {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		expected  string
	}{
		{name: "正常拼接", firstName: "John", lastName: "Doe", expected: "John Doe"},
		{name: "仅名", firstName: "John", lastName: "", expected: "John"},
		{name: "仅姓", firstName: "", lastName: "Doe", expected: "Doe"},
		{name: "都为空", firstName: "", lastName: "", expected: ""},
		{name: "含空格", firstName: "  John  ", lastName: "  Doe  ", expected: "John Doe"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, JoinName(tt.firstName, tt.lastName))
		})
	}
}

func (s *StrxTestSuite) TestTrimAndLower() {
	s.Equal("hello", TrimAndLower("  Hello  "))
	s.Equal("", TrimAndLower(""))
	s.Equal("abc", TrimAndLower("ABC"))
}

func (s *StrxTestSuite) TestTrimAndUpper() {
	s.Equal("HELLO", TrimAndUpper("  hello  "))
	s.Equal("", TrimAndUpper(""))
}

func (s *StrxTestSuite) TestIsEmpty() {
	s.True(IsEmpty(""))
	s.True(IsEmpty("   "))
	s.True(IsEmpty("\t\n"))
	s.False(IsEmpty("a"))
	s.False(IsEmpty(" a "))
}

func (s *StrxTestSuite) TestIsNotEmpty() {
	s.False(IsNotEmpty(""))
	s.True(IsNotEmpty("a"))
}

func (s *StrxTestSuite) TestToTitle() {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "小写", input: "hello", expected: "Hello"},
		{name: "大写", input: "HELLO", expected: "Hello"},
		{name: "混合", input: "hELLO", expected: "Hello"},
		{name: "空字符串", input: "", expected: ""},
		{name: "单字符", input: "a", expected: "A"},
		{name: "Unicode 中文", input: "你好世界", expected: "你好世界"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, ToTitle(tt.input))
		})
	}
}

func (s *StrxTestSuite) TestTruncate() {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{name: "不截断", input: "hello", maxLen: 10, expected: "hello"},
		{name: "刚好等长", input: "hello", maxLen: 5, expected: "hello"},
		{name: "截断加省略号", input: "hello world", maxLen: 8, expected: "hello..."},
		{name: "maxLen <= 3", input: "hello", maxLen: 3, expected: "hel"},
		{name: "maxLen 为 0", input: "hello", maxLen: 0, expected: ""},
		{name: "负数长度", input: "hello", maxLen: -1, expected: ""},
		{name: "Unicode 截断", input: "你好世界测试", maxLen: 5, expected: "你好..."},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, Truncate(tt.input, tt.maxLen))
		})
	}
}

func (s *StrxTestSuite) TestDefaultIfEmpty() {
	s.Equal("default", DefaultIfEmpty("", "default"))
	s.Equal("default", DefaultIfEmpty("   ", "default"))
	s.Equal("value", DefaultIfEmpty("value", "default"))
}

func (s *StrxTestSuite) TestUnsafeToBytes() {
	tests := []struct {
		name     string
		input    string
		expected []byte
	}{
		{name: "空字符串", input: "", expected: nil},
		{name: "普通字符串", input: "hello", expected: []byte("hello")},
		{name: "Unicode", input: "你好", expected: []byte("你好")},
		{name: "单字符", input: "a", expected: []byte("a")},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := UnsafeToBytes(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *StrxTestSuite) TestUnsafeToString() {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{name: "nil 切片", input: nil, expected: ""},
		{name: "空切片", input: []byte{}, expected: ""},
		{name: "普通字节", input: []byte("hello"), expected: "hello"},
		{name: "Unicode 字节", input: []byte("你好"), expected: "你好"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := UnsafeToString(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *StrxTestSuite) TestUnsafeRoundTrip() {
	original := "hello world"
	bytes := UnsafeToBytes(original)
	restored := UnsafeToString(bytes)
	s.Equal(original, restored)
}
