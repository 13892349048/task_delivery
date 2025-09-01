package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
)

type permissionAssignmentHistoryRepository struct {
	db *gorm.DB
}

func NewPermissionAssignmentHistoryRepository(db *gorm.DB) repository.PermissionAssignmentHistoryRepository {
	return &permissionAssignmentHistoryRepository{db: db}
}

func (r *permissionAssignmentHistoryRepository) Create(ctx context.Context, history *database.PermissionAssignmentHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *permissionAssignmentHistoryRepository) GetByID(ctx context.Context, id uint) (*database.PermissionAssignmentHistory, error) {
	var history database.PermissionAssignmentHistory
	err := r.db.WithContext(ctx).
		Preload("Assignment").
		Preload("Assignment.User").
		Preload("Assignment.Template").
		Preload("Operator").
		First(&history, id).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *permissionAssignmentHistoryRepository) List(ctx context.Context, filter *repository.PermissionAssignmentHistoryFilter) ([]*database.PermissionAssignmentHistory, error) {
	var histories []*database.PermissionAssignmentHistory
	query := r.db.WithContext(ctx).
		Preload("Assignment").
		Preload("Assignment.User").
		Preload("Assignment.Template").
		Preload("Operator")

	// 应用过滤条件
	if filter.AssignmentID != nil {
		query = query.Where("assignment_id = ?", *filter.AssignmentID)
	}
	if filter.UserID != nil {
		query = query.Joins("LEFT JOIN permission_assignments ON permission_assignments.id = permission_assignment_histories.assignment_id").
			Where("permission_assignments.user_id = ?", *filter.UserID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.OperatorID != nil {
		query = query.Where("operator_id = ?", *filter.OperatorID)
	}
	if filter.Search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("reason LIKE ? OR notes LIKE ?", searchPattern, searchPattern)
	}

	// 分页
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("operated_at DESC").Find(&histories).Error
	return histories, err
}

func (r *permissionAssignmentHistoryRepository) GetByAssignmentID(ctx context.Context, assignmentID uint) ([]*database.PermissionAssignmentHistory, error) {
	var histories []*database.PermissionAssignmentHistory
	err := r.db.WithContext(ctx).
		Preload("Operator").
		Where("assignment_id = ?", assignmentID).
		Order("operated_at DESC").
		Find(&histories).Error
	return histories, err
}

func (r *permissionAssignmentHistoryRepository) GetByUserID(ctx context.Context, userID uint) ([]*database.PermissionAssignmentHistory, error) {
	var histories []*database.PermissionAssignmentHistory
	err := r.db.WithContext(ctx).
		Preload("Assignment").
		Preload("Assignment.Template").
		Preload("Operator").
		Joins("LEFT JOIN permission_assignments ON permission_assignments.id = permission_assignment_histories.assignment_id").
		Where("permission_assignments.user_id = ?", userID).
		Order("operated_at DESC").
		Find(&histories).Error
	return histories, err
}
