package rdbms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type JsonColumnTestSuite struct {
	suite.Suite
}

func TestJsonColumnSuite(t *testing.T) {
	suite.Run(t, new(JsonColumnTestSuite))
}

type address struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

func (s *JsonColumnTestSuite) TestNewJsonColumn() {
	jc := NewJsonColumn(address{City: "Beijing", Country: "CN"})
	s.True(jc.Valid)
	s.Equal("Beijing", jc.Val.City)
}

func (s *JsonColumnTestSuite) TestNullJsonColumn() {
	jc := NullJsonColumn[address]()
	s.False(jc.Valid)
}

func (s *JsonColumnTestSuite) TestValue_Valid() {
	jc := NewJsonColumn(address{City: "Shanghai", Country: "CN"})
	val, err := jc.Value()
	s.NoError(err)
	s.Contains(val.(string), `"city":"Shanghai"`)
}

func (s *JsonColumnTestSuite) TestValue_Null() {
	jc := NullJsonColumn[address]()
	val, err := jc.Value()
	s.NoError(err)
	s.Nil(val)
}

func (s *JsonColumnTestSuite) TestScan_String() {
	jc := NullJsonColumn[address]()
	err := jc.Scan(`{"city":"Tokyo","country":"JP"}`)
	s.NoError(err)
	s.True(jc.Valid)
	s.Equal("Tokyo", jc.Val.City)
	s.Equal("JP", jc.Val.Country)
}

func (s *JsonColumnTestSuite) TestScan_Bytes() {
	jc := NullJsonColumn[address]()
	err := jc.Scan([]byte(`{"city":"Seoul","country":"KR"}`))
	s.NoError(err)
	s.True(jc.Valid)
	s.Equal("Seoul", jc.Val.City)
}

func (s *JsonColumnTestSuite) TestScan_Nil() {
	jc := NewJsonColumn(address{City: "old"})
	err := jc.Scan(nil)
	s.NoError(err)
	s.False(jc.Valid)
}

func (s *JsonColumnTestSuite) TestScan_InvalidJSON() {
	jc := NullJsonColumn[address]()
	err := jc.Scan(`{invalid json}`)
	s.Error(err)
}

func (s *JsonColumnTestSuite) TestScan_UnsupportedType() {
	jc := NullJsonColumn[address]()
	err := jc.Scan(42)
	s.Error(err)
}

func (s *JsonColumnTestSuite) TestRoundTrip() {
	original := NewJsonColumn(address{City: "London", Country: "UK"})

	val, err := original.Value()
	s.NoError(err)

	restored := NullJsonColumn[address]()
	err = restored.Scan(val)
	s.NoError(err)
	s.True(restored.Valid)
	s.Equal(original.Val, restored.Val)
}

func (s *JsonColumnTestSuite) TestJsonColumn_Map() {
	data := map[string]any{"key": "value", "num": float64(42)}
	jc := NewJsonColumn(data)

	val, err := jc.Value()
	s.NoError(err)

	var restored JsonColumn[map[string]any]
	err = restored.Scan(val)
	s.NoError(err)
	s.Equal("value", restored.Val["key"])
	s.Equal(float64(42), restored.Val["num"])
}

func (s *JsonColumnTestSuite) TestJsonColumn_Slice() {
	jc := NewJsonColumn([]string{"a", "b", "c"})

	val, err := jc.Value()
	s.NoError(err)

	var restored JsonColumn[[]string]
	err = restored.Scan(val)
	s.NoError(err)
	s.Equal([]string{"a", "b", "c"}, restored.Val)
}
