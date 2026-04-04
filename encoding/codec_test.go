package encoding

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

// CodecTestSuite 编解码器测试套件.
type CodecTestSuite struct {
	suite.Suite
}

func TestCodecSuite(t *testing.T) {
	suite.Run(t, new(CodecTestSuite))
}

// mockCodec 模拟编解码器.
type mockCodec struct {
	name string
}

func (c mockCodec) Marshal(any) ([]byte, error)       { return nil, nil }
func (c mockCodec) Unmarshal([]byte, any) error        { return nil }
func (c mockCodec) Name() string                       { return c.name }

func (s *CodecTestSuite) SetupTest() {
	// 清理全局注册表
	registryMu.Lock()
	registry = make(map[string]Codec)
	registryMu.Unlock()
}

func (s *CodecTestSuite) TestRegisterAndGet() {
	c := mockCodec{name: "json"}
	RegisterCodec(c)

	got := GetCodec("json")
	s.NotNil(got)
	s.Equal("json", got.Name())
}

func (s *CodecTestSuite) TestGetCodec_NotFound() {
	got := GetCodec("nonexistent")
	s.Nil(got)
}

func (s *CodecTestSuite) TestRegisterCodec_Override() {
	c1 := mockCodec{name: "json"}
	c2 := mockCodec{name: "json"}
	RegisterCodec(c1)
	RegisterCodec(c2)

	got := GetCodec("json")
	s.NotNil(got)
	s.Equal("json", got.Name())
}

func (s *CodecTestSuite) TestCodecForRequest_JSON() {
	RegisterCodec(mockCodec{name: "json"})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "application/json")

	c := CodecForRequest(r, "Accept")
	s.NotNil(c)
	s.Equal("json", c.Name())
}

func (s *CodecTestSuite) TestCodecForRequest_XML() {
	RegisterCodec(mockCodec{name: "json"})
	RegisterCodec(mockCodec{name: "xml"})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "application/xml")

	c := CodecForRequest(r, "Accept")
	s.NotNil(c)
	s.Equal("xml", c.Name())
}

func (s *CodecTestSuite) TestCodecForRequest_Protobuf() {
	RegisterCodec(mockCodec{name: "json"})
	RegisterCodec(mockCodec{name: "proto"})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Content-Type", "application/x-protobuf")

	c := CodecForRequest(r, "Content-Type")
	s.NotNil(c)
	s.Equal("proto", c.Name())
}

func (s *CodecTestSuite) TestCodecForRequest_WithCharset() {
	RegisterCodec(mockCodec{name: "json"})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "application/json; charset=utf-8")

	c := CodecForRequest(r, "Accept")
	s.NotNil(c)
	s.Equal("json", c.Name())
}

func (s *CodecTestSuite) TestCodecForRequest_VendorType() {
	RegisterCodec(mockCodec{name: "json"})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "application/vnd.api+json")

	c := CodecForRequest(r, "Accept")
	s.NotNil(c)
	s.Equal("json", c.Name())
}

func (s *CodecTestSuite) TestCodecForRequest_EmptyHeader_FallbackJSON() {
	RegisterCodec(mockCodec{name: "json"})

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	c := CodecForRequest(r, "Accept")
	s.NotNil(c)
	s.Equal("json", c.Name())
}

func (s *CodecTestSuite) TestCodecForRequest_UnknownType_FallbackJSON() {
	RegisterCodec(mockCodec{name: "json"})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "application/unknown-format")

	c := CodecForRequest(r, "Accept")
	s.NotNil(c)
	s.Equal("json", c.Name())
}

func (s *CodecTestSuite) TestCodecForRequest_TextXML() {
	RegisterCodec(mockCodec{name: "json"})
	RegisterCodec(mockCodec{name: "xml"})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "text/xml")

	c := CodecForRequest(r, "Accept")
	s.NotNil(c)
	s.Equal("xml", c.Name())
}

func (s *CodecTestSuite) TestSubtypeFromHeader() {
	tests := []struct {
		input    string
		expected string
	}{
		{"application/json", "json"},
		{"application/xml", "xml"},
		{"application/json; charset=utf-8", "json"},
		{"application/x-protobuf", "proto"},
		{"application/vnd.api+json", "json"},
		{"text/xml", "xml"},
		{"", ""},
		{"json", "json"},
	}

	for _, tc := range tests {
		s.Equal(tc.expected, subtypeFromHeader(tc.input), "input: %s", tc.input)
	}
}
