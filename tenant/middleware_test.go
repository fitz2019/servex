package tenant

import (
	"context"
	"errors"
	"testing"
)

// mockResolver 模拟解析器.
type mockResolver struct {
	tenant Tenant
	err    error
}

func (m *mockResolver) Resolve(_ context.Context, _ string) (Tenant, error) {
	return m.tenant, m.err
}

func TestMiddleware_Success(t *testing.T) {
	resolver := &mockResolver{
		tenant: &testTenant{id: "t1", enabled: true},
	}

	mw := Middleware(resolver, WithTokenExtractor(func(_ context.Context, _ any) (string, error) {
		return "token", nil
	}))

	var gotID string
	ep := mw(func(ctx context.Context, _ any) (any, error) {
		gotID = ID(ctx)
		return "ok", nil
	})

	resp, err := ep(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("resp = %v, want %q", resp, "ok")
	}
	if gotID != "t1" {
		t.Fatalf("gotID = %q, want %q", gotID, "t1")
	}
}

func TestMiddleware_Skipper(t *testing.T) {
	resolver := &mockResolver{
		err: errors.New("should not be called"),
	}

	mw := Middleware(resolver,
		WithSkipper(func(_ context.Context, _ any) bool { return true }),
	)

	ep := mw(func(_ context.Context, _ any) (any, error) {
		return "skipped", nil
	})

	resp, err := ep(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "skipped" {
		t.Fatalf("resp = %v, want %q", resp, "skipped")
	}
}

func TestMiddleware_MissingToken(t *testing.T) {
	resolver := &mockResolver{}

	mw := Middleware(resolver) // 无 token extractor

	ep := mw(func(_ context.Context, _ any) (any, error) {
		return nil, nil
	})

	_, err := ep(t.Context(), nil)
	if !errors.Is(err, ErrMissingToken) {
		t.Fatalf("err = %v, want ErrMissingToken", err)
	}
}

func TestMiddleware_ResolveError(t *testing.T) {
	resolver := &mockResolver{err: ErrTenantNotFound}

	mw := Middleware(resolver, WithTokenExtractor(func(_ context.Context, _ any) (string, error) {
		return "token", nil
	}))

	ep := mw(func(_ context.Context, _ any) (any, error) {
		return nil, nil
	})

	_, err := ep(t.Context(), nil)
	if !errors.Is(err, ErrTenantNotFound) {
		t.Fatalf("err = %v, want ErrTenantNotFound", err)
	}
}

func TestMiddleware_Disabled(t *testing.T) {
	resolver := &mockResolver{
		tenant: &testTenant{id: "t1", enabled: false},
	}

	mw := Middleware(resolver, WithTokenExtractor(func(_ context.Context, _ any) (string, error) {
		return "token", nil
	}))

	ep := mw(func(_ context.Context, _ any) (any, error) {
		return nil, nil
	})

	_, err := ep(t.Context(), nil)
	if !errors.Is(err, ErrTenantDisabled) {
		t.Fatalf("err = %v, want ErrTenantDisabled", err)
	}
}

func TestMiddleware_ErrorHandler(t *testing.T) {
	resolver := &mockResolver{err: ErrTenantNotFound}
	customErr := errors.New("custom error")

	mw := Middleware(resolver,
		WithTokenExtractor(func(_ context.Context, _ any) (string, error) {
			return "token", nil
		}),
		WithErrorHandler(func(_ context.Context, _ error) error {
			return customErr
		}),
	)

	ep := mw(func(_ context.Context, _ any) (any, error) {
		return nil, nil
	})

	_, err := ep(t.Context(), nil)
	if !errors.Is(err, customErr) {
		t.Fatalf("err = %v, want custom error", err)
	}
}

func TestMiddleware_PanicOnNilResolver(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("应在 nil resolver 时 panic")
		}
	}()
	Middleware(nil)
}
