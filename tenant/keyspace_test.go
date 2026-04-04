package tenant

import (
	"testing"
)

func TestPrefixKey(t *testing.T) {
	ctx := WithTenant(t.Context(), &testTenant{id: "t1", enabled: true})

	got := PrefixKey(ctx, "session:abc")
	want := "t1/session:abc"
	if got != want {
		t.Fatalf("PrefixKey = %q, want %q", got, want)
	}
}

func TestPrefixKey_NoTenant(t *testing.T) {
	got := PrefixKey(t.Context(), "session:abc")
	if got != "session:abc" {
		t.Fatalf("PrefixKey = %q, want %q", got, "session:abc")
	}
}

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		input   string
		wantID  string
		wantKey string
		wantOK  bool
	}{
		{"t1/session:abc", "t1", "session:abc", true},
		{"abc/def/ghi", "abc", "def/ghi", true},
		{"noprefix", "", "noprefix", false},
	}

	for _, tt := range tests {
		id, key, ok := StripPrefix(tt.input)
		if id != tt.wantID || key != tt.wantKey || ok != tt.wantOK {
			t.Errorf("StripPrefix(%q) = (%q, %q, %v), want (%q, %q, %v)",
				tt.input, id, key, ok, tt.wantID, tt.wantKey, tt.wantOK)
		}
	}
}
