package tenant

import (
	"net/http"
	"testing"
)

func TestTenantHTTPKeyFunc(t *testing.T) {
	keyFunc := TenantHTTPKeyFunc()

	ctx := WithTenant(t.Context(), &testTenant{id: "t1", enabled: true})
	r, _ := http.NewRequestWithContext(ctx, "GET", "/", nil)

	key := keyFunc(r)
	if key != "t1" {
		t.Fatalf("key = %q, want %q", key, "t1")
	}
}

func TestTenantHTTPKeyFunc_NoTenant(t *testing.T) {
	keyFunc := TenantHTTPKeyFunc()

	r, _ := http.NewRequest("GET", "/", nil)
	key := keyFunc(r)
	if key != "" {
		t.Fatalf("key = %q, want empty", key)
	}
}

func TestTenantKeyFunc(t *testing.T) {
	keyFunc := TenantKeyFunc()

	ctx := WithTenant(t.Context(), &testTenant{id: "t1", enabled: true})
	key := keyFunc(ctx, nil)
	if key != "t1" {
		t.Fatalf("key = %q, want %q", key, "t1")
	}
}

func TestTenantKeyFunc_NoTenant(t *testing.T) {
	keyFunc := TenantKeyFunc()
	key := keyFunc(t.Context(), nil)
	if key != "" {
		t.Fatalf("key = %q, want empty", key)
	}
}
