package service

import (
	"context"

	"taskmanage/internal/models"
	"taskmanage/internal/repository"
	"taskmanage/internal/workflow"
)

// UserService 用户服务接口
type UserService interface {
	// 用户管理
	CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	GetUser(ctx context.Context, userID uint) (*UserResponse, error)
	UpdateUser(ctx context.Context, userID uint, req *UpdateUserRequest) (*UserResponse, error)
	DeleteUser(ctx context.Context, userID uint) error
	ListUsers(ctx context.Context, filter repository.ListFilter) ([]*UserResponse, int64, error)

	// 认证相关
	AuthenticateUser(ctx context.Context, username, password string) (*UserResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	Logout(ctx context.Context, userID uint) error
	ChangePassword(ctx context.Context, userID uint, req *ChangePasswordRequest) error
	HasPermission(ctx context.Context, userID uint, resource, action string) (bool, error)

	// 角色权限
	AssignRoles(ctx context.Context, userID uint, roleIDs []uint) error
	RemoveRoles(ctx context.Context, userID uint, roleIDs []uint) error
	GetUserPermissions(ctx context.Context, userID uint) ([]*PermissionResponse, error)
}

// TaskService 任务服务接口
type TaskService interface {
	// 任务管理
	CreateTask(ctx context.Context, req *CreateTaskRequest) (*TaskResponse, error)
	GetTask(ctx context.Context, taskID uint) (*TaskResponse, error)
	UpdateTask(ctx context.Context, taskID uint, req *UpdateTaskRequest) (*TaskResponse, error)
	DeleteTask(ctx context.Context, taskID uint) error
	ListTasks(ctx context.Context, filter TaskListFilter) ([]*TaskResponse, int64, error)

	// 任务分配
	AssignTask(ctx context.Context, req *AssignTaskRequest) (*AssignmentResponse, error)
	ReassignTask(ctx context.Context, taskID uint, req *ReassignTaskRequest) (*AssignmentResponse, error)
	ApproveAssignment(ctx context.Context, assignmentID uint, req *ApproveAssignmentRequest) error
	RejectAssignment(ctx context.Context, assignmentID uint, req *RejectAssignmentRequest) error

	// 任务状态管理
	StartTask(ctx context.Context, taskID uint, userID uint) error
	CompleteTask(ctx context.Context, taskID uint, userID uint, req *CompleteTaskRequest) error
	CancelTask(ctx context.Context, taskID uint, userID uint, reason string) error

	// 智能分配
	AutoAssignTask(ctx context.Context, taskID uint, strategy AssignmentStrategy) (*AssignmentResponse, error)
	GetAssignmentSuggestions(ctx context.Context, taskID uint) ([]*AssignmentSuggestion, error)

	// 工作流集成
	CompleteTaskAssignmentWorkflow(ctx context.Context, workflowInstanceID string, approved bool, approverID uint) error
}

// EmployeeService 员工服务接口
type EmployeeService interface {
	// 员工管理
	CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*EmployeeResponse, error)
	GetEmployee(ctx context.Context, employeeID uint) (*EmployeeResponse, error)
	UpdateEmployee(ctx context.Context, employeeID uint, req *UpdateEmployeeRequest) (*EmployeeResponse, error)
	DeleteEmployee(ctx context.Context, employeeID uint) error
	ListEmployees(ctx context.Context, filter EmployeeListFilter) ([]*EmployeeResponse, int64, error)
	GetAvailableEmployees(ctx context.Context) ([]*EmployeeResponse, error)
	GetEmployeeWorkload(ctx context.Context, employeeID uint) (*WorkloadResponse, error)
	AddSkill(ctx context.Context, employeeID uint, req *SkillRequest) error
	RemoveSkill(ctx context.Context, employeeID uint, req *RemoveSkillRequest) error

	// 员工状态管理
	UpdateEmployeeStatus(ctx context.Context, employeeID uint, status string) error
	GetEmployeesByStatus(ctx context.Context, status string) ([]*EmployeeResponse, error)

	// 工作负载统计
	GetWorkloadStats(ctx context.Context, req *WorkloadStatsRequest) ([]*WorkloadResponse, error)
	GetDepartmentWorkload(ctx context.Context, departmentID uint) (*DepartmentWorkloadResponse, error)
}

// NotificationService 通知服务接口
type NotificationService interface {
	// 创建任务分配通知
	CreateTaskAssignmentNotification(ctx context.Context, taskID, recipientID, senderID uint) error
	// 创建任务状态变更通知
	CreateTaskStatusNotification(ctx context.Context, taskID, recipientID uint, notificationType models.TaskNotificationType, title, content string) error
	// 获取用户通知
	GetUserNotifications(ctx context.Context, userID uint, status string, page, pageSize int) ([]models.TaskNotification, int64, error)
	// 获取通知列表 (为Handler提供)
	ListNotifications(ctx context.Context, userID uint, unreadOnly bool) ([]*NotificationResponse, int64, error)
	// 标记通知为已读
	MarkAsRead(ctx context.Context, notificationID, userID uint) error
	// 批量标记为已读
	MarkAllAsRead(ctx context.Context, userID uint) error
	// 获取未读数量
	GetUnreadCount(ctx context.Context, userID uint) (int64, error)
	// 接受任务通知
	AcceptTaskNotification(ctx context.Context, notificationID, taskID, userID uint, reason *string) error
	// 拒绝任务通知
	RejectTaskNotification(ctx context.Context, notificationID, taskID, userID uint, reason *string) error
}

// AssignmentService 任务分配服务接口
type AssignmentService interface {
	// 手动分配
	ManualAssign(ctx context.Context, req *ManualAssignmentRequest) (*AssignmentHistory, error)

	// 自动分配
	AutoAssign(ctx context.Context, taskID uint, strategy string) (*AssignmentResponse, error)

	// 分配建议
	GetAssignmentSuggestions(ctx context.Context, req *AssignmentSuggestionRequest) ([]*AssignmentSuggestion, error)

	// 分配历史
	GetAssignmentHistory(ctx context.Context, taskID uint) ([]*AssignmentHistory, error)

	// 分配管理
	ReassignTask(ctx context.Context, taskID uint, newEmployeeID uint, reason string, assignedBy uint) error
	CancelAssignment(ctx context.Context, taskID uint, reason string, cancelledBy uint) error

	// 冲突检查
	CheckAssignmentConflicts(ctx context.Context, taskID uint, employeeID uint) ([]*AssignmentConflict, error)
}

// SkillService 技能服务接口
type SkillService interface {
	// 技能管理
	CreateSkill(ctx context.Context, req *CreateSkillRequest) (*SkillResponse, error)
	UpdateSkill(ctx context.Context, id uint, req *UpdateSkillRequest) (*SkillResponse, error)
	DeleteSkill(ctx context.Context, id uint) error
	GetSkill(ctx context.Context, id uint) (*SkillResponse, error)
	ListSkills(ctx context.Context, req *ListSkillsRequest) (*ListSkillsResponse, error)
	GetSkillsByCategory(ctx context.Context, category string) ([]*SkillResponse, error)
	GetAllCategories(ctx context.Context) ([]string, error)
	AssignSkillToEmployee(ctx context.Context, employeeID, skillID uint, level int) error
	RemoveSkillFromEmployee(ctx context.Context, employeeID, skillID uint) error
	GetEmployeeSkills(ctx context.Context, employeeID uint) ([]*SkillResponse, error)
}

// DepartmentService 部门服务接口
type DepartmentService interface {
	CreateDepartment(ctx context.Context, req *CreateDepartmentRequest) (*DepartmentResponse, error)
	UpdateDepartment(ctx context.Context, id uint, req *UpdateDepartmentRequest) (*DepartmentResponse, error)
	DeleteDepartment(ctx context.Context, id uint) error
	GetDepartment(ctx context.Context, id uint) (*DepartmentResponse, error)
	ListDepartments(ctx context.Context, req *ListRequest) (*ListResponse[*DepartmentResponse], error)
	GetDepartmentTree(ctx context.Context) ([]*DepartmentResponse, error)
	GetRootDepartments(ctx context.Context) ([]*DepartmentResponse, error)
	GetSubDepartments(ctx context.Context, departmentID uint) ([]*DepartmentResponse, error)
	UpdateManager(ctx context.Context, departmentID, managerID uint) error
}

// PositionService 职位服务接口
type PositionService interface {
	CreatePosition(ctx context.Context, req *CreatePositionRequest) (*PositionResponse, error)
	UpdatePosition(ctx context.Context, id uint, req *UpdatePositionRequest) (*PositionResponse, error)
	DeletePosition(ctx context.Context, id uint) error
	GetPosition(ctx context.Context, id uint) (*PositionResponse, error)
	ListPositions(ctx context.Context, req *ListRequest) (*ListResponse[*PositionResponse], error)
	GetPositionsByCategory(ctx context.Context, category string) ([]*PositionResponse, error)
	GetPositionsByLevel(ctx context.Context, level int) ([]*PositionResponse, error)
	GetAllCategories(ctx context.Context) ([]string, error)
}

// ProjectService 项目服务接口
type ProjectService interface {
	CreateProject(ctx context.Context, req *CreateProjectRequest) (*ProjectResponse, error)
	UpdateProject(ctx context.Context, id uint, req *UpdateProjectRequest) (*ProjectResponse, error)
	DeleteProject(ctx context.Context, id uint) error
	GetProject(ctx context.Context, id uint) (*ProjectResponse, error)
	ListProjects(ctx context.Context, req *ListRequest) (*ListResponse[*ProjectResponse], error)
	GetProjectsByDepartment(ctx context.Context, departmentID uint) ([]*ProjectResponse, error)
	GetProjectsByManager(ctx context.Context, managerID uint) ([]*ProjectResponse, error)
	GetProjectsByStatus(ctx context.Context, status string) ([]*ProjectResponse, error)
	AddProjectMember(ctx context.Context, projectID uint, req *AddProjectMemberRequest) error
	RemoveProjectMember(ctx context.Context, projectID uint, req *RemoveProjectMemberRequest) error
	GetProjectMembers(ctx context.Context, projectID uint) ([]*EmployeeResponse, error)
	UpdateProjectManager(ctx context.Context, projectID, managerID uint) error
}

// WorkflowService 工作流服务接口
type WorkflowService interface {
	// 工作流定义管理
	CreateWorkflowDefinition(ctx context.Context, req *workflow.CreateWorkflowRequest) (*workflow.WorkflowDefinition, error)
	GetWorkflowDefinitions(ctx context.Context) ([]*workflow.WorkflowDefinition, error)
	GetWorkflowDefinition(ctx context.Context, id string) (*workflow.WorkflowDefinition, error)
	UpdateWorkflowDefinition(ctx context.Context, id string, req *workflow.UpdateWorkflowRequest) (*workflow.WorkflowDefinition, error)
	DeleteWorkflowDefinition(ctx context.Context, id string) error
	ValidateWorkflowDefinition(ctx context.Context, req *workflow.CreateWorkflowRequest) error

	// 启动任务分配审批流程
	StartTaskAssignmentApproval(ctx context.Context, req *workflow.TaskAssignmentApprovalRequest) (*workflow.WorkflowInstance, error)

	// 处理任务分配审批
	ProcessTaskAssignmentApproval(ctx context.Context, req *workflow.ApprovalRequest) (*workflow.ApprovalResult, error)

	// 启动入职审批流程
	StartOnboardingApproval(ctx context.Context, req *workflow.OnboardingApprovalRequest) (*workflow.WorkflowInstance, error)

	// 处理入职审批
	ProcessOnboardingApproval(ctx context.Context, req *workflow.ApprovalRequest) (*workflow.ApprovalResult, error)

	// 获取流程实例
	GetWorkflowInstance(ctx context.Context, instanceID string) (*workflow.WorkflowInstance, error)

	// 获取待审批任务分配
	GetPendingTaskAssignmentApprovals(ctx context.Context, userID uint) ([]*workflow.PendingApproval, error)

	// 获取待审批任务
	GetPendingApprovals(ctx context.Context, userID uint) ([]*workflow.PendingApproval, error)

	// 取消流程
	CancelWorkflow(ctx context.Context, instanceID string, reason string) error

	// 获取流程历史
	GetWorkflowHistory(ctx context.Context, instanceID string) ([]workflow.ExecutionHistory, error)
}

// ServiceManager 服务管理器接口
type ServiceManager interface {
	UserService() UserService
	TaskService() TaskService
	EmployeeService() EmployeeService
	SkillService() SkillService
	NotificationService() NotificationService
	WorkflowService() WorkflowService
	DepartmentService() DepartmentService
	PositionService() PositionService
	ProjectService() ProjectService
	OnboardingService() OnboardingService
	PermissionAssignmentService() PermissionAssignmentService
	HealthCheck(ctx context.Context) error
}
