package handlers

import (
	"strconv"

	"taskmanage/internal/container"
	"taskmanage/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 基础处理器结构体
type BaseHandler struct {
	container *container.ApplicationContainer
	logger    *logrus.Logger
}

// TaskHandler 任务处理器 - 已移动到单独的文件 task.go

// 任务状态流转方法已移动到task.go文件中实现

// AssignmentHandler 分配处理器
// type AssignmentHandler struct {
// 	*BaseHandler
// }

// func NewAssignmentHandler(container *container.ApplicationContainer, logger *logrus.Logger) *AssignmentHandler {
// 	return &AssignmentHandler{
// 		BaseHandler: &BaseHandler{container: container, logger: logger},
// 	}
// }

// func (h *AssignmentHandler) ListAssignments(c *gin.Context) {
// 	response.Success(c, gin.H{"message": "ListAssignments - TODO: implement"})
// }

// func (h *AssignmentHandler) ApproveAssignment(c *gin.Context) {
// 	response.Success(c, gin.H{"message": "ApproveAssignment - TODO: implement"})
// }

// func (h *AssignmentHandler) RejectAssignment(c *gin.Context) {
// 	response.Success(c, gin.H{"message": "RejectAssignment - TODO: implement"})
// }

// func (h *AssignmentHandler) GetAssignmentHistory(c *gin.Context) {
// 	response.Success(c, gin.H{"message": "GetAssignmentHistory - TODO: implement"})
// }

// EmployeeHandler is implemented in employee.go

// NotificationHandler 通知处理器
type NotificationHandler struct {
	*BaseHandler
}

func NewNotificationHandler(container *container.ApplicationContainer, logger *logrus.Logger) *NotificationHandler {
	return &NotificationHandler{
		BaseHandler: &BaseHandler{container: container, logger: logger},
	}
}

// GetNotifications 获取用户通知列表
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	// 解析查询参数
	unreadOnly := c.Query("unread_only") == "true"

	// 使用NotificationService获取通知
	notificationService := h.container.GetServiceManager().NotificationService()
	notifications, total, err := notificationService.ListNotifications(c.Request.Context(), userID.(uint), unreadOnly)
	if err != nil {
		h.logger.WithError(err).Error("获取通知列表失败")
		response.InternalError(c, "获取通知列表失败")
		return
	}

	response.Success(c, gin.H{
		"notifications": notifications,
		"total":         total,
	})
}

func (h *NotificationHandler) SendNotification(c *gin.Context) {
	response.Success(c, gin.H{"message": "SendNotification - TODO: implement"})
}

func (h *NotificationHandler) BroadcastNotification(c *gin.Context) {
	response.Success(c, gin.H{"message": "BroadcastNotification - TODO: implement"})
}

// MarkAsRead 标记通知为已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的通知ID")
		return
	}

	// 使用NotificationService标记已读
	notificationService := h.container.GetServiceManager().NotificationService()
	err = notificationService.MarkAsRead(c.Request.Context(), uint(notificationID), userID.(uint))
	if err != nil {
		h.logger.WithError(err).Error("标记通知已读失败")
		response.InternalError(c, "标记通知已读失败")
		return
	}

	response.Success(c, gin.H{"message": "通知已标记为已读"})
}

// MarkAllAsRead 标记所有通知为已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	// 使用NotificationService批量标记已读
	notificationService := h.container.GetServiceManager().NotificationService()
	err := notificationService.MarkAllAsRead(c.Request.Context(), userID.(uint))
	if err != nil {
		h.logger.WithError(err).Error("批量标记通知已读失败")
		response.InternalError(c, "批量标记通知已读失败")
		return
	}

	response.Success(c, gin.H{
		"message": "批量标记完成",
	})
}

// GetUnreadCount 获取未读通知数量
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	// 使用NotificationService获取未读数量
	notificationService := h.container.GetServiceManager().NotificationService()
	count, err := notificationService.GetUnreadCount(c.Request.Context(), userID.(uint))
	if err != nil {
		h.logger.WithError(err).Error("获取未读通知数量失败")
		response.InternalError(c, "获取未读通知数量失败")
		return
	}

	response.Success(c, gin.H{"count": count})
}

// AcceptTask 接受任务
func (h *NotificationHandler) AcceptTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	var req struct {
		NotificationID uint    `json:"notification_id" binding:"required"`
		TaskID         uint    `json:"task_id" binding:"required"`
		Reason         *string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 使用NotificationService处理任务接受
	notificationService := h.container.GetServiceManager().NotificationService()
	err := notificationService.AcceptTaskNotification(c.Request.Context(), req.NotificationID, req.TaskID, userID.(uint), req.Reason)
	if err != nil {
		h.logger.WithError(err).Error("接受任务失败")
		response.InternalError(c, "接受任务失败")
		return
	}

	response.Success(c, gin.H{"message": "任务已接受"})
}

// RejectTask 拒绝任务
func (h *NotificationHandler) RejectTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	var req struct {
		NotificationID uint    `json:"notification_id" binding:"required"`
		TaskID         uint    `json:"task_id" binding:"required"`
		Reason         *string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.Reason == nil || *req.Reason == "" {
		response.BadRequest(c, "拒绝任务必须提供理由")
		return
	}

	// 使用NotificationService处理任务拒绝
	notificationService := h.container.GetServiceManager().NotificationService()
	err := notificationService.RejectTaskNotification(c.Request.Context(), req.NotificationID, req.TaskID, userID.(uint), req.Reason)
	if err != nil {
		h.logger.WithError(err).Error("拒绝任务失败")
		response.InternalError(c, "拒绝任务失败")
		return
	}

	response.Success(c, gin.H{"message": "任务已拒绝"})
}
