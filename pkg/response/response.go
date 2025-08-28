package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    ErrCodeSuccess,
		Message: "操作成功",
		Data:    data,
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    ErrCodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	if appErr := GetAppError(err); appErr != nil {
		c.JSON(appErr.HTTPStatus, Response{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		})
		return
	}

	// 未知错误，返回内部服务器错误
	c.JSON(http.StatusInternalServerError, Response{
		Code:    ErrCodeInternalError,
		Message: "内部服务器错误",
	})
}

// ErrorWithCode 指定错误码的错误响应
func ErrorWithCode(c *gin.Context, code ErrorCode, message string) {
	httpStatus := getHTTPStatus(code)
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 400错误响应
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    ErrCodeInvalidRequest,
		Message: message,
	})
}

// Unauthorized 401错误响应
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    ErrCodeUnauthorized,
		Message: message,
	})
}

// Forbidden 403错误响应
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    ErrCodeForbidden,
		Message: message,
	})
}

// NotFound 404错误响应
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    ErrCodeNotFound,
		Message: message,
	})
}

// Conflict 409错误响应
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Code:    ErrCodeConflict,
		Message: message,
	})
}

// ValidationError 参数验证错误响应
func ValidationError(c *gin.Context, details interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    ErrCodeValidationFailed,
		Message: "参数验证失败",
		Details: details,
	})
}

// InternalError 内部服务器错误响应
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    ErrCodeInternalError,
		Message: message,
	})
}

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	Code       ErrorCode   `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination 分页信息
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// SuccessWithPagination 带分页的成功响应
func SuccessWithPagination(c *gin.Context, data interface{}, page, pageSize int, total int64) {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	
	c.JSON(http.StatusOK, PaginationResponse{
		Code:    ErrCodeSuccess,
		Message: "操作成功",
		Data:    data,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
