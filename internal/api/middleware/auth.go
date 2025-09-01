package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	
	"taskmanage/internal/container"
	"taskmanage/pkg/jwt"
	"taskmanage/pkg/response"
)

// Auth 认证中间件
func Auth(appContainer *container.ApplicationContainer) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := appContainer.GetLogger()
		
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header")
			response.Unauthorized(c, "缺少认证头")
			c.Abort()
			return
		}
		
		// 检查Bearer token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid authorization header format")
			response.Unauthorized(c, "认证头格式错误")
			c.Abort()
			return
		}
		
		token := parts[1]
		if token == "" {
			logger.Warn("Missing token")
			response.Unauthorized(c, "缺少令牌")
			c.Abort()
			return
		}
		
		// 获取JWT管理器
		jwtManager, err := appContainer.GetJWTManager()
		if err != nil {
			logger.WithError(err).Error("Failed to get JWT manager")
			response.InternalError(c, "认证服务不可用")
			c.Abort()
			return
		}
		
		// 验证JWT token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			logger.WithError(err).Warn("Invalid JWT token")
			switch err {
			case jwt.ErrTokenExpired:
				response.Unauthorized(c, "令牌已过期")
			case jwt.ErrTokenMalformed:
				response.Unauthorized(c, "令牌格式错误")
			case jwt.ErrTokenNotValidYet:
				response.Unauthorized(c, "令牌尚未生效")
			default:
				response.Unauthorized(c, "无效令牌")
			}
			c.Abort()
			return
		}
		
		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("token", token)
		
		logger.WithFields(logrus.Fields{
			"user_id":  claims.UserID,
			"username": claims.Username,
			"role":     claims.Role,
		}).Debug("User authenticated successfully")
		
		c.Next()
	}
}

// OptionalAuth 可选认证中间件
func OptionalAuth(appContainer *container.ApplicationContainer) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := appContainer.GetLogger()
		authHeader := c.GetHeader("Authorization")
		
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token := parts[1]
				
				// 获取JWT管理器
				jwtManager, err := appContainer.GetJWTManager()
				if err != nil {
					logger.WithError(err).Warn("Failed to get JWT manager in optional auth")
					c.Next()
					return
				}
				
				// 验证JWT token
				claims, err := jwtManager.ValidateToken(token)
				if err == nil {
					// 设置用户信息到上下文
					c.Set("user_id", claims.UserID)
					c.Set("username", claims.Username)
					c.Set("email", claims.Email)
					c.Set("role", claims.Role)
					c.Set("token", token)
					
					logger.WithFields(logrus.Fields{
						"user_id":  claims.UserID,
						"username": claims.Username,
					}).Debug("Optional auth: user authenticated")
				} else {
					logger.WithError(err).Debug("Optional auth: invalid token")
				}
			}
		}
		
		c.Next()
	}
}
