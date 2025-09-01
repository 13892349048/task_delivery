package mysql

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
)

type permissionAssignmentRepository struct {
	db *gorm.DB
}

func NewPermissionAssignmentRepository(db *gorm.DB) repository.PermissionAssignmentRepository {
	return &permissionAssignmentRepository{db: db}
}

func (r *permissionAssignmentRepository) Create(ctx context.Context, assignment *database.PermissionAssignment) error {
	return r.db.WithContext(ctx).Create(assignment).Error
}

func (r *permissionAssignmentRepository) GetByID(ctx context.Context, id uint) (*database.PermissionAssignment, error) {
	var assignment database.PermissionAssignment
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Template").
		Preload("Permission").
		Preload("Rule").
		Preload("AssignedByUser").
		Preload("ApprovedByUser").
		First(&assignment, id).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *permissionAssignmentRepository) List(ctx context.Context, filter *repository.PermissionAssignmentFilter) ([]*database.PermissionAssignment, error) {
	var assignments []*database.PermissionAssignment
	query := r.db.WithContext(ctx).
		Preload("User").
		Preload("Template").
		Preload("Permission").
		Preload("AssignedByUser")

	// 应用过滤条件
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.TemplateID != nil {
		query = query.Where("template_id = ?", *filter.TemplateID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.ApprovalStatus != "" {
		query = query.Where("approval_status = ?", filter.ApprovalStatus)
	}
	if filter.AssignedBy != nil {
		query = query.Where("assigned_by = ?", *filter.AssignedBy)
	}
	if filter.Search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Joins("LEFT JOIN users ON users.id = permission_assignments.user_id").
			Where("users.real_name LIKE ? OR users.username LIKE ?", searchPattern, searchPattern)
	}

	// 分页
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Order("assigned_at DESC").Find(&assignments).Error
	return assignments, err
}

func (r *permissionAssignmentRepository) Update(ctx context.Context, assignment *database.PermissionAssignment) error {
	return r.db.WithContext(ctx).Save(assignment).Error
}

func (r *permissionAssignmentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&database.PermissionAssignment{}, id).Error
}

func (r *permissionAssignmentRepository) GetByUserID(ctx context.Context, userID uint) ([]*database.PermissionAssignment, error) {
	var assignments []*database.PermissionAssignment
	err := r.db.WithContext(ctx).
		Preload("Template").
		Preload("Permission").
		Where("user_id = ?", userID).
		Order("assigned_at DESC").
		Find(&assignments).Error
	return assignments, err
}

func (r *permissionAssignmentRepository) GetActiveByUserID(ctx context.Context, userID uint) ([]*database.PermissionAssignment, error) {
	var assignments []*database.PermissionAssignment
	now := time.Now()
	err := r.db.WithContext(ctx).
		Preload("Template").
		Preload("Permission").
		Where("user_id = ? AND status = ? AND approval_status = ? AND (expires_at IS NULL OR expires_at > ?)", 
			userID, database.PermissionStatusActive, database.ApprovalStatusApproved, now).
		Order("assigned_at DESC").
		Find(&assignments).Error
	return assignments, err
}

func (r *permissionAssignmentRepository) GetByTemplateID(ctx context.Context, templateID uint) ([]*database.PermissionAssignment, error) {
	var assignments []*database.PermissionAssignment
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("template_id = ?", templateID).
		Order("assigned_at DESC").
		Find(&assignments).Error
	return assignments, err
}

func (r *permissionAssignmentRepository) GetPendingApprovals(ctx context.Context) ([]*database.PermissionAssignment, error) {
	var assignments []*database.PermissionAssignment
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Template").
		Preload("Permission").
		Preload("AssignedByUser").
		Where("approval_status = ?", database.ApprovalStatusPending).
		Order("assigned_at ASC").
		Find(&assignments).Error
	return assignments, err
}

func (r *permissionAssignmentRepository) GetExpiredAssignments(ctx context.Context) ([]*database.PermissionAssignment, error) {
	var assignments []*database.PermissionAssignment
	now := time.Now()
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Template").
		Where("status = ? AND expires_at IS NOT NULL AND expires_at <= ?", 
			database.PermissionStatusActive, now).
		Find(&assignments).Error
	return assignments, err
}

func (r *permissionAssignmentRepository) BulkUpdateStatus(ctx context.Context, ids []uint, status string) error {
	return r.db.WithContext(ctx).
		Model(&database.PermissionAssignment{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}
