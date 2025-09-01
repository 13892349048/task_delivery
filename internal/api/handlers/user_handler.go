package handlers

import (
	"fmt"
	"strconv"

	"taskmanage/internal/container"
	"taskmanage/internal/service"
	"taskmanage/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// UserHandler 用户处理器
type UserHandler struct {
	*BaseHandler
}

// NewUserHandler 创建用户处理器
func NewUserHandler(container *container.ApplicationContainer, logger *logrus.Logger) *UserHandler {
	return &UserHandler{
		BaseHandler: &BaseHandler{container: container, logger: logger},
	}
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	keyword := c.Query("keyword")
	status := c.Query("status")
	role := c.Query("role")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	// 构建查询条件
	conditions := make(map[string]interface{})
	if status != "" {
		conditions["status"] = status
	}
	if role != "" {
		conditions["role"] = role
	}

	userRepo := h.container.GetRepositoryManager().UserRepository()
	users, total, err := userRepo.ListUsers(c, page, limit, conditions, keyword)
	if err != nil {
		h.logger.WithError(err).Error("获取用户列表失败")
		response.InternalError(c, "获取用户列表失败")
		return
	}

	// 构建响应数据
	responseData := gin.H{
		"users": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	}

	response.Success(c, responseData)
}

// GetUser 获取单个用户信息
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	// 转换用户ID
	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	userRepo := h.container.GetRepositoryManager().UserRepository()
	user, err := userRepo.GetByID(c, id)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", id).Error("获取用户失败")
		response.NotFound(c, "用户不存在")
		return
	}

	response.Success(c, user)
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	// 绑定请求数据
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		RealName string `json:"real_name" binding:"required"`
		Role     string `json:"role"`
		Phone    string `json:"phone"`
		Avatar   string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 使用用户服务创建用户
	userService := h.container.GetServiceManager().UserService()
	createReq := &service.CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		RealName: req.RealName,
	}

	userResp, err := userService.CreateUser(c, createReq)
	if err != nil {
		h.logger.WithError(err).Error("创建用户失败")
		response.InternalError(c, err.Error())
		return
	}

	h.logger.WithField("user_id", userResp.ID).Info("用户创建成功")
	response.Success(c, userResp)
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	// 转换用户ID
	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	// 绑定请求数据
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		RealName string `json:"real_name"`
		Role     string `json:"role"`
		Phone    string `json:"phone"`
		Avatar   string `json:"avatar"`
		Status   string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userRepo := h.container.GetRepositoryManager().UserRepository()

	// 获取现有用户
	user, err := userRepo.GetByID(c, id)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", id).Error("获取用户失败")
		response.NotFound(c, "用户不存在")
		return
	}

	// 检查用户名和邮箱唯一性（如果有更新）
	if req.Username != "" && req.Username != user.Username {
		if existingUser, _ := userRepo.GetByUsername(c, req.Username); existingUser != nil {
			response.BadRequest(c, "用户名已存在")
			return
		}
		user.Username = req.Username
	}

	if req.Email != "" && req.Email != user.Email {
		if existingUser, _ := userRepo.GetByEmail(c, req.Email); existingUser != nil {
			response.BadRequest(c, "邮箱已存在")
			return
		}
		user.Email = req.Email
	}

	// 更新其他字段
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	// 保存更新
	if err := userRepo.Update(c, user); err != nil {
		h.logger.WithError(err).WithField("user_id", id).Error("更新用户失败")
		response.InternalError(c, "更新用户失败")
		return
	}

	h.logger.WithField("user_id", id).Info("用户更新成功")
	response.Success(c, user)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	// 转换用户ID
	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	userRepo := h.container.GetRepositoryManager().UserRepository()
	if err := userRepo.Delete(c, id); err != nil {
		h.logger.WithError(err).WithField("user_id", id).Error("删除用户失败")
		response.InternalError(c, "删除用户失败")
		return
	}

	h.logger.WithField("user_id", id).Info("用户删除成功")
	response.Success(c, gin.H{"message": "用户删除成功"})
}

// AssignRoles 分配角色给用户
func (h *UserHandler) AssignRoles(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	// 转换用户ID
	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	// 绑定请求数据
	var req struct {
		RoleIDs []uint `json:"role_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userRepo := h.container.GetRepositoryManager().UserRepository()

	// 检查用户是否存在
	_, err := userRepo.GetByID(c, id)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", id).Error("获取用户失败")
		response.NotFound(c, "用户不存在")
		return
	}

	// 分配角色
	if err := userRepo.AssignRoles(c, id, req.RoleIDs); err != nil {
		h.logger.WithError(err).WithField("user_id", id).Error("分配角色失败")
		response.InternalError(c, "分配角色失败")
		return
	}

	h.logger.WithField("user_id", id).Info("角色分配成功")
	response.Success(c, gin.H{
		"message":  "角色分配成功",
		"user_id":  id,
		"role_ids": req.RoleIDs,
	})
}

// RemoveRoles 移除用户角色
func (h *UserHandler) RemoveRoles(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	// 转换用户ID
	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	// 绑定请求数据
	var req struct {
		RoleIDs []uint `json:"role_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userRepo := h.container.GetRepositoryManager().UserRepository()

	// 检查用户是否存在
	_, err := userRepo.GetByID(c, id)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", id).Error("获取用户失败")
		response.NotFound(c, "用户不存在")
		return
	}

	// 移除角色
	if err := userRepo.RemoveRoles(c, id, req.RoleIDs); err != nil {
		h.logger.WithError(err).WithField("user_id", id).Error("移除角色失败")
		response.InternalError(c, "移除角色失败")
		return
	}

	h.logger.WithField("user_id", id).Info("角色移除成功")
	response.Success(c, gin.H{
		"message":  "角色移除成功",
		"user_id":  id,
		"role_ids": req.RoleIDs,
	})
}
