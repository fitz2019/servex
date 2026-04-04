// oauth2/google/provider_test.go
package google

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Tsukikage7/servex/oauth2"
)

func TestProvider_AuthURL(t *testing.T) {
	p := NewProvider(WithClientID("goog-id"), WithRedirectURL("http://localhost/cb"))
	url := p.AuthURL("s1")
	if !strings.Contains(url, "client_id=goog-id") {
		t.Error("missing client_id")
	}
	if !strings.Contains(url, "response_type=code") {
		t.Error("missing response_type")
	}
	if !strings.Contains(url, "access_type=offline") {
		t.Error("missing access_type")
	}
}

func TestProvider_Exchange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "ya29.test",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "1//test",
		})
	}))
	defer server.Close()

	p := NewProvider(WithClientID("id"), WithClientSecret("secret"))
	p.tokenURL = server.URL

	token, err := p.Exchange(t.Context(), "code")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "ya29.test" {
		t.Errorf("access_token = %s", token.AccessToken)
	}
	if token.RefreshToken != "1//test" {
		t.Errorf("refresh_token = %s", token.RefreshToken)
	}
}

func TestProvider_UserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"id": "123", "name": "Test User", "email": "test@gmail.com", "picture": "https://pic.test",
		})
	}))
	defer server.Close()

	p := NewProvider()
	p.userInfoURL = server.URL

	user, err := p.UserInfo(t.Context(), &oauth2.Token{AccessToken: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if user.Provider != "google" {
		t.Errorf("provider = %s", user.Provider)
	}
	if user.Email != "test@gmail.com" {
		t.Errorf("email = %s", user.Email)
	}
}

func TestProvider_ImplementsInterface(t *testing.T) {
	var _ oauth2.Provider = (*Provider)(nil)
}
