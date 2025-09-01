package handlers

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

// GetUserIDFromContext 从Gin上下文中获取当前用户ID
func GetUserIDFromContext(c *gin.Context) (uint, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("用户ID不存在于上下文中")
	}
	
	// JWT claims中的UserID是uint类型
	if id, ok := userID.(uint); ok {
		return id, nil
	}
	
	return 0, fmt.Errorf("用户ID类型错误")
}

// GetUsernameFromContext 从Gin上下文中获取当前用户名
func GetUsernameFromContext(c *gin.Context) (string, error) {
	username, exists := c.Get("username")
	if !exists {
		return "", fmt.Errorf("用户名不存在于上下文中")
	}
	
	if name, ok := username.(string); ok {
		return name, nil
	}
	
	return "", fmt.Errorf("用户名类型错误")
}

// GetUserRoleFromContext 从Gin上下文中获取当前用户角色
func GetUserRoleFromContext(c *gin.Context) (string, error) {
	role, exists := c.Get("role")
	if !exists {
		return "", fmt.Errorf("用户角色不存在于上下文中")
	}
	
	if r, ok := role.(string); ok {
		return r, nil
	}
	
	return "", fmt.Errorf("用户角色类型错误")
}

// SetUserIDInContext 将用户ID设置到标准context中（用于服务层）
func SetUserIDInContext(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, "user_id", userID)
}

// GetUserIDFromStandardContext 从标准context中获取用户ID（用于服务层）
func GetUserIDFromStandardContext(ctx context.Context) (uint, error) {
	userID := ctx.Value("user_id")
	if userID == nil {
		return 0, fmt.Errorf("用户ID不存在于上下文中")
	}
	
	if id, ok := userID.(uint); ok {
		return id, nil
	}
	
	return 0, fmt.Errorf("用户ID类型错误")
}
