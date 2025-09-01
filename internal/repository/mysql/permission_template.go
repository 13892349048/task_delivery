package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
)

type permissionTemplateRepository struct {
	db *gorm.DB
}

func NewPermissionTemplateRepository(db *gorm.DB) repository.PermissionTemplateRepository {
	return &permissionTemplateRepository{db: db}
}

func (r *permissionTemplateRepository) Create(ctx context.Context, template *database.PermissionTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

func (r *permissionTemplateRepository) GetByID(ctx context.Context, id uint) (*database.PermissionTemplate, error) {
	var template database.PermissionTemplate
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Preload("Rules").
		First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *permissionTemplateRepository) GetByCode(ctx context.Context, code string) (*database.PermissionTemplate, error) {
	var template database.PermissionTemplate
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Preload("Rules").
		Where("code = ?", code).
		First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *permissionTemplateRepository) List(ctx context.Context, filter *repository.PermissionTemplateFilter) ([]*database.PermissionTemplate, error) {
	var templates []*database.PermissionTemplate
	query := r.db.WithContext(ctx).Preload("Permissions").Preload("Rules")

	// 应用过滤条件
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.Level != nil {
		query = query.Where("level = ?", *filter.Level)
	}
	if filter.DepartmentID != nil {
		query = query.Where("department_id = ? OR department_id IS NULL", *filter.DepartmentID)
	}
	if filter.PositionID != nil {
		query = query.Where("position_id = ? OR position_id IS NULL", *filter.PositionID)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	// 分页
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("level ASC, created_at DESC").Find(&templates).Error
	return templates, err
}

func (r *permissionTemplateRepository) Update(ctx context.Context, template *database.PermissionTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

func (r *permissionTemplateRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&database.PermissionTemplate{}, id).Error
}

func (r *permissionTemplateRepository) GetByCategory(ctx context.Context, category string) ([]*database.PermissionTemplate, error) {
	var templates []*database.PermissionTemplate
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Where("category = ? AND is_active = ?", category, true).
		Order("level ASC").
		Find(&templates).Error
	return templates, err
}

func (r *permissionTemplateRepository) GetByDepartmentAndPosition(ctx context.Context, departmentID, positionID *uint) ([]*database.PermissionTemplate, error) {
	var templates []*database.PermissionTemplate
	query := r.db.WithContext(ctx).Preload("Permissions").Where("is_active = ?", true)

	// 构建复杂的查询条件
	if departmentID != nil && positionID != nil {
		query = query.Where("(department_id = ? AND position_id = ?) OR (department_id = ? AND position_id IS NULL) OR (department_id IS NULL AND position_id = ?) OR (department_id IS NULL AND position_id IS NULL)",
			*departmentID, *positionID, *departmentID, *positionID)
	} else if departmentID != nil {
		query = query.Where("(department_id = ? AND position_id IS NULL) OR (department_id IS NULL)", *departmentID)
	} else if positionID != nil {
		query = query.Where("(position_id = ? AND department_id IS NULL) OR (department_id IS NULL AND position_id IS NULL)", *positionID)
	} else {
		query = query.Where("department_id IS NULL AND position_id IS NULL")
	}

	err := query.Order("level ASC").Find(&templates).Error
	return templates, err
}
