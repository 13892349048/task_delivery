package repository

import (
	"context"
	"taskmanage/internal/database"
)

// PermissionTemplateRepository 权限模板仓储接口
type PermissionTemplateRepository interface {
	Create(ctx context.Context, template *database.PermissionTemplate) error
	GetByID(ctx context.Context, id uint) (*database.PermissionTemplate, error)
	GetByCode(ctx context.Context, code string) (*database.PermissionTemplate, error)
	List(ctx context.Context, filter *PermissionTemplateFilter) ([]*database.PermissionTemplate, error)
	Update(ctx context.Context, template *database.PermissionTemplate) error
	Delete(ctx context.Context, id uint) error
	GetByCategory(ctx context.Context, category string) ([]*database.PermissionTemplate, error)
	GetByDepartmentAndPosition(ctx context.Context, departmentID, positionID *uint) ([]*database.PermissionTemplate, error)
}

// PermissionRuleRepository 权限规则仓储接口
type PermissionRuleRepository interface {
	Create(ctx context.Context, rule *database.PermissionRule) error
	GetByID(ctx context.Context, id uint) (*database.PermissionRule, error)
	List(ctx context.Context, filter *PermissionRuleFilter) ([]*database.PermissionRule, error)
	Update(ctx context.Context, rule *database.PermissionRule) error
	Delete(ctx context.Context, id uint) error
	GetByTemplateID(ctx context.Context, templateID uint) ([]*database.PermissionRule, error)
	GetByTriggerCondition(ctx context.Context, condition, value string) ([]*database.PermissionRule, error)
	GetActiveRules(ctx context.Context) ([]*database.PermissionRule, error)
}

// PermissionAssignmentRepository 权限分配仓储接口
type PermissionAssignmentRepository interface {
	Create(ctx context.Context, assignment *database.PermissionAssignment) error
	GetByID(ctx context.Context, id uint) (*database.PermissionAssignment, error)
	List(ctx context.Context, filter *PermissionAssignmentFilter) ([]*database.PermissionAssignment, error)
	Update(ctx context.Context, assignment *database.PermissionAssignment) error
	Delete(ctx context.Context, id uint) error
	GetByUserID(ctx context.Context, userID uint) ([]*database.PermissionAssignment, error)
	GetActiveByUserID(ctx context.Context, userID uint) ([]*database.PermissionAssignment, error)
	GetByTemplateID(ctx context.Context, templateID uint) ([]*database.PermissionAssignment, error)
	GetPendingApprovals(ctx context.Context) ([]*database.PermissionAssignment, error)
	GetExpiredAssignments(ctx context.Context) ([]*database.PermissionAssignment, error)
	BulkUpdateStatus(ctx context.Context, ids []uint, status string) error
}

// PermissionAssignmentHistoryRepository 权限分配历史仓储接口
type PermissionAssignmentHistoryRepository interface {
	Create(ctx context.Context, history *database.PermissionAssignmentHistory) error
	GetByID(ctx context.Context, id uint) (*database.PermissionAssignmentHistory, error)
	List(ctx context.Context, filter *PermissionAssignmentHistoryFilter) ([]*database.PermissionAssignmentHistory, error)
	GetByAssignmentID(ctx context.Context, assignmentID uint) ([]*database.PermissionAssignmentHistory, error)
	GetByUserID(ctx context.Context, userID uint) ([]*database.PermissionAssignmentHistory, error)
}

// OnboardingPermissionConfigRepository 入职权限配置仓储接口
type OnboardingPermissionConfigRepository interface {
	Create(ctx context.Context, config *database.OnboardingPermissionConfig) error
	GetByID(ctx context.Context, id uint) (*database.OnboardingPermissionConfig, error)
	List(ctx context.Context, filter *OnboardingPermissionConfigFilter) ([]*database.OnboardingPermissionConfig, error)
	Update(ctx context.Context, config *database.OnboardingPermissionConfig) error
	Delete(ctx context.Context, id uint) error
	GetByStatus(ctx context.Context, status string) ([]*database.OnboardingPermissionConfig, error)
	GetByStatusAndDepartment(ctx context.Context, status string, departmentID *uint, positionID *uint) (*database.OnboardingPermissionConfig, error)
	GetGlobalConfig(ctx context.Context, status string) (*database.OnboardingPermissionConfig, error)
}

// 过滤器结构体
type PermissionTemplateFilter struct {
	Page         int
	PageSize     int
	Category     string
	Level        *int
	DepartmentID *uint
	PositionID   *uint
	IsActive     *bool
	Search       string
}

type PermissionRuleFilter struct {
	Page             int
	PageSize         int
	TemplateID       *uint
	TriggerCondition string
	Action           string
	IsActive         *bool
	Search           string
}

type PermissionAssignmentFilter struct {
	Page           int
	PageSize       int
	UserID         *uint
	TemplateID     *uint
	Status         string
	ApprovalStatus string
	AssignedBy     *uint
	Search         string
}

type PermissionAssignmentHistoryFilter struct {
	Page         int
	PageSize     int
	AssignmentID *uint
	UserID       *uint
	Action       string
	OperatorID   *uint
	Search       string
}

type OnboardingPermissionConfigFilter struct {
	Page             int
	PageSize         int
	OnboardingStatus string
	DepartmentID     *uint
	PositionID       *uint
	AutoAssign       *bool
	Search           string
}
