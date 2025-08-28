package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"taskmanage/pkg/logger"
	"taskmanage/pkg/response"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 记录panic信息
				logger.WithFields(map[string]interface{}{
					"panic":      r,
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"client_ip":  c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
					"stack":      string(debug.Stack()),
				}).Error("Panic recovered in ErrorHandler")

				// 返回内部服务器错误
				if !c.Writer.Written() {
					response.InternalError(c, "系统发生严重错误")
				}
				c.Abort()
				return
			}

			// 处理错误
			if len(c.Errors) > 0 {
				err := c.Errors.Last().Err

				// 记录错误
				logger.WithFields(map[string]interface{}{
					"error":      err.Error(),
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"client_ip":  c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				}).Error("Request error")

				// 如果响应还没有写入，处理错误
				if !c.Writer.Written() {
					response.Error(c, err)
				}
			}
		}()

		c.Next()
	}
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			logger.WithFields(map[string]interface{}{
				"panic":      err,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"client_ip":  c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"stack":      string(debug.Stack()),
			}).Error("String panic recovered")
		} else {
			logger.WithFields(map[string]interface{}{
				"panic":      fmt.Sprintf("%v", recovered),
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"client_ip":  c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"stack":      string(debug.Stack()),
			}).Error("Unknown panic recovered")
		}

		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// NotFoundHandler 404处理中间件
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.NotFound(c, "请求的资源不存在")
	}
}

// MethodNotAllowedHandler 405处理中间件
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.ErrorWithCode(c, response.ErrCodeInvalidRequest, "请求方法不被允许")
	}
}

// ValidationErrorHandler 参数验证错误处理
func ValidationErrorHandler(err error) *response.AppError {
	return response.NewValidationError("request", err.Error())
}

// DatabaseErrorHandler 数据库错误处理
func DatabaseErrorHandler(operation string, err error) *response.AppError {
	logger.WithFields(map[string]interface{}{
		"operation": operation,
		"error":     err.Error(),
	}).Error("Database operation failed")

	return response.NewDatabaseError(operation, err)
}

// ExternalServiceErrorHandler 外部服务错误处理
func ExternalServiceErrorHandler(service string, err error) *response.AppError {
	logger.WithFields(map[string]interface{}{
		"service": service,
		"error":   err.Error(),
	}).Error("External service call failed")

	return response.NewErrorWithCause(
		response.ErrCodeExternalServiceError,
		fmt.Sprintf("外部服务 %s 调用失败", service),
		err,
	)
}

// BusinessErrorHandler 业务逻辑错误处理
func BusinessErrorHandler(module string, err error) *response.AppError {
	logger.WithFields(map[string]interface{}{
		"module": module,
		"error":  err.Error(),
	}).Warn("Business logic error")

	if appErr := response.GetAppError(err); appErr != nil {
		return appErr
	}

	return response.WrapError(err, response.ErrCodeInternalError, "业务逻辑错误")
}
