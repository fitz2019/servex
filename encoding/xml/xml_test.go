package xml

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// XMLCodecTestSuite XML 编解码器测试套件.
type XMLCodecTestSuite struct {
	suite.Suite
}

func TestXMLCodecSuite(t *testing.T) {
	suite.Run(t, new(XMLCodecTestSuite))
}

type testItem struct {
	Name  string `xml:"name"`
	Price int    `xml:"price"`
}

func (s *XMLCodecTestSuite) TestMarshal() {
	c := codec{}
	data, err := c.Marshal(testItem{Name: "Pen", Price: 10})
	s.NoError(err)
	s.Contains(string(data), "<name>Pen</name>")
}

func (s *XMLCodecTestSuite) TestUnmarshal() {
	c := codec{}
	var item testItem
	err := c.Unmarshal([]byte(`<testItem><name>Pen</name><price>10</price></testItem>`), &item)
	s.NoError(err)
	s.Equal("Pen", item.Name)
	s.Equal(10, item.Price)
}

func (s *XMLCodecTestSuite) TestUnmarshal_InvalidXML() {
	c := codec{}
	var item testItem
	err := c.Unmarshal([]byte(`<invalid`), &item)
	s.Error(err)
}

func (s *XMLCodecTestSuite) TestName() {
	c := codec{}
	s.Equal("xml", c.Name())
}
