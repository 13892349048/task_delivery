package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
)

type onboardingPermissionConfigRepository struct {
	db *gorm.DB
}

func NewOnboardingPermissionConfigRepository(db *gorm.DB) repository.OnboardingPermissionConfigRepository {
	return &onboardingPermissionConfigRepository{db: db}
}

func (r *onboardingPermissionConfigRepository) Create(ctx context.Context, config *database.OnboardingPermissionConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

func (r *onboardingPermissionConfigRepository) GetByID(ctx context.Context, id uint) (*database.OnboardingPermissionConfig, error) {
	var config database.OnboardingPermissionConfig
	err := r.db.WithContext(ctx).
		Preload("Department").
		Preload("Position").
		Preload("DefaultTemplate").
		Preload("NextLevelTemplate").
		First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *onboardingPermissionConfigRepository) List(ctx context.Context, filter *repository.OnboardingPermissionConfigFilter) ([]*database.OnboardingPermissionConfig, error) {
	var configs []*database.OnboardingPermissionConfig
	query := r.db.WithContext(ctx).
		Preload("Department").
		Preload("Position").
		Preload("DefaultTemplate").
		Preload("NextLevelTemplate")

	// 应用过滤条件
	if filter.OnboardingStatus != "" {
		query = query.Where("onboarding_status = ?", filter.OnboardingStatus)
	}
	if filter.DepartmentID != nil {
		query = query.Where("department_id = ?", *filter.DepartmentID)
	}
	if filter.PositionID != nil {
		query = query.Where("position_id = ?", *filter.PositionID)
	}
	if filter.AutoAssign != nil {
		query = query.Where("auto_assign = ?", *filter.AutoAssign)
	}
	if filter.Search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("onboarding_status LIKE ?", searchPattern)
	}

	// 分页
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("onboarding_status ASC, department_id ASC, position_id ASC").Find(&configs).Error
	return configs, err
}

func (r *onboardingPermissionConfigRepository) Update(ctx context.Context, config *database.OnboardingPermissionConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *onboardingPermissionConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&database.OnboardingPermissionConfig{}, id).Error
}

func (r *onboardingPermissionConfigRepository) GetByStatus(ctx context.Context, status string) ([]*database.OnboardingPermissionConfig, error) {
	var configs []*database.OnboardingPermissionConfig
	err := r.db.WithContext(ctx).
		Preload("Department").
		Preload("Position").
		Preload("DefaultTemplate").
		Preload("NextLevelTemplate").
		Where("onboarding_status = ?", status).
		Order("department_id ASC, position_id ASC").
		Find(&configs).Error
	return configs, err
}

func (r *onboardingPermissionConfigRepository) GetByStatusAndDepartment(ctx context.Context, status string, departmentID, positionID *uint) (*database.OnboardingPermissionConfig, error) {
	var config database.OnboardingPermissionConfig
	query := r.db.WithContext(ctx).
		Preload("Department").
		Preload("Position").
		Preload("DefaultTemplate").
		Preload("NextLevelTemplate").
		Where("onboarding_status = ?", status)

	// 优先级查找：具体部门+职位 > 具体部门 > 全局配置
	if departmentID != nil && positionID != nil {
		// 先查找具体部门和职位的配置
		err := query.Where("department_id = ? AND position_id = ?", *departmentID, *positionID).First(&config).Error
		if err == nil {
			return &config, nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}

		// 再查找具体部门的配置
		err = query.Where("department_id = ? AND position_id IS NULL", *departmentID).First(&config).Error
		if err == nil {
			return &config, nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	} else if departmentID != nil {
		// 查找具体部门的配置
		err := query.Where("department_id = ? AND position_id IS NULL", *departmentID).First(&config).Error
		if err == nil {
			return &config, nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	// 最后查找全局配置
	err := query.Where("department_id IS NULL AND position_id IS NULL").First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *onboardingPermissionConfigRepository) GetGlobalConfig(ctx context.Context, status string) (*database.OnboardingPermissionConfig, error) {
	var config database.OnboardingPermissionConfig
	err := r.db.WithContext(ctx).
		Preload("DefaultTemplate").
		Preload("NextLevelTemplate").
		Where("onboarding_status = ? AND department_id IS NULL AND position_id IS NULL", status).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}
