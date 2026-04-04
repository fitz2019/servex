package version

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type VersionTestSuite struct {
	suite.Suite
}

func TestVersionSuite(t *testing.T) {
	suite.Run(t, new(VersionTestSuite))
}

func (s *VersionTestSuite) TestGet_DefaultValues() {
	info := Get()

	s.Equal("dev", info.Version)
	s.Equal("unknown", info.GitCommit)
	s.Equal("unknown", info.BuildTime)
	s.Equal(runtime.Version(), info.GoVersion)
}

func (s *VersionTestSuite) TestGet_CustomValues() {
	// 模拟 ldflags 注入
	origVersion := Version
	origCommit := GitCommit
	origBuildTime := BuildTime
	defer func() {
		Version = origVersion
		GitCommit = origCommit
		BuildTime = origBuildTime
	}()

	Version = "v1.2.3"
	GitCommit = "abc1234"
	BuildTime = "2025-01-01T00:00:00Z"

	info := Get()
	s.Equal("v1.2.3", info.Version)
	s.Equal("abc1234", info.GitCommit)
	s.Equal("2025-01-01T00:00:00Z", info.BuildTime)
}

func (s *VersionTestSuite) TestString_Format() {
	info := Info{
		Version:   "v1.0.0",
		GitCommit: "abc1234",
		BuildTime: "2025-01-01",
		GoVersion: "go1.25.1",
	}

	str := info.String()
	s.True(strings.Contains(str, "version=v1.0.0"))
	s.True(strings.Contains(str, "commit=abc1234"))
	s.True(strings.Contains(str, "built=2025-01-01"))
	s.True(strings.Contains(str, "go=go1.25.1"))
}
