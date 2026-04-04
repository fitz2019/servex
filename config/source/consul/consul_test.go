//go:build integration

package consul

import (
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/suite"
)

// ConsulSourceTestSuite Consul 配置源集成测试套件.
// 需要本地 Consul 实例运行.
type ConsulSourceTestSuite struct {
	suite.Suite
	client *api.Client
	key    string
}

func TestConsulSourceSuite(t *testing.T) {
	suite.Run(t, new(ConsulSourceTestSuite))
}

func (s *ConsulSourceTestSuite) SetupSuite() {
	client, err := api.NewClient(api.DefaultConfig())
	s.Require().NoError(err)
	s.client = client
	s.key = "servex/test/config"

	// 写入测试数据
	_, err = s.client.KV().Put(&api.KVPair{
		Key:   s.key,
		Value: []byte(`{"name":"consul-test","port":8080}`),
	}, nil)
	s.Require().NoError(err)
}

func (s *ConsulSourceTestSuite) TearDownSuite() {
	s.client.KV().Delete(s.key, nil)
}

func (s *ConsulSourceTestSuite) TestLoad_Success() {
	src := New(s.client, s.key)

	kvs, err := src.Load()
	s.NoError(err)
	s.Len(kvs, 1)
	s.Equal(s.key, kvs[0].Key)
	s.Equal("json", kvs[0].Format)
	s.Contains(string(kvs[0].Value), "consul-test")
}

func (s *ConsulSourceTestSuite) TestLoad_KeyNotFound() {
	src := New(s.client, "nonexistent/key")

	_, err := src.Load()
	s.Error(err)
}

func (s *ConsulSourceTestSuite) TestLoad_WithFormat() {
	src := New(s.client, s.key, WithFormat("yaml"))

	kvs, err := src.Load()
	s.NoError(err)
	s.Equal("yaml", kvs[0].Format)
}

func (s *ConsulSourceTestSuite) TestWatch_Stop() {
	src := New(s.client, s.key)

	watcher, err := src.Watch()
	s.Require().NoError(err)

	err = watcher.Stop()
	s.NoError(err)
}
