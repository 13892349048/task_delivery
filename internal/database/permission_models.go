package database

import (
	"time"
)

// PermissionTemplate 权限模板
type PermissionTemplate struct {
	BaseModel
	Name         string `gorm:"size:100;not null" json:"name"`
	Code         string `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Description  string `gorm:"size:255" json:"description"`
	Category     string `gorm:"size:50" json:"category"` // basic, intermediate, advanced, manager
	Level        int    `gorm:"not null" json:"level"`   // 权限等级，数字越大权限越高
	DepartmentID *uint  `gorm:"index" json:"department_id,omitempty"` // 部门特定权限
	PositionID   *uint  `gorm:"index" json:"position_id,omitempty"`   // 职位特定权限
	IsActive     bool   `gorm:"default:true" json:"is_active"`
	
	// 任务分配系统特定字段
	ProjectScope     string `gorm:"size:20;default:'assigned'" json:"project_scope"`     // assigned, team, department, all
	TaskScope        string `gorm:"size:20;default:'assigned'" json:"task_scope"`        // assigned, created, team, all
	CanAssignToLevel int    `gorm:"default:0" json:"can_assign_to_level"`                // 可分配给的最高职级
	CrossDepartment  bool   `gorm:"default:false" json:"cross_department"`               // 是否可跨部门分配
	MaxTasksPerDay   int    `gorm:"default:0" json:"max_tasks_per_day"`                  // 每日最大任务分配数，0表示无限制
	
	// 关联关系
	Permissions []Permission `gorm:"many2many:template_permissions;" json:"permissions,omitempty"`
	Rules       []PermissionRule `gorm:"foreignKey:TemplateID" json:"rules,omitempty"`
}

// PermissionRule 权限分配规则
type PermissionRule struct {
	BaseModel
	TemplateID       uint   `gorm:"not null;index" json:"template_id"`
	Name             string `gorm:"size:100;not null" json:"name"`
	Description      string `gorm:"size:255" json:"description"`
	TriggerCondition string `gorm:"size:100;not null" json:"trigger_condition"` // onboarding_status, work_days, department_change
	ConditionValue   string `gorm:"size:100;not null" json:"condition_value"`   // pending_onboard, onboarding, probation, active
	Action           string `gorm:"size:50;not null" json:"action"`             // grant, revoke, upgrade
	Priority         int    `gorm:"default:0" json:"priority"`                  // 规则优先级
	IsActive         bool   `gorm:"default:true" json:"is_active"`
	
	// 延迟执行配置
	DelayDays    int  `gorm:"default:0" json:"delay_days"`    // 延迟天数
	AutoExecute  bool `gorm:"default:true" json:"auto_execute"` // 是否自动执行
	RequireApproval bool `gorm:"default:false" json:"require_approval"` // 是否需要审批
	
	// 关联关系
	Template PermissionTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
}

// PermissionAssignment 权限分配记录
type PermissionAssignment struct {
	BaseModel
	UserID       uint   `gorm:"not null;index" json:"user_id"`
	TemplateID   *uint  `gorm:"index" json:"template_id,omitempty"`   // 通过模板分配
	PermissionID *uint  `gorm:"index" json:"permission_id,omitempty"` // 直接分配权限
	RuleID       *uint  `gorm:"index" json:"rule_id,omitempty"`       // 触发的规则
	
	// 分配信息
	AssignedBy   uint      `gorm:"not null;index" json:"assigned_by"`   // 分配者
	AssignedAt   time.Time `gorm:"not null" json:"assigned_at"`         // 分配时间
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`               // 过期时间
	Status       string    `gorm:"size:20;default:active" json:"status"` // active, expired, revoked
	
	// 触发信息
	TriggerEvent string `gorm:"size:100" json:"trigger_event"` // 触发事件
	TriggerData  string `gorm:"type:text" json:"trigger_data"` // 触发数据（JSON格式）
	
	// 审批信息
	ApprovalStatus   string     `gorm:"size:20;default:approved" json:"approval_status"` // pending, approved, rejected
	ApprovedBy       *uint      `gorm:"index" json:"approved_by,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
	ApprovalComments string     `gorm:"size:500" json:"approval_comments"`
	
	// 关联关系
	User       User                `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Template   *PermissionTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Permission *Permission         `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
	Rule       *PermissionRule     `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	AssignedByUser User            `gorm:"foreignKey:AssignedBy" json:"assigned_by_user,omitempty"`
	ApprovedByUser *User           `gorm:"foreignKey:ApprovedBy" json:"approved_by_user,omitempty"`
}

// PermissionAssignmentHistory 权限分配历史
type PermissionAssignmentHistory struct {
	BaseModel
	AssignmentID uint      `gorm:"not null;index" json:"assignment_id"`
	Action       string    `gorm:"size:50;not null" json:"action"` // granted, revoked, expired, upgraded
	Reason       string    `gorm:"size:255" json:"reason"`
	OperatorID   uint      `gorm:"not null;index" json:"operator_id"`
	OperatedAt   time.Time `gorm:"not null" json:"operated_at"`
	OldStatus    string    `gorm:"size:20" json:"old_status"`
	NewStatus    string    `gorm:"size:20" json:"new_status"`
	Notes        string    `gorm:"size:500" json:"notes"`
	
	// 关联关系
	Assignment PermissionAssignment `gorm:"foreignKey:AssignmentID" json:"assignment,omitempty"`
	Operator   User                 `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
}

// OnboardingPermissionConfig 入职权限配置
type OnboardingPermissionConfig struct {
	BaseModel
	OnboardingStatus string `gorm:"size:30;not null;index" json:"onboarding_status"` // 入职状态
	DepartmentID     *uint  `gorm:"index" json:"department_id,omitempty"`            // 部门ID，null表示全局配置
	PositionID       *uint  `gorm:"index" json:"position_id,omitempty"`              // 职位ID，null表示全局配置
	
	// 权限配置
	DefaultTemplateID *uint `gorm:"index" json:"default_template_id,omitempty"` // 默认权限模板
	AutoAssign        bool  `gorm:"default:true" json:"auto_assign"`            // 是否自动分配
	RequireApproval   bool  `gorm:"default:false" json:"require_approval"`      // 是否需要审批
	
	// 时间配置
	EffectiveAfterDays int `gorm:"default:0" json:"effective_after_days"` // 生效延迟天数
	ExpirationDays     int `gorm:"default:0" json:"expiration_days"`      // 过期天数，0表示永不过期
	
	// 升级配置
	NextLevelTemplateID *uint `gorm:"index" json:"next_level_template_id,omitempty"` // 下一级权限模板
	UpgradeAfterDays    int   `gorm:"default:0" json:"upgrade_after_days"`           // 自动升级天数
	
	// 关联关系
	Department       *Department         `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Position         *Position           `gorm:"foreignKey:PositionID" json:"position,omitempty"`
	DefaultTemplate  *PermissionTemplate `gorm:"foreignKey:DefaultTemplateID" json:"default_template,omitempty"`
	NextLevelTemplate *PermissionTemplate `gorm:"foreignKey:NextLevelTemplateID" json:"next_level_template,omitempty"`
}

// 权限分配状态常量
const (
	PermissionStatusActive   = "active"
	PermissionStatusExpired  = "expired"
	PermissionStatusRevoked  = "revoked"
	PermissionStatusPending  = "pending"
)

// 审批状态常量
const (
	ApprovalStatusPending  = "pending"
	ApprovalStatusApproved = "approved"
	ApprovalStatusRejected = "rejected"
)

// 权限分配动作常量
const (
	PermissionActionGrant   = "grant"
	PermissionActionRevoke  = "revoke"
	PermissionActionUpgrade = "upgrade"
	PermissionActionExpire  = "expire"
)

// 触发条件常量
const (
	TriggerOnboardingStatus = "onboarding_status"
	TriggerWorkDays         = "work_days"
	TriggerDepartmentChange = "department_change"
	TriggerPositionChange   = "position_change"
	TriggerManagerAssign    = "manager_assign"
)
