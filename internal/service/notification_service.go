package service

import (
	"context"
	"fmt"

	"taskmanage/internal/database"
	"taskmanage/internal/models"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// NotificationServiceImpl 通知服务实现
type NotificationServiceImpl struct {
	notificationRepo repository.NotificationRepository
}

// NewNotificationService 创建通知服务
func NewNotificationService(repoManager repository.RepositoryManager) NotificationService {
	return &NotificationServiceImpl{
		notificationRepo: repoManager.NotificationRepository(),
	}
}

// SendNotification 发送通知
func (s *NotificationServiceImpl) SendNotification(ctx context.Context, req *SendNotificationRequest) (*NotificationResponse, error) {
	// TODO: 实现发送通知逻辑
	return nil, fmt.Errorf("未实现")
}

// ListNotifications 获取通知列表 (为Handler提供的方法)
func (s *NotificationServiceImpl) ListNotifications(ctx context.Context, userID uint, unreadOnly bool) ([]*NotificationResponse, int64, error) {
	status := ""
	if unreadOnly {
		status = "unread"
	}
	
	notifications, total, err := s.notificationRepo.GetUserNotifications(ctx, userID, status, 1, 20)
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	responses := make([]*NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = &NotificationResponse{
			ID:        n.ID,
			Type:      string(n.Type),
			Title:     n.Title,
			Content:   n.Content,
			CreatedAt: n.CreatedAt,
		}
	}

	return responses, total, nil
}

// MarkAsRead 标记为已读
func (s *NotificationServiceImpl) MarkAsRead(ctx context.Context, notificationID, userID uint) error {
	return s.notificationRepo.MarkAsRead(ctx, notificationID, userID)
}

// MarkAllAsRead 标记所有为已读
func (s *NotificationServiceImpl) MarkAllAsRead(ctx context.Context, userID uint) error {
	_, err := s.notificationRepo.MarkAllAsRead(ctx, userID)
	return err
}

// GetUnreadCount 获取未读数量
func (s *NotificationServiceImpl) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	return s.notificationRepo.GetUnreadCount(ctx, userID)
}

// BroadcastNotification 广播通知
func (s *NotificationServiceImpl) BroadcastNotification(ctx context.Context, req *BroadcastNotificationRequest) error {
	// TODO: 实现广播通知逻辑
	return fmt.Errorf("未实现")
}

// CreateTaskAssignmentNotification 创建任务分配通知
func (s *NotificationServiceImpl) CreateTaskAssignmentNotification(ctx context.Context, taskID, recipientID, senderID uint) error {
	if err := s.notificationRepo.CreateTaskAssignmentNotification(ctx, taskID, recipientID, senderID); err != nil {
		logger.Errorf("创建任务分配通知失败: %v", err)
		return fmt.Errorf("创建任务分配通知失败: %w", err)
	}

	logger.Infof("成功创建任务分配通知: task_id=%d, recipient_id=%d", taskID, recipientID)
	return nil
}

// CreateTaskStatusNotification 创建任务状态变更通知
func (s *NotificationServiceImpl) CreateTaskStatusNotification(ctx context.Context, taskID, recipientID uint, notificationType models.TaskNotificationType, title, content string) error {
	notification := &database.TaskNotification{
		Type:        string(notificationType),
		Title:       title,
		Content:     content,
		RecipientID: recipientID,
		TaskID:      &taskID,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		logger.Errorf("创建任务状态通知失败: %v", err)
		return fmt.Errorf("创建任务状态通知失败: %w", err)
	}

	logger.Infof("成功创建任务状态通知: task_id=%d, recipient_id=%d, type=%s", taskID, recipientID, notificationType)
	return nil
}

// GetUserNotifications 获取用户通知
func (s *NotificationServiceImpl) GetUserNotifications(ctx context.Context, userID uint, status string, page, pageSize int) ([]models.TaskNotification, int64, error) {
	notifications, total, err := s.notificationRepo.GetUserNotifications(ctx, userID, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 转换为models.TaskNotification
	result := make([]models.TaskNotification, len(notifications))
	for i, n := range notifications {
		result[i] = models.TaskNotification{
			ID:          n.ID,
			Type:        models.TaskNotificationType(n.Type),
			Title:       n.Title,
			Content:     n.Content,
			RecipientID: n.RecipientID,
			SenderID:    n.SenderID,
			TaskID:      n.TaskID,
			Priority:    models.NotificationPriority(n.Priority),
			Status:      models.TaskNotificationStatus(n.Status),
			ActionType:  n.ActionType,
			ActionData:  n.ActionData,
			ExpiresAt:   n.ExpiresAt,
			ReadAt:      n.ReadAt,
			ActionAt:    n.ActionAt,
			CreatedAt:   n.CreatedAt,
			UpdatedAt:   n.UpdatedAt,
		}
	}

	return result, total, nil
}

// AcceptTaskNotification 接受任务通知
func (s *NotificationServiceImpl) AcceptTaskNotification(ctx context.Context, notificationID, taskID, userID uint, reason *string) error {
	if err := s.notificationRepo.AcceptTaskNotification(ctx, notificationID, taskID, userID, reason); err != nil {
		logger.Errorf("接受任务通知失败: %v", err)
		return fmt.Errorf("接受任务通知失败: %w", err)
	}

	logger.Infof("用户 %d 成功接受任务 %d", userID, taskID)
	return nil
}

// RejectTaskNotification 拒绝任务通知
func (s *NotificationServiceImpl) RejectTaskNotification(ctx context.Context, notificationID, taskID, userID uint, reason *string) error {
	if err := s.notificationRepo.RejectTaskNotification(ctx, notificationID, taskID, userID, reason); err != nil {
		logger.Errorf("拒绝任务通知失败: %v", err)
		return fmt.Errorf("拒绝任务通知失败: %w", err)
	}

	logger.Infof("用户 %d 拒绝了任务 %d，理由: %s", userID, taskID, *reason)
	return nil
}
