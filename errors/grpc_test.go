package errors

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func TestToGRPCStatus(t *testing.T) {
	errTokenExpired := New(100401, "auth.token.expired", "令牌已过期").
		WithHTTP(http.StatusUnauthorized).WithGRPC(codes.Unauthenticated)

	t.Run("from *Error", func(t *testing.T) {
		st := ToGRPCStatus(errTokenExpired)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Contains(t, st.Message(), "令牌已过期")
	})

	t.Run("from *Error without GRPC set", func(t *testing.T) {
		err := New(999, "unknown", "未知错误")
		st := ToGRPCStatus(err)
		assert.Equal(t, codes.Internal, st.Code())
	})

	t.Run("from standard error", func(t *testing.T) {
		st := ToGRPCStatus(fmt.Errorf("plain error"))
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "plain error", st.Message())
	})

	t.Run("from nil", func(t *testing.T) {
		st := ToGRPCStatus(nil)
		assert.Equal(t, codes.OK, st.Code())
	})
}

func TestFromGRPCStatus(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		original := New(100401, "auth.token.expired", "令牌已过期").
			WithHTTP(401).WithGRPC(codes.Unauthenticated)
		st := ToGRPCStatus(original)

		restored := FromGRPCStatus(st)
		require.NotNil(t, restored)
		assert.Equal(t, 100401, restored.Code)
		assert.Equal(t, "auth.token.expired", restored.Key)
		assert.Equal(t, "令牌已过期", restored.Message)
	})

	t.Run("from plain grpc status", func(t *testing.T) {
		st := grpcstatus.New(codes.NotFound, "not found")
		restored := FromGRPCStatus(st)
		require.NotNil(t, restored)
		assert.Equal(t, int(codes.NotFound), restored.Code)
		assert.Equal(t, "not found", restored.Message)
	})
}

func TestUnaryServerInterceptor(t *testing.T) {
	errTokenExpired := New(100401, "auth.token.expired", "令牌已过期").
		WithHTTP(http.StatusUnauthorized).WithGRPC(codes.Unauthenticated)
	interceptor := UnaryServerInterceptor()

	t.Run("handler returns *Error", func(t *testing.T) {
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, errTokenExpired
		}
		_, err := interceptor(t.Context(), nil, &grpc.UnaryServerInfo{}, handler)
		require.Error(t, err)

		st, ok := grpcstatus.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("handler returns nil", func(t *testing.T) {
		handler := func(ctx context.Context, req any) (any, error) {
			return "ok", nil
		}
		resp, err := interceptor(t.Context(), nil, &grpc.UnaryServerInfo{}, handler)
		assert.NoError(t, err)
		assert.Equal(t, "ok", resp)
	})

	t.Run("handler returns standard error", func(t *testing.T) {
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, fmt.Errorf("plain error")
		}
		_, err := interceptor(t.Context(), nil, &grpc.UnaryServerInfo{}, handler)
		require.Error(t, err)

		st, ok := grpcstatus.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
	})
}
