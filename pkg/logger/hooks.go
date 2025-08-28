package logger

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// ContextHook 上下文钩子，添加额外的上下文信息
type ContextHook struct {
	AppName    string
	Version    string
	Environment string
}

// NewContextHook 创建上下文钩子
func NewContextHook(appName, version, environment string) *ContextHook {
	return &ContextHook{
		AppName:     appName,
		Version:     version,
		Environment: environment,
	}
}

// Levels 返回钩子适用的日志级别
func (hook *ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 执行钩子
func (hook *ContextHook) Fire(entry *logrus.Entry) error {
	entry.Data["app"] = hook.AppName
	entry.Data["version"] = hook.Version
	entry.Data["environment"] = hook.Environment
	return nil
}

// CallerHook 调用者信息钩子
type CallerHook struct {
	Field  string
	Skip   int
	levels []logrus.Level
}

// NewCallerHook 创建调用者钩子
func NewCallerHook(levels ...logrus.Level) *CallerHook {
	hook := CallerHook{
		Field: "caller",
		Skip:  5,
	}
	if len(levels) == 0 {
		hook.levels = logrus.AllLevels
	} else {
		hook.levels = levels
	}
	return &hook
}

// Levels 返回钩子适用的日志级别
func (hook *CallerHook) Levels() []logrus.Level {
	return hook.levels
}

// Fire 执行钩子
func (hook *CallerHook) Fire(entry *logrus.Entry) error {
	if pc, file, line, ok := runtime.Caller(hook.Skip); ok {
		funcName := runtime.FuncForPC(pc).Name()
		
		// 简化文件路径
		if idx := strings.LastIndex(file, "/"); idx != -1 {
			file = file[idx+1:]
		}
		
		// 简化函数名
		if idx := strings.LastIndex(funcName, "."); idx != -1 {
			funcName = funcName[idx+1:]
		}
		
		entry.Data[hook.Field] = fmt.Sprintf("%s:%d %s()", file, line, funcName)
	}
	return nil
}

// ErrorStackHook 错误堆栈钩子
type ErrorStackHook struct {
	levels []logrus.Level
}

// NewErrorStackHook 创建错误堆栈钩子
func NewErrorStackHook() *ErrorStackHook {
	return &ErrorStackHook{
		levels: []logrus.Level{
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		},
	}
}

// Levels 返回钩子适用的日志级别
func (hook *ErrorStackHook) Levels() []logrus.Level {
	return hook.levels
}

// Fire 执行钩子
func (hook *ErrorStackHook) Fire(entry *logrus.Entry) error {
	if err, ok := entry.Data[logrus.ErrorKey]; ok {
		if e, ok := err.(error); ok {
			// 添加错误堆栈信息
			entry.Data["error_stack"] = fmt.Sprintf("%+v", e)
		}
	}
	return nil
}

// RequestIDHook 请求ID钩子
type RequestIDHook struct{}

// NewRequestIDHook 创建请求ID钩子
func NewRequestIDHook() *RequestIDHook {
	return &RequestIDHook{}
}

// Levels 返回钩子适用的日志级别
func (hook *RequestIDHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 执行钩子
func (hook *RequestIDHook) Fire(entry *logrus.Entry) error {
	// 如果没有request_id，尝试从上下文中获取
	if _, exists := entry.Data["request_id"]; !exists {
		// 这里可以从context中获取request_id
		// entry.Data["request_id"] = getRequestIDFromContext(entry.Context)
	}
	return nil
}

// AddHooks 添加默认钩子
func AddHooks(logger *logrus.Logger, appName, version, environment string) {
	// 添加上下文钩子
	logger.AddHook(NewContextHook(appName, version, environment))
	
	// 添加错误堆栈钩子
	logger.AddHook(NewErrorStackHook())
	
	// 添加请求ID钩子
	logger.AddHook(NewRequestIDHook())
}
