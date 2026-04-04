package rdbms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type EncryptColumnTestSuite struct {
	suite.Suite
}

func TestEncryptColumnSuite(t *testing.T) {
	suite.Run(t, new(EncryptColumnTestSuite))
}

// AES-256 需要 32 字节密钥.
const testKey = "01234567890123456789012345678901"

func (s *EncryptColumnTestSuite) TestNewEncryptColumn() {
	ec := NewEncryptColumn("secret-ssn", testKey)
	s.True(ec.Valid)
	s.Equal("secret-ssn", ec.Val)
	s.Equal(testKey, ec.Key)
}

func (s *EncryptColumnTestSuite) TestNullEncryptColumn() {
	ec := NullEncryptColumn[string](testKey)
	s.False(ec.Valid)
}

func (s *EncryptColumnTestSuite) TestValue_Valid() {
	ec := NewEncryptColumn("hello", testKey)
	val, err := ec.Value()
	s.NoError(err)
	s.NotNil(val)
	s.IsType("", val)
}

func (s *EncryptColumnTestSuite) TestValue_Null() {
	ec := NullEncryptColumn[string](testKey)
	val, err := ec.Value()
	s.NoError(err)
	s.Nil(val)
}

func (s *EncryptColumnTestSuite) TestRoundTrip_String() {
	original := NewEncryptColumn("sensitive-data-123", testKey)

	val, err := original.Value()
	s.NoError(err)

	restored := NullEncryptColumn[string](testKey)
	err = restored.Scan(val)
	s.NoError(err)
	s.True(restored.Valid)
	s.Equal("sensitive-data-123", restored.Val)
}

func (s *EncryptColumnTestSuite) TestRoundTrip_Struct() {
	type secret struct {
		SSN  string `json:"ssn"`
		Code int    `json:"code"`
	}

	original := NewEncryptColumn(secret{SSN: "123-45-6789", Code: 42}, testKey)

	val, err := original.Value()
	s.NoError(err)

	restored := NullEncryptColumn[secret](testKey)
	err = restored.Scan(val)
	s.NoError(err)
	s.True(restored.Valid)
	s.Equal("123-45-6789", restored.Val.SSN)
	s.Equal(42, restored.Val.Code)
}

func (s *EncryptColumnTestSuite) TestScan_Nil() {
	ec := NewEncryptColumn("old", testKey)
	err := ec.Scan(nil)
	s.NoError(err)
	s.False(ec.Valid)
}

func (s *EncryptColumnTestSuite) TestScan_UnsupportedType() {
	ec := NullEncryptColumn[string](testKey)
	err := ec.Scan(42)
	s.Error(err)
}

func (s *EncryptColumnTestSuite) TestScan_WithoutKey() {
	original := NewEncryptColumn("data", testKey)
	val, err := original.Value()
	s.NoError(err)

	noKey := NullEncryptColumn[string]("")
	err = noKey.Scan(val)
	s.NoError(err)
	s.False(noKey.Valid)
	s.NotEmpty(noKey.Ciphertext)
}

func (s *EncryptColumnTestSuite) TestDecrypt_AfterScanWithoutKey() {
	original := NewEncryptColumn("secret-data", testKey)
	val, err := original.Value()
	s.NoError(err)

	ec := NullEncryptColumn[string]("")
	err = ec.Scan(val)
	s.NoError(err)
	s.False(ec.Valid)

	ec.Key = testKey
	err = ec.Decrypt()
	s.NoError(err)
	s.True(ec.Valid)
	s.Equal("secret-data", ec.Val)
	s.Empty(ec.Ciphertext)
}

func (s *EncryptColumnTestSuite) TestDecrypt_AlreadyValid() {
	ec := NewEncryptColumn("data", testKey)
	err := ec.Decrypt()
	s.NoError(err)
}

func (s *EncryptColumnTestSuite) TestDecrypt_NoKey() {
	ec := &EncryptColumn[string]{Ciphertext: "something"}
	err := ec.Decrypt()
	s.Error(err)
}

func (s *EncryptColumnTestSuite) TestDecrypt_NoCiphertext() {
	ec := &EncryptColumn[string]{Key: testKey}
	err := ec.Decrypt()
	s.Error(err)
}

func (s *EncryptColumnTestSuite) TestScan_WrongKey() {
	original := NewEncryptColumn("data", testKey)
	val, err := original.Value()
	s.NoError(err)

	wrongKey := "99999999999999999999999999999999"
	restored := NullEncryptColumn[string](wrongKey)
	err = restored.Scan(val)
	s.Error(err)
}

func (s *EncryptColumnTestSuite) TestScan_InvalidBase64() {
	ec := NullEncryptColumn[string](testKey)
	err := ec.Scan("not-valid-base64!!!")
	s.Error(err)
}

func (s *EncryptColumnTestSuite) TestScan_ByteInput() {
	original := NewEncryptColumn("test", testKey)
	val, err := original.Value()
	s.NoError(err)

	restored := NullEncryptColumn[string](testKey)
	err = restored.Scan([]byte(val.(string)))
	s.NoError(err)
	s.True(restored.Valid)
	s.Equal("test", restored.Val)
}

func (s *EncryptColumnTestSuite) TestValue_InvalidKey() {
	// Key 长度不合法（非 16/24/32 字节）
	ec := NewEncryptColumn("data", "short-key")
	_, err := ec.Value()
	s.Error(err)
}

func (s *EncryptColumnTestSuite) TestDifferentEncryptions() {
	// 相同数据加密两次，密文应不同（随机 nonce）
	ec1 := NewEncryptColumn("same-data", testKey)
	ec2 := NewEncryptColumn("same-data", testKey)

	val1, err1 := ec1.Value()
	val2, err2 := ec2.Value()

	s.NoError(err1)
	s.NoError(err2)
	s.NotEqual(val1, val2)
}
