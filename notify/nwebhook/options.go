// notification/webhook/options.go
package nwebhook

import (
	"net/http"
	"time"

	"github.com/Tsukikage7/servex/observability/logger"
)

type senderOptions struct {
	httpClient *http.Client
	timeout    time.Duration
	maxRetry   int
	logger     logger.Logger
}

type Option func(*senderOptions)

func WithTimeout(d time.Duration) Option   { return func(o *senderOptions) { o.timeout = d } }
func WithRetry(n int) Option               { return func(o *senderOptions) { o.maxRetry = n } }
func WithHTTPClient(c *http.Client) Option { return func(o *senderOptions) { o.httpClient = c } }
func WithLogger(log logger.Logger) Option  { return func(o *senderOptions) { o.logger = log } }
