// Package logger 提供结构化日志记录功能.
package logger

import "context"

const (
	// TypeZap 基于 zap 的日志实现.
	TypeZap = "zap"
)

const (
	// LevelDebug 调试级别.
	LevelDebug = "debug"
	// LevelInfo 信息级别.
	LevelInfo = "info"
	// LevelWarn 警告级别.
	LevelWarn = "warn"
	// LevelError 错误级别.
	LevelError = "error"
	// LevelFatal 致命级别.
	LevelFatal = "fatal"
	// LevelPanic 恐慌级别.
	LevelPanic = "panic"
)

const (
	// FormatJSON JSON 输出格式.
	FormatJSON = "json"
	// FormatConsole 控制台输出格式.
	FormatConsole = "console"
)

const (
	// OutputConsole 输出到控制台.
	OutputConsole = "console"
	// OutputFile 输出到文件.
	OutputFile = "file"
	// OutputBoth 同时输出到控制台和文件.
	OutputBoth = "both"
)

const (
	// RotationDaily 按天轮转.
	RotationDaily = "daily"
	// RotationHourly 按小时轮转.
	RotationHourly = "hourly"
)

const (
	// TimeFormatISO8601 ISO8601 时间格式.
	TimeFormatISO8601 = "iso8601"
	// TimeFormatRFC3339 RFC3339 时间格式.
	TimeFormatRFC3339 = "rfc3339"
	// TimeFormatRFC3339Nano RFC3339 纳秒精度时间格式.
	TimeFormatRFC3339Nano = "rfc3339nano"
	// TimeFormatEpoch Unix 时间戳格式.
	TimeFormatEpoch = "epoch"
	// TimeFormatEpochMillis 毫秒时间戳格式.
	TimeFormatEpochMillis = "epochmillis"
	// TimeFormatEpochNanos 纳秒时间戳格式.
	TimeFormatEpochNanos = "epochnanos"
	// TimeFormatDateTime 日期时间格式.
	TimeFormatDateTime = "datetime"
)

const (
	// EncodeLevelCapital 大写级别编码.
	EncodeLevelCapital = "capital"
	// EncodeLevelCapitalColor 大写彩色级别编码.
	EncodeLevelCapitalColor = "capitalcolor"
	// EncodeLevelLower 小写级别编码.
	EncodeLevelLower = "lower"
	// EncodeLevelLowerColor 小写彩色级别编码.
	EncodeLevelLowerColor = "lowercolor"
)

const (
	// EncodeCallerShort 短路径调用者编码.
	EncodeCallerShort = "short"
	// EncodeCallerFull 完整路径调用者编码.
	EncodeCallerFull = "full"
)

// contextKey context 键类型.
type contextKey string

const (
	// TraceIDKey 用于在 context 中存储 traceId.
	TraceIDKey contextKey = "logger:traceId"
	// SpanIDKey 用于在 context 中存储 spanId.
	SpanIDKey contextKey = "logger:spanId"
)

// Field 表示一个日志字段.
type Field struct {
	Key   string
	Value any
}

// Logger 日志记录器接口.
type Logger interface {
	// 基础日志方法
	Debug(args ...any)
	Debugf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Panic(args ...any)
	Panicf(format string, args ...any)

	// 结构化日志方法
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger

	// 生命周期管理
	Sync() error
	Close() error
}

// ContextWithTraceID 将 traceId 注入到 context.
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// ContextWithSpanID 将 spanId 注入到 context.
func ContextWithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, SpanIDKey, spanID)
}

// NewLogger 创建 logger 实例.
func NewLogger(config *Config) (Logger, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	config.ApplyDefaults()

	switch config.Type {
	case TypeZap, "":
		return newZapLogger(config)
	default:
		return nil, &ConfigError{Field: "type", Message: "unsupported logger type: " + config.Type}
	}
}

// MustNewLogger 创建 logger 实例，失败时 panic.
func MustNewLogger(config *Config) Logger {
	l, err := NewLogger(config)
	if err != nil {
		panic(err)
	}
	return l
}
