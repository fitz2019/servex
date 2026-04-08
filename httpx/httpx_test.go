package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromContextEmpty(t *testing.T) {
	info := FromContext(t.Context())
	if info == nil {
		t.Fatal("FromContext returned nil")
	}
	// All fields should be nil when context has no values.
	if info.IP != nil {
		t.Error("expected nil IP")
	}
	if info.UserAgent != nil {
		t.Error("expected nil UserAgent")
	}
	if info.Locale != nil {
		t.Error("expected nil Locale")
	}
}

func TestHTTPMiddlewareDefault(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := HTTPMiddleware()
	wrapped := mw(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHTTPMiddlewareWithAll(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := HTTPMiddleware(WithAll())
	wrapped := mw(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.RemoteAddr = "10.0.0.1:5000"
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestDisableOptions(t *testing.T) {
	opts := defaultOptions()

	DisableClientIP()(opts)
	if opts.enableClientIP {
		t.Error("expected ClientIP disabled")
	}

	DisableUserAgent()(opts)
	if opts.enableUserAgent {
		t.Error("expected UserAgent disabled")
	}

	DisableLocale()(opts)
	if opts.enableLocale {
		t.Error("expected Locale disabled")
	}

	DisableReferer()(opts)
	if opts.enableReferer {
		t.Error("expected Referer disabled")
	}
}

func TestWithOptions(t *testing.T) {
	opts := &options{}

	WithBot()(opts)
	if !opts.enableBot {
		t.Error("expected Bot enabled")
	}

	WithDevice()(opts)
	if !opts.enableDevice {
		t.Error("expected Device enabled")
	}

	WithLocale()(opts)
	if !opts.enableLocale {
		t.Error("expected Locale enabled")
	}

	WithReferer()(opts)
	if !opts.enableReferer {
		t.Error("expected Referer enabled")
	}

	WithUserAgent()(opts)
	if !opts.enableUserAgent {
		t.Error("expected UserAgent enabled")
	}

	WithClientIP()(opts)
	if !opts.enableClientIP {
		t.Error("expected ClientIP enabled")
	}
}
