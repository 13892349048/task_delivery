package database

import (
	"time"

	"gorm.io/gorm"
)

// 任务状态常量
const (
	TaskStatusPending    = "pending"
	TaskStatusAssigned   = "assigned"
	TaskStatusInProgress = "in_progress"
	TaskStatusCompleted  = "completed"
	TaskStatusCancelled  = "cancelled"
)

// BaseModel 基础模型，包含通用字段
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// User 用户表
type User struct {
	BaseModel
	Username     string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string     `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Password     string     `gorm:"size:255;not null;column:password" json:"-"` // 兼容数据库中的password字段
	PasswordHash string     `gorm:"size:255;not null;column:password_hash" json:"-"`
	RealName     string     `gorm:"size:100" json:"real_name"`
	Role         string     `gorm:"size:50;default:employee" json:"role"`
	Phone        string     `gorm:"size:20" json:"phone"`
	Avatar       string     `gorm:"size:255" json:"avatar"`
	Status       string     `gorm:"size:20;default:active" json:"status"` // active, inactive, suspended
	LastLoginAt  *time.Time `json:"last_login_at"`
	LastLoginIP  string     `gorm:"size:45" json:"last_login_ip"`

	// 关联关系
	Roles        []Role    `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	Tasks        []Task    `gorm:"foreignKey:AssigneeID" json:"tasks,omitempty"`
	CreatedTasks []Task    `gorm:"foreignKey:CreatorID" json:"created_tasks,omitempty"`
	Employee     *Employee `gorm:"foreignKey:UserID" json:"employee,omitempty"`
}

// Role 角色表
type Role struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;size:50;not null" json:"name"`
	DisplayName string `gorm:"size:100" json:"display_name"`
	Description string `gorm:"size:255" json:"description"`

	// 关联关系
	Users       []User       `gorm:"many2many:user_roles;" json:"users,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

// Permission 权限表
type Permission struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;size:100;not null" json:"name"`
	DisplayName string `gorm:"size:100" json:"display_name"`
	Description string `gorm:"size:255" json:"description"`
	Resource    string `gorm:"size:50" json:"resource"`
	Action      string `gorm:"size:50" json:"action"`

	// 关联关系
	Roles []Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}

// Department 部门表
type Department struct {
	BaseModel
	Name        string `gorm:"size:100;not null" json:"name"`
	Code        string `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Description string `gorm:"size:255" json:"description"`
	ParentID    *uint  `gorm:"index" json:"parent_id"`
	ManagerID   *uint  `gorm:"index" json:"manager_id"`
	Level       int    `gorm:"default:1" json:"level"`
	Path        string `gorm:"size:500" json:"path"` // 部门路径，如: /1/3/5
	Status      string `gorm:"size:20;default:active" json:"status"`

	// 关联关系
	Parent    *Department  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children  []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Manager   *Employee    `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	Employees []Employee   `gorm:"foreignKey:DepartmentID" json:"employees,omitempty"`
	Projects  []Project    `gorm:"foreignKey:DepartmentID" json:"projects,omitempty"`
}

// Position 职位表
type Position struct {
	BaseModel
	Name        string `gorm:"size:100;not null" json:"name"`
	Code        string `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Description string `gorm:"size:255" json:"description"`
	Level       int    `gorm:"not null" json:"level"`   // 职级等级，数字越大级别越高
	Category    string `gorm:"size:50" json:"category"` // 职位类别：管理类、技术类、业务类等
	Status      string `gorm:"size:20;default:active" json:"status"`

	// 关联关系
	Employees []Employee `gorm:"foreignKey:PositionID" json:"employees,omitempty"`
}

// Project 项目表
type Project struct {
	BaseModel
	Name         string     `gorm:"size:200;not null" json:"name"`
	Code         string     `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Description  string     `gorm:"type:text" json:"description"`
	Status       string     `gorm:"size:20;default:planning" json:"status"` // planning, active, suspended, completed, cancelled
	Priority     string     `gorm:"size:20;default:medium" json:"priority"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Budget       float64    `gorm:"default:0" json:"budget"`
	DepartmentID uint       `gorm:"not null;index" json:"department_id"`
	ManagerID    uint       `gorm:"not null;index" json:"manager_id"`

	// 关联关系
	Department Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Manager    Employee   `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	Members    []Employee `gorm:"many2many:project_members;" json:"members,omitempty"`
	Tasks      []Task     `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`
}

// Employee 员工表
type Employee struct {
	BaseModel
	UserID          uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	EmployeeNo      string `gorm:"uniqueIndex;size:50;not null" json:"employee_no"`
	DepartmentID    *uint  `gorm:"index" json:"department_id"`     // 改为可选，支持待入职状态
	PositionID      *uint  `gorm:"index" json:"position_id"`       // 改为可选，支持待入职状态
	DirectManagerID *uint  `gorm:"index" json:"direct_manager_id"` // 直接上级

	// 入职流程相关字段
	OnboardingStatus string     `gorm:"size:30;default:pending_onboard" json:"onboarding_status"` // 入职状态
	ExpectedDate     *time.Time `json:"expected_date,omitempty"`                                  // 预期入职日期
	HireDate         *time.Time `json:"hire_date,omitempty"`                                      // 实际入职日期
	ProbationEndDate *time.Time `json:"probation_end_date,omitempty"`                             // 试用期结束日期
	ConfirmDate      *time.Time `json:"confirm_date,omitempty"`                                   // 转正日期

	// 工作信息
	WorkLocation string `gorm:"size:100" json:"work_location"`
	WorkType     string `gorm:"size:20;default:fulltime" json:"work_type"` // fulltime, parttime, contract, intern
	Status       string `gorm:"size:20;default:available" json:"status"`   // available, busy, offline, leave, resigned
	MaxTasks     int    `gorm:"default:5" json:"max_tasks"`
	CurrentTasks int    `gorm:"default:0" json:"current_tasks"`

	// 入职备注
	OnboardingNotes string `gorm:"type:text" json:"onboarding_notes,omitempty"`

	// 关联关系
	User            User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Department      Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Position        Position   `gorm:"foreignKey:PositionID" json:"position,omitempty"`
	DirectManager   *Employee  `gorm:"foreignKey:DirectManagerID" json:"direct_manager,omitempty"`
	Subordinates    []Employee `gorm:"foreignKey:DirectManagerID" json:"subordinates,omitempty"`
	Skills          []Skill    `gorm:"many2many:employee_skills;" json:"skills,omitempty"`
	Projects        []Project  `gorm:"many2many:project_members;" json:"projects,omitempty"`
	ManagedProjects []Project  `gorm:"foreignKey:ManagerID" json:"managed_projects,omitempty"`
}

// Skill 技能表
type Skill struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Category    string `gorm:"size:50" json:"category"`
	Description string `gorm:"size:255" json:"description"`

	// 关联关系
	Employees []Employee `gorm:"many2many:employee_skills;" json:"employees,omitempty"`
	Tasks     []Task     `gorm:"many2many:task_skills;" json:"tasks,omitempty"`
}

// Task 任务表
type Task struct {
	BaseModel
	Title          string     `gorm:"size:200;not null" json:"title"`
	Description    string     `gorm:"type:text" json:"description"`
	Priority       string     `gorm:"size:20;default:medium" json:"priority"` // low, medium, high, urgent
	Status         string     `gorm:"size:20;default:pending" json:"status"`  // pending, assigned, in_progress, completed, cancelled
	Type           string     `gorm:"size:50" json:"type"`
	EstimatedHours float64    `gorm:"default:0" json:"estimated_hours"`
	ActualHours    float64    `gorm:"default:0" json:"actual_hours"`
	DueDate        *time.Time `json:"due_date"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`

	// 外键
	CreatorID  uint  `gorm:"not null" json:"creator_id"`
	AssigneeID *uint `json:"assignee_id"`
	ParentID   *uint `json:"parent_id"`
	ProjectID  *uint `gorm:"index" json:"project_id"`

	// 关联关系
	Creator     User             `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Assignee    *User            `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Parent      *Task            `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	SubTasks    []Task           `gorm:"foreignKey:ParentID" json:"sub_tasks,omitempty"`
	Project     *Project         `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Skills      []Skill          `gorm:"many2many:task_skills;" json:"skills,omitempty"`
	Assignments []Assignment     `gorm:"foreignKey:TaskID" json:"assignments,omitempty"`
	Comments    []TaskComment    `gorm:"foreignKey:TaskID" json:"comments,omitempty"`
	Attachments []TaskAttachment `gorm:"foreignKey:TaskID" json:"attachments,omitempty"`
}

// Assignment 任务分配表
type Assignment struct {
	BaseModel
	TaskID             uint       `gorm:"not null" json:"task_id"`
	AssigneeID         uint       `gorm:"not null" json:"assignee_id"`
	AssignerID         uint       `gorm:"not null" json:"assigner_id"`
	Method             string     `gorm:"size:50" json:"method"`                 // manual, auto_round_robin, auto_load_balance, auto_skill_match
	Status             string     `gorm:"size:20;default:pending" json:"status"` // pending, approved, rejected, completed
	AssignedAt         time.Time  `json:"assigned_at"`
	ApprovedAt         *time.Time `json:"approved_at"`
	ApproverID         *uint      `json:"approver_id"`
	Reason             string     `gorm:"size:255" json:"reason"`
	WorkflowInstanceID *string    `gorm:"size:100" json:"workflow_instance_id"` // 关联的工作流实例ID

	// 关联关系
	Task     Task  `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	Assignee User  `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Assigner User  `gorm:"foreignKey:AssignerID" json:"assigner,omitempty"`
	Approver *User `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
}

// TaskComment 任务评论表
type TaskComment struct {
	BaseModel
	TaskID   uint   `gorm:"not null" json:"task_id"`
	UserID   uint   `gorm:"not null" json:"user_id"`
	Content  string `gorm:"type:text;not null" json:"content"`
	ParentID *uint  `json:"parent_id"`

	// 关联关系
	Task    Task          `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	User    User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Parent  *TaskComment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies []TaskComment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TaskAttachment 任务附件表
type TaskAttachment struct {
	BaseModel
	TaskID   uint   `gorm:"not null" json:"task_id"`
	UserID   uint   `gorm:"not null" json:"user_id"`
	Filename string `gorm:"size:255;not null" json:"filename"`
	FileSize int64  `json:"file_size"`
	FilePath string `gorm:"size:500;not null" json:"file_path"`
	MimeType string `gorm:"size:100" json:"mime_type"`

	// 关联关系
	Task Task `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TaskNotification 任务通知表
type TaskNotification struct {
	BaseModel
	Type        string     `gorm:"type:varchar(50);not null" json:"type"`
	Title       string     `gorm:"type:varchar(200);not null" json:"title"`
	Content     string     `gorm:"type:text" json:"content"`
	RecipientID uint       `gorm:"not null;index" json:"recipient_id"`
	SenderID    *uint      `gorm:"index" json:"sender_id,omitempty"`
	TaskID      *uint      `gorm:"index" json:"task_id,omitempty"`
	Priority    string     `gorm:"type:varchar(20);default:'medium'" json:"priority"`
	Status      string     `gorm:"type:varchar(20);default:'unread'" json:"status"`
	ActionType  *string    `gorm:"type:varchar(50)" json:"action_type,omitempty"`
	ActionData  *string    `gorm:"type:json" json:"action_data,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	ActionAt    *time.Time `json:"action_at,omitempty"`

	// 关联关系
	Recipient User  `gorm:"foreignKey:RecipientID" json:"recipient,omitempty"`
	Sender    *User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Task      *Task `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

// TaskNotificationAction 任务通知操作记录表
type TaskNotificationAction struct {
	BaseModel
	NotificationID uint    `gorm:"not null;index" json:"notification_id"`
	UserID         uint    `gorm:"not null;index" json:"user_id"`
	ActionType     string  `gorm:"type:varchar(50);not null" json:"action_type"`
	ActionData     *string `gorm:"type:json" json:"action_data,omitempty"`
	Reason         *string `gorm:"type:text" json:"reason,omitempty"`

	// 关联关系
	Notification TaskNotification `gorm:"foreignKey:NotificationID" json:"notification,omitempty"`
	User         User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Notification 通知表
type Notification struct {
	BaseModel
	UserID  uint       `gorm:"not null" json:"user_id"`
	Type    string     `gorm:"size:50;not null" json:"type"` // task_assigned, task_completed, etc.
	Title   string     `gorm:"size:200;not null" json:"title"`
	Content string     `gorm:"type:text" json:"content"`
	IsRead  bool       `gorm:"default:false" json:"is_read"`
	ReadAt  *time.Time `json:"read_at"`

	// 关联数据 (JSON格式存储)
	Data string `gorm:"type:json" json:"data"`

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// AuditLog 审计日志表
type AuditLog struct {
	BaseModel
	UserID     uint   `json:"user_id"`
	Action     string `gorm:"size:100;not null" json:"action"`
	Resource   string `gorm:"size:100;not null" json:"resource"`
	ResourceID uint   `json:"resource_id"`
	Method     string `gorm:"size:10" json:"method"`
	Path       string `gorm:"size:255" json:"path"`
	IP         string `gorm:"size:45" json:"ip"`
	UserAgent  string `gorm:"size:500" json:"user_agent"`

	// 请求和响应数据 (JSON格式存储)
	RequestData  string `gorm:"type:json" json:"request_data"`
	ResponseData string `gorm:"type:json" json:"response_data"`

	// 关联关系
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// SystemConfig 系统配置表
type SystemConfig struct {
	BaseModel
	Key         string `gorm:"uniqueIndex;size:100;not null" json:"key"`
	Value       string `gorm:"type:text" json:"value"`
	Type        string `gorm:"size:50;default:string" json:"type"` // string, int, float, bool, json
	Category    string `gorm:"size:50" json:"category"`
	Description string `gorm:"size:255" json:"description"`
	IsPublic    bool   `gorm:"default:false" json:"is_public"`
}

// 中间表结构定义

// UserRole 用户角色关联表
type UserRole struct {
	UserID    uint `gorm:"primaryKey"`
	RoleID    uint `gorm:"primaryKey"`
	CreatedAt time.Time
}

// RolePermission 角色权限关联表
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
	CreatedAt    time.Time
}

// EmployeeSkill 员工技能关联表
type EmployeeSkill struct {
	EmployeeID uint `gorm:"primaryKey"`
	SkillID    uint `gorm:"primaryKey"`
	Level      int  `gorm:"default:1"` // 技能等级 1-5
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TaskSkill 任务技能关联表
type TaskSkill struct {
	TaskID    uint `gorm:"primaryKey"`
	SkillID   uint `gorm:"primaryKey"`
	Required  bool `gorm:"default:true"` // 是否必需
	Level     int  `gorm:"default:1"`    // 所需技能等级
	CreatedAt time.Time
}

// GetAllModels 返回所有模型，用于数据库迁移
func GetAllModels() []interface{} {
	return []interface{}{
		&User{},
		&Role{},
		&Permission{},
		&UserRole{},
		&RolePermission{},
		&Department{},
		&Position{},
		&Project{},
		&Task{},
		&Employee{},
		&Skill{},
		&EmployeeSkill{},
		&Assignment{},
		&TaskNotification{},
		&TaskNotificationAction{},
		&WorkflowDefinition{},
		&WorkflowInstance{},
		&WorkflowExecutionHistory{},
		&WorkflowPendingApproval{},
		&AuditLog{},
		&SystemConfig{},
		&OnboardingHistory{},
		// 权限分配相关模型
		&PermissionTemplate{},
		&PermissionRule{},
		&PermissionAssignment{},
		&PermissionAssignmentHistory{},
		&OnboardingPermissionConfig{},
		&Notification{},
	}
}

// OnboardingHistory 入职工作流历史记录表
type OnboardingHistory struct {
	BaseModel
	EmployeeID    uint       `gorm:"not null;index" json:"employee_id"`
	FromStatus    string     `gorm:"size:30" json:"from_status"`
	ToStatus      string     `gorm:"size:30;not null" json:"to_status"`
	OperatorID    uint       `gorm:"not null;index" json:"operator_id"`
	Reason        string     `gorm:"size:255" json:"reason,omitempty"`
	Notes         string     `gorm:"type:text" json:"notes,omitempty"`
	EffectiveDate *time.Time `json:"effective_date,omitempty"`

	// 关联关系
	Employee Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Operator User     `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
}
