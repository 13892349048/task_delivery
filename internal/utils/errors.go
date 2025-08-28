package utils

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"taskmanage/pkg/logger"
	"taskmanage/pkg/response"
)

// ErrorHandler 错误处理器
type ErrorHandler struct {
	logger *logger.SystemLogger
}

// NewErrorHandler 创建错误处理器
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		logger: logger.NewSystemLogger("error_handler"),
	}
}

// Handle 处理错误
func (h *ErrorHandler) Handle(err error) *response.AppError {
	if err == nil {
		return nil
	}

	// 如果已经是应用程序错误，直接返回
	if appErr := response.GetAppError(err); appErr != nil {
		h.logError(appErr)
		return appErr
	}

	// 包装为内部错误
	appErr := response.WrapError(err, response.ErrCodeInternalError, "系统内部错误")
	h.logError(appErr)
	return appErr
}

// HandleWithCode 使用指定错误码处理错误
func (h *ErrorHandler) HandleWithCode(err error, code response.ErrorCode, message string) *response.AppError {
	if err == nil {
		return nil
	}

	appErr := response.WrapError(err, code, message)
	h.logError(appErr)
	return appErr
}

// logError 记录错误日志
func (h *ErrorHandler) logError(appErr *response.AppError) {
	fields := map[string]interface{}{
		"error_code": appErr.Code,
		"message":    appErr.Message,
	}

	if appErr.Details != nil {
		fields["details"] = appErr.Details
	}

	if appErr.Cause != nil {
		fields["cause"] = appErr.Cause.Error()
		fields["stack_trace"] = getStackTrace()
	}

	// 根据错误类型选择日志级别
	switch appErr.HTTPStatus {
	case 500, 502, 503, 504:
		h.logger.WithFields(fields).Error("Application Error")
	case 400, 401, 403, 404, 409:
		h.logger.WithFields(fields).Warn("Client Error")
	default:
		h.logger.WithFields(fields).Info("Application Error")
	}
}

// getStackTrace 获取堆栈跟踪
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var builder strings.Builder
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&builder, "%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	return builder.String()
}

// Recovery 恢复函数，用于捕获panic
func Recovery() func() {
	return func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic recovered: %v", r)
			handler := NewErrorHandler()
			appErr := handler.HandleWithCode(err, response.ErrCodeInternalError, "系统发生严重错误")
			
			// 记录panic堆栈
			handler.logger.WithFields(map[string]interface{}{
				"panic_value": r,
				"stack_trace": getStackTrace(),
			}).Error("Panic Recovered")
			
			// 这里可以添加额外的恢复逻辑，比如发送告警
			_ = appErr
		}
	}
}

// Must 必须成功，否则panic
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustNot 必须不能有错误，否则panic
func MustNot(condition bool, message string) {
	if condition {
		panic(errors.New(message))
	}
}

// Chain 错误链处理
type Chain struct {
	errors []error
}

// NewChain 创建错误链
func NewChain() *Chain {
	return &Chain{
		errors: make([]error, 0),
	}
}

// Add 添加错误到链中
func (c *Chain) Add(err error) *Chain {
	if err != nil {
		c.errors = append(c.errors, err)
	}
	return c
}

// HasError 检查是否有错误
func (c *Chain) HasError() bool {
	return len(c.errors) > 0
}

// Error 获取第一个错误
func (c *Chain) Error() error {
	if len(c.errors) == 0 {
		return nil
	}
	return c.errors[0]
}

// Errors 获取所有错误
func (c *Chain) Errors() []error {
	return c.errors
}

// Join 合并所有错误为一个错误
func (c *Chain) Join() error {
	if len(c.errors) == 0 {
		return nil
	}
	
	if len(c.errors) == 1 {
		return c.errors[0]
	}
	
	var messages []string
	for _, err := range c.errors {
		messages = append(messages, err.Error())
	}
	
	return errors.New(strings.Join(messages, "; "))
}

// Retry 重试机制
type RetryConfig struct {
	MaxAttempts int
	OnRetry     func(attempt int, err error)
}

// Retry 执行重试
func Retry(fn func() error, config RetryConfig) error {
	var lastErr error
	
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		if config.OnRetry != nil {
			config.OnRetry(attempt, err)
		}
		
		// 最后一次尝试失败后不再重试
		if attempt == config.MaxAttempts {
			break
		}
	}
	
	return lastErr
}

// SafeExecute 安全执行函数，捕获panic
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in safe execute: %v", r)
			
			// 记录panic
			logger := logger.NewSystemLogger("safe_execute")
			logger.WithFields(map[string]interface{}{
				"panic_value": r,
				"stack_trace": getStackTrace(),
			}).Error("Panic in SafeExecute")
		}
	}()
	
	return fn()
}

// ValidateRequired 验证必需字段
func ValidateRequired(value interface{}, fieldName string) error {
	if value == nil {
		return response.NewValidationError(fieldName, "字段不能为空")
	}
	
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return response.NewValidationError(fieldName, "字段不能为空")
		}
	case []interface{}:
		if len(v) == 0 {
			return response.NewValidationError(fieldName, "数组不能为空")
		}
	}
	
	return nil
}

// ValidateRange 验证数值范围
func ValidateRange(value, min, max int, fieldName string) error {
	if value < min || value > max {
		return response.NewValidationError(fieldName, fmt.Sprintf("值必须在 %d 到 %d 之间", min, max))
	}
	return nil
}

// ValidateLength 验证字符串长度
func ValidateLength(value string, min, max int, fieldName string) error {
	length := len(strings.TrimSpace(value))
	if length < min || length > max {
		return response.NewValidationError(fieldName, fmt.Sprintf("长度必须在 %d 到 %d 之间", min, max))
	}
	return nil
}
