package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取请求ID
		requestID := c.GetHeader("X-Request-ID")
		
		// 如果没有，则生成新的UUID
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// 设置到上下文和响应头
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}
