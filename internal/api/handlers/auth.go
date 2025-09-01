package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"taskmanage/internal/container"
	"taskmanage/internal/service"
	"taskmanage/pkg/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	container *container.ApplicationContainer
	logger    *logrus.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(container *container.ApplicationContainer, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		container: container,
		logger:    logger,
	}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid login request")
		response.BadRequest(c, "请求参数无效")
		return
	}

	// 获取用户服务
	userService := h.container.GetServiceManager().UserService()

	// 验证用户凭据
	user, err := userService.AuthenticateUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.WithError(err).WithField("username", req.Username).Warn("User authentication failed")
		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 获取JWT管理器
	jwtManager, err := h.container.GetJWTManager()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get JWT manager")
		response.InternalError(c, "认证服务不可用")
		return
	}

	// 生成访问令牌
	accessToken, err := jwtManager.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate access token")
		response.InternalError(c, "令牌生成失败")
		return
	}

	// 生成刷新令牌
	refreshToken, err := jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate refresh token")
		response.InternalError(c, "刷新令牌生成失败")
		return
	}

	// 构建响应
	loginResp := &service.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(jwtManager.GetTokenExpiry().Seconds()),
		User: service.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			RealName: user.RealName,
			Status:   user.Status,
			Role:     user.Role,
		},
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
	}).Info("User login successful")

	response.Success(c, loginResp)
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid registration request")
		response.BadRequest(c, "请求参数无效")
		return
	}

	// 获取用户服务
	userService := h.container.GetServiceManager().UserService()

	// 创建用户
	user, err := userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).WithField("username", req.Username).Warn("User registration failed")
		response.BadRequest(c, err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}).Info("User registration successful")

	response.Success(c, user)
}

// RefreshToken 刷新令牌
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid refresh token request")
		response.BadRequest(c, "请求参数无效")
		return
	}

	// 获取JWT管理器
	jwtManager, err := h.container.GetJWTManager()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get JWT manager")
		response.InternalError(c, "认证服务不可用")
		return
	}

	// 验证刷新令牌并获取用户ID
	userID, err := jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.WithError(err).Warn("Invalid refresh token")
		response.Unauthorized(c, "刷新令牌无效或已过期")
		return
	}

	// 获取用户信息
	userService := h.container.GetServiceManager().UserService()
	user, err := userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user for token refresh")
		response.Unauthorized(c, "用户不存在")
		return
	}

	// 生成新的访问令牌和刷新令牌
	newAccessToken, newRefreshToken, err := jwtManager.RefreshToken(req.RefreshToken, user.Username, user.Email, user.Role)
	if err != nil {
		h.logger.WithError(err).Error("Failed to refresh tokens")
		response.InternalError(c, "令牌刷新失败")
		return
	}

	// 构建响应
	loginResp := &service.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(jwtManager.GetTokenExpiry().Seconds()),
		User:         *user,
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
	}).Info("Token refresh successful")

	response.Success(c, loginResp)
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 获取用户服务
	userService := h.container.GetServiceManager().UserService()

	// 执行登出操作
	if err := userService.Logout(c.Request.Context(), userID); err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Warn("Logout operation failed")
		// 即使登出操作失败，也返回成功，因为客户端可以丢弃令牌
	}

	h.logger.WithField("user_id", userID).Info("User logout successful")
	response.Success(c, gin.H{"message": "登出成功"})
}
