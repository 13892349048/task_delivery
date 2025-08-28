package response

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode string

// 系统级错误码
const (
	// 通用错误
	ErrCodeSuccess         ErrorCode = "SUCCESS"
	ErrCodeInternalError   ErrorCode = "INTERNAL_ERROR"
	ErrCodeInvalidRequest  ErrorCode = "INVALID_REQUEST"
	ErrCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden       ErrorCode = "FORBIDDEN"
	ErrCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrCodeConflict        ErrorCode = "CONFLICT"
	ErrCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// 参数验证错误
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeMissingParameter ErrorCode = "MISSING_PARAMETER"
	ErrCodeInvalidParameter ErrorCode = "INVALID_PARAMETER"

	// 数据库错误
	ErrCodeDatabaseError     ErrorCode = "DATABASE_ERROR"
	ErrCodeRecordNotFound    ErrorCode = "RECORD_NOT_FOUND"
	ErrCodeDuplicateRecord   ErrorCode = "DUPLICATE_RECORD"
	ErrCodeConstraintViolation ErrorCode = "CONSTRAINT_VIOLATION"

	// 认证授权错误
	ErrCodeInvalidToken    ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired    ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodePermissionDenied   ErrorCode = "PERMISSION_DENIED"

	// 业务逻辑错误
	ErrCodeTaskNotFound        ErrorCode = "TASK_NOT_FOUND"
	ErrCodeTaskAlreadyAssigned ErrorCode = "TASK_ALREADY_ASSIGNED"
	ErrCodeTaskStatusInvalid   ErrorCode = "TASK_STATUS_INVALID"
	ErrCodeEmployeeNotFound    ErrorCode = "EMPLOYEE_NOT_FOUND"
	ErrCodeEmployeeNotAvailable ErrorCode = "EMPLOYEE_NOT_AVAILABLE"
	ErrCodeAssignmentFailed     ErrorCode = "ASSIGNMENT_FAILED"
	ErrCodeApprovalRequired     ErrorCode = "APPROVAL_REQUIRED"
	ErrCodeApprovalNotFound     ErrorCode = "APPROVAL_NOT_FOUND"
	ErrCodeApprovalAlreadyProcessed ErrorCode = "APPROVAL_ALREADY_PROCESSED"

	// 外部服务错误
	ErrCodeExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrCodeServiceUnavailable   ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout              ErrorCode = "TIMEOUT"
)

// AppError 应用程序错误结构
type AppError struct {
	Code       ErrorCode   `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	Cause      error       `json:"-"`
	HTTPStatus int         `json:"-"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 支持errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithDetails 添加错误详情
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// WithCause 添加原因错误
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// NewError 创建新的应用程序错误
func NewError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
	}
}

// NewErrorWithCause 创建带原因的错误
func NewErrorWithCause(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Cause:      cause,
		HTTPStatus: getHTTPStatus(code),
	}
}

// getHTTPStatus 根据错误码获取HTTP状态码
func getHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrCodeSuccess:
		return http.StatusOK
	case ErrCodeInvalidRequest, ErrCodeValidationFailed, ErrCodeMissingParameter, ErrCodeInvalidParameter:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeInvalidToken, ErrCodeTokenExpired, ErrCodeInvalidCredentials:
		return http.StatusUnauthorized
	case ErrCodeForbidden, ErrCodePermissionDenied:
		return http.StatusForbidden
	case ErrCodeNotFound, ErrCodeRecordNotFound, ErrCodeTaskNotFound, ErrCodeEmployeeNotFound, ErrCodeApprovalNotFound:
		return http.StatusNotFound
	case ErrCodeConflict, ErrCodeDuplicateRecord, ErrCodeTaskAlreadyAssigned, ErrCodeApprovalAlreadyProcessed:
		return http.StatusConflict
	case ErrCodeTooManyRequests:
		return http.StatusTooManyRequests
	case ErrCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrCodeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// 预定义的常用错误
var (
	ErrInternalError = NewError(ErrCodeInternalError, "内部服务器错误")
	ErrInvalidRequest = NewError(ErrCodeInvalidRequest, "无效的请求")
	ErrUnauthorized = NewError(ErrCodeUnauthorized, "未授权访问")
	ErrForbidden = NewError(ErrCodeForbidden, "禁止访问")
	ErrNotFound = NewError(ErrCodeNotFound, "资源未找到")
	ErrValidationFailed = NewError(ErrCodeValidationFailed, "参数验证失败")
	ErrDatabaseError = NewError(ErrCodeDatabaseError, "数据库操作失败")
	ErrRecordNotFound = NewError(ErrCodeRecordNotFound, "记录未找到")
	ErrDuplicateRecord = NewError(ErrCodeDuplicateRecord, "记录已存在")
	ErrInvalidToken = NewError(ErrCodeInvalidToken, "无效的令牌")
	ErrTokenExpired = NewError(ErrCodeTokenExpired, "令牌已过期")
	ErrPermissionDenied = NewError(ErrCodePermissionDenied, "权限不足")
)

// 业务错误构造函数
func NewTaskNotFoundError(taskID interface{}) *AppError {
	return NewError(ErrCodeTaskNotFound, fmt.Sprintf("任务不存在: %v", taskID))
}

func NewEmployeeNotFoundError(employeeID interface{}) *AppError {
	return NewError(ErrCodeEmployeeNotFound, fmt.Sprintf("员工不存在: %v", employeeID))
}

func NewTaskAlreadyAssignedError(taskID interface{}) *AppError {
	return NewError(ErrCodeTaskAlreadyAssigned, fmt.Sprintf("任务已被分配: %v", taskID))
}

func NewEmployeeNotAvailableError(employeeID interface{}) *AppError {
	return NewError(ErrCodeEmployeeNotAvailable, fmt.Sprintf("员工不可用: %v", employeeID))
}

func NewValidationError(field string, message string) *AppError {
	return NewError(ErrCodeValidationFailed, fmt.Sprintf("字段 %s 验证失败: %s", field, message))
}

func NewDatabaseError(operation string, cause error) *AppError {
	return NewErrorWithCause(ErrCodeDatabaseError, fmt.Sprintf("数据库操作失败: %s", operation), cause)
}

// IsAppError 检查是否为应用程序错误
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError 获取应用程序错误
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}

// WrapError 包装普通错误为应用程序错误
func WrapError(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	
	if appErr := GetAppError(err); appErr != nil {
		return appErr
	}
	
	return NewErrorWithCause(code, message, err)
}
