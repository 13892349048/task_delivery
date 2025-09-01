package mysql

import (
	"context"

	"gorm.io/gorm"
	"taskmanage/internal/database"
	"taskmanage/internal/repository"
)

// OnboardingHistoryRepositoryImpl 入职历史仓储MySQL实现
type OnboardingHistoryRepositoryImpl struct {
	db *gorm.DB
}

// NewOnboardingHistoryRepository 创建入职历史仓储
func NewOnboardingHistoryRepository(db *gorm.DB) repository.OnboardingHistoryRepository {
	return &OnboardingHistoryRepositoryImpl{db: db}
}

// Create 创建入职历史记录
func (r *OnboardingHistoryRepositoryImpl) Create(ctx context.Context, history *database.OnboardingHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetByEmployeeID 根据员工ID获取历史记录
func (r *OnboardingHistoryRepositoryImpl) GetByEmployeeID(ctx context.Context, employeeID uint) ([]*database.OnboardingHistory, error) {
	var histories []*database.OnboardingHistory
	err := r.db.WithContext(ctx).
		Preload("Employee").
		Preload("Employee.User").
		Preload("Operator").
		Where("employee_id = ?", employeeID).
		Order("created_at DESC").
		Find(&histories).Error
	return histories, err
}

// GetByID 根据ID获取历史记录
func (r *OnboardingHistoryRepositoryImpl) GetByID(ctx context.Context, id uint) (*database.OnboardingHistory, error) {
	var history database.OnboardingHistory
	err := r.db.WithContext(ctx).
		Preload("Employee").
		Preload("Employee.User").
		Preload("Operator").
		First(&history, id).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// List 获取历史记录列表
func (r *OnboardingHistoryRepositoryImpl) List(ctx context.Context, filter *repository.OnboardingHistoryFilter) ([]*database.OnboardingHistory, error) {
	var histories []*database.OnboardingHistory
	query := r.db.WithContext(ctx).
		Preload("Employee").
		Preload("Employee.User").
		Preload("Operator")

	// 应用过滤条件
	if filter.EmployeeID != nil {
		query = query.Where("employee_id = ?", *filter.EmployeeID)
	}
	if filter.FromDate != nil {
		query = query.Where("created_at >= ?", *filter.FromDate)
	}
	if filter.ToDate != nil {
		query = query.Where("created_at <= ?", *filter.ToDate)
	}

	// 分页
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("created_at DESC").Find(&histories).Error
	return histories, err
}
