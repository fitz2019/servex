package tenant

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryServerInterceptor_Success(t *testing.T) {
	resolver := &mockResolver{
		tenant: &testTenant{id: "grpc-t", enabled: true},
	}

	interceptor := UnaryServerInterceptor(resolver)

	md := metadata.New(map[string]string{"x-tenant-token": "token-123"})
	ctx := metadata.NewIncomingContext(t.Context(), md)

	var gotID string
	handler := func(ctx context.Context, _ any) (any, error) {
		gotID = ID(ctx)
		return "ok", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test/Method"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("resp = %v, want %q", resp, "ok")
	}
	if gotID != "grpc-t" {
		t.Fatalf("gotID = %q, want %q", gotID, "grpc-t")
	}
}

func TestUnaryServerInterceptor_MissingToken(t *testing.T) {
	resolver := &mockResolver{}
	interceptor := UnaryServerInterceptor(resolver)

	handler := func(_ context.Context, _ any) (any, error) { return nil, nil }
	_, err := interceptor(t.Context(), nil, &grpc.UnaryServerInfo{FullMethod: "/test/Method"}, handler)

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestUnaryServerInterceptor_Disabled(t *testing.T) {
	resolver := &mockResolver{
		tenant: &testTenant{id: "disabled-t", enabled: false},
	}
	interceptor := UnaryServerInterceptor(resolver)

	md := metadata.New(map[string]string{"x-tenant-token": "token"})
	ctx := metadata.NewIncomingContext(t.Context(), md)

	handler := func(_ context.Context, _ any) (any, error) { return nil, nil }
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test/Method"}, handler)

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Fatalf("code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestUnaryServerInterceptor_Skipper(t *testing.T) {
	resolver := &mockResolver{err: errors.New("should not be called")}
	interceptor := UnaryServerInterceptor(resolver,
		WithSkipper(func(_ context.Context, _ any) bool { return true }),
	)

	handler := func(_ context.Context, _ any) (any, error) { return "skipped", nil }
	resp, err := interceptor(t.Context(), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "skipped" {
		t.Fatalf("resp = %v, want %q", resp, "skipped")
	}
}

func TestUnaryServerInterceptor_PanicOnNilResolver(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("应在 nil resolver 时 panic")
		}
	}()
	UnaryServerInterceptor(nil)
}

func TestStreamServerInterceptor_PanicOnNilResolver(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("应在 nil resolver 时 panic")
		}
	}()
	StreamServerInterceptor(nil)
}
