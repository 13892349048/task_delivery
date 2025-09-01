package service

import (
	"time"

	"taskmanage/internal/database"
)

// 权限模板相关DTO

// CreatePermissionTemplateRequest 创建权限模板请求
type CreatePermissionTemplateRequest struct {
	Name             string                    `json:"name" binding:"required"`
	Code             string                    `json:"code" binding:"required"`
	Description      string                    `json:"description"`
	Category         string                    `json:"category" binding:"required"`
	Level            int                       `json:"level"`
	DepartmentID     *uint                     `json:"department_id"`
	PositionID       *uint                     `json:"position_id"`
	ProjectScope     string                    `json:"project_scope" binding:"omitempty,oneof=assigned team department all"`
	TaskScope        string                    `json:"task_scope" binding:"omitempty,oneof=assigned created team all"`
	CanAssignToLevel int                       `json:"can_assign_to_level"`
	CrossDepartment  bool                      `json:"cross_department"`
	MaxTasksPerDay   int                       `json:"max_tasks_per_day"`
	Permissions      []database.Permission     `json:"permissions"`
}

// UpdatePermissionTemplateRequest 更新权限模板请求
type UpdatePermissionTemplateRequest struct {
	Name             *string                   `json:"name"`
	Description      *string                   `json:"description"`
	Category         *string                   `json:"category"`
	Level            *int                      `json:"level"`
	DepartmentID     *uint                     `json:"department_id"`
	PositionID       *uint                     `json:"position_id"`
	ProjectScope     *string                   `json:"project_scope" binding:"omitempty,oneof=assigned team department all"`
	TaskScope        *string                   `json:"task_scope" binding:"omitempty,oneof=assigned created team all"`
	CanAssignToLevel *int                      `json:"can_assign_to_level"`
	CrossDepartment  *bool                     `json:"cross_department"`
	MaxTasksPerDay   *int                      `json:"max_tasks_per_day"`
	Permissions      *[]database.Permission    `json:"permissions"`
	IsActive         *bool                     `json:"is_active"`
}

// ListPermissionTemplatesRequest 获取权限模板列表请求
type ListPermissionTemplatesRequest struct {
	Page         int    `json:"page" form:"page"`
	PageSize     int    `json:"page_size" form:"page_size"`
	Category     string `json:"category" form:"category"`
	Level        *int   `json:"level" form:"level"`
	DepartmentID *uint  `json:"department_id" form:"department_id"`
	PositionID   *uint  `json:"position_id" form:"position_id"`
	IsActive     *bool  `json:"is_active" form:"is_active"`
	Search       string `json:"search" form:"search"`
}

// PermissionTemplateResponse 权限模板响应
type PermissionTemplateResponse struct {
	ID               uint                  `json:"id"`
	Name             string                `json:"name"`
	Code             string                `json:"code"`
	Description      string                `json:"description"`
	Category         string                `json:"category"`
	Level            int                   `json:"level"`
	DepartmentID     *uint                 `json:"department_id"`
	PositionID       *uint                 `json:"position_id"`
	ProjectScope     string                `json:"project_scope"`
	TaskScope        string                `json:"task_scope"`
	CanAssignToLevel int                   `json:"can_assign_to_level"`
	CrossDepartment  bool                  `json:"cross_department"`
	MaxTasksPerDay   int                   `json:"max_tasks_per_day"`
	Permissions      []database.Permission `json:"permissions"`
	IsActive         bool                  `json:"is_active"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
}

// ListPermissionTemplatesResponse 权限模板列表响应
type ListPermissionTemplatesResponse struct {
	Items []*PermissionTemplateResponse `json:"items"`
	Total int                           `json:"total"`
	Page  int                           `json:"page"`
	Size  int                           `json:"size"`
}

// 权限规则相关DTO

// CreatePermissionRuleRequest 创建权限规则请求
type CreatePermissionRuleRequest struct {
	TemplateID       uint   `json:"template_id" binding:"required"`
	Name             string `json:"name" binding:"required"`
	Description      string `json:"description"`
	TriggerCondition string `json:"trigger_condition" binding:"required"`
	ConditionValue   string `json:"condition_value" binding:"required"`
	Action           string `json:"action" binding:"required"`
	Priority         int    `json:"priority"`
	DelayDays        int    `json:"delay_days"`
	RequireApproval  bool   `json:"require_approval"`
}

// UpdatePermissionRuleRequest 更新权限规则请求
type UpdatePermissionRuleRequest struct {
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	TriggerCondition *string `json:"trigger_condition"`
	ConditionValue   *string `json:"condition_value"`
	Action           *string `json:"action"`
	Priority         *int    `json:"priority"`
	DelayDays        *int    `json:"delay_days"`
	RequireApproval  *bool   `json:"require_approval"`
	IsActive         *bool   `json:"is_active"`
}

// ListPermissionRulesRequest 获取权限规则列表请求
type ListPermissionRulesRequest struct {
	Page             int    `json:"page" form:"page"`
	PageSize         int    `json:"page_size" form:"page_size"`
	TemplateID       *uint  `json:"template_id" form:"template_id"`
	TriggerCondition string `json:"trigger_condition" form:"trigger_condition"`
	Action           string `json:"action" form:"action"`
	IsActive         *bool  `json:"is_active" form:"is_active"`
	Search           string `json:"search" form:"search"`
}

// PermissionRuleResponse 权限规则响应
type PermissionRuleResponse struct {
	ID               uint                        `json:"id"`
	TemplateID       uint                        `json:"template_id"`
	Template         *PermissionTemplateResponse `json:"template,omitempty"`
	Name             string                      `json:"name"`
	Description      string                      `json:"description"`
	TriggerCondition string                      `json:"trigger_condition"`
	ConditionValue   string                      `json:"condition_value"`
	Action           string                      `json:"action"`
	Priority         int                         `json:"priority"`
	DelayDays        int                         `json:"delay_days"`
	RequireApproval  bool                        `json:"require_approval"`
	IsActive         bool                        `json:"is_active"`
	CreatedAt        time.Time                   `json:"created_at"`
	UpdatedAt        time.Time                   `json:"updated_at"`
}

// ListPermissionRulesResponse 权限规则列表响应
type ListPermissionRulesResponse struct {
	Items []*PermissionRuleResponse `json:"items"`
	Total int                       `json:"total"`
	Page  int                       `json:"page"`
	Size  int                       `json:"size"`
}

// 权限分配相关DTO

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	UserID       uint   `json:"user_id" binding:"required"`
	TemplateID   *uint  `json:"template_id"`
	PermissionID *uint  `json:"permission_id"`
	Reason       string `json:"reason" binding:"required"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

// UpdatePermissionAssignmentRequest 更新权限分配请求
type UpdatePermissionAssignmentRequest struct {
	Status    *string    `json:"status"`
	ExpiresAt *time.Time `json:"expires_at"`
	Reason    *string    `json:"reason"`
}

// ListPermissionAssignmentsRequest 获取权限分配列表请求
type ListPermissionAssignmentsRequest struct {
	Page           int    `json:"page" form:"page"`
	PageSize       int    `json:"page_size" form:"page_size"`
	UserID         *uint  `json:"user_id" form:"user_id"`
	TemplateID     *uint  `json:"template_id" form:"template_id"`
	Status         string `json:"status" form:"status"`
	ApprovalStatus string `json:"approval_status" form:"approval_status"`
	AssignedBy     *uint  `json:"assigned_by" form:"assigned_by"`
	Search         string `json:"search" form:"search"`
}

// PermissionAssignmentResponse 权限分配响应
type PermissionAssignmentResponse struct {
	ID             uint                        `json:"id"`
	UserID         uint                        `json:"user_id"`
	TemplateID     uint                        `json:"template_id,omitempty"`
	PermissionID   uint                        `json:"permission_id,omitempty"`
	Template       *PermissionTemplateResponse `json:"template,omitempty"`
	Permission     *database.Permission        `json:"permission,omitempty"`
	Status         string                      `json:"status"`
	ApprovalStatus string                      `json:"approval_status"`
	AssignedAt     time.Time                   `json:"assigned_at"`
	ExpiresAt      *time.Time                  `json:"expires_at,omitempty"`
	ApprovedAt     *time.Time                  `json:"approved_at,omitempty"`
	Reason         string                      `json:"reason"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
}

// ListPermissionAssignmentsResponse 权限分配列表响应
type ListPermissionAssignmentsResponse struct {
	Items []*PermissionAssignmentResponse `json:"items"`
	Total int                             `json:"total"`
	Page  int                             `json:"page"`
	Size  int                             `json:"size"`
}

// 权限分配历史相关DTO

// GetPermissionAssignmentHistoryRequest 获取权限分配历史请求
type GetPermissionAssignmentHistoryRequest struct {
	Page         int   `json:"page" form:"page"`
	PageSize     int   `json:"page_size" form:"page_size"`
	AssignmentID *uint `json:"assignment_id" form:"assignment_id"`
	UserID       *uint `json:"user_id" form:"user_id"`
	Action       string `json:"action" form:"action"`
	OperatorID   *uint `json:"operator_id" form:"operator_id"`
	Search       string `json:"search" form:"search"`
}

// PermissionAssignmentHistoryResponse 权限分配历史响应
type PermissionAssignmentHistoryResponse struct {
	ID           uint                          `json:"id"`
	AssignmentID uint                          `json:"assignment_id"`
	Assignment   *PermissionAssignmentResponse `json:"assignment,omitempty"`
	Action       string                        `json:"action"`
	Reason       string                        `json:"reason"`
	Notes        string                        `json:"notes"`
	OperatorID   *uint                         `json:"operator_id"`
	OperatedAt   time.Time                     `json:"operated_at"`
	CreatedAt    time.Time                     `json:"created_at"`
}

// ListPermissionAssignmentHistoryResponse 权限分配历史列表响应
type ListPermissionAssignmentHistoryResponse struct {
	Items []*PermissionAssignmentHistoryResponse `json:"items"`
	Total int                                    `json:"total"`
	Page  int                                    `json:"page"`
	Size  int                                    `json:"size"`
}

// 入职权限配置相关DTO

// CreateOnboardingPermissionConfigRequest 创建入职权限配置请求
type CreateOnboardingPermissionConfigRequest struct {
	OnboardingStatus      string `json:"onboarding_status" binding:"required"`
	DepartmentID          *uint  `json:"department_id"`
	PositionID            *uint  `json:"position_id"`
	DefaultTemplateID     *uint  `json:"default_template_id"`
	NextLevelTemplateID   *uint  `json:"next_level_template_id"`
	DelayDays             int    `json:"delay_days"`
	AutoAssign            bool   `json:"auto_assign"`
	RequireApproval       bool   `json:"require_approval"`
	Description           string `json:"description"`
}

// UpdateOnboardingPermissionConfigRequest 更新入职权限配置请求
type UpdateOnboardingPermissionConfigRequest struct {
	DefaultTemplateID   *uint   `json:"default_template_id"`
	NextLevelTemplateID *uint   `json:"next_level_template_id"`
	DelayDays           *int    `json:"delay_days"`
	AutoAssign          *bool   `json:"auto_assign"`
	RequireApproval     *bool   `json:"require_approval"`
	Description         *string `json:"description"`
}

// ListOnboardingPermissionConfigsRequest 获取入职权限配置列表请求
type ListOnboardingPermissionConfigsRequest struct {
	Page             int    `json:"page" form:"page"`
	PageSize         int    `json:"page_size" form:"page_size"`
	OnboardingStatus string `json:"onboarding_status" form:"onboarding_status"`
	DepartmentID     *uint  `json:"department_id" form:"department_id"`
	PositionID       *uint  `json:"position_id" form:"position_id"`
	AutoAssign       *bool  `json:"auto_assign" form:"auto_assign"`
	Search           string `json:"search" form:"search"`
}

// OnboardingPermissionConfigResponse 入职权限配置响应
type OnboardingPermissionConfigResponse struct {
	ID                    uint                        `json:"id"`
	OnboardingStatus      string                      `json:"onboarding_status"`
	DepartmentID          *uint                       `json:"department_id"`
	PositionID            *uint                       `json:"position_id"`
	DefaultTemplateID     *uint                       `json:"default_template_id"`
	NextLevelTemplateID   *uint                       `json:"next_level_template_id"`
	DefaultTemplate       *PermissionTemplateResponse `json:"default_template,omitempty"`
	NextLevelTemplate     *PermissionTemplateResponse `json:"next_level_template,omitempty"`
	DelayDays             int                         `json:"delay_days"`
	AutoAssign            bool                        `json:"auto_assign"`
	RequireApproval       bool                        `json:"require_approval"`
	Description           string                      `json:"description"`
	CreatedAt             time.Time                   `json:"created_at"`
	UpdatedAt             time.Time                   `json:"updated_at"`
}

// ListOnboardingPermissionConfigsResponse 入职权限配置列表响应
type ListOnboardingPermissionConfigsResponse struct {
	Items []*OnboardingPermissionConfigResponse `json:"items"`
	Total int                                   `json:"total"`
	Page  int                                   `json:"page"`
	Size  int                                   `json:"size"`
}
