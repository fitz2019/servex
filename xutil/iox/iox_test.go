package iox

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type IoxTestSuite struct {
	suite.Suite
}

func TestIoxSuite(t *testing.T) {
	suite.Run(t, new(IoxTestSuite))
}

func (s *IoxTestSuite) TestReadAll() {
	data, err := ReadAll(strings.NewReader("hello"))
	s.NoError(err)
	s.Equal([]byte("hello"), data)
}

func (s *IoxTestSuite) TestReadString() {
	str, err := ReadString(strings.NewReader("world"))
	s.NoError(err)
	s.Equal("world", str)
}

func (s *IoxTestSuite) TestReadLines() {
	input := "line1\nline2\r\nline3"
	lines, err := ReadLines(strings.NewReader(input))
	s.NoError(err)
	s.Equal([]string{"line1", "line2", "line3"}, lines)
}

func (s *IoxTestSuite) TestReadLinesEmpty() {
	lines, err := ReadLines(strings.NewReader(""))
	s.NoError(err)
	s.Empty(lines)
}

func (s *IoxTestSuite) TestDrain() {
	n, err := Drain(strings.NewReader("12345"))
	s.NoError(err)
	s.Equal(int64(5), n)
}

func (s *IoxTestSuite) TestWriteString() {
	var sb strings.Builder
	n, err := WriteString(&sb, "hello")
	s.NoError(err)
	s.Equal(5, n)
	s.Equal("hello", sb.String())
}

func (s *IoxTestSuite) TestMultiCloser_AllClosed() {
	var closed [2]bool
	c1 := &mockCloser{onClose: func() error { closed[0] = true; return nil }}
	c2 := &mockCloser{onClose: func() error { closed[1] = true; return nil }}

	mc := MultiCloser(c1, c2)
	s.NoError(mc.Close())
	s.True(closed[0])
	s.True(closed[1])
}

func (s *IoxTestSuite) TestMultiCloser_CollectsErrors() {
	c1 := &mockCloser{onClose: func() error { return io.ErrClosedPipe }}
	c2 := &mockCloser{onClose: func() error { return io.EOF }}

	mc := MultiCloser(c1, c2)
	err := mc.Close()
	s.Error(err)
}

func (s *IoxTestSuite) TestCloseAndLog() {
	var logged error
	c := &mockCloser{onClose: func() error { return io.ErrClosedPipe }}
	CloseAndLog(c, func(err error) { logged = err })
	s.ErrorIs(logged, io.ErrClosedPipe)
}

func (s *IoxTestSuite) TestLimitReadCloser_WithinLimit() {
	rc := io.NopCloser(strings.NewReader("hello world"))
	limited := LimitReadCloser(rc, 5)
	data, _ := io.ReadAll(limited)
	s.Equal([]byte("hello"), data)
}

func (s *IoxTestSuite) TestLimitReadCloser_ExceedsLimit() {
	rc := io.NopCloser(strings.NewReader("hello world"))
	limited := LimitReadCloser(rc, 3)
	buf := make([]byte, 10)
	n, err := limited.Read(buf)
	s.Equal(3, n)
	s.Error(err)
}

func (s *IoxTestSuite) TestLimitReadCloser_Close() {
	var closed bool
	mc := &mockCloser{onClose: func() error { closed = true; return nil }}
	rc := &mockReadCloser{ReadCloser: io.NopCloser(strings.NewReader("data")), mockCloser: mc}
	limited := LimitReadCloser(rc, 10)
	s.NoError(limited.Close())
	s.True(closed)
}

// mockCloser 用于测试的 io.Closer 实现.
type mockCloser struct {
	onClose func() error
}

func (m *mockCloser) Close() error { return m.onClose() }

// mockReadCloser 将读取委托给内嵌 io.ReadCloser，将关闭委托给 mockCloser.
type mockReadCloser struct {
	io.ReadCloser
	*mockCloser
}

func (m *mockReadCloser) Close() error { return m.mockCloser.Close() }
