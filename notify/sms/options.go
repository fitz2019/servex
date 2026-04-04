// notification/sms/options.go
package sms

import "github.com/Tsukikage7/servex/observability/logger"

type senderOptions struct {
	signName string
	logger   logger.Logger
}

type Option func(*senderOptions)

func WithSignName(name string) Option {
	return func(o *senderOptions) { o.signName = name }
}

func WithLogger(log logger.Logger) Option {
	return func(o *senderOptions) { o.logger = log }
}
