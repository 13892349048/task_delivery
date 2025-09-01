package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
)

type permissionRuleRepository struct {
	db *gorm.DB
}

func NewPermissionRuleRepository(db *gorm.DB) repository.PermissionRuleRepository {
	return &permissionRuleRepository{db: db}
}

func (r *permissionRuleRepository) Create(ctx context.Context, rule *database.PermissionRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *permissionRuleRepository) GetByID(ctx context.Context, id uint) (*database.PermissionRule, error) {
	var rule database.PermissionRule
	err := r.db.WithContext(ctx).
		Preload("Template").
		First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *permissionRuleRepository) List(ctx context.Context, filter *repository.PermissionRuleFilter) ([]*database.PermissionRule, error) {
	var rules []*database.PermissionRule
	query := r.db.WithContext(ctx).Preload("Template")

	// 应用过滤条件
	if filter.TemplateID != nil {
		query = query.Where("template_id = ?", *filter.TemplateID)
	}
	if filter.TriggerCondition != "" {
		query = query.Where("trigger_condition = ?", filter.TriggerCondition)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
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

	err := query.Order("priority DESC, created_at DESC").Find(&rules).Error
	return rules, err
}

func (r *permissionRuleRepository) Update(ctx context.Context, rule *database.PermissionRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *permissionRuleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&database.PermissionRule{}, id).Error
}

func (r *permissionRuleRepository) GetByTemplateID(ctx context.Context, templateID uint) ([]*database.PermissionRule, error) {
	var rules []*database.PermissionRule
	err := r.db.WithContext(ctx).
		Where("template_id = ? AND is_active = ?", templateID, true).
		Order("priority DESC").
		Find(&rules).Error
	return rules, err
}

func (r *permissionRuleRepository) GetByTriggerCondition(ctx context.Context, condition, value string) ([]*database.PermissionRule, error) {
	var rules []*database.PermissionRule
	err := r.db.WithContext(ctx).
		Preload("Template").
		Where("trigger_condition = ? AND condition_value = ? AND is_active = ?", condition, value, true).
		Order("priority DESC").
		Find(&rules).Error
	return rules, err
}

func (r *permissionRuleRepository) GetActiveRules(ctx context.Context) ([]*database.PermissionRule, error) {
	var rules []*database.PermissionRule
	err := r.db.WithContext(ctx).
		Preload("Template").
		Where("is_active = ?", true).
		Order("priority DESC, template_id ASC").
		Find(&rules).Error
	return rules, err
}
