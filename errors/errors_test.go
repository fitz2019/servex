package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestNew(t *testing.T) {
	err := New(100401, "auth.token.expired", "令牌已过期")
	assert.Equal(t, 100401, err.Code)
	assert.Equal(t, "auth.token.expired", err.Key)
	assert.Equal(t, "令牌已过期", err.Message)
	assert.Equal(t, 0, err.HTTP)
	assert.Equal(t, codes.OK, err.GRPC)
	assert.Nil(t, err.Metadata)
	assert.Nil(t, err.cause)
}

func TestError_Error(t *testing.T) {
	err := New(100401, "auth.token.expired", "令牌已过期")
	assert.Equal(t, "[100401] auth.token.expired: 令牌已过期", err.Error())
}

func TestError_WithHTTP(t *testing.T) {
	err := New(100401, "auth.token.expired", "令牌已过期").WithHTTP(401)
	assert.Equal(t, 401, err.HTTP)
}

func TestError_WithGRPC(t *testing.T) {
	err := New(100401, "auth.token.expired", "令牌已过期").WithGRPC(codes.Unauthenticated)
	assert.Equal(t, codes.Unauthenticated, err.GRPC)
}

func TestError_ChainedBuilder(t *testing.T) {
	err := New(100401, "auth.token.expired", "令牌已过期").
		WithHTTP(401).
		WithGRPC(codes.Unauthenticated)
	assert.Equal(t, 100401, err.Code)
	assert.Equal(t, 401, err.HTTP)
	assert.Equal(t, codes.Unauthenticated, err.GRPC)
}

func TestError_WithCause(t *testing.T) {
	original := New(100401, "auth.token.expired", "令牌已过期").WithHTTP(401)
	cause := fmt.Errorf("token parse failed")
	wrapped := original.WithCause(cause)

	assert.Equal(t, "[100401] auth.token.expired: 令牌已过期: token parse failed", wrapped.Error())
	assert.ErrorIs(t, wrapped, cause)

	assert.Equal(t, "[100401] auth.token.expired: 令牌已过期", original.Error())
	assert.Nil(t, original.cause)
}

func TestError_WithMeta(t *testing.T) {
	original := New(100401, "auth.token.expired", "令牌已过期")
	withMeta := original.WithMeta("user_id", "123")

	assert.Equal(t, "123", withMeta.Metadata["user_id"])
	assert.Nil(t, original.Metadata)
}

func TestError_WithMeta_Multiple(t *testing.T) {
	err := New(100401, "auth.token.expired", "令牌已过期").
		WithMeta("user_id", "123").
		WithMeta("ip", "1.2.3.4")

	assert.Equal(t, "123", err.Metadata["user_id"])
	assert.Equal(t, "1.2.3.4", err.Metadata["ip"])
}

func TestError_WithMessage(t *testing.T) {
	original := New(100401, "auth.token.expired", "令牌已过期").WithHTTP(401)
	replaced := original.WithMessage("Token expired")

	assert.Equal(t, "Token expired", replaced.Message)
	assert.Equal(t, "令牌已过期", original.Message)
	assert.Equal(t, 401, replaced.HTTP)
}

func TestError_Is_StandardLibrary(t *testing.T) {
	ErrAuth := New(100401, "auth.token.expired", "令牌已过期")
	wrapped := ErrAuth.WithCause(fmt.Errorf("bad token"))

	assert.True(t, errors.Is(wrapped, ErrAuth))

	ErrOther := New(200404, "user.not_found", "用户不存在")
	assert.False(t, errors.Is(wrapped, ErrOther))
}

func TestError_As(t *testing.T) {
	ErrAuth := New(100401, "auth.token.expired", "令牌已过期").WithHTTP(401)
	wrapped := fmt.Errorf("outer: %w", ErrAuth)

	var target *Error
	require.True(t, errors.As(wrapped, &target))
	assert.Equal(t, 100401, target.Code)
	assert.Equal(t, 401, target.HTTP)
}

func TestFromError(t *testing.T) {
	t.Run("from *Error", func(t *testing.T) {
		err := New(100401, "auth.token.expired", "令牌已过期")
		got, ok := FromError(err)
		assert.True(t, ok)
		assert.Equal(t, 100401, got.Code)
	})

	t.Run("from wrapped *Error", func(t *testing.T) {
		err := New(100401, "auth.token.expired", "令牌已过期")
		wrapped := fmt.Errorf("outer: %w", err)
		got, ok := FromError(wrapped)
		assert.True(t, ok)
		assert.Equal(t, 100401, got.Code)
	})

	t.Run("from WithCause", func(t *testing.T) {
		err := New(100401, "auth.token.expired", "令牌已过期").WithCause(fmt.Errorf("bad"))
		got, ok := FromError(err)
		assert.True(t, ok)
		assert.Equal(t, 100401, got.Code)
	})

	t.Run("from nil", func(t *testing.T) {
		_, ok := FromError(nil)
		assert.False(t, ok)
	})

	t.Run("from standard error", func(t *testing.T) {
		_, ok := FromError(fmt.Errorf("plain error"))
		assert.False(t, ok)
	})
}

func TestCodeIs(t *testing.T) {
	ErrAuth := New(100401, "auth.token.expired", "令牌已过期")

	t.Run("direct match", func(t *testing.T) {
		assert.True(t, CodeIs(ErrAuth, ErrAuth))
	})

	t.Run("wrapped match", func(t *testing.T) {
		wrapped := ErrAuth.WithCause(fmt.Errorf("bad"))
		assert.True(t, CodeIs(wrapped, ErrAuth))
	})

	t.Run("fmt wrapped match", func(t *testing.T) {
		wrapped := fmt.Errorf("outer: %w", ErrAuth)
		assert.True(t, CodeIs(wrapped, ErrAuth))
	})

	t.Run("no match", func(t *testing.T) {
		other := New(200404, "user.not_found", "用户不存在")
		assert.False(t, CodeIs(other, ErrAuth))
	})

	t.Run("nil error", func(t *testing.T) {
		assert.False(t, CodeIs(nil, ErrAuth))
	})
}
