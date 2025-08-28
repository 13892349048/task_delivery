package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"taskmanage/internal/config"
)

// Logger 全局日志实例
var Logger *logrus.Logger

// Init 初始化日志系统
func Init(cfg *config.Config) error {
	Logger = logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		return fmt.Errorf("无效的日志级别 %s: %w", cfg.Log.Level, err)
	}
	Logger.SetLevel(level)

	// 设置日志格式
	switch strings.ToLower(cfg.Log.Format) {
	case "json":
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	case "text":
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
	default:
		return fmt.Errorf("不支持的日志格式: %s", cfg.Log.Format)
	}

	// 设置输出目标
	output, err := getLogOutput(cfg)
	if err != nil {
		return fmt.Errorf("设置日志输出失败: %w", err)
	}
	Logger.SetOutput(output)

	// 设置调用者信息
	Logger.SetReportCaller(true)

	return nil
}

// getLogOutput 获取日志输出目标
func getLogOutput(cfg *config.Config) (io.Writer, error) {
	switch strings.ToLower(cfg.Log.Output) {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	case "file":
		if cfg.Log.Filename == "" {
			return nil, fmt.Errorf("文件输出模式下必须指定文件名")
		}
		
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.Log.Filename)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %w", err)
		}

		// 使用lumberjack进行日志轮转
		return &lumberjack.Logger{
			Filename:   cfg.Log.Filename,
			MaxSize:    cfg.Log.MaxSize,    // MB
			MaxBackups: cfg.Log.MaxBackups,
			MaxAge:     cfg.Log.MaxAge,     // days
			Compress:   cfg.Log.Compress,
		}, nil
	default:
		return nil, fmt.Errorf("不支持的输出类型: %s", cfg.Log.Output)
	}
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	if Logger == nil {
		// 如果未初始化，返回默认logger
		return logrus.StandardLogger()
	}
	return Logger
}

// WithFields 创建带字段的日志条目
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// WithField 创建带单个字段的日志条目
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

// WithError 创建带错误的日志条目
func WithError(err error) *logrus.Entry {
	return GetLogger().WithError(err)
}

// Debug 调试日志
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf 格式化调试日志
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info 信息日志
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof 格式化信息日志
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn 警告日志
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf 格式化警告日志
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error 错误日志
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf 格式化错误日志
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal 致命错误日志
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf 格式化致命错误日志
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// Panic panic日志
func Panic(args ...interface{}) {
	GetLogger().Panic(args...)
}

// Panicf 格式化panic日志
func Panicf(format string, args ...interface{}) {
	GetLogger().Panicf(format, args...)
}

// LogLevel 日志级别类型
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)

// SetLevel 动态设置日志级别
func SetLevel(level LogLevel) error {
	logrusLevel, err := logrus.ParseLevel(string(level))
	if err != nil {
		return err
	}
	GetLogger().SetLevel(logrusLevel)
	return nil
}

// GetLevel 获取当前日志级别
func GetLevel() LogLevel {
	return LogLevel(GetLogger().GetLevel().String())
}

// IsDebugEnabled 检查是否启用调试日志
func IsDebugEnabled() bool {
	return GetLogger().IsLevelEnabled(logrus.DebugLevel)
}

// IsInfoEnabled 检查是否启用信息日志
func IsInfoEnabled() bool {
	return GetLogger().IsLevelEnabled(logrus.InfoLevel)
}

// HTTPRequestLogger 请求日志中间件使用的结构
type HTTPRequestLogger struct {
	*logrus.Entry
}

// NewHTTPRequestLogger 创建请求日志记录器
func NewHTTPRequestLogger(requestID string) *HTTPRequestLogger {
	return &HTTPRequestLogger{
		Entry: GetLogger().WithField("request_id", requestID),
	}
}

// BusinessLogger 业务日志记录器
type BusinessLogger struct {
	*logrus.Entry
}

// NewBusinessLogger 创建业务日志记录器
func NewBusinessLogger(module string) *BusinessLogger {
	return &BusinessLogger{
		Entry: GetLogger().WithField("module", module),
	}
}

// SystemLogger 系统日志记录器
type SystemLogger struct {
	*logrus.Entry
}

// NewSystemLogger 创建系统日志记录器
func NewSystemLogger(component string) *SystemLogger {
	return &SystemLogger{
		Entry: GetLogger().WithField("component", component),
	}
}
