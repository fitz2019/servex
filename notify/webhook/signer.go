package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type hmacSigner struct{}

// NewHMACSigner 返回默认的 HMAC-SHA256 签名器.
func NewHMACSigner() Signer {
	return &hmacSigner{}
}

// Sign 使用 HMAC-SHA256 对载荷签名.
func (s *hmacSigner) Sign(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify 验证载荷签名是否匹配.
func (s *hmacSigner) Verify(payload []byte, secret string, signature string) bool {
	expected := s.Sign(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}
