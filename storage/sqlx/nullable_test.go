package sqlx

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

type NullableTestSuite struct {
	suite.Suite
}

func TestNullableSuite(t *testing.T) {
	suite.Run(t, new(NullableTestSuite))
}

func (s *NullableTestSuite) TestOf() {
	n := Of(42)
	s.True(n.Valid)
	s.Equal(42, n.Val)
}

func (s *NullableTestSuite) TestNull() {
	n := Null[int]()
	s.False(n.Valid)
}

func (s *NullableTestSuite) TestValueOr() {
	s.Equal(42, Of(42).ValueOr(0))
	s.Equal(0, Null[int]().ValueOr(0))
}

func (s *NullableTestSuite) TestMarshalJSON_Valid() {
	n := Of("hello")
	data, err := json.Marshal(n)
	s.NoError(err)
	s.Equal(`"hello"`, string(data))
}

func (s *NullableTestSuite) TestMarshalJSON_Null() {
	n := Null[string]()
	data, err := json.Marshal(n)
	s.NoError(err)
	s.Equal("null", string(data))
}

func (s *NullableTestSuite) TestUnmarshalJSON_Valid() {
	var n Nullable[int]
	err := json.Unmarshal([]byte("123"), &n)
	s.NoError(err)
	s.True(n.Valid)
	s.Equal(123, n.Val)
}

func (s *NullableTestSuite) TestUnmarshalJSON_Null() {
	var n Nullable[int]
	err := json.Unmarshal([]byte("null"), &n)
	s.NoError(err)
	s.False(n.Valid)
}

func (s *NullableTestSuite) TestMarshalUnmarshalRoundtrip() {
	original := Of(3.14)
	data, err := json.Marshal(original)
	s.NoError(err)

	var restored Nullable[float64]
	err = json.Unmarshal(data, &restored)
	s.NoError(err)
	s.True(restored.Valid)
	s.InDelta(3.14, restored.Val, 1e-9)
}

func (s *NullableTestSuite) TestValue_Valid() {
	n := Of(int64(100))
	v, err := n.Value()
	s.NoError(err)
	s.Equal(int64(100), v)
}

func (s *NullableTestSuite) TestValue_Null() {
	n := Null[string]()
	v, err := n.Value()
	s.NoError(err)
	s.Nil(v)
}

func (s *NullableTestSuite) TestScan_Valid() {
	var n Nullable[string]
	err := n.Scan("test")
	s.NoError(err)
	s.True(n.Valid)
	s.Equal("test", n.Val)
}

func (s *NullableTestSuite) TestScan_Null() {
	var n Nullable[string]
	err := n.Scan(nil)
	s.NoError(err)
	s.False(n.Valid)
}

func (s *NullableTestSuite) TestNullableString() {
	n := NullableString(sql.NullString{String: "hello", Valid: true})
	s.True(n.Valid)
	s.Equal("hello", n.Val)

	n2 := NullableString(sql.NullString{})
	s.False(n2.Valid)
}

func (s *NullableTestSuite) TestNullableInt64() {
	n := NullableInt64(sql.NullInt64{Int64: 42, Valid: true})
	s.True(n.Valid)
	s.Equal(int64(42), n.Val)

	n2 := NullableInt64(sql.NullInt64{})
	s.False(n2.Valid)
}
