// notification/sms/aliyun.go
package sms

import (
	"context"

	"github.com/google/uuid"
)

type AliyunConfig struct {
	AccessKeyID     string `json:"access_key_id"     yaml:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret" yaml:"access_key_secret"`
	SignName        string `json:"sign_name"         yaml:"sign_name"`
	Endpoint        string `json:"endpoint"          yaml:"endpoint"`
}

type AliyunProvider struct{ config AliyunConfig }

func NewAliyunProvider(cfg AliyunConfig) *AliyunProvider { return &AliyunProvider{config: cfg} }
func (p *AliyunProvider) Name() string                   { return "aliyun" }

// Send 桩实现。TODO: 接入阿里云 SMS SDK。
func (p *AliyunProvider) Send(_ context.Context, _ *SendRequest) (string, error) {
	return "aliyun-stub-" + uuid.New().String(), nil
}
