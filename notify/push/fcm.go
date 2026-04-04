// notification/push/fcm.go
package push

import (
	"context"

	"github.com/google/uuid"
)

type FCMConfig struct {
	ProjectID       string `json:"project_id"       yaml:"project_id"`
	CredentialsJSON []byte `json:"credentials_json" yaml:"credentials_json"`
}

type FCMProvider struct{ config FCMConfig }

func NewFCMProvider(cfg FCMConfig) *FCMProvider { return &FCMProvider{config: cfg} }
func (p *FCMProvider) Name() string             { return "fcm" }

// Send 桩实现。TODO: 接入 Firebase Admin SDK。
func (p *FCMProvider) Send(_ context.Context, _ string, _ *Payload) (string, error) {
	return "fcm-stub-" + uuid.New().String(), nil
}
