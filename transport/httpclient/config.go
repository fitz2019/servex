package httpclient

import (
	"time"

	"github.com/Tsukikage7/servex/middleware/circuitbreaker"
	"github.com/Tsukikage7/servex/middleware/retry"
)

// Config 配置驱动的客户端创建.
type Config struct {
	BaseURL        string        `json:"base_url" yaml:"base_url" mapstructure:"base_url"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	MaxRetries     int           `json:"max_retries" yaml:"max_retries" mapstructure:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay" yaml:"retry_delay" mapstructure:"retry_delay"`
	CircuitBreaker bool          `json:"circuit_breaker" yaml:"circuit_breaker" mapstructure:"circuit_breaker"`
	Tracing        bool          `json:"tracing" yaml:"tracing" mapstructure:"tracing"`
}

// NewFromConfig 从配置创建简单客户端（不使用服务发现）.
func NewFromConfig(cfg *Config, additionalOpts ...Option) *Client {
	var opts []Option

	if cfg.BaseURL != "" {
		opts = append(opts, WithBaseURL(cfg.BaseURL))
	}
	if cfg.Timeout > 0 {
		opts = append(opts, WithTimeout(cfg.Timeout))
	}
	if cfg.MaxRetries > 0 {
		opts = append(opts, WithRetry(&retry.Config{
			MaxAttempts: cfg.MaxRetries,
			Delay:       cfg.RetryDelay,
			Backoff:     retry.FixedBackoff,
			Retryable:   retry.AlwaysRetry,
		}))
	}
	if cfg.CircuitBreaker {
		opts = append(opts, WithCircuitBreaker(circuitbreaker.New()))
	}
	if cfg.Tracing {
		opts = append(opts, WithTracing("httpclient"))
	}

	opts = append(opts, additionalOpts...)
	return NewSimple(opts...)
}
