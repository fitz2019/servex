package validation

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testUser struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=150"`
}

func TestValidate_Required(t *testing.T) {
	v := New()

	err := v.Validate(&testUser{})
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(ve.Errors), 1)

	// name 应该在错误中
	var found bool
	for _, e := range ve.Errors {
		if e.Field == "name" && e.Tag == "required" {
			found = true
			assert.Contains(t, e.Message, "name")
			break
		}
	}
	assert.True(t, found, "应包含 name 字段的 required 错误")
}

func TestValidate_Email(t *testing.T) {
	v := New()

	err := v.Validate(&testUser{
		Name:  "test",
		Email: "invalid-email",
		Age:   25,
	})
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Len(t, ve.Errors, 1)
	assert.Equal(t, "email", ve.Errors[0].Field)
	assert.Equal(t, "email", ve.Errors[0].Tag)
}

func TestValidate_Multiple(t *testing.T) {
	v := New()

	// 所有字段都不合法
	err := v.Validate(&testUser{
		Email: "bad",
		Age:   -1,
	})
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok)
	// name required, email invalid, age < 0
	assert.GreaterOrEqual(t, len(ve.Errors), 2)

	// 验证 Error() 输出
	errMsg := ve.Error()
	assert.Contains(t, errMsg, "validation failed:")
}

func TestValidate_Success(t *testing.T) {
	v := New()

	err := v.Validate(&testUser{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   25,
	})
	assert.NoError(t, err)
}

func TestValidateField(t *testing.T) {
	v := New()

	// 合法 email
	err := v.ValidateField("test@example.com", "required,email")
	assert.NoError(t, err)

	// 非法 email
	err = v.ValidateField("not-email", "email")
	assert.Error(t, err)

	// required 空字符串
	err = v.ValidateField("", "required")
	assert.Error(t, err)
}

func TestCustomMessages(t *testing.T) {
	v := New(WithMessages(map[string]string{
		"required": "{field}是必填项",
	}))

	err := v.Validate(&testUser{})
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok)

	for _, e := range ve.Errors {
		if e.Tag == "required" {
			assert.Contains(t, e.Message, "是必填项")
		}
	}
}

func TestRegisterValidation(t *testing.T) {
	v := New()

	// 注册自定义规则：必须以 "test_" 开头
	err := v.RegisterValidation("test_prefix", func(fl validator.FieldLevel) bool {
		val, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}
		return len(val) > 5 && val[:5] == "test_"
	})
	require.NoError(t, err)

	type item struct {
		Code string `json:"code" validate:"test_prefix"`
	}

	// 不合法
	err = v.Validate(&item{Code: "abc"})
	assert.Error(t, err)

	// 合法
	err = v.Validate(&item{Code: "test_ok"})
	assert.NoError(t, err)
}

func TestParseErrors(t *testing.T) {
	v := New()
	err := v.Validate(&testUser{})
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.NotEmpty(t, ve.Errors)

	for _, e := range ve.Errors {
		assert.NotEmpty(t, e.Field)
		assert.NotEmpty(t, e.Tag)
		assert.NotEmpty(t, e.Message)
	}
}

func TestWithLocale_English(t *testing.T) {
	v := New(WithLocale("en"))

	err := v.Validate(&testUser{})
	require.Error(t, err)

	ve, ok := err.(*ValidationError)
	require.True(t, ok)

	for _, e := range ve.Errors {
		if e.Tag == "required" {
			assert.Contains(t, e.Message, "is required")
		}
	}
}

func TestWithTagName(t *testing.T) {
	type item struct {
		Name string `json:"name" binding:"required"`
	}

	v := New(WithTagName("binding"))
	err := v.Validate(&item{})
	assert.Error(t, err)

	err = v.Validate(&item{Name: "ok"})
	assert.NoError(t, err)
}
