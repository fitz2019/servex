// webhook/signer_test.go
package webhook

import "testing"

func TestHMACSigner_Sign(t *testing.T) {
	s := NewHMACSigner()
	payload := []byte(`{"type":"order.created"}`)
	secret := "test-secret"

	sig := s.Sign(payload, secret)
	if sig == "" {
		t.Fatal("signature should not be empty")
	}

	// 相同输入产生相同签名
	sig2 := s.Sign(payload, secret)
	if sig != sig2 {
		t.Error("same input should produce same signature")
	}

	// 不同 secret 产生不同签名
	sig3 := s.Sign(payload, "other-secret")
	if sig == sig3 {
		t.Error("different secret should produce different signature")
	}
}

func TestHMACSigner_Verify(t *testing.T) {
	s := NewHMACSigner()
	payload := []byte(`{"type":"order.created"}`)
	secret := "test-secret"

	sig := s.Sign(payload, secret)

	if !s.Verify(payload, secret, sig) {
		t.Error("valid signature should verify")
	}

	if s.Verify(payload, secret, "invalid-sig") {
		t.Error("invalid signature should not verify")
	}

	if s.Verify(payload, "wrong-secret", sig) {
		t.Error("wrong secret should not verify")
	}

	if s.Verify([]byte("tampered"), secret, sig) {
		t.Error("tampered payload should not verify")
	}
}
