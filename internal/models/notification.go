package models

import (
	"time"
	"gorm.io/gorm"
)

// TaskNotification 任务通知
type TaskNotification struct {
	ID          uint                   `gorm:"primaryKey" json:"id"`
	Type        TaskNotificationType   `gorm:"type:varchar(50);not null" json:"type"`
	Title       string                 `gorm:"type:varchar(200);not null" json:"title"`
	Content     string                 `gorm:"type:text" json:"content"`
	RecipientID uint                   `gorm:"not null;index" json:"recipient_id"`
	SenderID    *uint                  `gorm:"index" json:"sender_id,omitempty"`
	TaskID      *uint                  `gorm:"index" json:"task_id,omitempty"`
	Priority    NotificationPriority   `gorm:"type:varchar(20);default:'medium'" json:"priority"`
	Status      TaskNotificationStatus `gorm:"type:varchar(20);default:'unread'" json:"status"`
	ActionType  *string                `gorm:"type:varchar(50)" json:"action_type,omitempty"` // accept, reject, view
	ActionData  *string                `gorm:"type:json" json:"action_data,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	ReadAt      *time.Time             `json:"read_at,omitempty"`
	ActionAt    *time.Time             `json:"action_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `gorm:"index" json:"-"`
}

// TaskNotificationAction 任务通知操作记录
type TaskNotificationAction struct {
	ID             uint                     `gorm:"primaryKey" json:"id"`
	NotificationID uint                     `gorm:"not null;index" json:"notification_id"`
	UserID         uint                     `gorm:"not null;index" json:"user_id"`
	ActionType     TaskNotificationActionType `gorm:"type:varchar(50);not null" json:"action_type"`
	ActionData     *string                  `gorm:"type:json" json:"action_data,omitempty"`
	Reason         *string                  `gorm:"type:text" json:"reason,omitempty"` // 拒绝理由等
	CreatedAt      time.Time                `json:"created_at"`
}

// 枚举类型定义
type TaskNotificationType string

const (
	NotificationTypeTaskAssigned   TaskNotificationType = "task_assigned"   // 任务分配
	NotificationTypeTaskStarted    TaskNotificationType = "task_started"    // 任务开始
	NotificationTypeTaskCompleted  TaskNotificationType = "task_completed"  // 任务完成
	NotificationTypeTaskCancelled  TaskNotificationType = "task_cancelled"  // 任务取消
	NotificationTypeTaskReassigned TaskNotificationType = "task_reassigned" // 任务重新分配
	NotificationTypeTaskOverdue    TaskNotificationType = "task_overdue"    // 任务逾期
	NotificationTypeTaskReminder   TaskNotificationType = "task_reminder"   // 任务提醒
	NotificationTypeSystemMessage  TaskNotificationType = "system_message"  // 系统消息
)

type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityMedium NotificationPriority = "medium"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

type TaskNotificationStatus string

const (
	NotificationStatusUnread  TaskNotificationStatus = "unread"
	NotificationStatusRead    TaskNotificationStatus = "read"
	NotificationStatusActed   TaskNotificationStatus = "acted" // 已操作（接受/拒绝等）
	NotificationStatusExpired TaskNotificationStatus = "expired"
)

type TaskNotificationActionType string

const (
	ActionTypeRead   TaskNotificationActionType = "read"
	ActionTypeAccept TaskNotificationActionType = "accept"
	ActionTypeReject TaskNotificationActionType = "reject"
	ActionTypeView   TaskNotificationActionType = "view"
)

// TableName 指定表名
func (TaskNotification) TableName() string {
	return "task_notifications"
}

func (TaskNotificationAction) TableName() string {
	return "task_notification_actions"
}

// 辅助方法
func (n *TaskNotification) IsExpired() bool {
	return n.ExpiresAt != nil && time.Now().After(*n.ExpiresAt)
}

func (n *TaskNotification) CanTakeAction() bool {
	return n.Status == NotificationStatusUnread || n.Status == NotificationStatusRead
}

func (n *TaskNotification) RequiresAction() bool {
	return n.ActionType != nil && n.CanTakeAction()
}
