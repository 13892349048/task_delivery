package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// SkillRepositoryImpl 技能Repository实现
type SkillRepositoryImpl struct {
	*BaseRepositoryImpl[database.Skill]
}

// NewSkillRepository 创建技能Repository实例
func NewSkillRepository(db *gorm.DB) repository.SkillRepository {
	return &SkillRepositoryImpl{
		BaseRepositoryImpl: NewBaseRepository[database.Skill](db),
	}
}

// GetByName 根据技能名称获取技能
func (r *SkillRepositoryImpl) GetByName(ctx context.Context, name string) (*database.Skill, error) {
	var skill database.Skill
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&skill).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrNotFound
		}
		logger.Errorf("根据名称获取技能失败: %v", err)
		return nil, fmt.Errorf("根据名称获取技能失败: %w", err)
	}
	return &skill, nil
}

// GetByCategory 根据分类获取技能列表
func (r *SkillRepositoryImpl) GetByCategory(ctx context.Context, category string) ([]*database.Skill, error) {
	var skills []*database.Skill
	query := r.db.WithContext(ctx)
	
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	err := query.Find(&skills).Error
	if err != nil {
		logger.Errorf("根据分类获取技能列表失败: %v", err)
		return nil, fmt.Errorf("根据分类获取技能列表失败: %w", err)
	}
	
	return skills, nil
}

// AssignToEmployee 为员工分配技能
func (r *SkillRepositoryImpl) AssignToEmployee(ctx context.Context, employeeID uint, skillID uint, level int) error {
	// 检查员工是否存在
	var employee database.Employee
	err := r.db.WithContext(ctx).Where("id = ?", employeeID).First(&employee).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return repository.ErrNotFound
		}
		return fmt.Errorf("获取员工信息失败: %w", err)
	}
	
	// 检查技能是否存在
	var skill database.Skill
	err = r.db.WithContext(ctx).Where("id = ?", skillID).First(&skill).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return repository.ErrNotFound
		}
		return fmt.Errorf("获取技能信息失败: %w", err)
	}
	
	// 使用原生SQL插入或更新员工技能关联
	// 如果已存在则更新等级，否则插入新记录
	err = r.db.WithContext(ctx).Exec(`
		INSERT INTO employee_skills (employee_id, skill_id, level, created_at, updated_at) 
		VALUES (?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE 
		level = VALUES(level), 
		updated_at = NOW()
	`, employeeID, skillID, level).Error
	
	if err != nil {
		logger.Errorf("为员工分配技能失败: %v", err)
		return fmt.Errorf("为员工分配技能失败: %w", err)
	}
	
	logger.Infof("为员工分配技能成功: EmployeeID=%d, SkillID=%d, Level=%d", employeeID, skillID, level)
	return nil
}

// RemoveFromEmployee 移除员工的技能
func (r *SkillRepositoryImpl) RemoveFromEmployee(ctx context.Context, employeeID uint, skillID uint) error {
	// 删除员工技能关联
	result := r.db.WithContext(ctx).Exec(`
		DELETE FROM employee_skills 
		WHERE employee_id = ? AND skill_id = ?
	`, employeeID, skillID)
	
	if result.Error != nil {
		logger.Errorf("移除员工技能失败: %v", result.Error)
		return fmt.Errorf("移除员工技能失败: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	
	logger.Infof("移除员工技能成功: EmployeeID=%d, SkillID=%d", employeeID, skillID)
	return nil
}

// GetEmployeeSkills 获取员工的所有技能
func (r *SkillRepositoryImpl) GetEmployeeSkills(ctx context.Context, employeeID uint) ([]*database.Skill, error) {
	var skills []*database.Skill
	err := r.db.WithContext(ctx).
		Joins("JOIN employee_skills es ON skills.id = es.skill_id").
		Where("es.employee_id = ?", employeeID).
		Find(&skills).Error
	
	if err != nil {
		logger.Errorf("获取员工技能列表失败: %v", err)
		return nil, fmt.Errorf("获取员工技能列表失败: %w", err)
	}
	
	return skills, nil
}

// GetEmployeeSkillLevel 获取员工某项技能的等级
func (r *SkillRepositoryImpl) GetEmployeeSkillLevel(ctx context.Context, employeeID uint, skillID uint) (int, error) {
	var level int
	err := r.db.WithContext(ctx).
		Table("employee_skills").
		Select("level").
		Where("employee_id = ? AND skill_id = ?", employeeID, skillID).
		Scan(&level).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, repository.ErrNotFound
		}
		logger.Errorf("获取员工技能等级失败: %v", err)
		return 0, fmt.Errorf("获取员工技能等级失败: %w", err)
	}
	
	return level, nil
}

// GetSkillsByEmployee 获取拥有某项技能的员工列表
func (r *SkillRepositoryImpl) GetSkillsByEmployee(ctx context.Context, skillID uint, minLevel int) ([]*database.Employee, error) {
	var employees []*database.Employee
	err := r.db.WithContext(ctx).
		Table("employees").
		Joins("JOIN employee_skills es ON employees.id = es.employee_id").
		Joins("JOIN users ON employees.user_id = users.id").
		Where("es.skill_id = ? AND es.level >= ?", skillID, minLevel).
		Preload("User").
		Find(&employees).Error
	
	if err != nil {
		logger.Errorf("获取拥有技能的员工列表失败: %v", err)
		return nil, fmt.Errorf("获取拥有技能的员工列表失败: %w", err)
	}
	
	return employees, nil
}

// GetAllCategories 获取所有技能分类
func (r *SkillRepositoryImpl) GetAllCategories(ctx context.Context) ([]string, error) {
	var categories []string
	err := r.db.WithContext(ctx).
		Model(&database.Skill{}).
		Distinct("category").
		Where("category IS NOT NULL AND category != ''").
		Pluck("category", &categories).Error
	
	if err != nil {
		logger.Errorf("获取技能分类列表失败: %v", err)
		return nil, fmt.Errorf("获取技能分类列表失败: %w", err)
	}
	
	return categories, nil
}

// BatchAssignSkills 批量为员工分配技能
func (r *SkillRepositoryImpl) BatchAssignSkills(ctx context.Context, employeeID uint, skillLevels map[uint]int) error {
	if len(skillLevels) == 0 {
		return nil
	}
	
	// 开始事务
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	for skillID, level := range skillLevels {
		err := tx.Exec(`
			INSERT INTO employee_skills (employee_id, skill_id, level, created_at, updated_at) 
			VALUES (?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE 
			level = VALUES(level), 
			updated_at = NOW()
		`, employeeID, skillID, level).Error
		
		if err != nil {
			tx.Rollback()
			logger.Errorf("批量分配技能失败: %v", err)
			return fmt.Errorf("批量分配技能失败: %w", err)
		}
	}
	
	if err := tx.Commit().Error; err != nil {
		logger.Errorf("批量分配技能事务提交失败: %v", err)
		return fmt.Errorf("批量分配技能事务提交失败: %w", err)
	}
	
	logger.Infof("批量分配技能成功: EmployeeID=%d, 技能数量=%d", employeeID, len(skillLevels))
	return nil
}
