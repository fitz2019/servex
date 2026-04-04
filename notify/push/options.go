// notification/push/options.go
package push

import "github.com/Tsukikage7/servex/observability/logger"

type senderOptions struct{ logger logger.Logger }
type Option func(*senderOptions)

func WithLogger(log logger.Logger) Option { return func(o *senderOptions) { o.logger = log } }
