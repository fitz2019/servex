// notification/sms/tencent.go
package sms

import (
	"context"

	"github.com/google/uuid"
)

type TencentConfig struct {
	SecretID  string `json:"secret_id"  yaml:"secret_id"`
	SecretKey string `json:"secret_key" yaml:"secret_key"`
	AppID     string `json:"app_id"     yaml:"app_id"`
	SignName  string `json:"sign_name"  yaml:"sign_name"`
	Endpoint  string `json:"endpoint"   yaml:"endpoint"`
}

type TencentProvider struct{ config TencentConfig }

func NewTencentProvider(cfg TencentConfig) *TencentProvider { return &TencentProvider{config: cfg} }
func (p *TencentProvider) Name() string                     { return "tencent" }

// Send 桩实现。TODO: 接入腾讯云 SMS SDK。
func (p *TencentProvider) Send(_ context.Context, _ *SendRequest) (string, error) {
	return "tencent-stub-" + uuid.New().String(), nil
}
