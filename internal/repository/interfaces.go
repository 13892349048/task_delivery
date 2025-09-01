package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"taskmanage/internal/database"
)

// BaseRepository 基础仓储接口
type BaseRepository[T any] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id uint) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter ListFilter) ([]*T, int64, error)
	Exists(ctx context.Context, id uint) (bool, error)
}

// ListFilter 列表查询过滤器
type ListFilter struct {
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Sort     string                 `json:"sort"`
	Order    string                 `json:"order"`
	Filters  map[string]interface{} `json:"filters"`
}

// UserRepository 用户仓储接口
type UserRepository interface {
	BaseRepository[database.User]
	GetByUsername(ctx context.Context, username string) (*database.User, error)
	GetByEmail(ctx context.Context, email string) (*database.User, error)
	UpdateLastLogin(ctx context.Context, userID uint, ip string) error
	GetUserWithRoles(ctx context.Context, userID uint) (*database.User, error)
	BatchUpdateStatus(ctx context.Context, userIDs []uint, status string) error
	ListUsers(ctx context.Context, page, limit int, conditions map[string]interface{}, keyword string) ([]*database.User, int64, error)
	AssignRoles(ctx context.Context, userID uint, roleIDs []uint) error
	RemoveRoles(ctx context.Context, userID uint, roleIDs []uint) error
	GetUsersByRole(ctx context.Context, role string) ([]*database.User, error)
}

// RoleRepository 角色仓储接口
type RoleRepository interface {
	BaseRepository[database.Role]
	GetByName(ctx context.Context, name string) (*database.Role, error)
	GetRoleWithPermissions(ctx context.Context, roleID uint) (*database.Role, error)
	AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
	RemovePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
}

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	BaseRepository[database.Permission]
	GetByResource(ctx context.Context, resource string) ([]*database.Permission, error)
	GetUserPermissions(ctx context.Context, userID uint) ([]*database.Permission, error)
	GetByName(ctx context.Context, name string) (*database.Permission, error)
}

// EmployeeRepository 员工仓储接口
type EmployeeRepository interface {
	BaseRepository[database.Employee]
	GetByUserID(ctx context.Context, userID uint) (*database.Employee, error)
	GetByEmployeeNo(ctx context.Context, employeeNo string) (*database.Employee, error)
	GetAvailableEmployees(ctx context.Context) ([]*database.Employee, error)
	UpdateTaskCount(ctx context.Context, employeeID uint, delta int) error
	GetEmployeeWithSkills(ctx context.Context, employeeID uint) (*database.Employee, error)
	GetBySkills(ctx context.Context, skillIDs []uint, minLevel int) ([]*database.Employee, error)
	
	// 员工状态管理
	UpdateStatus(ctx context.Context, employeeID uint, status string) error
	GetByStatus(ctx context.Context, status string) ([]*database.Employee, error)
	
	// 部门和批量查询
	GetByDepartment(ctx context.Context, department string) ([]*database.Employee, error)
	GetAll(ctx context.Context) ([]*database.Employee, error)
}

// SkillRepository 技能仓储接口
type SkillRepository interface {
	BaseRepository[database.Skill]
	GetByCategory(ctx context.Context, category string) ([]*database.Skill, error)
	GetByName(ctx context.Context, name string) (*database.Skill, error)
	AssignToEmployee(ctx context.Context, employeeID, skillID uint, level int) error
	RemoveFromEmployee(ctx context.Context, employeeID, skillID uint) error
	GetAllCategories(ctx context.Context) ([]string, error)
	GetEmployeeSkills(ctx context.Context, employeeID uint) ([]*database.Skill, error)
	GetEmployeeSkillLevel(ctx context.Context, employeeID, skillID uint) (int, error)
}

// DepartmentRepository 部门仓储接口
type DepartmentRepository interface {
	BaseRepository[database.Department]
	GetByName(ctx context.Context, name string) (*database.Department, error)
	GetByParentID(ctx context.Context, parentID uint) ([]*database.Department, error)
	GetRootDepartments(ctx context.Context) ([]*database.Department, error)
	GetDepartmentTree(ctx context.Context) ([]*database.Department, error)
	GetDepartmentWithManager(ctx context.Context, id uint) (*database.Department, error)
	UpdateManager(ctx context.Context, departmentID, managerID uint) error
	GetSubDepartments(ctx context.Context, departmentID uint) ([]*database.Department, error)
}

// PositionRepository 职位仓储接口
type PositionRepository interface {
	BaseRepository[database.Position]
	GetByName(ctx context.Context, name string) (*database.Position, error)
	GetByCategory(ctx context.Context, category string) ([]*database.Position, error)
	GetByLevel(ctx context.Context, level int) ([]*database.Position, error)
	GetAllCategories(ctx context.Context) ([]string, error)
}

// ProjectRepository 项目仓储接口
type ProjectRepository interface {
	BaseRepository[database.Project]
	GetByName(ctx context.Context, name string) (*database.Project, error)
	GetByDepartmentID(ctx context.Context, departmentID uint) ([]*database.Project, error)
	GetByManagerID(ctx context.Context, managerID uint) ([]*database.Project, error)
	GetByStatus(ctx context.Context, status string) ([]*database.Project, error)
	GetProjectWithMembers(ctx context.Context, id uint) (*database.Project, error)
	AddMember(ctx context.Context, projectID, employeeID uint) error
	RemoveMember(ctx context.Context, projectID, employeeID uint) error
	GetProjectMembers(ctx context.Context, projectID uint) ([]*database.Employee, error)
	UpdateManager(ctx context.Context, projectID, managerID uint) error
}

// TaskRepository 任务仓储接口
type TaskRepository interface {
	BaseRepository[database.Task]
	GetByStatus(ctx context.Context, status string) ([]*database.Task, error)
	GetByAssignee(ctx context.Context, assigneeID uint, status string) ([]*database.Task, error)
	GetByCreator(ctx context.Context, creatorID uint) ([]*database.Task, error)
	GetOverdueTasks(ctx context.Context) ([]*database.Task, error)
	GetTaskWithDetails(ctx context.Context, taskID uint) (*database.Task, error)
	UpdateStatus(ctx context.Context, taskID uint, status string) error
	AssignTask(ctx context.Context, taskID, assigneeID uint) error
	GetTasksByPriority(ctx context.Context, priority string) ([]*database.Task, error)
	GetTasksInDateRange(ctx context.Context, start, end time.Time) ([]*database.Task, error)
	
	// Assignment management methods
	GetActiveTasksByEmployee(ctx context.Context, employeeID uint) ([]*database.Task, error)
	UpdateAssignee(ctx context.Context, taskID, assigneeID uint) error
}

// AssignmentRepository 任务分配仓储接口
type AssignmentRepository interface {
	BaseRepository[database.Assignment]
	GetByTask(ctx context.Context, taskID uint) ([]*database.Assignment, error)
	GetByTaskID(ctx context.Context, taskID uint) ([]*database.Assignment, error)
	GetActiveByTaskID(ctx context.Context, taskID uint) (*database.Assignment, error)
	GetByAssignee(ctx context.Context, assigneeID uint, status string) ([]*database.Assignment, error)
	GetPendingAssignments(ctx context.Context) ([]*database.Assignment, error)
	ApproveAssignment(ctx context.Context, assignmentID, approverID uint, reason string) error
	RejectAssignment(ctx context.Context, assignmentID, approverID uint, reason string) error
	GetAssignmentHistory(ctx context.Context, taskID uint) ([]*database.Assignment, error)
}

// NotificationRepository 通知仓储接口
type NotificationRepository interface {
	BaseRepository[database.TaskNotification]
	GetUserNotifications(ctx context.Context, userID uint, status string, page, pageSize int) ([]*database.TaskNotification, int64, error)
	MarkAsRead(ctx context.Context, notificationID, userID uint) error
	MarkAllAsRead(ctx context.Context, userID uint) (int64, error)
	GetUnreadCount(ctx context.Context, userID uint) (int64, error)
	CreateTaskAssignmentNotification(ctx context.Context, taskID, recipientID, senderID uint) error
	UpdateNotificationStatus(ctx context.Context, notificationID, userID uint, status string) error
	AcceptTaskNotification(ctx context.Context, notificationID, taskID, userID uint, reason *string) error
	RejectTaskNotification(ctx context.Context, notificationID, taskID, userID uint, reason *string) error
}

// AuditLogRepository 审计日志仓储接口
type AuditLogRepository interface {
	BaseRepository[database.AuditLog]
	GetByUser(ctx context.Context, userID uint, start, end time.Time) ([]*database.AuditLog, error)
	GetByResource(ctx context.Context, resource string, resourceID uint) ([]*database.AuditLog, error)
	GetByAction(ctx context.Context, action string, start, end time.Time) ([]*database.AuditLog, error)
	CleanupOldLogs(ctx context.Context, before time.Time) error
}

// SystemConfigRepository 系统配置仓储接口
type SystemConfigRepository interface {
	BaseRepository[database.SystemConfig]
	GetByKey(ctx context.Context, key string) (*database.SystemConfig, error)
	GetByCategory(ctx context.Context, category string) ([]*database.SystemConfig, error)
	GetPublicConfigs(ctx context.Context) ([]*database.SystemConfig, error)
	SetValue(ctx context.Context, key, value string) error
	BatchSet(ctx context.Context, configs map[string]string) error
}

// WorkflowRepository 工作流定义仓库接口
type WorkflowRepository interface {
	// GetWorkflowDefinition 获取流程定义
	GetWorkflowDefinition(ctx context.Context, workflowID string) (*database.WorkflowDefinition, error)
	
	// SaveWorkflowDefinition 保存流程定义
	SaveWorkflowDefinition(ctx context.Context, definition *database.WorkflowDefinition) error
	
	// ListWorkflowDefinitions 列出流程定义
	ListWorkflowDefinitions(ctx context.Context, isActive *bool, limit, offset int) ([]*database.WorkflowDefinition, error)
	
	// UpdateWorkflowStatus 更新流程状态
	UpdateWorkflowStatus(ctx context.Context, workflowID string, isActive bool) error
}

// WorkflowInstanceRepository 工作流实例仓库接口
type WorkflowInstanceRepository interface {
	// SaveInstance 保存流程实例
	SaveInstance(ctx context.Context, instance *database.WorkflowInstance) error
	
	// GetInstance 获取流程实例
	GetInstance(ctx context.Context, instanceID string) (*database.WorkflowInstance, error)
	
	// UpdateInstanceStatus 更新实例状态
	UpdateInstanceStatus(ctx context.Context, instanceID string, status string) error
	
	// UpdateInstance 更新实例
	UpdateInstance(ctx context.Context, instance *database.WorkflowInstance) error
	
	// UpdateInstanceNodes 更新实例当前节点
	UpdateInstanceNodes(ctx context.Context, instanceID string, currentNodes []string) error
	
	// AddExecutionHistory 添加执行历史
	AddExecutionHistory(ctx context.Context, history *database.WorkflowExecutionHistory) error
	
	// GetExecutionHistory 获取执行历史
	GetExecutionHistory(ctx context.Context, instanceID string) ([]*database.WorkflowExecutionHistory, error)
	
	// GetPendingApprovals 获取待审批任务
	GetPendingApprovals(ctx context.Context, userID uint) ([]*database.WorkflowPendingApproval, error)
	
	// CreatePendingApproval 创建待审批任务
	CreatePendingApproval(ctx context.Context, approval *database.WorkflowPendingApproval) error
	
	// CompletePendingApproval 完成待审批任务
	CompletePendingApproval(ctx context.Context, instanceID, nodeID string, userID uint) error
	
	// SavePendingApproval 保存待审批记录
	SavePendingApproval(ctx context.Context, approval *database.WorkflowPendingApproval) error
	
	// DeletePendingApproval 删除待审批记录
	DeletePendingApproval(ctx context.Context, instanceID, nodeID string, userID uint) error
	
	// GetInstancesByBusinessID 根据业务ID获取实例
	GetInstancesByBusinessID(ctx context.Context, businessID, businessType string) ([]*database.WorkflowInstance, error)
}

// OnboardingHistoryRepository 入职历史仓储接口
type OnboardingHistoryRepository interface {
	// Create 创建入职历史记录
	Create(ctx context.Context, history *database.OnboardingHistory) error
	
	// GetByEmployeeID 根据员工ID获取历史记录
	GetByEmployeeID(ctx context.Context, employeeID uint) ([]*database.OnboardingHistory, error)
	
	// GetByID 根据ID获取历史记录
	GetByID(ctx context.Context, id uint) (*database.OnboardingHistory, error)
	
	// List 获取历史记录列表
	List(ctx context.Context, filter *OnboardingHistoryFilter) ([]*database.OnboardingHistory, error)
}

// OnboardingHistoryFilter 入职历史过滤器
type OnboardingHistoryFilter struct {
	Page       int
	PageSize   int
	EmployeeID *uint
	FromDate   *time.Time
	ToDate     *time.Time
}

// RepositoryManager 仓储管理器接口
type RepositoryManager interface {
	UserRepository() UserRepository
	RoleRepository() RoleRepository
	
	// OnboardingHistoryRepository 入职历史仓储接口
	OnboardingHistoryRepository() OnboardingHistoryRepository
	TaskRepository() TaskRepository
	EmployeeRepository() EmployeeRepository
	AssignmentRepository() AssignmentRepository
	NotificationRepository() NotificationRepository
	WorkflowRepository() WorkflowRepository
	SkillRepository() SkillRepository
	PermissionRepository() PermissionRepository
	DepartmentRepository() DepartmentRepository
	PositionRepository() PositionRepository
	ProjectRepository() ProjectRepository
	AuditLogRepository() AuditLogRepository
	SystemConfigRepository() SystemConfigRepository
	WorkflowInstanceRepository() WorkflowInstanceRepository
	
	// 权限分配相关仓储
	PermissionTemplateRepository() PermissionTemplateRepository
	PermissionRuleRepository() PermissionRuleRepository
	PermissionAssignmentRepository() PermissionAssignmentRepository
	PermissionAssignmentHistoryRepository() PermissionAssignmentHistoryRepository
	OnboardingPermissionConfigRepository() OnboardingPermissionConfigRepository
	
	// 事务支持
	WithTx(ctx context.Context, fn func(ctx context.Context, repos RepositoryManager) error) error
	
	// 健康检查
	HealthCheck(ctx context.Context) error
	
	// 获取数据库连接
	GetDB() *gorm.DB
}
