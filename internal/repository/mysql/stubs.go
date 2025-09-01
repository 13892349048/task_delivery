package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"gorm.io/gorm"
)

// 临时存根实现，后续需要完整实现

type RoleRepositoryImpl struct {
	*BaseRepositoryImpl[database.Role]
}

type PermissionRepositoryImpl struct {
	*BaseRepositoryImpl[database.Permission]
}

// EmployeeRepositoryImpl 和 SkillRepositoryImpl 已移动到独立文件

// type TaskRepositoryImpl struct {
// 	*BaseRepositoryImpl[database.Task]
// }

// type AssignmentRepositoryImpl struct {
// 	*BaseRepositoryImpl[database.Assignment]
// }

type NotificationRepositoryImpl struct {
	*BaseRepositoryImpl[database.TaskNotification]
}

type AuditLogRepositoryImpl struct {
	*BaseRepositoryImpl[database.AuditLog]
}

type SystemConfigRepositoryImpl struct {
	*BaseRepositoryImpl[database.SystemConfig]
}

// 实现接口方法的存根 - 这些需要后续完整实现

// RoleRepositoryImpl 方法存根
func (r *RoleRepositoryImpl) GetByName(ctx context.Context, name string) (*database.Role, error) {
	var role database.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepositoryImpl) GetRoleWithPermissions(ctx context.Context, roleID uint) (*database.Role, error) {
	var role database.Role
	err := r.db.Preload("Permissions").First(&role, roleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepositoryImpl) AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	// 获取角色
	var role database.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	// 获取权限
	var permissions []database.Permission
	if err := r.db.Find(&permissions, permissionIDs).Error; err != nil {
		return err
	}

	// 先清除现有权限关联，再重新分配（确保权限是最新的）
	if err := r.db.Model(&role).Association("Permissions").Clear(); err != nil {
		return err
	}

	// 分配权限（使用GORM的Association方法）
	return r.db.Model(&role).Association("Permissions").Append(&permissions)
}

func (r *RoleRepositoryImpl) RemovePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	// 获取角色
	var role database.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	// 获取权限
	var permissions []database.Permission
	if err := r.db.Find(&permissions, permissionIDs).Error; err != nil {
		return err
	}

	// 移除权限（使用GORM的Association方法）
	return r.db.Model(&role).Association("Permissions").Delete(&permissions)
}

// PermissionRepositoryImpl 方法存根
func (r *PermissionRepositoryImpl) GetByResource(ctx context.Context, resource string) ([]*database.Permission, error) {
	// TODO: 实现
	return nil, nil
}

func (r *PermissionRepositoryImpl) GetUserPermissions(ctx context.Context, userID uint) ([]*database.Permission, error) {
	var permissions []*database.Permission
	
	// 通过用户角色获取权限
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND permissions.deleted_at IS NULL", userID).
		Find(&permissions).Error
	
	if err != nil {
		return nil, err
	}
	
	return permissions, nil
}

func (r *PermissionRepositoryImpl) GetByName(ctx context.Context, name string) (*database.Permission, error) {
	var permission database.Permission
	err := r.db.Where("name = ?", name).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &permission, nil
}

// Employee 和 Skill 相关方法已移动到独立文件实现

// TaskRepositoryImpl 方法存根
// func (r *TaskRepositoryImpl) AssignTask(ctx context.Context, taskID uint, employeeID uint) error {
// 	// TODO: 实现
// 	return nil
// }

// func (r *TaskRepositoryImpl) GetByAssignee(ctx context.Context, assigneeID uint, status string) ([]*database.Task, error) {
// 	// TODO: 实现
// 	return nil, nil
// }

// func (r *TaskRepositoryImpl) GetByCreator(ctx context.Context, creatorID uint) ([]*database.Task, error) {
// 	// TODO: 实现
// 	return nil, nil
// }

// func (r *TaskRepositoryImpl) GetByStatus(ctx context.Context, status string) ([]*database.Task, error) {
// 	// TODO: 实现
// 	return nil, nil
// }

// func (r *TaskRepositoryImpl) GetOverdueTasks(ctx context.Context) ([]*database.Task, error) {
// 	// TODO: 实现
// 	return nil, nil
// }

// func (r *TaskRepositoryImpl) GetTaskWithDetails(ctx context.Context, taskID uint) (*database.Task, error) {
// 	// TODO: 实现
// 	return nil, repository.ErrNotFound
// }

// func (r *TaskRepositoryImpl) UpdateStatus(ctx context.Context, taskID uint, status string) error {
// 	// TODO: 实现
// 	return nil
// }

// func (r *TaskRepositoryImpl) GetTasksByPriority(ctx context.Context, priority string) ([]*database.Task, error) {
// 	// TODO: 实现
// 	return nil, nil
// }

// func (r *TaskRepositoryImpl) GetTasksInDateRange(ctx context.Context, start, end time.Time) ([]*database.Task, error) {
// 	// TODO: 实现
// 	return nil, nil
// }

// AssignmentRepositoryImpl 方法存根
// func (r *AssignmentRepositoryImpl) ApproveAssignment(ctx context.Context, assignmentID uint, approverID uint, comment string) error {
// 	// TODO: 实现
// 	return nil
// }

// func (r *AssignmentRepositoryImpl) RejectAssignment(ctx context.Context, assignmentID uint, approverID uint, reason string) error {
// 	// TODO: 实现
// 	return nil
// }

// func (r *AssignmentRepositoryImpl) GetByTask(ctx context.Context, taskID uint) ([]*database.Assignment, error) {
// 	// TODO: 实现
// 	return nil, nil
// }

// func (r *AssignmentRepositoryImpl) GetByAssignee(ctx context.Context, assigneeID uint, status string) ([]*database.Assignment, error) {
// 	// TODO: 实现
// 	return nil, nil
// }

// func (r *AssignmentRepositoryImpl) GetPendingAssignments(ctx context.Context) ([]*database.Assignment, error) {
// 	// TODO: 实现
// 	return nil, nil
// }

func (r *AssignmentRepositoryImpl) CompleteAssignment(ctx context.Context, assignmentID uint, completedBy uint) error {
	// TODO: 实现
	return nil
}

// NotificationRepositoryImpl 方法实现
func (n *NotificationRepositoryImpl) GetUserNotifications(ctx context.Context, userID uint, status string, page, pageSize int) ([]*database.TaskNotification, int64, error) {
	query := n.db.WithContext(ctx).Where("recipient_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 只显示未过期的通知
	query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())

	// 计算总数
	var total int64
	if err := query.Model(&database.TaskNotification{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var notifications []*database.TaskNotification
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (n *NotificationRepositoryImpl) MarkAsRead(ctx context.Context, notificationID, userID uint) error {
	now := time.Now()
	result := n.db.WithContext(ctx).Model(&database.TaskNotification{}).
		Where("id = ? AND recipient_id = ? AND status = ?", notificationID, userID, "unread").
		Updates(map[string]interface{}{
			"status":  "read",
			"read_at": &now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (n *NotificationRepositoryImpl) MarkAllAsRead(ctx context.Context, userID uint) (int64, error) {
	now := time.Now()
	result := n.db.WithContext(ctx).Model(&database.TaskNotification{}).
		Where("recipient_id = ? AND status = ?", userID, "unread").
		Where("expires_at IS NULL OR expires_at > ?", now).
		Updates(map[string]interface{}{
			"status":  "read",
			"read_at": &now,
		})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (n *NotificationRepositoryImpl) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := n.db.WithContext(ctx).Model(&database.TaskNotification{}).
		Where("recipient_id = ? AND status = ?", userID, "unread").
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Count(&count).Error

	return count, err
}

func (n *NotificationRepositoryImpl) CreateTaskAssignmentNotification(ctx context.Context, taskID, recipientID, senderID uint) error {
	// 获取任务信息
	var task database.Task
	if err := n.db.WithContext(ctx).First(&task, taskID).Error; err != nil {
		return err
	}

	// 创建通知
	notification := &database.TaskNotification{
		Type:        "task_assigned",
		Title:       "新任务分配",
		Content:     "您收到了新任务：" + task.Title,
		RecipientID: recipientID,
		SenderID:    &senderID,
		TaskID:      &taskID,
		Priority:    "medium",
		Status:      "unread",
		ActionType:  stringPtr("accept_reject"),
		ExpiresAt:   timePtr(time.Now().Add(24 * time.Hour)),
	}

	return n.db.WithContext(ctx).Create(notification).Error
}

func (n *NotificationRepositoryImpl) UpdateNotificationStatus(ctx context.Context, notificationID, userID uint, status string) error {
	now := time.Now()
	result := n.db.WithContext(ctx).Model(&database.TaskNotification{}).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":    status,
			"action_at": &now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// AcceptTaskNotification 接受任务通知
func (n *NotificationRepositoryImpl) AcceptTaskNotification(ctx context.Context, notificationID, taskID, userID uint, reason *string) error {
	// 开始事务
	tx := n.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新通知状态
	now := time.Now()
	if err := tx.WithContext(ctx).Model(&database.TaskNotification{}).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":    "acted",
			"action_at": &now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新任务状态为进行中
	if err := tx.WithContext(ctx).Model(&database.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":     "in_progress",
		"started_at": &now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 记录操作
	action := &database.TaskNotificationAction{
		NotificationID: notificationID,
		UserID:         userID,
		ActionType:     "accept",
		Reason:         reason,
	}
	if err := tx.WithContext(ctx).Create(action).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// RejectTaskNotification 拒绝任务通知
func (n *NotificationRepositoryImpl) RejectTaskNotification(ctx context.Context, notificationID, taskID, userID uint, reason *string) error {
	if reason == nil || *reason == "" {
		return fmt.Errorf("拒绝任务必须提供理由")
	}

	// 开始事务
	tx := n.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新通知状态
	now := time.Now()
	if err := tx.WithContext(ctx).Model(&database.TaskNotification{}).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":    "acted",
			"action_at": &now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新任务状态为待分配
	if err := tx.WithContext(ctx).Model(&database.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":      "pending",
		"assignee_id": nil,
		"assigned_at": nil,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 记录操作
	action := &database.TaskNotificationAction{
		NotificationID: notificationID,
		UserID:         userID,
		ActionType:     "reject",
		Reason:         reason,
	}
	if err := tx.WithContext(ctx).Create(action).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// AuditLogRepositoryImpl 方法存根
func (r *AuditLogRepositoryImpl) GetByUserID(ctx context.Context, userID uint) ([]*database.AuditLog, error) {
	// TODO: 实现
	return nil, nil
}

func (r *AuditLogRepositoryImpl) GetByUser(ctx context.Context, userID uint, start, end time.Time) ([]*database.AuditLog, error) {
	// TODO: 实现
	return nil, nil
}

func (r *AuditLogRepositoryImpl) GetByAction(ctx context.Context, action string, start, end time.Time) ([]*database.AuditLog, error) {
	// TODO: 实现
	return nil, nil
}

func (r *AuditLogRepositoryImpl) GetByResource(ctx context.Context, resource string, resourceID uint) ([]*database.AuditLog, error) {
	// TODO: 实现
	return nil, nil
}

func (r *AuditLogRepositoryImpl) CleanupOldLogs(ctx context.Context, before time.Time) error {
	// TODO: 实现
	return nil
}

// SystemConfigRepositoryImpl 方法存根
func (r *SystemConfigRepositoryImpl) GetByKey(ctx context.Context, key string) (*database.SystemConfig, error) {
	// TODO: 实现
	return nil, repository.ErrNotFound
}

func (r *SystemConfigRepositoryImpl) SetValue(ctx context.Context, key, value string) error {
	// TODO: 实现
	return nil
}

func (r *SystemConfigRepositoryImpl) BatchSet(ctx context.Context, configs map[string]string) error {
	// TODO: 实现
	return nil
}

func (r *SystemConfigRepositoryImpl) GetByCategory(ctx context.Context, category string) ([]*database.SystemConfig, error) {
	// TODO: 实现
	return nil, nil
}

func (r *SystemConfigRepositoryImpl) GetPublicConfigs(ctx context.Context) ([]*database.SystemConfig, error) {
	// TODO: 实现
	return nil, nil
}
