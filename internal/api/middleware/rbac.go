package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"taskmanage/internal/container"
	"taskmanage/pkg/response"
)

// RequireRole 要求特定角色的中间件
func RequireRole(appContainer *container.ApplicationContainer, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := appContainer.GetLogger()
		
		// 获取用户角色
		userRole, exists := c.Get("role")
		if !exists {
			logger.Warn("User role not found in context")
			response.Forbidden(c, "访问被拒绝：角色信息缺失")
			c.Abort()
			return
		}
		
		roleStr, ok := userRole.(string)
		if !ok {
			logger.Warn("Invalid role type in context")
			response.Forbidden(c, "访问被拒绝：角色信息无效")
			c.Abort()
			return
		}
		
		// 检查角色权限
		for _, requiredRole := range roles {
			if roleStr == requiredRole {
				logger.WithFields(logrus.Fields{
					"user_role":     roleStr,
					"required_role": requiredRole,
				}).Debug("Role check passed")
				c.Next()
				return
			}
		}
		
		logger.WithFields(logrus.Fields{
			"user_role":      roleStr,
			"required_roles": roles,
		}).Warn("Role check failed")
		
		response.Forbidden(c, "访问被拒绝：权限不足")
		c.Abort()
	}
}

// RequirePermission 要求特定权限的中间件
func RequirePermission(appContainer *container.ApplicationContainer, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := appContainer.GetLogger()
		
		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			logger.Warn("User ID not found in context")
			response.Forbidden(c, "访问被拒绝：用户信息缺失")
			c.Abort()
			return
		}
		
		userIDUint, ok := userID.(uint)
		if !ok {
			logger.Warn("Invalid user ID type in context")
			response.Forbidden(c, "访问被拒绝：用户信息无效")
			c.Abort()
			return
		}
		
		// 获取服务管理器
		serviceManager := appContainer.GetServiceManager()
		
		// 检查用户权限
		hasPermission, err := serviceManager.UserService().HasPermission(c.Request.Context(), userIDUint, resource, action)
		if err != nil {
			logger.WithError(err).Error("Failed to check user permission")
			response.InternalError(c, "权限检查失败")
			c.Abort()
			return
		}
		
		if !hasPermission {
			logger.WithFields(logrus.Fields{
				"user_id":  userIDUint,
				"resource": resource,
				"action":   action,
			}).Warn("Permission check failed")
			
			response.Forbidden(c, "访问被拒绝：权限不足")
			c.Abort()
			return
		}
		
		logger.WithFields(logrus.Fields{
			"user_id":  userIDUint,
			"resource": resource,
			"action":   action,
		}).Debug("Permission check passed")
		
		c.Next()
	}
}

// RequireAdmin 要求管理员权限的中间件
func RequireAdmin(appContainer *container.ApplicationContainer) gin.HandlerFunc {
	return RequireRole(appContainer, "admin")
}

// RequireManager 要求经理权限的中间件
func RequireManager(appContainer *container.ApplicationContainer) gin.HandlerFunc {
	return RequireRole(appContainer, "admin", "manager")
}

// RequireOwnerOrAdmin 要求资源所有者或管理员权限的中间件
func RequireOwnerOrAdmin(appContainer *container.ApplicationContainer, resourceIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := appContainer.GetLogger()
		
		// 获取用户信息
		userID, exists := c.Get("user_id")
		if !exists {
			logger.Warn("User ID not found in context")
			response.Forbidden(c, "访问被拒绝：用户信息缺失")
			c.Abort()
			return
		}
		
		userIDUint, ok := userID.(uint)
		if !ok {
			logger.Warn("Invalid user ID type in context")
			response.Forbidden(c, "访问被拒绝：用户信息无效")
			c.Abort()
			return
		}
		
		userRole, exists := c.Get("role")
		if !exists {
			logger.Warn("User role not found in context")
			response.Forbidden(c, "访问被拒绝：角色信息缺失")
			c.Abort()
			return
		}
		
		roleStr, ok := userRole.(string)
		if !ok {
			logger.Warn("Invalid role type in context")
			response.Forbidden(c, "访问被拒绝：角色信息无效")
			c.Abort()
			return
		}
		
		// 管理员直接通过
		if roleStr == "admin" {
			logger.WithFields(logrus.Fields{
				"user_id": userIDUint,
				"role":    roleStr,
			}).Debug("Admin access granted")
			c.Next()
			return
		}
		
		// 检查是否为资源所有者
		resourceIDStr := c.Param(resourceIDParam)
		if resourceIDStr == "" {
			logger.Warn("Resource ID parameter not found")
			response.BadRequest(c, "资源ID参数缺失")
			c.Abort()
			return
		}
		
		// 简单的所有者检查（这里可以根据具体业务逻辑扩展）
		// 例如检查任务的创建者、分配者等
		if resourceIDStr == "me" || strings.Contains(c.Request.URL.Path, "/users/"+string(rune(userIDUint))) {
			logger.WithFields(logrus.Fields{
				"user_id":     userIDUint,
				"resource_id": resourceIDStr,
			}).Debug("Owner access granted")
			c.Next()
			return
		}
		
		logger.WithFields(logrus.Fields{
			"user_id":     userIDUint,
			"role":        roleStr,
			"resource_id": resourceIDStr,
		}).Warn("Owner/Admin check failed")
		
		response.Forbidden(c, "访问被拒绝：只能访问自己的资源")
		c.Abort()
	}
}

// RequireTaskAccess 要求任务访问权限的中间件
func RequireTaskAccess(appContainer *container.ApplicationContainer) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := appContainer.GetLogger()
		
		// 获取用户信息
		userID, exists := c.Get("user_id")
		if !exists {
			logger.Warn("User ID not found in context")
			response.Forbidden(c, "访问被拒绝：用户信息缺失")
			c.Abort()
			return
		}
		
		userIDUint, ok := userID.(uint)
		if !ok {
			logger.Warn("Invalid user ID type in context")
			response.Forbidden(c, "访问被拒绝：用户信息无效")
			c.Abort()
			return
		}
		
		userRole, exists := c.Get("role")
		if !exists {
			logger.Warn("User role not found in context")
			response.Forbidden(c, "访问被拒绝：角色信息缺失")
			c.Abort()
			return
		}
		
		roleStr, ok := userRole.(string)
		if !ok {
			logger.Warn("Invalid role type in context")
			response.Forbidden(c, "访问被拒绝：角色信息无效")
			c.Abort()
			return
		}
		
		// 管理员和经理直接通过
		if roleStr == "admin" || roleStr == "manager" {
			logger.WithFields(logrus.Fields{
				"user_id": userIDUint,
				"role":    roleStr,
			}).Debug("Admin/Manager task access granted")
			c.Next()
			return
		}
		
		// 获取任务ID
		taskIDStr := c.Param("id")
		if taskIDStr == "" {
			logger.Warn("Task ID parameter not found")
			response.BadRequest(c, "任务ID参数缺失")
			c.Abort()
			return
		}
		
		// 这里应该检查用户是否有访问该任务的权限
		// 例如：用户是任务的创建者、分配者或被分配者
		// 为了简化，这里先允许所有认证用户访问
		logger.WithFields(logrus.Fields{
			"user_id": userIDUint,
			"task_id": taskIDStr,
		}).Debug("Task access granted for authenticated user")
		
		c.Next()
	}
}
