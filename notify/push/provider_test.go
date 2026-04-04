// notification/push/provider_test.go
package push

import (
	"testing"
)

func TestFCMProvider_ImplementsInterface(t *testing.T)  { var _ Provider = (*FCMProvider)(nil) }
func TestAPNsProvider_ImplementsInterface(t *testing.T) { var _ Provider = (*APNsProvider)(nil) }

func TestFCMProvider_Name(t *testing.T) {
	p := NewFCMProvider(FCMConfig{ProjectID: "proj"})
	if p.Name() != "fcm" {
		t.Errorf("name = %q", p.Name())
	}
}

func TestFCMProvider_Send_Stub(t *testing.T) {
	p := NewFCMProvider(FCMConfig{ProjectID: "proj"})
	id, err := p.Send(t.Context(), "token", &Payload{Title: "T", Body: "B"})
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Error("expected non-empty stub ID")
	}
}

func TestAPNsProvider_Name(t *testing.T) {
	p := NewAPNsProvider(APNsConfig{BundleID: "com.example.app"})
	if p.Name() != "apns" {
		t.Errorf("name = %q", p.Name())
	}
}

func TestAPNsProvider_Send_Stub(t *testing.T) {
	p := NewAPNsProvider(APNsConfig{BundleID: "com.example.app"})
	id, err := p.Send(t.Context(), "token", &Payload{Title: "T", Body: "B", Badge: 1, Sound: "default"})
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Error("expected non-empty stub ID")
	}
}
