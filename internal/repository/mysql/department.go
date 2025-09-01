package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"taskmanage/internal/database"
	"taskmanage/internal/repository"
)

// DepartmentRepositoryImpl 部门仓储MySQL实现
type DepartmentRepositoryImpl struct {
	db *gorm.DB
}

// NewDepartmentRepository 创建部门仓储实例
func NewDepartmentRepository(db *gorm.DB) repository.DepartmentRepository {
	return &DepartmentRepositoryImpl{db: db}
}

// Create 创建部门
func (r *DepartmentRepositoryImpl) Create(ctx context.Context, department *database.Department) error {
	return r.db.WithContext(ctx).Create(department).Error
}

// GetByID 根据ID获取部门
func (r *DepartmentRepositoryImpl) GetByID(ctx context.Context, id uint) (*database.Department, error) {
	var department database.Department
	err := r.db.WithContext(ctx).
		Preload("Manager").
		Preload("Parent").
		Preload("Children").
		First(&department, id).Error
	if err != nil {
		return nil, err
	}
	return &department, nil
}

// Update 更新部门
func (r *DepartmentRepositoryImpl) Update(ctx context.Context, department *database.Department) error {
	return r.db.WithContext(ctx).Save(department).Error
}

// Delete 删除部门
func (r *DepartmentRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&database.Department{}, id).Error
}

// List 获取部门列表
func (r *DepartmentRepositoryImpl) List(ctx context.Context, filter repository.ListFilter) ([]*database.Department, int64, error) {
	var departments []*database.Department
	var total int64

	query := r.db.WithContext(ctx).Model(&database.Department{})

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	err := query.
		Preload("Manager").
		Preload("Parent").
		Offset(offset).
		Limit(filter.PageSize).
		Order(fmt.Sprintf("%s %s", filter.Sort, filter.Order)).
		Find(&departments).Error

	return departments, total, err
}

// Exists 检查部门是否存在
func (r *DepartmentRepositoryImpl) Exists(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&database.Department{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// GetByName 根据名称获取部门
func (r *DepartmentRepositoryImpl) GetByName(ctx context.Context, name string) (*database.Department, error) {
	var department database.Department
	err := r.db.WithContext(ctx).
		Preload("Manager").
		Preload("Parent").
		Where("name = ?", name).
		First(&department).Error
	if err != nil {
		return nil, err
	}
	return &department, nil
}

// GetByParentID 根据父部门ID获取子部门
func (r *DepartmentRepositoryImpl) GetByParentID(ctx context.Context, parentID uint) ([]*database.Department, error) {
	var departments []*database.Department
	err := r.db.WithContext(ctx).
		Preload("Manager").
		Where("parent_id = ?", parentID).
		Find(&departments).Error
	return departments, err
}

// GetRootDepartments 获取根部门（无父部门）
func (r *DepartmentRepositoryImpl) GetRootDepartments(ctx context.Context) ([]*database.Department, error) {
	var departments []*database.Department
	err := r.db.WithContext(ctx).
		Preload("Manager").
		Where("parent_id IS NULL").
		Find(&departments).Error
	return departments, err
}

// GetDepartmentTree 获取部门树结构
func (r *DepartmentRepositoryImpl) GetDepartmentTree(ctx context.Context) ([]*database.Department, error) {
	var departments []*database.Department
	err := r.db.WithContext(ctx).
		Preload("Manager").
		Preload("Children").
		Find(&departments).Error
	return departments, err
}

// GetDepartmentWithManager 获取部门及其管理者信息
func (r *DepartmentRepositoryImpl) GetDepartmentWithManager(ctx context.Context, id uint) (*database.Department, error) {
	var department database.Department
	err := r.db.WithContext(ctx).
		Preload("Manager").
		Preload("Manager.User").
		First(&department, id).Error
	if err != nil {
		return nil, err
	}
	return &department, nil
}

// UpdateManager 更新部门管理者
func (r *DepartmentRepositoryImpl) UpdateManager(ctx context.Context, departmentID, managerID uint) error {
	return r.db.WithContext(ctx).
		Model(&database.Department{}).
		Where("id = ?", departmentID).
		Update("manager_id", managerID).Error
}

// GetSubDepartments 获取子部门（递归）
func (r *DepartmentRepositoryImpl) GetSubDepartments(ctx context.Context, departmentID uint) ([]*database.Department, error) {
	var departments []*database.Department
	
	// 使用递归CTE查询所有子部门
	query := `
		WITH RECURSIVE sub_departments AS (
			SELECT * FROM departments WHERE parent_id = ?
			UNION ALL
			SELECT d.* FROM departments d
			INNER JOIN sub_departments sd ON d.parent_id = sd.id
		)
		SELECT * FROM sub_departments
	`
	
	err := r.db.WithContext(ctx).Raw(query, departmentID).Scan(&departments).Error
	return departments, err
}
