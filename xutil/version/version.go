// Package version 提供编译时版本信息注入与查询.
package version

import (
	"fmt"
	"runtime"
)

// 编译时通过 -ldflags 注入的变量.
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

// Info 版本信息.
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildTime string `json:"buildTime"`
	GoVersion string `json:"goVersion"`
}

// Get 获取版本信息.
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
	}
}

// String 返回格式化的版本字符串.
func (i Info) String() string {
	return fmt.Sprintf("version=%s commit=%s built=%s go=%s",
		i.Version, i.GitCommit, i.BuildTime, i.GoVersion)
}
