package repository

import (
	"errors"
	"fmt"
)

// 预定义错误
var (
	ErrNotFound          = errors.New("记录未找到")
	ErrDuplicateKey      = errors.New("重复键冲突")
	ErrInvalidInput      = errors.New("无效输入")
	ErrConnectionFailed  = errors.New("数据库连接失败")
	ErrTransactionFailed = errors.New("事务执行失败")
	ErrTimeout           = errors.New("操作超时")
	ErrPermissionDenied  = errors.New("权限不足")
	ErrResourceLocked    = errors.New("资源被锁定")
	ErrConcurrentUpdate  = errors.New("并发更新冲突")
)

// RepositoryError 仓储层错误
type RepositoryError struct {
	Code    string
	Message string
	Cause   error
}

func (e *RepositoryError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *RepositoryError) Unwrap() error {
	return e.Cause
}

// NewRepositoryError 创建仓储错误
func NewRepositoryError(code, message string, cause error) *RepositoryError {
	return &RepositoryError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// 错误码常量
const (
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeDuplicateKey     = "DUPLICATE_KEY"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeConnectionFailed = "CONNECTION_FAILED"
	ErrCodeTransactionFailed = "TRANSACTION_FAILED"
	ErrCodeTimeout          = "TIMEOUT"
	ErrCodePermissionDenied = "PERMISSION_DENIED"
	ErrCodeResourceLocked   = "RESOURCE_LOCKED"
	ErrCodeConcurrentUpdate = "CONCURRENT_UPDATE"
)

// IsNotFoundError 检查是否为未找到错误
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsDuplicateKeyError 检查是否为重复键错误
func IsDuplicateKeyError(err error) bool {
	return errors.Is(err, ErrDuplicateKey)
}

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsConnectionError 检查是否为连接错误
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed)
}
