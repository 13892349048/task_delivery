package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// EmployeeRepositoryImpl 员工Repository实现
type EmployeeRepositoryImpl struct {
	*BaseRepositoryImpl[database.Employee]
}

// NewEmployeeRepository 创建员工Repository实例
func NewEmployeeRepository(db *gorm.DB) repository.EmployeeRepository {
	return &EmployeeRepositoryImpl{
		BaseRepositoryImpl: NewBaseRepository[database.Employee](db),
	}
}

// GetByUserID 根据用户ID获取员工信息
func (r *EmployeeRepositoryImpl) GetByUserID(ctx context.Context, userID uint) (*database.Employee, error) {
	var employee database.Employee
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&employee).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		logger.Errorf("根据用户ID获取员工失败: %v", err)
		return nil, fmt.Errorf("根据用户ID获取员工失败: %w", err)
	}
	return &employee, nil
}

// GetByEmployeeNo 根据员工编号获取员工信息
func (r *EmployeeRepositoryImpl) GetByEmployeeNo(ctx context.Context, employeeNo string) (*database.Employee, error) {
	var employee database.Employee
	err := r.db.WithContext(ctx).Where("employee_no = ?", employeeNo).First(&employee).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		logger.Errorf("根据员工编号获取员工失败: %v", err)
		return nil, fmt.Errorf("根据员工编号获取员工失败: %w", err)
	}
	return &employee, nil
}

// GetAvailableEmployees 获取可用的员工列表
func (r *EmployeeRepositoryImpl) GetAvailableEmployees(ctx context.Context) ([]*database.Employee, error) {
	var employees []*database.Employee
	err := r.db.WithContext(ctx).
		Where("status = ?", "available").
		Where("current_tasks < max_tasks").
		Preload("User").
		Find(&employees).Error
	
	if err != nil {
		logger.Errorf("获取可用员工列表失败: %v", err)
		return nil, fmt.Errorf("获取可用员工列表失败: %w", err)
	}
	
	return employees, nil
}

// UpdateTaskCount 更新员工任务数量
func (r *EmployeeRepositoryImpl) UpdateTaskCount(ctx context.Context, employeeID uint, delta int) error {
	result := r.db.WithContext(ctx).
		Model(&database.Employee{}).
		Where("id = ?", employeeID).
		Update("current_tasks", gorm.Expr("current_tasks + ?", delta))
	
	if result.Error != nil {
		logger.Errorf("更新员工任务数量失败: %v", result.Error)
		return fmt.Errorf("更新员工任务数量失败: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	
	return nil
}

// GetEmployeeWithSkills 获取员工及其技能信息
func (r *EmployeeRepositoryImpl) GetEmployeeWithSkills(ctx context.Context, employeeID uint) (*database.Employee, error) {
	var employee database.Employee
	err := r.db.WithContext(ctx).
		Preload("Skills").
		Preload("User").
		Where("id = ?", employeeID).
		First(&employee).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		logger.Errorf("获取员工技能信息失败: %v", err)
		return nil, fmt.Errorf("获取员工技能信息失败: %w", err)
	}
	
	return &employee, nil
}

// GetBySkills 根据技能获取员工列表
func (r *EmployeeRepositoryImpl) GetBySkills(ctx context.Context, skillIDs []uint, minLevel int) ([]*database.Employee, error) {
	if len(skillIDs) == 0 {
		return []*database.Employee{}, nil
	}
	
	var employees []*database.Employee
	
	// 构建查询：员工必须拥有所有指定的技能
	query := r.db.WithContext(ctx).
		Preload("Skills").
		Preload("User").
		Joins("JOIN employee_skills es ON employees.id = es.employee_id").
		Where("es.skill_id IN ?", skillIDs).
		Where("es.level >= ?", minLevel).
		Group("employees.id").
		Having("COUNT(DISTINCT es.skill_id) = ?", len(skillIDs))
	
	err := query.Find(&employees).Error
	if err != nil {
		logger.Errorf("根据技能获取员工列表失败: %v", err)
		return nil, fmt.Errorf("根据技能获取员工列表失败: %w", err)
	}
	
	return employees, nil
}

// GetByDepartment 根据部门获取员工列表
func (r *EmployeeRepositoryImpl) GetByDepartment(ctx context.Context, department string) ([]*database.Employee, error) {
	var employees []*database.Employee
	err := r.db.WithContext(ctx).
		Where("department = ?", department).
		Preload("User").
		Find(&employees).Error
	
	if err != nil {
		logger.Errorf("根据部门获取员工列表失败: %v", err)
		return nil, fmt.Errorf("根据部门获取员工列表失败: %w", err)
	}
	
	return employees, nil
}

// GetByStatus 根据状态获取员工列表
func (r *EmployeeRepositoryImpl) GetByStatus(ctx context.Context, status string) ([]*database.Employee, error) {
	var employees []*database.Employee
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Preload("User").
		Find(&employees).Error
	
	if err != nil {
		logger.Errorf("根据状态获取员工列表失败: %v", err)
		return nil, fmt.Errorf("根据状态获取员工列表失败: %w", err)
	}
	
	return employees, nil
}

// GetWorkloadStats 获取员工工作负载统计
func (r *EmployeeRepositoryImpl) GetWorkloadStats(ctx context.Context, employeeID uint) (map[string]interface{}, error) {
	var employee database.Employee
	err := r.db.WithContext(ctx).Where("id = ?", employeeID).First(&employee).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("获取员工信息失败: %w", err)
	}
	
	stats := map[string]interface{}{
		"employee_id":    employee.ID,
		"current_tasks":  employee.CurrentTasks,
		"max_tasks":      employee.MaxTasks,
		"utilization":    float64(employee.CurrentTasks) / float64(employee.MaxTasks) * 100,
		"available_slots": employee.MaxTasks - employee.CurrentTasks,
	}
	
	return stats, nil
}

// UpdateStatus 更新员工状态
func (r *EmployeeRepositoryImpl) UpdateStatus(ctx context.Context, employeeID uint, status string) error {
	result := r.db.WithContext(ctx).
		Model(&database.Employee{}).
		Where("id = ?", employeeID).
		Update("status", status)
	
	if result.Error != nil {
		logger.Errorf("更新员工状态失败: %v", result.Error)
		return fmt.Errorf("更新员工状态失败: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	
	return nil
}

// GetAll 获取所有员工
func (r *EmployeeRepositoryImpl) GetAll(ctx context.Context) ([]*database.Employee, error) {
	var employees []*database.Employee
	err := r.db.WithContext(ctx).
		Preload("User").
		Find(&employees).Error
	
	if err != nil {
		logger.Errorf("获取所有员工失败: %v", err)
		return nil, fmt.Errorf("获取所有员工失败: %w", err)
	}
	
	return employees, nil
}

// BatchUpdateStatus 批量更新员工状态
func (r *EmployeeRepositoryImpl) BatchUpdateStatus(ctx context.Context, employeeIDs []uint, status string) error {
	if len(employeeIDs) == 0 {
		return nil
	}
	
	result := r.db.WithContext(ctx).
		Model(&database.Employee{}).
		Where("id IN ?", employeeIDs).
		Update("status", status)
	
	if result.Error != nil {
		logger.Errorf("批量更新员工状态失败: %v", result.Error)
		return fmt.Errorf("批量更新员工状态失败: %w", result.Error)
	}
	
	logger.Infof("批量更新员工状态成功，影响行数: %d", result.RowsAffected)
	return nil
}
