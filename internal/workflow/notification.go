package workflow

import (
	"context"
	"fmt"
	"taskmanage/pkg/logger"
	"time"
)

// NotificationService 通知服务
type NotificationService struct {
	instanceRepo WorkflowInstanceRepository
}

// NewNotificationService 创建通知服务
func NewNotificationService(instanceRepo WorkflowInstanceRepository) *NotificationService {
	return &NotificationService{
		instanceRepo: instanceRepo,
	}
}

// SendApprovalNotification 发送审批通知
func (s *NotificationService) SendApprovalNotification(ctx context.Context, req *ApprovalNotificationRequest) error {
	logger.Infof("发送审批通知: 实例=%s, 用户=%d", req.InstanceID, req.UserID)

	// 获取流程实例信息
	instance, err := s.instanceRepo.GetInstance(ctx, req.InstanceID)
	if err != nil {
		return fmt.Errorf("获取流程实例失败: %w", err)
	}

	// 构建通知内容
	notification := &Notification{
		ID:         generateNotificationID(),
		Type:       NotificationTypeApproval,
		Title:      req.Title,
		Content:    req.Content,
		Recipient:  req.UserID,
		InstanceID: req.InstanceID,
		NodeID:     req.NodeID,
		Priority:   req.Priority,
		CreatedAt:  time.Now(),
		ExpiresAt:  req.ExpiresAt,
		Metadata: map[string]interface{}{
			"workflow_id":   instance.WorkflowID,
			"business_id":   instance.BusinessID,
			"business_type": instance.BusinessType,
		},
	}

	// 发送通知
	if err := s.sendNotification(ctx, notification); err != nil {
		return fmt.Errorf("发送通知失败: %w", err)
	}

	logger.Infof("审批通知发送成功: %s", notification.ID)
	return nil
}

// SendReminderNotification 发送提醒通知
func (s *NotificationService) SendReminderNotification(ctx context.Context, req *ReminderNotificationRequest) error {
	logger.Infof("发送提醒通知: 实例=%s, 用户=%d", req.InstanceID, req.UserID)

	notification := &Notification{
		ID:         generateNotificationID(),
		Type:       NotificationTypeReminder,
		Title:      req.Title,
		Content:    req.Content,
		Recipient:  req.UserID,
		InstanceID: req.InstanceID,
		NodeID:     req.NodeID,
		Priority:   req.Priority,
		CreatedAt:  time.Now(),
		Metadata: map[string]interface{}{
			"reminder_count": req.ReminderCount,
			"original_date":  req.OriginalDate,
		},
	}

	if err := s.sendNotification(ctx, notification); err != nil {
		return fmt.Errorf("发送提醒通知失败: %w", err)
	}

	logger.Infof("提醒通知发送成功: %s", notification.ID)
	return nil
}

// SendCompletionNotification 发送完成通知
func (s *NotificationService) SendCompletionNotification(ctx context.Context, req *CompletionNotificationRequest) error {
	logger.Infof("发送完成通知: 实例=%s", req.InstanceID)

	// 获取所有相关用户
	recipients := s.getCompletionRecipients(ctx, req)

	for _, userID := range recipients {
		notification := &Notification{
			ID:         generateNotificationID(),
			Type:       NotificationTypeCompletion,
			Title:      req.Title,
			Content:    req.Content,
			Recipient:  userID,
			InstanceID: req.InstanceID,
			Priority:   NotificationPriorityMedium,
			CreatedAt:  time.Now(),
			Metadata: map[string]interface{}{
				"result":       req.Result,
				"completed_at": req.CompletedAt,
			},
		}

		if err := s.sendNotification(ctx, notification); err != nil {
			logger.Errorf("发送完成通知失败: 用户=%d, error=%v", userID, err)
		}
	}

	return nil
}

// ScheduleReminders 安排提醒任务
func (s *NotificationService) ScheduleReminders(ctx context.Context, instanceID string, nodeID string, userID uint, deadline *time.Time) error {
	if deadline == nil {
		return nil // 没有截止时间，不需要提醒
	}

	logger.Infof("安排提醒任务: 实例=%s, 节点=%s, 用户=%d", instanceID, nodeID, userID)

	// 计算提醒时间点
	reminderTimes := s.calculateReminderTimes(*deadline)

	for i, reminderTime := range reminderTimes {
		reminder := &ScheduledReminder{
			ID:           generateReminderID(),
			InstanceID:   instanceID,
			NodeID:       nodeID,
			UserID:       userID,
			ScheduledAt:  reminderTime,
			ReminderType: s.getReminderType(i),
			Status:       ReminderStatusPending,
			CreatedAt:    time.Now(),
		}

		// 这里应该将提醒任务保存到数据库或消息队列
		logger.Infof("安排提醒: %s 在 %s", reminder.ReminderType, reminderTime.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// ProcessScheduledReminders 处理预定的提醒
func (s *NotificationService) ProcessScheduledReminders(ctx context.Context) error {
	// 这里应该从数据库查询到期的提醒任务
	// 简化实现，模拟处理逻辑
	logger.Info("处理预定的提醒任务")

	// 模拟查询到期的提醒
	dueReminders := s.getDueReminders(ctx)

	for _, reminder := range dueReminders {
		if err := s.processReminder(ctx, reminder); err != nil {
			logger.Errorf("处理提醒失败: %s, error: %v", reminder.ID, err)
		}
	}

	return nil
}

// sendNotification 发送通知
func (s *NotificationService) sendNotification(ctx context.Context, notification *Notification) error {
	// 根据通知类型选择发送渠道
	switch notification.Type {
	case NotificationTypeApproval, NotificationTypeReminder:
		// 发送系统通知
		if err := s.sendSystemNotification(ctx, notification); err != nil {
			return err
		}

		// 如果是高优先级，同时发送邮件
		if notification.Priority == NotificationPriorityHigh {
			if err := s.sendEmailNotification(ctx, notification); err != nil {
				logger.Errorf("发送邮件通知失败: %v", err)
			}
		}
	case NotificationTypeCompletion:
		// 发送系统通知
		if err := s.sendSystemNotification(ctx, notification); err != nil {
			return err
		}
	}

	return nil
}

// sendSystemNotification 发送系统通知
func (s *NotificationService) sendSystemNotification(ctx context.Context, notification *Notification) error {
	// 这里应该调用系统通知服务
	logger.Infof("发送系统通知: 用户=%d, 标题=%s", notification.Recipient, notification.Title)
	return nil
}

// sendEmailNotification 发送邮件通知
func (s *NotificationService) sendEmailNotification(ctx context.Context, notification *Notification) error {
	// 这里应该调用邮件服务
	logger.Infof("发送邮件通知: 用户=%d, 标题=%s", notification.Recipient, notification.Title)
	return nil
}

// calculateReminderTimes 计算提醒时间点
func (s *NotificationService) calculateReminderTimes(deadline time.Time) []time.Time {
	var times []time.Time
	now := time.Now()

	// 提前24小时提醒
	reminder24h := deadline.Add(-24 * time.Hour)
	if reminder24h.After(now) {
		times = append(times, reminder24h)
	}

	// 提前4小时提醒
	reminder4h := deadline.Add(-4 * time.Hour)
	if reminder4h.After(now) {
		times = append(times, reminder4h)
	}

	// 提前1小时提醒
	reminder1h := deadline.Add(-1 * time.Hour)
	if reminder1h.After(now) {
		times = append(times, reminder1h)
	}

	// 超时提醒（截止时间后1小时）
	overtimeReminder := deadline.Add(1 * time.Hour)
	times = append(times, overtimeReminder)

	return times
}

// getReminderType 获取提醒类型
func (s *NotificationService) getReminderType(index int) ReminderType {
	types := []ReminderType{
		ReminderType24Hours,
		ReminderType4Hours,
		ReminderType1Hour,
		ReminderTypeOvertime,
	}
	if index < len(types) {
		return types[index]
	}
	return ReminderTypeOvertime
}

// getCompletionRecipients 获取完成通知接收人
func (s *NotificationService) getCompletionRecipients(ctx context.Context, req *CompletionNotificationRequest) []uint {
	// 这里应该根据流程实例获取相关人员
	// 简化实现，返回模拟数据
	return []uint{req.InitiatorID}
}

// getDueReminders 获取到期的提醒
func (s *NotificationService) getDueReminders(ctx context.Context) []*ScheduledReminder {
	// 这里应该从数据库查询到期的提醒
	// 简化实现，返回空列表
	return []*ScheduledReminder{}
}

// processReminder 处理提醒
func (s *NotificationService) processReminder(ctx context.Context, reminder *ScheduledReminder) error {
	// 构建提醒通知
	title := s.buildReminderTitle(reminder.ReminderType)
	content := s.buildReminderContent(reminder)

	req := &ReminderNotificationRequest{
		InstanceID:    reminder.InstanceID,
		NodeID:        reminder.NodeID,
		UserID:        reminder.UserID,
		Title:         title,
		Content:       content,
		Priority:      s.getReminderPriority(reminder.ReminderType),
		ReminderCount: 1,
		OriginalDate:  reminder.CreatedAt,
	}

	return s.SendReminderNotification(ctx, req)
}

// buildReminderTitle 构建提醒标题
func (s *NotificationService) buildReminderTitle(reminderType ReminderType) string {
	switch reminderType {
	case ReminderType24Hours:
		return "审批提醒：任务即将到期（24小时内）"
	case ReminderType4Hours:
		return "审批提醒：任务即将到期（4小时内）"
	case ReminderType1Hour:
		return "紧急审批提醒：任务即将到期（1小时内）"
	case ReminderTypeOvertime:
		return "超时提醒：审批任务已超期"
	default:
		return "审批提醒"
	}
}

// buildReminderContent 构建提醒内容
func (s *NotificationService) buildReminderContent(reminder *ScheduledReminder) string {
	return fmt.Sprintf("您有一个待审批的任务，请及时处理。实例ID: %s", reminder.InstanceID)
}

// getReminderPriority 获取提醒优先级
func (s *NotificationService) getReminderPriority(reminderType ReminderType) NotificationPriority {
	switch reminderType {
	case ReminderType24Hours:
		return NotificationPriorityLow
	case ReminderType4Hours:
		return NotificationPriorityMedium
	case ReminderType1Hour, ReminderTypeOvertime:
		return NotificationPriorityHigh
	default:
		return NotificationPriorityMedium
	}
}

// 辅助函数
func generateNotificationID() string {
	return fmt.Sprintf("notif_%d", time.Now().UnixNano())
}

func generateReminderID() string {
	return fmt.Sprintf("remind_%d", time.Now().UnixNano())
}

// 通知相关类型定义

const (
	NotificationTypeApproval   NotificationType = "approval"
	NotificationTypeReminder   NotificationType = "reminder"
	NotificationTypeCompletion NotificationType = "completion"
)

type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityMedium NotificationPriority = "medium"
	NotificationPriorityHigh   NotificationPriority = "high"
)

type ReminderType string

const (
	ReminderType24Hours  ReminderType = "24_hours"
	ReminderType4Hours   ReminderType = "4_hours"
	ReminderType1Hour    ReminderType = "1_hour"
	ReminderTypeOvertime ReminderType = "overtime"
)

type ReminderStatus string

const (
	ReminderStatusPending   ReminderStatus = "pending"
	ReminderStatusSent      ReminderStatus = "sent"
	ReminderStatusCancelled ReminderStatus = "cancelled"
)

// Notification 通知
type Notification struct {
	ID         string                 `json:"id"`
	Type       NotificationType       `json:"type"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Recipient  uint                   `json:"recipient"`
	InstanceID string                 `json:"instance_id"`
	NodeID     string                 `json:"node_id,omitempty"`
	Priority   NotificationPriority   `json:"priority"`
	CreatedAt  time.Time              `json:"created_at"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ScheduledReminder 预定提醒
type ScheduledReminder struct {
	ID           string         `json:"id"`
	InstanceID   string         `json:"instance_id"`
	NodeID       string         `json:"node_id"`
	UserID       uint           `json:"user_id"`
	ScheduledAt  time.Time      `json:"scheduled_at"`
	ReminderType ReminderType   `json:"reminder_type"`
	Status       ReminderStatus `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	ProcessedAt  *time.Time     `json:"processed_at,omitempty"`
}

// ApprovalNotificationRequest 审批通知请求
type ApprovalNotificationRequest struct {
	InstanceID string               `json:"instance_id"`
	NodeID     string               `json:"node_id"`
	UserID     uint                 `json:"user_id"`
	Title      string               `json:"title"`
	Content    string               `json:"content"`
	Priority   NotificationPriority `json:"priority"`
	ExpiresAt  *time.Time           `json:"expires_at,omitempty"`
}

// ReminderNotificationRequest 提醒通知请求
type ReminderNotificationRequest struct {
	InstanceID    string               `json:"instance_id"`
	NodeID        string               `json:"node_id"`
	UserID        uint                 `json:"user_id"`
	Title         string               `json:"title"`
	Content       string               `json:"content"`
	Priority      NotificationPriority `json:"priority"`
	ReminderCount int                  `json:"reminder_count"`
	OriginalDate  time.Time            `json:"original_date"`
}

// CompletionNotificationRequest 完成通知请求
type CompletionNotificationRequest struct {
	InstanceID  string    `json:"instance_id"`
	InitiatorID uint      `json:"initiator_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Result      string    `json:"result"`
	CompletedAt time.Time `json:"completed_at"`
}
