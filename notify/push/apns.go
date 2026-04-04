// notification/push/apns.go
package push

import (
	"context"

	"github.com/google/uuid"
)

type APNsConfig struct {
	BundleID   string `json:"bundle_id"  yaml:"bundle_id"`
	TeamID     string `json:"team_id"    yaml:"team_id"`
	KeyID      string `json:"key_id"     yaml:"key_id"`
	KeyFile    string `json:"key_file"   yaml:"key_file"`
	Production bool   `json:"production" yaml:"production"`
}

type APNsProvider struct{ config APNsConfig }

func NewAPNsProvider(cfg APNsConfig) *APNsProvider { return &APNsProvider{config: cfg} }
func (p *APNsProvider) Name() string               { return "apns" }

// Send 桩实现。TODO: 接入 Apple APNs HTTP/2 API。
func (p *APNsProvider) Send(_ context.Context, _ string, _ *Payload) (string, error) {
	return "apns-stub-" + uuid.New().String(), nil
}
