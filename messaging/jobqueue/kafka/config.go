// jobqueue/kafka/config.go
package kafka

import (
	"github.com/IBM/sarama"
)

// NewStoreFromConfig 根据 broker 列表创建 Kafka Store。
func NewStoreFromConfig(brokers []string, prefix string) (*Store, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	client, err := sarama.NewClient(brokers, cfg)
	if err != nil {
		return nil, err
	}
	var opts []Option
	if prefix != "" {
		opts = append(opts, WithPrefix(prefix))
	}
	return NewStore(client, opts...)
}
