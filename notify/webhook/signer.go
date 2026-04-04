// webhook/signer.go
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// hmacSigner 使用 HMAC-SHA256 签名。
type hmacSigner struct{}

// NewHMACSigner 返回默认的 HMAC-SHA256 签名器。
func NewHMACSigner() Signer {
	return &hmacSigner{}
}

func (s *hmacSigner) Sign(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *hmacSigner) Verify(payload []byte, secret string, signature string) bool {
	expected := s.Sign(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}
