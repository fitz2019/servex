package copier

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type CopierTestSuite struct {
	suite.Suite
}

func TestCopierSuite(t *testing.T) {
	suite.Run(t, new(CopierTestSuite))
}

type srcUser struct {
	Name  string
	Email string
	Age   int
}

type dstUser struct {
	Name  string
	Email string
	Age   int
}

type dstPartial struct {
	Name  string
	Email string
}

type srcNested struct {
	Name    string
	Address srcAddress
}

type srcAddress struct {
	City    string
	Country string
}

type dstNested struct {
	Name    string
	Address dstAddress
}

type dstAddress struct {
	City    string
	Country string
}

type srcWithPtr struct {
	Name    string
	Address *srcAddress
}

type dstWithPtr struct {
	Name    string
	Address *dstAddress
}

func (s *CopierTestSuite) TestCopy_Basic() {
	src := &srcUser{Name: "Alice", Email: "alice@test.com", Age: 30}
	dst, err := Copy[dstUser](src)
	s.NoError(err)
	s.Equal("Alice", dst.Name)
	s.Equal("alice@test.com", dst.Email)
	s.Equal(30, dst.Age)
}

func (s *CopierTestSuite) TestCopy_PartialFields() {
	src := &srcUser{Name: "Bob", Email: "bob@test.com", Age: 25}
	dst, err := Copy[dstPartial](src)
	s.NoError(err)
	s.Equal("Bob", dst.Name)
	s.Equal("bob@test.com", dst.Email)
}

func (s *CopierTestSuite) TestCopy_NilSource() {
	_, err := Copy[dstUser, srcUser](nil)
	s.ErrorIs(err, ErrNilSource)
}

func (s *CopierTestSuite) TestCopy_NestedStruct() {
	src := &srcNested{
		Name: "Charlie",
		Address: srcAddress{
			City:    "Beijing",
			Country: "CN",
		},
	}
	dst, err := Copy[dstNested](src)
	s.NoError(err)
	s.Equal("Charlie", dst.Name)
	s.Equal("Beijing", dst.Address.City)
	s.Equal("CN", dst.Address.Country)
}

func (s *CopierTestSuite) TestCopy_NestedPointer() {
	src := &srcWithPtr{
		Name: "Dave",
		Address: &srcAddress{
			City:    "Shanghai",
			Country: "CN",
		},
	}
	dst, err := Copy[dstWithPtr](src)
	s.NoError(err)
	s.Equal("Dave", dst.Name)
	s.NotNil(dst.Address)
	s.Equal("Shanghai", dst.Address.City)
}

func (s *CopierTestSuite) TestCopy_NilNestedPointer() {
	src := &srcWithPtr{Name: "Eve", Address: nil}
	dst, err := Copy[dstWithPtr](src)
	s.NoError(err)
	s.Equal("Eve", dst.Name)
	s.Nil(dst.Address)
}

func (s *CopierTestSuite) TestCopyTo_Basic() {
	src := &srcUser{Name: "Frank", Email: "frank@test.com", Age: 40}
	dst := &dstUser{}
	err := CopyTo(src, dst)
	s.NoError(err)
	s.Equal("Frank", dst.Name)
	s.Equal("frank@test.com", dst.Email)
	s.Equal(40, dst.Age)
}

func (s *CopierTestSuite) TestCopyTo_NilSource() {
	dst := &dstUser{}
	err := CopyTo[dstUser, srcUser](nil, dst)
	s.ErrorIs(err, ErrNilSource)
}

func (s *CopierTestSuite) TestCopyTo_NilDest() {
	src := &srcUser{Name: "Grace"}
	err := CopyTo[dstUser](src, nil)
	s.ErrorIs(err, ErrNilDestination)
}

func (s *CopierTestSuite) TestCopyTo_OverwritesAllFields() {
	src := &srcUser{Name: "Hank"}
	dst := &dstUser{Email: "old@test.com", Age: 50}

	err := CopyTo(src, dst)
	s.NoError(err)
	s.Equal("Hank", dst.Name)
	s.Equal("", dst.Email)
	s.Equal(0, dst.Age)
}

func (s *CopierTestSuite) TestCopyWithOptions_IgnoreFields() {
	src := &srcUser{Name: "Ivy", Email: "ivy@test.com", Age: 28}
	dst, err := CopyWithOptions[dstUser](src, IgnoreFields("Age"))
	s.NoError(err)
	s.Equal("Ivy", dst.Name)
	s.Equal("ivy@test.com", dst.Email)
	s.Equal(0, dst.Age)
}

func (s *CopierTestSuite) TestCopyWithOptions_FieldMapping() {
	type srcOrder struct {
		OrderID string
		Amount  float64
	}
	type dstOrder struct {
		ID     string
		Amount float64
	}

	src := &srcOrder{OrderID: "ORD-001", Amount: 99.9}
	dst, err := CopyWithOptions[dstOrder](src, FieldMapping("OrderID", "ID"))
	s.NoError(err)
	s.Equal("ORD-001", dst.ID)
	s.InDelta(99.9, dst.Amount, 0.01)
}

func (s *CopierTestSuite) TestCopyWithOptions_MultipleOptions() {
	src := &srcUser{Name: "Jack", Email: "jack@test.com", Age: 35}
	dst, err := CopyWithOptions[dstUser](src,
		IgnoreFields("Email"),
		IgnoreFields("Age"),
	)
	s.NoError(err)
	s.Equal("Jack", dst.Name)
	s.Equal("", dst.Email)
	s.Equal(0, dst.Age)
}

func (s *CopierTestSuite) TestCopyWithOptions_NilSource() {
	_, err := CopyWithOptions[dstUser, srcUser](nil)
	s.ErrorIs(err, ErrNilSource)
}

func (s *CopierTestSuite) TestCopyToWithOptions() {
	src := &srcUser{Name: "Kate", Email: "kate@test.com", Age: 22}
	dst := &dstUser{}
	err := CopyToWithOptions(src, dst, IgnoreFields("Age"))
	s.NoError(err)
	s.Equal("Kate", dst.Name)
	s.Equal("kate@test.com", dst.Email)
	s.Equal(0, dst.Age)
}

func (s *CopierTestSuite) TestCopyToWithOptions_NilSource() {
	dst := &dstUser{}
	err := CopyToWithOptions[dstUser, srcUser](nil, dst)
	s.ErrorIs(err, ErrNilSource)
}

func (s *CopierTestSuite) TestCopyToWithOptions_NilDest() {
	src := &srcUser{}
	err := CopyToWithOptions[dstUser](src, nil)
	s.ErrorIs(err, ErrNilDestination)
}

func (s *CopierTestSuite) TestCopy_TypeMismatch_Silent() {
	type src struct {
		Name string
		Data []int
	}
	type dst struct {
		Name string
		Data string
	}

	from := &src{Name: "test", Data: []int{1, 2, 3}}
	result, err := Copy[dst](from)
	s.NoError(err)
	s.Equal("test", result.Name)
	s.Equal("", result.Data)
}

func (s *CopierTestSuite) TestCopy_NoMatchingFields() {
	type src struct {
		Foo string
	}
	type dst struct {
		Bar string
	}

	from := &src{Foo: "hello"}
	result, err := Copy[dst](from)
	s.NoError(err)
	s.Equal("", result.Bar)
}

func (s *CopierTestSuite) TestCopy_SameType() {
	src := &srcUser{Name: "Same", Email: "same@test.com", Age: 99}
	dst, err := Copy[srcUser](src)
	s.NoError(err)
	s.Equal(src.Name, dst.Name)
	s.Equal(src.Email, dst.Email)
	s.Equal(src.Age, dst.Age)

	dst.Name = "Changed"
	s.Equal("Same", src.Name)
}

func (s *CopierTestSuite) TestCopy_EmptyStruct() {
	type empty struct{}
	src := &empty{}
	dst, err := Copy[empty](src)
	s.NoError(err)
	s.NotNil(dst)
}

func (s *CopierTestSuite) TestCopy_ConvertibleTypes() {
	type src struct {
		Value int32
	}
	type dst struct {
		Value int64
	}

	from := &src{Value: 42}
	result, err := Copy[dst](from)
	s.NoError(err)
	s.Equal(int64(42), result.Value)
}
