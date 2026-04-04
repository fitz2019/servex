package randx

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/suite"
)

type RandxTestSuite struct {
	suite.Suite
	r *Rand
}

func TestRandxSuite(t *testing.T) {
	suite.Run(t, new(RandxTestSuite))
}

func (s *RandxTestSuite) SetupTest() {
	s.r = New()
}

func (s *RandxTestSuite) TestRandInt() {
	for range 100 {
		v := s.r.RandInt(0, 10)
		s.GreaterOrEqual(v, 0)
		s.Less(v, 10)
	}
}

func (s *RandxTestSuite) TestRandIntEqualMinMax() {
	v := s.r.RandInt(5, 5)
	s.Equal(5, v)
}

func (s *RandxTestSuite) TestRandInt64() {
	for range 100 {
		v := s.r.RandInt64(100, 200)
		s.GreaterOrEqual(v, int64(100))
		s.Less(v, int64(200))
	}
}

func (s *RandxTestSuite) TestRandString() {
	str := s.r.RandString(20)
	s.Len(str, 20)
	for _, c := range str {
		s.True(unicode.IsPrint(c), "字符 %q 应为可打印 ASCII", c)
	}
}

func (s *RandxTestSuite) TestRandStringEmpty() {
	s.Equal("", s.r.RandString(0))
	s.Equal("", s.r.RandString(-1))
}

func (s *RandxTestSuite) TestRandAlphanumeric() {
	str := s.r.RandAlphanumeric(50)
	s.Len(str, 50)
	for _, c := range str {
		s.True(unicode.IsLetter(c) || unicode.IsDigit(c), "字符 %q 应为字母或数字", c)
	}
}

func (s *RandxTestSuite) TestRandAlpha() {
	str := s.r.RandAlpha(30)
	s.Len(str, 30)
	for _, c := range str {
		s.True(unicode.IsLetter(c), "字符 %q 应为字母", c)
	}
}

func (s *RandxTestSuite) TestRandDigits() {
	str := s.r.RandDigits(20)
	s.Len(str, 20)
	for _, c := range str {
		s.True(unicode.IsDigit(c), "字符 %q 应为数字", c)
	}
}

func (s *RandxTestSuite) TestRandElement() {
	slice := []int{1, 2, 3, 4, 5}
	v, ok := RandElement(s.r, slice)
	s.True(ok)
	s.Contains(slice, v)
}

func (s *RandxTestSuite) TestRandElementEmpty() {
	v, ok := RandElement(s.r, []int{})
	s.False(ok)
	s.Equal(0, v)
}

func (s *RandxTestSuite) TestSample() {
	slice := []int{1, 2, 3, 4, 5}
	result := Sample(s.r, slice, 3)
	s.Len(result, 3)
	for _, v := range result {
		s.Contains(slice, v)
	}
}

func (s *RandxTestSuite) TestSampleAllWhenNLarger() {
	slice := []int{1, 2, 3}
	result := Sample(s.r, slice, 10)
	s.Len(result, 3)
}

func (s *RandxTestSuite) TestSampleEmpty() {
	result := Sample(s.r, []int{}, 3)
	s.Nil(result)
}

func (s *RandxTestSuite) TestShuffle() {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	original := make([]int, len(slice))
	copy(original, slice)
	Shuffle(s.r, slice)
	s.ElementsMatch(original, slice)
}

func (s *RandxTestSuite) TestNewSecure() {
	r := NewSecure()
	str := r.RandAlphanumeric(16)
	s.Len(str, 16)
}
