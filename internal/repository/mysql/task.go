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

// TaskRepositoryImpl 任务仓储实现
type TaskRepositoryImpl struct {
	*BaseRepositoryImpl[database.Task]
}

// NewTaskRepository 创建任务仓储实例
func NewTaskRepository(db *gorm.DB) repository.TaskRepository {
	return &TaskRepositoryImpl{
		BaseRepositoryImpl: NewBaseRepository[database.Task](db),
	}
}

// GetByStatus 根据状态获取任务列表
func (r *TaskRepositoryImpl) GetByStatus(ctx context.Context, status string) ([]*database.Task, error) {
	var tasks []*database.Task
	if err := r.db.WithContext(ctx).Where("status = ?", status).Find(&tasks).Error; err != nil {
		logger.Errorf("根据状态查询任务失败: %v", err)
		return nil, fmt.Errorf("根据状态查询任务失败: %w", err)
	}
	return tasks, nil
}

// GetByAssignee 根据分配人和状态获取任务列表
func (r *TaskRepositoryImpl) GetByAssignee(ctx context.Context, assigneeID uint, status string) ([]*database.Task, error) {
	var tasks []*database.Task
	query := r.db.WithContext(ctx).Joins("JOIN assignments ON tasks.id = assignments.task_id").
		Where("assignments.assignee_id = ?", assigneeID)
	
	if status != "" {
		query = query.Where("assignments.status = ?", status)
	}
	
	if err := query.Find(&tasks).Error; err != nil {
		logger.Errorf("根据分配人查询任务失败: %v", err)
		return nil, fmt.Errorf("根据分配人查询任务失败: %w", err)
	}
	return tasks, nil
}

// GetByCreator 根据创建者获取任务列表
func (r *TaskRepositoryImpl) GetByCreator(ctx context.Context, creatorID uint) ([]*database.Task, error) {
	var tasks []*database.Task
	if err := r.db.WithContext(ctx).Where("created_by = ?", creatorID).Find(&tasks).Error; err != nil {
		logger.Errorf("根据创建者查询任务失败: %v", err)
		return nil, fmt.Errorf("根据创建者查询任务失败: %w", err)
	}
	return tasks, nil
}

// GetOverdueTasks 获取逾期任务列表
func (r *TaskRepositoryImpl) GetOverdueTasks(ctx context.Context) ([]*database.Task, error) {
	var tasks []*database.Task
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Where("due_date < ? AND status NOT IN (?)", now, []string{"completed", "cancelled"}).
		Find(&tasks).Error; err != nil {
		logger.Errorf("查询逾期任务失败: %v", err)
		return nil, fmt.Errorf("查询逾期任务失败: %w", err)
	}
	return tasks, nil
}

// GetTaskWithDetails 获取任务详情（包含关联信息）
func (r *TaskRepositoryImpl) GetTaskWithDetails(ctx context.Context, taskID uint) (*database.Task, error) {
	var task database.Task
	if err := r.db.WithContext(ctx).
		Preload("Assignments").
		Preload("Assignments.Assignee").
		First(&task, taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		logger.Errorf("查询任务详情失败: %v", err)
		return nil, fmt.Errorf("查询任务详情失败: %w", err)
	}
	return &task, nil
}

// UpdateStatus 更新任务状态
func (r *TaskRepositoryImpl) UpdateStatus(ctx context.Context, taskID uint, status string) error {
	result := r.db.WithContext(ctx).Model(&database.Task{}).
		Where("id = ?", taskID).
		Update("status", status)
	
	if result.Error != nil {
		logger.Errorf("更新任务状态失败: %v", result.Error)
		return fmt.Errorf("更新任务状态失败: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	
	return nil
}

// AssignTask 分配任务给员工
func (r *TaskRepositoryImpl) AssignTask(ctx context.Context, taskID, assigneeID uint) error {
	// 创建分配记录
	assignment := &database.Assignment{
		TaskID:     taskID,
		AssigneeID: assigneeID,
		Status:     "pending",
		AssignedAt: time.Now(),
	}
	
	if err := r.db.WithContext(ctx).Create(assignment).Error; err != nil {
		logger.Errorf("分配任务失败: %v", err)
		return fmt.Errorf("分配任务失败: %w", err)
	}
	
	// 更新任务状态为已分配
	if err := r.UpdateStatus(ctx, taskID, "assigned"); err != nil {
		logger.Errorf("更新任务分配状态失败: %v", err)
		return fmt.Errorf("更新任务分配状态失败: %w", err)
	}
	
	return nil
}

// GetTasksByPriority 根据优先级获取任务列表
func (r *TaskRepositoryImpl) GetTasksByPriority(ctx context.Context, priority string) ([]*database.Task, error) {
	var tasks []*database.Task
	if err := r.db.WithContext(ctx).Where("priority = ?", priority).Find(&tasks).Error; err != nil {
		logger.Errorf("根据优先级查询任务失败: %v", err)
		return nil, fmt.Errorf("根据优先级查询任务失败: %w", err)
	}
	return tasks, nil
}

// GetTasksInDateRange 获取指定日期范围内的任务
func (r *TaskRepositoryImpl) GetTasksInDateRange(ctx context.Context, start, end time.Time) ([]*database.Task, error) {
	var tasks []*database.Task
	if err := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", start, end).
		Find(&tasks).Error; err != nil {
		logger.Errorf("根据日期范围查询任务失败: %v", err)
		return nil, fmt.Errorf("根据日期范围查询任务失败: %w", err)
	}
	return tasks, nil
}

// GetActiveTasksByEmployee 获取员工的活跃任务
func (r *TaskRepositoryImpl) GetActiveTasksByEmployee(ctx context.Context, employeeID uint) ([]*database.Task, error) {
	var tasks []*database.Task
	if err := r.db.WithContext(ctx).
		Joins("JOIN assignments ON tasks.id = assignments.task_id").
		Where("assignments.assignee_id = ? AND assignments.status IN (?)", employeeID, []string{"approved", "pending"}).
		Where("tasks.status NOT IN (?)", []string{"completed", "cancelled"}).
		Find(&tasks).Error; err != nil {
		logger.Errorf("获取员工活跃任务失败: %v", err)
		return nil, fmt.Errorf("获取员工活跃任务失败: %w", err)
	}
	return tasks, nil
}

// UpdateAssignee 更新任务分配人
func (r *TaskRepositoryImpl) UpdateAssignee(ctx context.Context, taskID, assigneeID uint) error {
	result := r.db.WithContext(ctx).Model(&database.Task{}).
		Where("id = ?", taskID).
		Update("assignee_id", assigneeID)
	
	if result.Error != nil {
		logger.Errorf("更新任务分配人失败: %v", result.Error)
		return fmt.Errorf("更新任务分配人失败: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	
	return nil
}
