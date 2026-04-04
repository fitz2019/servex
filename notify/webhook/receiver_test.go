// webhook/receiver_test.go
package webhook

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"
)

func TestReceiver_Handle(t *testing.T) {
	secret := "test-secret"
	signer := NewHMACSigner()
	payload := []byte(`{"id":1}`)
	sig := signer.Sign(payload, secret)

	r := NewReceiver(WithSecret(secret))

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-Webhook-Signature", sig)
	req.Header.Set("X-Webhook-Event", "order.created")
	req.Header.Set("X-Webhook-ID", "evt-1")

	event, err := r.Handle(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != "order.created" {
		t.Errorf("type = %s", event.Type)
	}
	if event.ID != "evt-1" {
		t.Errorf("id = %s", event.ID)
	}
	if string(event.Payload) != `{"id":1}` {
		t.Errorf("payload = %s", event.Payload)
	}
}

func TestReceiver_Handle_InvalidSignature(t *testing.T) {
	r := NewReceiver(WithSecret("secret"))

	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader([]byte("body")))
	req.Header.Set("X-Webhook-Signature", "invalid")

	_, err := r.Handle(t.Context(), req)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("got %v, want ErrInvalidSignature", err)
	}
}

func TestReceiver_Handle_EmptyBody(t *testing.T) {
	r := NewReceiver(WithSecret("secret"))
	req := httptest.NewRequest("POST", "/webhook", nil)

	_, err := r.Handle(t.Context(), req)
	if !errors.Is(err, ErrEmptyBody) {
		t.Errorf("got %v, want ErrEmptyBody", err)
	}
}

func TestReceiver_Handle_NoSecret(t *testing.T) {
	r := NewReceiver()
	payload := []byte(`{"id":1}`)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-Webhook-Event", "test")

	event, err := r.Handle(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != "test" {
		t.Errorf("type = %s", event.Type)
	}
}

func TestReceiver_ImplementsInterface(t *testing.T) {
	var _ Receiver = (*receiver)(nil)
}
