package service

import (
	"time"

	"taskmanage/internal/database"
)

// 用户相关DTO
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	RealName string `json:"real_name" binding:"required"`
}

type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty" binding:"omitempty,email"`
	RealName *string `json:"real_name,omitempty"`
	Status   *string `json:"status,omitempty"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	RealName  string    `json:"real_name"`
	Status    string    `json:"status"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         UserResponse `json:"user"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// 通用请求和响应类型
type ListRequest struct {
	Page     int    `json:"page" form:"page" binding:"omitempty,min=1"`
	PageSize int    `json:"page_size" form:"page_size" binding:"omitempty,min=1,max=100"`
	Sort     string `json:"sort" form:"sort"`
	Order    string `json:"order" form:"order" binding:"omitempty,oneof=asc desc"`
	Keyword  string `json:"keyword" form:"keyword"`
}

type ListResponse[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

// 任务分配审批相关DTO
type StartTaskAssignmentApprovalRequest struct {
	TaskID         uint   `json:"task_id" binding:"required"`
	AssigneeID     uint   `json:"assignee_id" binding:"required"`
	AssignmentType string `json:"assignment_type" binding:"required"`
	Priority       string `json:"priority"`
	RequesterID    uint   `json:"requester_id" binding:"required"`
	Reason         string `json:"reason,omitempty"`
}

type ProcessTaskAssignmentApprovalRequest struct {
	InstanceID string                 `json:"instance_id" binding:"required"`
	NodeID     string                 `json:"node_id" binding:"required"`
	Action     string                 `json:"action" binding:"required"`
	Comment    string                 `json:"comment,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	ApprovedBy uint                   `json:"approved_by" binding:"required"`
}

type TaskAssignmentApprovalResponse struct {
	WorkflowInstanceID string    `json:"workflow_instance_id"`
	Status             string    `json:"status"`
	Message            string    `json:"message"`
	CreatedAt          time.Time `json:"created_at"`
}

type PendingTaskAssignmentApproval struct {
	InstanceID     string     `json:"instance_id"`
	WorkflowName   string     `json:"workflow_name"`
	NodeID         string     `json:"node_id"`
	NodeName       string     `json:"node_name"`
	TaskID         string     `json:"task_id"`
	Priority       int        `json:"priority"`
	AssignedTo     uint       `json:"assigned_to"`
	CreatedAt      time.Time  `json:"created_at"`
	Deadline       *time.Time `json:"deadline,omitempty"`
	CanDelegate    bool       `json:"can_delegate"`
	RequiredAction []string   `json:"required_actions"`
}

type PermissionResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// 任务相关DTO
type CreateTaskRequest struct {
	Title          string    `json:"title" binding:"required"`
	Description    string    `json:"description"`
	Priority       string    `json:"priority" binding:"required,oneof=low medium high urgent"`
	DueDate        time.Time `json:"due_date"`
	RequiredSkills []string  `json:"required_skills"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Priority    *string    `json:"priority,omitempty" binding:"omitempty,oneof=low medium high urgent"`
	Status      *string    `json:"status,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

type TaskResponse struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	DueDate     *time.Time `json:"due_date"`
	CreatedBy   uint       `json:"created_by"`
	AssignedTo  *uint      `json:"assigned_to,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type AssignTaskRequest struct {
	TaskID     uint   `json:"task_id"`
	AssigneeID uint   `json:"assignee_id" binding:"required"` // 被分配人ID
	Method     string `json:"method,omitempty"`               // 分配方式
	Reason     string `json:"reason,omitempty"`               // 分配原因
}

type ReassignTaskRequest struct {
	FromEmployeeID uint   `json:"from_employee_id" binding:"required"`
	ToEmployeeID   uint   `json:"to_employee_id" binding:"required"`
	Reason         string `json:"reason" binding:"required"`
}

type ApproveAssignmentRequest struct {
	Comment string `json:"comment"`
}

type RejectAssignmentRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type CompleteTaskRequest struct {
	Comment string   `json:"comment"`
	Files   []string `json:"files"`
}

type AssignmentResponse struct {
	ID                 uint      `json:"id"`
	TaskID             uint      `json:"task_id"`
	EmployeeID         uint      `json:"employee_id"`
	Status             string    `json:"status"`
	AssignedBy         uint      `json:"assigned_by"`
	AssignedAt         time.Time `json:"assigned_at"`
	WorkflowInstanceID string    `json:"workflow_instance_id,omitempty"` // 工作流实例ID
	Comment            string    `json:"comment"`
}

// type AssignmentSuggestion struct {
// 	EmployeeID   uint    `json:"employee_id"`
// 	EmployeeName string  `json:"employee_name"`
// 	Score        float64 `json:"score"`
// 	Reason       string  `json:"reason"`
// 	Workload     int     `json:"workload"`
// }

type AssignmentStrategy string

const (
	StrategyRoundRobin    AssignmentStrategy = "round_robin"
	StrategyLoadBalancing AssignmentStrategy = "load_balancing"
	StrategySkillMatching AssignmentStrategy = "skill_matching"
	StrategyPriorityBased AssignmentStrategy = "priority_based"
)

// 员工相关DTO
type CreateEmployeeRequest struct {
	Name               string           `json:"name" binding:"required"`
	Email              string           `json:"email" binding:"required,email"`
	DepartmentID       uint             `json:"department_id" binding:"required"`
	PositionID         uint             `json:"position_id" binding:"required"`
	Projects           []string         `json:"projects"`
	Skills             []SkillRequest   `json:"skills"`
	MaxConcurrentTasks int              `json:"max_concurrent_tasks"`
	WorkHours          WorkHoursRequest `json:"work_hours"`
}

type UpdateEmployeeRequest struct {
	Name               *string           `json:"name"`
	Email              *string           `json:"email"`
	DepartmentID       *uint             `json:"department_id"`
	PositionID         *uint             `json:"position_id"`
	Projects           []string          `json:"projects"`
	MaxConcurrentTasks *int              `json:"max_concurrent_tasks"`
	WorkHours          *WorkHoursRequest `json:"work_hours"`
}

type EmployeeResponse struct {
	ID                 uint            `json:"id"`
	Name               string          `json:"name"`
	Email              string          `json:"email"`
	Department         string          `json:"department"`
	Position           string          `json:"position"`
	Status             string          `json:"status"`
	Projects           []string        `json:"projects"`
	Skills             []SkillResponse `json:"skills"`
	MaxConcurrentTasks int             `json:"max_concurrent_tasks"`
	CurrentTasks       int             `json:"current_tasks"`
	CreatedAt          string          `json:"created_at"`
	UpdatedAt          string          `json:"updated_at"`
}

type SkillRequest struct {
	SkillID uint `json:"skill_id" binding:"required"`
	Level   int  `json:"level" binding:"required,min=1,max=5"`
}

type WorkHoursRequest struct {
	Start    string `json:"start" binding:"required"`
	End      string `json:"end" binding:"required"`
	Timezone string `json:"timezone" binding:"required"`
}

type AddSkillRequest struct {
	Name  string `json:"name" binding:"required"`
	Level int    `json:"level" binding:"required,min=1,max=5"`
}

type RemoveSkillRequest struct {
	SkillID uint `json:"skill_id" binding:"required"`
}

// 员工列表过滤器已在下方定义

// 技能相关DTO
type CreateSkillRequest struct {
	Name        string   `json:"name" binding:"required,min=1,max=50"`
	Category    string   `json:"category" binding:"required,min=1,max=50"`
	Description string   `json:"description" binding:"max=255"`
	Tags        []string `json:"tags"`
}

type UpdateSkillRequest struct {
	Name        *string   `json:"name,omitempty" binding:"omitempty,min=1,max=50"`
	Category    *string   `json:"category,omitempty" binding:"omitempty,min=1,max=50"`
	Description *string   `json:"description,omitempty" binding:"omitempty,max=255"`
	Tags        *[]string `json:"tags,omitempty"`
}

type SkillResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	//Tags        []string `json:"tags"`
	Level     int    `json:"level,omitempty"` // 员工技能级别 (1-5)
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// 员工状态枚举
const (
	// 入职流程状态
	EmployeeStatusPendingOnboard = "pending_onboard" // 待入职（HR已创建档案，等待报到）
	EmployeeStatusOnboarding     = "onboarding"      // 入职中（已报到，完成入职手续）
	EmployeeStatusProbation      = "probation"       // 试用期（已分配部门和领导，试用期内）

	// 正式员工状态
	EmployeeStatusActive   = "active"   // 在职（试用期通过，正式员工）
	EmployeeStatusInactive = "inactive" // 离职（已办理离职手续）

	// 工作状态
	EmployeeStatusOnLeave   = "on_leave"  // 请假（临时状态）
	EmployeeStatusBusy      = "busy"      // 忙碌（任务较多）
	EmployeeStatusAvailable = "available" // 空闲（可接受新任务）

	// 特殊状态
	EmployeeStatusSuspended    = "suspended"    // 停职（纪律处分等）
	EmployeeStatusTransferring = "transferring" // 调岗中（部门或职位变更）
)

type EmployeeListFilter struct {
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
	Department   string `json:"department"`
	Position     string `json:"position"`
	Status       string `json:"status"`
	SkillKeyword string `json:"skill_keyword"`
	Available    *bool  `json:"available,omitempty"` // 是否可用
}

type SkillListFilter struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Category string `json:"category"`
}

type AssignSkillRequest struct {
	EmployeeID uint `json:"employee_id" binding:"required"`
	SkillID    uint `json:"skill_id" binding:"required"`
	Level      int  `json:"level" binding:"required,min=1,max=5"`
}

type EmployeeSkillResponse struct {
	SkillID     uint   `json:"skill_id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Level       int    `json:"level"`
}

type ListSkillsRequest struct {
	ListRequest
	Category string `json:"category" form:"category"`
}

type ListSkillsResponse struct {
	ListResponse[*SkillResponse]
}

// 员工相关DTO
type EmployeeSkill struct {
	SkillID uint `json:"skill_id" binding:"required"`
	Level   int  `json:"level" binding:"required,min=1,max=5"`
}

// 部门相关DTO
type CreateDepartmentRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description"`
	ParentID    *uint  `json:"parent_id"`
	ManagerID   *uint  `json:"manager_id"`
	Path        string `json:"path"`
}

type UpdateDepartmentRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=100"`
	Description *string `json:"description,omitempty"`
	ParentID    *uint   `json:"parent_id,omitempty"`
	ManagerID   *uint   `json:"manager_id,omitempty"`
	Path        *string `json:"path,omitempty"`
}

type DepartmentResponse struct {
	ID          uint                  `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	ParentID    *uint                 `json:"parent_id"`
	ManagerID   *uint                 `json:"manager_id"`
	Path        string                `json:"path"`
	Manager     *EmployeeResponse     `json:"manager,omitempty"`
	Parent      *DepartmentResponse   `json:"parent,omitempty"`
	Children    []*DepartmentResponse `json:"children,omitempty"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at"`
}

// 职位相关DTO
type CreatePositionRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description"`
	Category    string `json:"category" binding:"required,max=50"`
	Level       int    `json:"level" binding:"required,min=1,max=10"`
}

type UpdatePositionRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=100"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty" binding:"omitempty,max=50"`
	Level       *int    `json:"level,omitempty" binding:"omitempty,min=1,max=10"`
}

type PositionResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Level       int    `json:"level"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// 项目相关DTO
type CreateProjectRequest struct {
	Name         string     `json:"name" binding:"required,max=200"`
	Description  string     `json:"description"`
	Status       string     `json:"status" binding:"required,oneof=planning active paused completed cancelled"`
	Priority     string     `json:"priority" binding:"required,oneof=low medium high urgent"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Budget       float64    `json:"budget"`
	DepartmentID uint       `json:"department_id" binding:"required"`
	ManagerID    uint       `json:"manager_id" binding:"required"`
}

type UpdateProjectRequest struct {
	Name         *string    `json:"name,omitempty" binding:"omitempty,max=200"`
	Description  *string    `json:"description,omitempty"`
	Status       *string    `json:"status,omitempty" binding:"omitempty,oneof=planning active paused completed cancelled"`
	Priority     *string    `json:"priority,omitempty" binding:"omitempty,oneof=low medium high urgent"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Budget       *float64   `json:"budget,omitempty"`
	DepartmentID *uint      `json:"department_id,omitempty"`
	ManagerID    *uint      `json:"manager_id,omitempty"`
}

type ProjectResponse struct {
	ID           uint                `json:"id"`
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	Status       string              `json:"status"`
	Priority     string              `json:"priority"`
	StartDate    *time.Time          `json:"start_date"`
	EndDate      *time.Time          `json:"end_date"`
	Budget       float64             `json:"budget"`
	DepartmentID uint                `json:"department_id"`
	ManagerID    uint                `json:"manager_id"`
	Department   *DepartmentResponse `json:"department,omitempty"`
	Manager      *EmployeeResponse   `json:"manager,omitempty"`
	Members      []*EmployeeResponse `json:"members,omitempty"`
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
}

type AddProjectMemberRequest struct {
	EmployeeID uint `json:"employee_id" binding:"required"`
}

type RemoveProjectMemberRequest struct {
	EmployeeID uint `json:"employee_id" binding:"required"`
}

type WorkloadResponse struct {
	EmployeeID      uint    `json:"employee_id"`
	EmployeeName    string  `json:"employee_name"`
	Department      string  `json:"department"`
	ActiveTasks     int     `json:"active_tasks"`
	PendingTasks    int     `json:"pending_tasks"`
	CompletedTasks  int     `json:"completed_tasks"`
	OverdueTasks    int     `json:"overdue_tasks"`
	MaxTasks        int     `json:"max_tasks"`
	WorkloadRate    float64 `json:"workload_rate"`     // 工作负载率 (0-1)
	EfficiencyRate  float64 `json:"efficiency_rate"`   // 效率率
	AvgTaskDuration float64 `json:"avg_task_duration"` // 平均任务完成时间(小时)
	Status          string  `json:"status"`            // 员工状态
	LastActiveTime  string  `json:"last_active_time"`  // 最后活跃时间
}

// 工作负载统计请求
type WorkloadStatsRequest struct {
	EmployeeIDs  []uint `json:"employee_ids,omitempty"`
	Department   string `json:"department,omitempty"`
	DepartmentID uint   `json:"department_id,omitempty"`
	StartDate    string `json:"start_date,omitempty"`
	EndDate      string `json:"end_date,omitempty"`
}

// 部门工作负载统计
type DepartmentWorkloadResponse struct {
	Department      string  `json:"department"`
	TotalEmployees  int     `json:"total_employees"`
	ActiveEmployees int     `json:"active_employees"`
	TotalTasks      int     `json:"total_tasks"`
	CompletedTasks  int     `json:"completed_tasks"`
	AvgWorkloadRate float64 `json:"avg_workload_rate"`
	OverloadedCount int     `json:"overloaded_count"` // 超负荷员工数量
}

// 通知相关DTO
type SendNotificationRequest struct {
	UserID  uint   `json:"user_id" binding:"required"`
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Type    string `json:"type" binding:"required,oneof=info warning error success"`
}

type BroadcastNotificationRequest struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Type       string `json:"type" binding:"required,oneof=info warning error success"`
	UserIDs    []uint `json:"user_ids,omitempty"`
	Department string `json:"department,omitempty"`
}

// 通知相关DTO
type NotificationResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// CancelTaskRequest 取消任务请求
type CancelTaskRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// AutoAssignRequest 自动分配请求
type AutoAssignRequest struct {
	Strategy string `json:"strategy"` // workload, skill, random
}

// 过滤器
type TaskListFilter struct {
	Status     string `form:"status"`
	Priority   string `form:"priority"`
	AssignedTo *uint  `form:"assigned_to"`
	CreatedBy  *uint  `form:"created_by"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}

// 转换函数
func UserToResponse(user *database.User) *UserResponse {
	return &UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RealName: user.RealName,
		Status:   user.Status,
		Role:     user.Role,
	}
}

func TaskToResponse(task *database.Task) *TaskResponse {
	resp := &TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Priority:    task.Priority,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}

	// 处理可能为nil的时间字段
	if task.DueDate != nil {
		resp.DueDate = task.DueDate
	}

	// 处理创建者信息
	resp.CreatedBy = task.CreatorID

	// 处理分配者信息
	resp.AssignedTo = task.AssigneeID

	return resp
}

func EmployeeToResponse(employee *database.Employee) *EmployeeResponse {
	resp := &EmployeeResponse{
		ID:                 employee.ID,
		Name:               employee.User.RealName,   // 从关联的User获取姓名
		Email:              employee.User.Email,      // 从关联的User获取邮箱
		Department:         employee.Department.Name, // 从关联的Department获取名称
		Position:           employee.Position.Name,   // 从关联的Position获取名称
		Status:             employee.Status,
		Projects:           []string{},        // TODO: 从关联表获取项目列表
		Skills:             []SkillResponse{}, // TODO: 从关联表获取技能列表
		MaxConcurrentTasks: employee.MaxTasks,
		CurrentTasks:       employee.CurrentTasks,
		CreatedAt:          employee.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          employee.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// 转换技能信息 - 包含技能级别
	for _, skill := range employee.Skills {
		skillResp := SkillResponse{
			ID:          skill.ID,
			Name:        skill.Name,
			Category:    skill.Category,
			Description: skill.Description,

			CreatedAt: skill.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: skill.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// TODO: 从 employee_skills 关联表获取技能级别
		// 需要通过 repository 查询 employee_skills 表获取具体的 level 信息

		resp.Skills = append(resp.Skills, skillResp)
	}

	return resp
}

func NotificationToResponse(notification *database.Notification) *NotificationResponse {
	return &NotificationResponse{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Title:     notification.Title,
		Content:   notification.Content,
		Type:      notification.Type,
		IsRead:    notification.IsRead,
		CreatedAt: notification.CreatedAt,
	}
}

// ==================== 入职工作流相关 DTO ====================

// 创建待入职员工请求
type CreatePendingEmployeeRequest struct {
	RealName     string `json:"real_name" binding:"required,min=2,max=50"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"required,min=11,max=15"`
	ExpectedDate string `json:"expected_date" binding:"required"` // 预期入职日期
	DepartmentID *uint  `json:"department_id,omitempty"`          // 可选，预分配部门
	PositionID   *uint  `json:"position_id,omitempty"`            // 可选，预分配职位
	Notes        string `json:"notes,omitempty"`                  // 备注信息
}

// 入职确认请求
type OnboardConfirmRequest struct {
	EmployeeID   uint   `json:"employee_id" binding:"required"`
	DepartmentID uint   `json:"department_id" binding:"required"`
	PositionID   uint   `json:"position_id" binding:"required"`
	ManagerID    *uint  `json:"manager_id,omitempty"`          // 直属领导
	StartDate    string `json:"start_date" binding:"required"` // 正式入职日期
	Notes        string `json:"notes,omitempty"`
}

// 试用期转正请求
type ProbationToActiveRequest struct {
	EmployeeID     uint   `json:"employee_id" binding:"required"`
	EvaluationNote string `json:"evaluation_note,omitempty"` // 试用期评价
	IsApproved     bool   `json:"is_approved" binding:"required"`
	EffectiveDate  string `json:"effective_date" binding:"required"` // 转正生效日期
}

// 员工状态变更请求
type EmployeeStatusChangeRequest struct {
	EmployeeID    uint   `json:"employee_id" binding:"required"`
	NewStatus     string `json:"new_status" binding:"required"`
	Reason        string `json:"reason,omitempty"`
	EffectiveDate string `json:"effective_date,omitempty"`
	Notes         string `json:"notes,omitempty"`
}

// 入职工作流状态响应
type OnboardingWorkflowResponse struct {
	ID            uint   `json:"id"`
	EmployeeID    uint   `json:"employee_id"`
	EmployeeName  string `json:"employee_name"`
	Email         string `json:"email"`
	CurrentStatus string `json:"current_status"`
	Department    string `json:"department,omitempty"`
	Position      string `json:"position,omitempty"`
	Manager       string `json:"manager,omitempty"`
	ExpectedDate  string `json:"expected_date,omitempty"`
	StartDate     string `json:"start_date,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// 入职工作流历史记录
type OnboardingHistoryResponse struct {
	ID           uint   `json:"id"`
	EmployeeID   uint   `json:"employee_id"`
	FromStatus   string `json:"from_status"`
	ToStatus     string `json:"to_status"`
	OperatorID   uint   `json:"operator_id"`
	OperatorName string `json:"operator_name"`
	Reason       string `json:"reason,omitempty"`
	Notes        string `json:"notes,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// 入职工作流列表过滤器
type OnboardingWorkflowFilter struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Status     string `json:"status,omitempty"`
	Department string `json:"department,omitempty"`
	DateFrom   string `json:"date_from,omitempty"`
	DateTo     string `json:"date_to,omitempty"`
}

// 权限分配请求
// type AssignPermissionsRequest struct {
// 	EmployeeID    uint   `json:"employee_id" binding:"required"`
// 	Permissions   []uint `json:"permissions" binding:"required"`   // 权限ID列表
// 	Reason        string `json:"reason,omitempty"`
// 	EffectiveDate string `json:"effective_date,omitempty"`
// }
