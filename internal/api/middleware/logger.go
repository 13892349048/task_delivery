package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger 日志中间件
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		// 获取请求ID
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = "unknown"
		}
		
		// 记录请求开始
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
		}).Info("HTTP request started")
		
		// 处理请求
		c.Next()
		
		// 计算延迟
		latency := time.Since(start)
		
		// 构建完整路径
		if raw != "" {
			path = path + "?" + raw
		}
		
		// 记录日志
		logFields := logrus.Fields{
			"request_id": requestID,
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"query":      raw,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"latency":    latency,
			"body_size":  c.Writer.Size(),
		}
		
		if c.Writer.Status() >= 400 {
			logger.WithFields(logFields).Error("HTTP request failed")
		} else {
			logger.WithFields(logFields).Info("HTTP request completed")
		}
	}
}
