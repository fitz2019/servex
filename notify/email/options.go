// notification/email/options.go
package email

import "github.com/Tsukikage7/servex/observability/logger"

type senderOptions struct {
	host     string
	port     int
	username string
	password string
	fromAddr string
	fromName string
	useTLS   bool
	logger   logger.Logger
}

type Option func(*senderOptions)

func WithSMTP(host string, port int) Option {
	return func(o *senderOptions) { o.host = host; o.port = port }
}

func WithAuth(username, password string) Option {
	return func(o *senderOptions) { o.username = username; o.password = password }
}

func WithFrom(addr, name string) Option {
	return func(o *senderOptions) { o.fromAddr = addr; o.fromName = name }
}

func WithTLS(enable bool) Option {
	return func(o *senderOptions) { o.useTLS = enable }
}

func WithLogger(log logger.Logger) Option {
	return func(o *senderOptions) { o.logger = log }
}
