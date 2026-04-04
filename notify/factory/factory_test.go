// notification/factory/factory_test.go
package factory

import "testing"

func TestNewDispatcher_NilConfig(t *testing.T) {
	_, err := NewDispatcher(nil, nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestNewDispatcher_EmptyConfig(t *testing.T) {
	d, err := NewDispatcher(&Config{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	d.Close()
}

func TestNewDispatcher_WithEmail(t *testing.T) {
	d, err := NewDispatcher(&Config{
		DefaultChannel: "email",
		Email:          &EmailConfig{Host: "smtp.example.com", Port: 587, From: "noreply@example.com", Name: "Test"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	d.Close()
}

func TestNewDispatcher_WithSMS_Aliyun(t *testing.T) {
	d, err := NewDispatcher(&Config{SMS: &SMSConfig{
		Provider: "aliyun", SignName: "Test",
		Aliyun: &AliyunSMSConfig{AccessKeyID: "ak", AccessKeySecret: "sk"},
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	d.Close()
}

func TestNewDispatcher_WithSMS_Tencent(t *testing.T) {
	d, err := NewDispatcher(&Config{SMS: &SMSConfig{
		Provider: "tencent", SignName: "Test",
		Tencent: &TencentSMSConfig{SecretID: "sid", SecretKey: "skey", AppID: "app"},
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	d.Close()
}

func TestNewDispatcher_WithSMS_UnknownProvider(t *testing.T) {
	_, err := NewDispatcher(&Config{SMS: &SMSConfig{Provider: "unknown"}}, nil)
	if err == nil {
		t.Error("expected error for unknown SMS provider")
	}
}

func TestNewDispatcher_WithWebhook(t *testing.T) {
	d, err := NewDispatcher(&Config{Webhook: &WebhookConfig{Timeout: 5, Retry: 3}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	d.Close()
}

func TestNewDispatcher_WithPush_FCM(t *testing.T) {
	d, err := NewDispatcher(&Config{Push: &PushConfig{Provider: "fcm", FCM: &FCMPushConfig{ProjectID: "proj"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	d.Close()
}

func TestNewDispatcher_WithPush_APNs(t *testing.T) {
	d, err := NewDispatcher(&Config{Push: &PushConfig{Provider: "apns", APNs: &APNsPushConfig{BundleID: "com.example.app"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	d.Close()
}

func TestNewDispatcher_WithPush_UnknownProvider(t *testing.T) {
	_, err := NewDispatcher(&Config{Push: &PushConfig{Provider: "unknown"}}, nil)
	if err == nil {
		t.Error("expected error for unknown push provider")
	}
}

func TestNewDispatcher_AllChannels(t *testing.T) {
	d, err := NewDispatcher(&Config{
		DefaultChannel: "email",
		Email:          &EmailConfig{Host: "smtp.example.com", Port: 587, From: "no@example.com", Name: "T"},
		SMS:            &SMSConfig{Provider: "aliyun", SignName: "T", Aliyun: &AliyunSMSConfig{AccessKeyID: "ak", AccessKeySecret: "sk"}},
		Webhook:        &WebhookConfig{Timeout: 10, Retry: 2},
		Push:           &PushConfig{Provider: "fcm", FCM: &FCMPushConfig{ProjectID: "p"}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	d.Close()
}
