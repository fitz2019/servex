package oauth2

import (
	"errors"
	"testing"
	"time"
)

func TestTokenIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{"zero time not expired", time.Time{}, false},
		{"future not expired", time.Now().Add(time.Hour), false},
		{"past expired", time.Now().Add(-time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &Token{ExpiresAt: tt.expiresAt}
			if got := token.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenFields(t *testing.T) {
	token := &Token{
		AccessToken:  "access",
		TokenType:    "Bearer",
		RefreshToken: "refresh",
		Scopes:       []string{"read", "write"},
		Raw:          map[string]any{"extra": "value"},
	}

	if token.AccessToken != "access" {
		t.Errorf("AccessToken = %q", token.AccessToken)
	}
	if token.TokenType != "Bearer" {
		t.Errorf("TokenType = %q", token.TokenType)
	}
	if len(token.Scopes) != 2 {
		t.Errorf("Scopes = %v", token.Scopes)
	}
}

func TestUserInfoFields(t *testing.T) {
	ui := &UserInfo{
		ProviderID: "123",
		Provider:   "github",
		Name:       "test",
		Email:      "test@example.com",
		AvatarURL:  "https://example.com/avatar.png",
		Extra:      map[string]any{"role": "admin"},
	}

	if ui.ProviderID != "123" {
		t.Errorf("ProviderID = %q", ui.ProviderID)
	}
	if ui.Provider != "github" {
		t.Errorf("Provider = %q", ui.Provider)
	}
}

func TestErrors(t *testing.T) {
	errs := []error{
		ErrInvalidState,
		ErrExchangeFailed,
		ErrRefreshFailed,
		ErrUserInfoFailed,
		ErrInvalidCode,
		ErrInvalidToken,
	}

	for _, e := range errs {
		if e == nil {
			t.Error("error should not be nil")
		}
		if e.Error() == "" {
			t.Error("error message should not be empty")
		}
	}

	// Verify errors are distinct.
	if errors.Is(ErrInvalidState, ErrExchangeFailed) {
		t.Error("ErrInvalidState should not equal ErrExchangeFailed")
	}
}

// --- Token Bearer convenience ---

func TestTokenIsExpired_JustExpired(t *testing.T) {
	// Token that expired 1 millisecond ago.
	token := &Token{ExpiresAt: time.Now().Add(-time.Millisecond)}
	if !token.IsExpired() {
		t.Error("token should be expired")
	}
}

func TestTokenIsExpired_FarFuture(t *testing.T) {
	token := &Token{ExpiresAt: time.Now().Add(365 * 24 * time.Hour)}
	if token.IsExpired() {
		t.Error("token should not be expired")
	}
}

func TestTokenWithAllFields(t *testing.T) {
	now := time.Now().Add(time.Hour)
	token := &Token{
		AccessToken:  "at",
		TokenType:    "Bearer",
		RefreshToken: "rt",
		ExpiresAt:    now,
		Scopes:       []string{"openid", "profile", "email"},
		Raw:          map[string]any{"foo": "bar"},
	}

	if token.AccessToken != "at" {
		t.Errorf("AccessToken = %q", token.AccessToken)
	}
	if len(token.Scopes) != 3 {
		t.Errorf("Scopes len = %d, want 3", len(token.Scopes))
	}
	if token.Raw["foo"] != "bar" {
		t.Errorf("Raw[foo] = %v", token.Raw["foo"])
	}
	if token.IsExpired() {
		t.Error("token should not be expired yet")
	}
}

// --- AuthURLOptions ---

func TestApplyAuthURLOptions_Empty(t *testing.T) {
	opts := ApplyAuthURLOptions(nil)
	if opts.Prompt != "" {
		t.Errorf("Prompt should be empty, got %q", opts.Prompt)
	}
	if len(opts.Scopes) != 0 {
		t.Errorf("Scopes should be empty, got %v", opts.Scopes)
	}
}

func TestApplyAuthURLOptions_WithValues(t *testing.T) {
	opts := ApplyAuthURLOptions([]AuthURLOption{
		WithExtraScopes("read", "write"),
		WithPrompt("consent"),
	})

	if opts.Prompt != "consent" {
		t.Errorf("Prompt = %q, want consent", opts.Prompt)
	}
	if len(opts.Scopes) != 2 {
		t.Errorf("Scopes len = %d, want 2", len(opts.Scopes))
	}
}

func TestWithExtraScopes_Appends(t *testing.T) {
	opts := ApplyAuthURLOptions([]AuthURLOption{
		WithExtraScopes("a"),
		WithExtraScopes("b", "c"),
	})
	if len(opts.Scopes) != 3 {
		t.Errorf("Scopes len = %d, want 3", len(opts.Scopes))
	}
}
