package optionx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OptionTestSuite struct {
	suite.Suite
}

func TestOptionSuite(t *testing.T) {
	suite.Run(t, new(OptionTestSuite))
}

type serverConfig struct {
	Host    string
	Port    int
	Debug   bool
	Timeout int
}

func (s *OptionTestSuite) TestApply() {
	cfg := &serverConfig{Host: "localhost", Port: 8080}

	Apply(cfg,
		func(c *serverConfig) { c.Host = "0.0.0.0" },
		func(c *serverConfig) { c.Port = 9090 },
		func(c *serverConfig) { c.Debug = true },
	)

	s.Equal("0.0.0.0", cfg.Host)
	s.Equal(9090, cfg.Port)
	s.True(cfg.Debug)
}

func (s *OptionTestSuite) TestApply_NoOptions() {
	cfg := &serverConfig{Host: "localhost", Port: 8080}
	Apply(cfg)
	s.Equal("localhost", cfg.Host)
	s.Equal(8080, cfg.Port)
}

func (s *OptionTestSuite) TestApplyErr_Success() {
	cfg := &serverConfig{}

	err := ApplyErr(cfg,
		func(c *serverConfig) error {
			c.Host = "127.0.0.1"
			return nil
		},
		func(c *serverConfig) error {
			c.Port = 3000
			return nil
		},
	)

	s.NoError(err)
	s.Equal("127.0.0.1", cfg.Host)
	s.Equal(3000, cfg.Port)
}

func (s *OptionTestSuite) TestApplyErr_FailFast() {
	errInvalid := errors.New("invalid port")
	cfg := &serverConfig{}

	err := ApplyErr(cfg,
		func(c *serverConfig) error {
			c.Host = "127.0.0.1"
			return nil
		},
		func(c *serverConfig) error {
			return errInvalid
		},
		func(c *serverConfig) error {
			c.Debug = true
			return nil
		},
	)

	s.ErrorIs(err, errInvalid)
	s.Equal("127.0.0.1", cfg.Host)
	s.False(cfg.Debug)
}

func (s *OptionTestSuite) TestApplyErr_NoOptions() {
	cfg := &serverConfig{Host: "localhost"}
	err := ApplyErr[serverConfig](cfg)
	s.NoError(err)
	s.Equal("localhost", cfg.Host)
}

func (s *OptionTestSuite) TestApply_WithDifferentTypes() {
	type dbConfig struct {
		DSN         string
		MaxConn     int
		EnableTrace bool
	}

	cfg := &dbConfig{}
	Apply(cfg,
		func(c *dbConfig) { c.DSN = "postgres://localhost/test" },
		func(c *dbConfig) { c.MaxConn = 10 },
		func(c *dbConfig) { c.EnableTrace = true },
	)

	s.Equal("postgres://localhost/test", cfg.DSN)
	s.Equal(10, cfg.MaxConn)
	s.True(cfg.EnableTrace)
}
