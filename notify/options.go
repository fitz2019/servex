package notify

import (
	"github.com/Tsukikage7/servex/messaging/jobqueue"
	"github.com/Tsukikage7/servex/observability/logger"
)

type dispatcherOptions struct {
	logger         logger.Logger
	templateEngine TemplateEngine
	jobClient      jobqueue.Client
	defaultChannel Channel
}

type Option func(*dispatcherOptions)

func WithLogger(log logger.Logger) Option {
	return func(o *dispatcherOptions) { o.logger = log }
}

func WithTemplateEngine(eng TemplateEngine) Option {
	return func(o *dispatcherOptions) { o.templateEngine = eng }
}

func WithJobQueue(client jobqueue.Client) Option {
	return func(o *dispatcherOptions) { o.jobClient = client }
}

func WithDefaultChannel(ch Channel) Option {
	return func(o *dispatcherOptions) { o.defaultChannel = ch }
}
