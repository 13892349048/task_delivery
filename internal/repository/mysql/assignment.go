package mysql

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// AssignmentRepositoryImpl 任务分配仓储实现
type AssignmentRepositoryImpl struct {
	*BaseRepositoryImpl[database.Assignment]
}

// NewAssignmentRepository 创建任务分配仓储实例
func NewAssignmentRepository(db *gorm.DB) repository.AssignmentRepository {
	return &AssignmentRepositoryImpl{
		BaseRepositoryImpl: NewBaseRepository[database.Assignment](db),
	}
}

// GetByTask 根据任务ID获取分配记录
func (r *AssignmentRepositoryImpl) GetByTask(ctx context.Context, taskID uint) ([]*database.Assignment, error) {
	var assignments []*database.Assignment
	if err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Preload("Assignee").
		Preload("Assigner").
		Preload("Approver").
		Find(&assignments).Error; err != nil {
		logger.Errorf("根据任务ID查询分配记录失败: %v", err)
		return nil, fmt.Errorf("根据任务ID查询分配记录失败: %w", err)
	}
	return assignments, nil
}

// GetByTaskID 根据任务ID获取分配记录（别名方法）
func (r *AssignmentRepositoryImpl) GetByTaskID(ctx context.Context, taskID uint) ([]*database.Assignment, error) {
	return r.GetByTask(ctx, taskID)
}

// GetActiveByTaskID 根据任务ID获取活跃的分配记录
func (r *AssignmentRepositoryImpl) GetActiveByTaskID(ctx context.Context, taskID uint) (*database.Assignment, error) {
	var assignment database.Assignment
	if err := r.db.WithContext(ctx).
		Where("task_id = ? AND status IN (?)", taskID, []string{"pending", "approved"}).
		Preload("Assignee").
		Preload("Assigner").
		Preload("Approver").
		First(&assignment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		logger.Errorf("根据任务ID查询活跃分配记录失败: %v", err)
		return nil, fmt.Errorf("根据任务ID查询活跃分配记录失败: %w", err)
	}
	return &assignment, nil
}

// GetByAssignee 根据分配人获取分配记录
func (r *AssignmentRepositoryImpl) GetByAssignee(ctx context.Context, assigneeID uint, status string) ([]*database.Assignment, error) {
	var assignments []*database.Assignment
	query := r.db.WithContext(ctx).Where("assignee_id = ?", assigneeID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.
		Preload("Task").
		Preload("Assigner").
		Preload("Approver").
		Find(&assignments).Error; err != nil {
		logger.Errorf("根据分配人查询分配记录失败: %v", err)
		return nil, fmt.Errorf("根据分配人查询分配记录失败: %w", err)
	}
	return assignments, nil
}

// GetPendingAssignments 获取待审批的分配记录
func (r *AssignmentRepositoryImpl) GetPendingAssignments(ctx context.Context) ([]*database.Assignment, error) {
	var assignments []*database.Assignment
	if err := r.db.WithContext(ctx).
		Where("status = ?", "pending").
		Preload("Task").
		Preload("Assignee").
		Preload("Assigner").
		Find(&assignments).Error; err != nil {
		logger.Errorf("查询待审批分配记录失败: %v", err)
		return nil, fmt.Errorf("查询待审批分配记录失败: %w", err)
	}
	return assignments, nil
}

// ApproveAssignment 审批通过分配
func (r *AssignmentRepositoryImpl) ApproveAssignment(ctx context.Context, assignmentID, approverID uint, reason string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&database.Assignment{}).
		Where("id = ?", assignmentID).
		Updates(map[string]interface{}{
			"status":      "approved",
			"approver_id": approverID,
			"approved_at": &now,
			"reason":      reason,
		})
	
	if result.Error != nil {
		logger.Errorf("审批通过分配失败: %v", result.Error)
		return fmt.Errorf("审批通过分配失败: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	
	return nil
}

// RejectAssignment 审批拒绝分配
func (r *AssignmentRepositoryImpl) RejectAssignment(ctx context.Context, assignmentID, approverID uint, reason string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&database.Assignment{}).
		Where("id = ?", assignmentID).
		Updates(map[string]interface{}{
			"status":      "rejected",
			"approver_id": approverID,
			"approved_at": &now,
			"reason":      reason,
		})
	
	if result.Error != nil {
		logger.Errorf("审批拒绝分配失败: %v", result.Error)
		return fmt.Errorf("审批拒绝分配失败: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	
	return nil
}

// GetAssignmentHistory 获取分配历史记录
func (r *AssignmentRepositoryImpl) GetAssignmentHistory(ctx context.Context, taskID uint) ([]*database.Assignment, error) {
	var assignments []*database.Assignment
	if err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("created_at DESC").
		Preload("Assignee").
		Preload("Assigner").
		Preload("Approver").
		Find(&assignments).Error; err != nil {
		logger.Errorf("查询分配历史记录失败: %v", err)
		return nil, fmt.Errorf("查询分配历史记录失败: %w", err)
	}
	return assignments, nil
}
