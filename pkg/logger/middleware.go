package logger

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GinLogger Gin框架的日志中间件
func GinLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		fields := logrus.Fields{
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
			"status_code": param.StatusCode,
			"latency":     param.Latency.String(),
			"user_agent":  param.Request.UserAgent(),
		}

		if param.ErrorMessage != "" {
			fields["error"] = param.ErrorMessage
		}

		entry := GetLogger().WithFields(fields)

		// 根据状态码选择日志级别
		switch {
		case param.StatusCode >= 500:
			entry.Error("HTTP Request")
		case param.StatusCode >= 400:
			entry.Warn("HTTP Request")
		default:
			entry.Info("HTTP Request")
		}

		return ""
	})
}

// GinRecovery Gin框架的恢复中间件
func GinRecovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(GetLogger().Writer(), func(c *gin.Context, recovered interface{}) {
		GetLogger().WithFields(logrus.Fields{
			"panic":      recovered,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}).Error("Panic recovered")

		c.AbortWithStatus(500)
	})
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 读取请求体
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 构建日志字段
		fields := logrus.Fields{
			"status_code": c.Writer.Status(),
			"method":      c.Request.Method,
			"path":        path,
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"latency":     latency.String(),
			"latency_ms":  float64(latency.Nanoseconds()) / 1000000,
		}

		if raw != "" {
			fields["query"] = raw
		}

		// 添加请求体（仅对POST/PUT/PATCH请求，且大小合理）
		if len(bodyBytes) > 0 && len(bodyBytes) < 1024 {
			if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
				fields["request_body"] = string(bodyBytes)
			}
		}

		// 添加响应大小
		fields["response_size"] = c.Writer.Size()

		// 添加错误信息
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// 根据状态码和延迟选择日志级别
		entry := GetLogger().WithFields(fields)
		switch {
		case c.Writer.Status() >= 500:
			entry.Error("HTTP Request Completed")
		case c.Writer.Status() >= 400:
			entry.Warn("HTTP Request Completed")
		case latency > time.Second:
			entry.Warn("Slow HTTP Request Completed")
		default:
			entry.Info("HTTP Request Completed")
		}
	}
}

// DatabaseLogger 数据库日志中间件
type DatabaseLogger struct {
	logger *logrus.Logger
}

// NewDatabaseLogger 创建数据库日志记录器
func NewDatabaseLogger() *DatabaseLogger {
	return &DatabaseLogger{
		logger: GetLogger(),
	}
}

// LogMode 设置日志模式
func (l *DatabaseLogger) LogMode(level logrus.Level) *DatabaseLogger {
	newLogger := *l
	newLogger.logger.SetLevel(level)
	return &newLogger
}

// Info 信息日志
func (l *DatabaseLogger) Info(msg string, data ...interface{}) {
	l.logger.WithFields(logrus.Fields{
		"component": "database",
		"data":      data,
	}).Info(msg)
}

// Warn 警告日志
func (l *DatabaseLogger) Warn(msg string, data ...interface{}) {
	l.logger.WithFields(logrus.Fields{
		"component": "database",
		"data":      data,
	}).Warn(msg)
}

// Error 错误日志
func (l *DatabaseLogger) Error(msg string, data ...interface{}) {
	l.logger.WithFields(logrus.Fields{
		"component": "database",
		"data":      data,
	}).Error(msg)
}

// Trace 跟踪日志
func (l *DatabaseLogger) Trace(begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	
	fields := logrus.Fields{
		"component": "database",
		"sql":       sql,
		"rows":      rows,
		"elapsed":   elapsed.String(),
		"elapsed_ms": float64(elapsed.Nanoseconds()) / 1000000,
	}

	if err != nil {
		fields["error"] = err.Error()
		l.logger.WithFields(fields).Error("Database Query Failed")
	} else {
		switch {
		case elapsed > time.Second:
			l.logger.WithFields(fields).Warn("Slow Database Query")
		default:
			l.logger.WithFields(fields).Debug("Database Query")
		}
	}
}
