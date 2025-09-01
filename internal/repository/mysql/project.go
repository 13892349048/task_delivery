package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"taskmanage/internal/database"
	"taskmanage/internal/repository"
)

// ProjectRepositoryImpl 项目仓储MySQL实现
type ProjectRepositoryImpl struct {
	db *gorm.DB
}

// NewProjectRepository 创建项目仓储实例
func NewProjectRepository(db *gorm.DB) repository.ProjectRepository {
	return &ProjectRepositoryImpl{db: db}
}

// Create 创建项目
func (r *ProjectRepositoryImpl) Create(ctx context.Context, project *database.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

// GetByID 根据ID获取项目
func (r *ProjectRepositoryImpl) GetByID(ctx context.Context, id uint) (*database.Project, error) {
	var project database.Project
	err := r.db.WithContext(ctx).
		Preload("Department").
		Preload("Manager").
		Preload("Manager.User").
		First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// Update 更新项目
func (r *ProjectRepositoryImpl) Update(ctx context.Context, project *database.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

// Delete 删除项目
func (r *ProjectRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&database.Project{}, id).Error
}

// List 获取项目列表
func (r *ProjectRepositoryImpl) List(ctx context.Context, filter repository.ListFilter) ([]*database.Project, int64, error) {
	var projects []*database.Project
	var total int64

	query := r.db.WithContext(ctx).Model(&database.Project{})

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	err := query.
		Preload("Department").
		Preload("Manager").
		Preload("Manager.User").
		Offset(offset).
		Limit(filter.PageSize).
		Order(fmt.Sprintf("%s %s", filter.Sort, filter.Order)).
		Find(&projects).Error

	return projects, total, err
}

// Exists 检查项目是否存在
func (r *ProjectRepositoryImpl) Exists(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&database.Project{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// GetByName 根据名称获取项目
func (r *ProjectRepositoryImpl) GetByName(ctx context.Context, name string) (*database.Project, error) {
	var project database.Project
	err := r.db.WithContext(ctx).
		Preload("Department").
		Preload("Manager").
		Preload("Manager.User").
		Where("name = ?", name).
		First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByDepartmentID 根据部门ID获取项目
func (r *ProjectRepositoryImpl) GetByDepartmentID(ctx context.Context, departmentID uint) ([]*database.Project, error) {
	var projects []*database.Project
	err := r.db.WithContext(ctx).
		Preload("Department").
		Preload("Manager").
		Preload("Manager.User").
		Where("department_id = ?", departmentID).
		Find(&projects).Error
	return projects, err
}

// GetByManagerID 根据管理者ID获取项目
func (r *ProjectRepositoryImpl) GetByManagerID(ctx context.Context, managerID uint) ([]*database.Project, error) {
	var projects []*database.Project
	err := r.db.WithContext(ctx).
		Preload("Department").
		Preload("Manager").
		Preload("Manager.User").
		Where("manager_id = ?", managerID).
		Find(&projects).Error
	return projects, err
}

// GetByStatus 根据状态获取项目
func (r *ProjectRepositoryImpl) GetByStatus(ctx context.Context, status string) ([]*database.Project, error) {
	var projects []*database.Project
	err := r.db.WithContext(ctx).
		Preload("Department").
		Preload("Manager").
		Preload("Manager.User").
		Where("status = ?", status).
		Find(&projects).Error
	return projects, err
}

// GetProjectWithMembers 获取项目及其成员信息
func (r *ProjectRepositoryImpl) GetProjectWithMembers(ctx context.Context, id uint) (*database.Project, error) {
	var project database.Project
	err := r.db.WithContext(ctx).
		Preload("Department").
		Preload("Manager").
		Preload("Manager.User").
		Preload("Members").
		Preload("Members.User").
		First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// AddMember 添加项目成员
func (r *ProjectRepositoryImpl) AddMember(ctx context.Context, projectID, employeeID uint) error {
	// 使用GORM的Association方法添加多对多关联
	project := &database.Project{}
	project.ID = projectID
	
	employee := &database.Employee{}
	employee.ID = employeeID
	
	return r.db.WithContext(ctx).Model(project).Association("Members").Append(employee)
}

// RemoveMember 移除项目成员
func (r *ProjectRepositoryImpl) RemoveMember(ctx context.Context, projectID, employeeID uint) error {
	// 使用GORM的Association方法删除多对多关联
	project := &database.Project{}
	project.ID = projectID
	
	employee := &database.Employee{}
	employee.ID = employeeID
	
	return r.db.WithContext(ctx).Model(project).Association("Members").Delete(employee)
}

// GetProjectMembers 获取项目成员
func (r *ProjectRepositoryImpl) GetProjectMembers(ctx context.Context, projectID uint) ([]*database.Employee, error) {
	var members []*database.Employee
	err := r.db.WithContext(ctx).
		Table("project_members").
		Select("employees.*").
		Joins("JOIN employees ON project_members.employee_id = employees.id").
		Preload("User").
		Preload("Department").
		Preload("Position").
		Where("project_members.project_id = ?", projectID).
		Find(&members).Error
	return members, err
}

// UpdateManager 更新项目管理者
func (r *ProjectRepositoryImpl) UpdateManager(ctx context.Context, projectID, managerID uint) error {
	return r.db.WithContext(ctx).
		Model(&database.Project{}).
		Where("id = ?", projectID).
		Update("manager_id", managerID).Error
}
