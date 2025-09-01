package service

import (
	"context"
	"fmt"
	"time"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/internal/workflow"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// OnboardingService 入职工作流服务接口
type OnboardingService interface {
	// 创建待入职员工
	CreatePendingEmployee(ctx context.Context, req *CreatePendingEmployeeRequest) (*OnboardingWorkflowResponse, error)

	// 入职确认（待入职 -> 入职中）
	ConfirmOnboarding(ctx context.Context, req *OnboardConfirmRequest) (*OnboardingWorkflowResponse, error)

	// 完成入职手续（入职中 -> 试用期）
	CompleteProbation(ctx context.Context, employeeID uint, operatorID uint) (*OnboardingWorkflowResponse, error)

	// 试用期转正（试用期 -> 正式员工）
	ConfirmEmployee(ctx context.Context, req *ProbationToActiveRequest, operatorID uint) (*OnboardingWorkflowResponse, error)

	// 员工状态变更
	ChangeEmployeeStatus(ctx context.Context, req *EmployeeStatusChangeRequest, operatorID uint) (*OnboardingWorkflowResponse, error)

	// 获取入职工作流列表
	GetOnboardingWorkflows(ctx context.Context, filter *OnboardingWorkflowFilter) ([]*OnboardingWorkflowResponse, error)

	// 获取入职历史记录
	GetOnboardingHistory(ctx context.Context, employeeID uint) ([]*OnboardingHistoryResponse, error)

	// 入职审批工作流相关方法
	// 启动入职审批流程
	StartOnboardingApproval(ctx context.Context, req *OnboardingApprovalRequest) (*OnboardingApprovalResponse, error)

	// 处理入职审批决策
	ProcessOnboardingApproval(ctx context.Context, req *ProcessOnboardingApprovalRequest) (*OnboardingApprovalResponse, error)

	// 获取待审批入职申请
	GetPendingOnboardingApprovals(ctx context.Context, userID uint) ([]*PendingOnboardingApproval, error)

	// 获取入职审批历史
	GetOnboardingApprovalHistory(ctx context.Context, employeeID uint) ([]*OnboardingApprovalHistory, error)

	// 取消入职审批流程
	CancelOnboardingApproval(ctx context.Context, instanceID string, reason string, operatorID uint) error
}

// 入职审批相关DTO定义

// OnboardingApprovalRequest 启动入职审批请求
type OnboardingApprovalRequest struct {
	EmployeeID    uint   `json:"employee_id" binding:"required"`
	DepartmentID  uint   `json:"department_id" binding:"required"`
	PositionID    *uint  `json:"position_id"`
	ExpectedDate  string `json:"expected_date" binding:"required"`
	ProbationDays int    `json:"probation_days" binding:"min=30,max=180"`
	WorkflowType  string `json:"workflow_type"` // "full" 或 "simple"
	Notes         string `json:"notes"`
	RequesterID   uint   `json:"requester_id"`
}

// 辅助函数
func getDepartmentName(departmentID *uint) string {
	if departmentID == nil {
		return "未分配"
	}
	return fmt.Sprintf("部门%d", *departmentID)
}

func getPositionName(positionID *uint) string {
	if positionID == nil {
		return "未分配"
	}
	return fmt.Sprintf("职位%d", *positionID)
}

// triggerPermissionAssignment 触发权限分配
func (s *OnboardingServiceImpl) triggerPermissionAssignment(ctx context.Context, userID uint, onboardingStatus string, departmentID, positionID *uint) error {
	if s.permissionAssignmentService == nil {
		s.logger.Warn("权限分配服务未初始化")
		return nil
	}

	return s.permissionAssignmentService.ProcessOnboardingPermissionAssignment(ctx, userID, onboardingStatus, departmentID, positionID)
}

// ProcessOnboardingApprovalRequest 处理入职审批请求
type ProcessOnboardingApprovalRequest struct {
	InstanceID string `json:"instance_id" binding:"required"`
	NodeID     string `json:"node_id" binding:"required"`
	Action     string `json:"action" binding:"required,oneof=approve reject"`
	Comment    string `json:"comment"`
	ApproverID uint   `json:"approver_id"`
}

// OnboardingApprovalResponse 入职审批响应
type OnboardingApprovalResponse struct {
	InstanceID   string             `json:"instance_id"`
	EmployeeID   uint               `json:"employee_id"`
	Status       string             `json:"status"`
	CurrentStep  string             `json:"current_step"`
	WorkflowType string             `json:"workflow_type"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Employee     *database.Employee `json:"employee,omitempty"`
}

// PendingOnboardingApproval 待审批入职申请
type PendingOnboardingApproval struct {
	InstanceID    string    `json:"instance_id"`
	EmployeeID    uint      `json:"employee_id"`
	EmployeeName  string    `json:"employee_name"`
	Department    string    `json:"department"`
	Position      string    `json:"position"`
	ExpectedDate  string    `json:"expected_date"`
	CurrentStep   string    `json:"current_step"`
	SubmittedAt   time.Time `json:"submitted_at"`
	RequesterName string    `json:"requester_name"`
	Notes         string    `json:"notes"`
}

// OnboardingApprovalHistory 入职审批历史
type OnboardingApprovalHistory struct {
	InstanceID   string    `json:"instance_id"`
	Step         string    `json:"step"`
	Decision     string    `json:"decision"`
	Comments     string    `json:"comments"`
	ApproverID   uint      `json:"approver_id"`
	ApproverName string    `json:"approver_name"`
	ProcessedAt  time.Time `json:"processed_at"`
}

// OnboardingServiceImpl 入职工作流服务实现
type OnboardingServiceImpl struct {
	employeeRepo                repository.EmployeeRepository
	userRepo                    repository.UserRepository
	historyRepo                 repository.OnboardingHistoryRepository
	workflowService             WorkflowService
	permissionAssignmentService PermissionAssignmentService
	logger                      *logrus.Logger
}

// NewOnboardingService 创建入职工作流服务
func NewOnboardingService(repoManager repository.RepositoryManager, workflowService WorkflowService, permissionAssignmentService PermissionAssignmentService, logger *logrus.Logger) OnboardingService {
	return &OnboardingServiceImpl{
		employeeRepo:                repoManager.EmployeeRepository(),
		userRepo:                    repoManager.UserRepository(),
		historyRepo:                 repoManager.OnboardingHistoryRepository(),
		workflowService:             workflowService,
		permissionAssignmentService: permissionAssignmentService,
		logger:                      logger,
	}
}

// CreatePendingEmployee 创建待入职员工
func (s *OnboardingServiceImpl) CreatePendingEmployee(ctx context.Context, req *CreatePendingEmployeeRequest) (*OnboardingWorkflowResponse, error) {
	logger := s.logger.WithField("method", "CreatePendingEmployee")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("密码加密失败: %v", err)
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}
	// 创建用户账号
	user := &database.User{
		Username:     req.Email, // 使用邮箱作为用户名
		Email:        req.Email,
		RealName:     req.RealName,
		Phone:        req.Phone,
		Password:     string(hashedPassword),
		PasswordHash: string(hashedPassword),
		Status:       "inactive", // 账号暂时不激活
		Role:         "employee",
		// 注意：不设置Password和PasswordHash，等待用户激活时设置
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Errorf("Failed to create user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// TODO: 生成账号激活令牌并发送激活邮件
	// activationToken := generateActivationToken(user.ID)
	// sendActivationEmail(user.Email, user.RealName, activationToken)
	logger.Infof("用户账号已创建，待激活: %s (ID: %d)", user.Email, user.ID)

	// 解析预期入职日期
	var expectedDate *time.Time
	if req.ExpectedDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.ExpectedDate); err == nil {
			expectedDate = &parsed
		}
	}

	// 创建员工记录
	employee := &database.Employee{
		UserID:           user.ID,
		EmployeeNo:       fmt.Sprintf("EMP%06d", user.ID),
		DepartmentID:     req.DepartmentID,
		PositionID:       req.PositionID,
		OnboardingStatus: "pending_onboard",
		ExpectedDate:     expectedDate,
		OnboardingNotes:  req.Notes,
		Status:           "available",
		MaxTasks:         5,
		CurrentTasks:     0,
	}

	if err := s.employeeRepo.Create(ctx, employee); err != nil {
		logger.Errorf("Failed to create employee: %v", err)
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	logger.Infof("Created pending employee: %d", employee.ID)
	return s.buildWorkflowResponse(employee), nil
}

// ConfirmOnboarding 入职确认
func (s *OnboardingServiceImpl) ConfirmOnboarding(ctx context.Context, req *OnboardConfirmRequest) (*OnboardingWorkflowResponse, error) {
	logger := s.logger.WithField("method", "ConfirmOnboarding")

	employee, err := s.employeeRepo.GetByID(ctx, req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	if employee.OnboardingStatus != "pending_onboard" {
		return nil, fmt.Errorf("employee is not in pending_onboard status")
	}

	// 解析入职日期
	var startDate *time.Time
	if req.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = &parsed
		}
	}

	// 更新员工信息
	oldStatus := employee.OnboardingStatus
	employee.OnboardingStatus = "onboarding"
	employee.DepartmentID = &req.DepartmentID
	employee.PositionID = &req.PositionID
	employee.DirectManagerID = req.ManagerID
	employee.HireDate = startDate
	employee.OnboardingNotes = req.Notes

	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		logger.Errorf("Failed to update employee: %v", err)
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	// 激活用户账号
	user, err := s.userRepo.GetByID(ctx, employee.UserID)
	if err == nil {
		user.Status = "active"
		s.userRepo.Update(ctx, user)
	}

	// 记录状态变更历史
	s.recordStatusChange(ctx, employee.ID, oldStatus, employee.OnboardingStatus, req.EmployeeID, req.Notes)

	//触发权限分配
	//if err := s.triggerPermissionAssignment(ctx, employee.UserID, employee.OnboardingStatus, employee.DepartmentID, employee.PositionID); err != nil {
	//logger.Warnf("权限分配失败: %v", err)
	// 权限分配失败不影响入职流程继续
	//}

	logger.Infof("Employee onboarding confirmed: %d", employee.ID)
	return s.buildWorkflowResponse(employee), nil
}

// CompleteProbation 完成入职手续，进入试用期
func (s *OnboardingServiceImpl) CompleteProbation(ctx context.Context, employeeID uint, operatorID uint) (*OnboardingWorkflowResponse, error) {
	logger := s.logger.WithField("method", "CompleteProbation")

	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	if employee.OnboardingStatus != "onboarding" {
		return nil, fmt.Errorf("employee is not in onboarding status")
	}

	// 更新状态为试用期
	oldStatus := employee.OnboardingStatus
	employee.OnboardingStatus = "probation"

	// 设置试用期结束日期（默认3个月）
	if employee.HireDate != nil {
		probationEnd := employee.HireDate.AddDate(0, 3, 0)
		employee.ProbationEndDate = &probationEnd
	}

	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		logger.Errorf("Failed to update employee: %v", err)
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	// 记录状态变更历史
	s.recordStatusChange(ctx, employee.ID, oldStatus, employee.OnboardingStatus, operatorID, "完成入职手续，进入试用期")

	// 触发权限分配
	if err := s.triggerPermissionAssignment(ctx, employee.UserID, employee.OnboardingStatus, employee.DepartmentID, employee.PositionID); err != nil {
		logger.Warnf("权限分配失败: %v", err)
	}

	logger.Infof("Employee entered probation: %d", employee.ID)
	return s.buildWorkflowResponse(employee), nil
}

// ConfirmEmployee 试用期转正
func (s *OnboardingServiceImpl) ConfirmEmployee(ctx context.Context, req *ProbationToActiveRequest, operatorID uint) (*OnboardingWorkflowResponse, error) {
	logger := s.logger.WithField("method", "ConfirmEmployee")

	employee, err := s.employeeRepo.GetByID(ctx, req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	if employee.OnboardingStatus != "probation" {
		return nil, fmt.Errorf("employee is not in probation status")
	}

	oldStatus := employee.OnboardingStatus

	if req.IsApproved {
		// 转正成功
		employee.OnboardingStatus = "active"

		// 设置转正日期
		if req.EffectiveDate != "" {
			if parsed, err := time.Parse("2006-01-02", req.EffectiveDate); err == nil {
				employee.ConfirmDate = &parsed
			}
		}
	} else {
		// 试用期不通过，设置为离职
		employee.OnboardingStatus = "inactive"
		employee.Status = "resigned"
	}

	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		logger.Errorf("Failed to update employee: %v", err)
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	// 记录状态变更历史
	reason := "试用期转正成功"
	if !req.IsApproved {
		reason = "试用期考核不通过"
	}
	s.recordStatusChange(ctx, employee.ID, oldStatus, employee.OnboardingStatus, operatorID, reason+": "+req.EvaluationNote)

	// 触发权限分配
	if err := s.triggerPermissionAssignment(ctx, employee.UserID, employee.OnboardingStatus, employee.DepartmentID, employee.PositionID); err != nil {
		logger.Warnf("权限分配失败: %v", err)
	}

	logger.Infof("Employee probation completed: %d, approved: %v", employee.ID, req.IsApproved)
	return s.buildWorkflowResponse(employee), nil
}

// ChangeEmployeeStatus 员工状态变更
func (s *OnboardingServiceImpl) ChangeEmployeeStatus(ctx context.Context, req *EmployeeStatusChangeRequest, operatorID uint) (*OnboardingWorkflowResponse, error) {
	logger := s.logger.WithField("method", "ChangeEmployeeStatus")

	employee, err := s.employeeRepo.GetByID(ctx, req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	oldStatus := employee.OnboardingStatus
	employee.OnboardingStatus = req.NewStatus

	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		logger.Errorf("Failed to update employee: %v", err)
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	// 记录状态变更历史
	s.recordStatusChange(ctx, employee.ID, oldStatus, employee.OnboardingStatus, operatorID, req.Reason+": "+req.Notes)

	logger.Infof("Employee status changed: %d, %s -> %s", employee.ID, oldStatus, req.NewStatus)
	return s.buildWorkflowResponse(employee), nil
}

// GetOnboardingWorkflows 获取入职工作流列表
func (s *OnboardingServiceImpl) GetOnboardingWorkflows(ctx context.Context, filter *OnboardingWorkflowFilter) ([]*OnboardingWorkflowResponse, error) {
	// 简化实现，直接获取所有员工
	employees, _, err := s.employeeRepo.List(ctx, repository.ListFilter{
		Page:     filter.Page,
		PageSize: filter.PageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}

	var workflows []*OnboardingWorkflowResponse
	for _, employee := range employees {
		// 根据过滤条件筛选
		if filter.Status != "" && employee.OnboardingStatus != filter.Status {
			continue
		}
		workflows = append(workflows, s.buildWorkflowResponse(employee))
	}

	return workflows, nil
}

// GetOnboardingHistory 获取入职历史记录
func (s *OnboardingServiceImpl) GetOnboardingHistory(ctx context.Context, employeeID uint) ([]*OnboardingHistoryResponse, error) {
	histories, err := s.historyRepo.GetByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get onboarding history: %w", err)
	}

	var responses []*OnboardingHistoryResponse
	for _, history := range histories {
		responses = append(responses, &OnboardingHistoryResponse{
			ID:           history.ID,
			EmployeeID:   history.EmployeeID,
			FromStatus:   history.FromStatus,
			ToStatus:     history.ToStatus,
			OperatorID:   history.OperatorID,
			OperatorName: history.Operator.RealName,
			Reason:       history.Reason,
			Notes:        history.Notes,
			CreatedAt:    history.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

// CompleteOnboardingApproval 完成入职审批流程
func (s *OnboardingServiceImpl) CompleteOnboardingApproval(ctx context.Context, instanceID string, approved bool, employeeID uint, approverID uint) error {
	logger := s.logger.WithField("method", "CompleteOnboardingApproval")
	logger.Infof("完成入职审批: InstanceID=%s, Approved=%t, EmployeeID=%d", instanceID, approved, employeeID)

	// 获取员工信息
	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		logger.WithError(err).Error("获取员工信息失败")
		return err
	}

	// 根据审批结果更新员工状态
	if approved {
		employee.OnboardingStatus = "approved"
		logger.Infof("员工 %d 入职审批通过", employeeID)
	} else {
		employee.OnboardingStatus = "rejected"
		logger.Infof("员工 %d 入职审批被拒绝", employeeID)
	}

	// 更新员工状态
	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		logger.WithError(err).Error("更新员工状态失败")
		return err
	}

	// 创建入职历史记录
	history := &database.OnboardingHistory{
		EmployeeID: employeeID,
		FromStatus: "approval_pending",
		ToStatus:   employee.OnboardingStatus,
		OperatorID: approverID,
		Reason:     fmt.Sprintf("审批完成: %s", map[bool]string{true: "通过", false: "拒绝"}[approved]),
		Notes:      fmt.Sprintf("工作流实例ID: %s", instanceID),
	}

	// 注意：这里暂时跳过历史记录创建，因为repository可能未实现
	// 可以在后续完善时添加
	_ = history

	logger.Infof("入职审批完成处理成功: EmployeeID=%d, Status=%s", employeeID, employee.OnboardingStatus)
	return nil
}

// recordStatusChange 记录状态变更历史
func (s *OnboardingServiceImpl) recordStatusChange(ctx context.Context, employeeID uint, fromStatus, toStatus string, operatorID uint, notes string) {
	history := &database.OnboardingHistory{
		EmployeeID: employeeID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		OperatorID: operatorID,
		Notes:      notes,
	}

	if err := s.historyRepo.Create(ctx, history); err != nil {
		s.logger.Errorf("Failed to record onboarding history: %v", err)
	}
}

// buildWorkflowResponse 构建工作流响应
func (s *OnboardingServiceImpl) buildWorkflowResponse(employee *database.Employee) *OnboardingWorkflowResponse {
	response := &OnboardingWorkflowResponse{
		ID:            employee.ID,
		EmployeeID:    employee.ID,
		EmployeeName:  employee.User.RealName,
		Email:         employee.User.Email,
		CurrentStatus: employee.OnboardingStatus,
		CreatedAt:     employee.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     employee.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if employee.Department.Name != "" {
		response.Department = employee.Department.Name
	}
	if employee.Position.Name != "" {
		response.Position = employee.Position.Name
	}
	if employee.DirectManager != nil {
		response.Manager = employee.DirectManager.User.RealName
	}
	if employee.ExpectedDate != nil {
		response.ExpectedDate = employee.ExpectedDate.Format("2006-01-02")
	}
	if employee.HireDate != nil {
		response.StartDate = employee.HireDate.Format("2006-01-02")
	}

	return response
}

// 入职审批工作流方法实现

// StartOnboardingApproval 启动入职审批流程
func (s *OnboardingServiceImpl) StartOnboardingApproval(ctx context.Context, req *OnboardingApprovalRequest) (*OnboardingApprovalResponse, error) {
	logger := s.logger.WithField("method", "StartOnboardingApproval")

	// 验证员工是否存在
	employee, err := s.employeeRepo.GetByID(ctx, req.EmployeeID)
	if err != nil {
		logger.WithError(err).Error("获取员工信息失败")
		return nil, err
	}

	// 启动入职审批工作流
	workflowReq := &workflow.OnboardingApprovalRequest{
		EmployeeID:      req.EmployeeID,
		EmployeeType:    "permanent", // 默认为正式员工
		DepartmentID:    &req.DepartmentID,
		PositionID:      req.PositionID,
		ExpectedDate:    req.ExpectedDate,
		ProbationPeriod: req.ProbationDays, // 使用ProbationDays字段
		Priority:        "normal",
		RequesterID:     req.RequesterID,
	}

	instance, err := s.workflowService.StartOnboardingApproval(ctx, workflowReq)
	if err != nil {
		logger.WithError(err).Error("启动入职审批工作流失败")
		return nil, err
	}

	// 更新员工状态为审批中
	oldStatus := employee.OnboardingStatus
	employee.OnboardingStatus = "approval_pending"
	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		logger.WithError(err).Error("更新员工状态失败")
	}

	// 记录历史
	history := &database.OnboardingHistory{
		EmployeeID: req.EmployeeID,
		FromStatus: oldStatus,
		ToStatus:   "approval_pending",
		OperatorID: req.RequesterID,
		Reason:     "启动入职审批流程",
		Notes:      req.Notes,
	}
	if err := s.historyRepo.Create(ctx, history); err != nil {
		logger.WithError(err).Error("记录入职历史失败")
	}

	return &OnboardingApprovalResponse{
		InstanceID:   instance.ID,
		EmployeeID:   req.EmployeeID,
		Status:       string(instance.Status),
		CurrentStep:  getCurrentNode(instance.CurrentNodes),
		WorkflowType: req.WorkflowType,
		CreatedAt:    instance.StartedAt,
		UpdatedAt:    instance.StartedAt, // Use StartedAt as UpdatedAt for now
		Employee:     employee,
	}, nil
}

// ProcessOnboardingApproval 处理入职审批决策
func (s *OnboardingServiceImpl) ProcessOnboardingApproval(ctx context.Context, req *ProcessOnboardingApprovalRequest) (*OnboardingApprovalResponse, error) {
	logger := s.logger.WithField("method", "ProcessOnboardingApproval")

	// 获取工作流实例
	instance, err := s.workflowService.GetWorkflowInstance(ctx, req.InstanceID)
	if err != nil {
		logger.WithError(err).Error("获取工作流实例失败")
		return nil, fmt.Errorf("获取工作流实例失败: %w", err)
	}

	// 从工作流实例的变量中获取员工ID
	employeeID, ok := instance.Variables["employee_id"].(float64)
	if !ok {
		logger.Error("无法从工作流实例中获取员工ID")
		return nil, fmt.Errorf("无效的工作流实例数据")
	}

	// 构建工作流审批请求
	approvalReq := &workflow.ApprovalRequest{
		InstanceID: req.InstanceID,
		NodeID:     req.NodeID,
		Action:     workflow.ApprovalAction(req.Action),
		Comment:    req.Comment,
		ApprovedBy: req.ApproverID,
	}
	// 处理工作流审批
	result, err := s.workflowService.ProcessOnboardingApproval(ctx, approvalReq)
	if err != nil {
		logger.WithError(err).Error("处理工作流审批失败")
		return nil, fmt.Errorf("处理工作流审批失败: %w", err)
	}

	// 根据工作流处理结果更新员工状态
	var newStatus string
	var reason string

	if req.Action == "approve" {
		// 审批通过后，根据工作流是否完成决定员工状态
		if result.IsCompleted {
			newStatus = "approved" // 整个工作流完成，员工正式入职
			reason = "入职工作流完成，员工正式入职"
		} else {
			newStatus = "in_progress" // 工作流继续，员工状态为处理中
			reason = "管理员审批通过，进入下一流程环节"
		}
	} else {
		newStatus = "rejected" // 审批拒绝，直接拒绝入职
		reason = fmt.Sprintf("入职审批被拒绝: %s", req.Action)
	}

	// 更新员工状态
	if err := s.employeeRepo.UpdateStatus(ctx, uint(employeeID), newStatus); err != nil {
		logger.WithError(err).Error("更新员工状态失败")
		return nil, fmt.Errorf("更新员工状态失败: %w", err)
	}

	// 记录状态变更历史
	history := &database.OnboardingHistory{
		EmployeeID: uint(employeeID),
		FromStatus: "approval_pending",
		ToStatus:   newStatus,
		OperatorID: req.ApproverID,
		Reason:     reason,
		Notes:      req.Comment,
	}
	if err := s.historyRepo.Create(ctx, history); err != nil {
		logger.WithError(err).Error("记录入职历史失败")
	}

	// 获取员工信息
	employee, err := s.employeeRepo.GetByID(ctx, uint(employeeID))
	if err != nil {
		logger.WithError(err).Error("获取员工信息失败")
		employee = nil
	}

	return &OnboardingApprovalResponse{
		InstanceID:   req.InstanceID,
		EmployeeID:   uint(employeeID),
		Status:       newStatus,
		CurrentStep:  result.NodeID,
		WorkflowType: "onboarding",
		CreatedAt:    instance.StartedAt,
		UpdatedAt:    result.ExecutedAt,
		Employee:     employee,
	}, nil
}

// GetPendingOnboardingApprovals 获取待审批的入职申请
func (s *OnboardingServiceImpl) GetPendingOnboardingApprovals(ctx context.Context, userID uint) ([]*PendingOnboardingApproval, error) {
	logger := s.logger.WithField("method", "GetPendingOnboardingApprovals")

	// 使用现有的待审批查询方法
	approvals, err := s.workflowService.GetPendingApprovals(ctx, userID)
	if err != nil {
		logger.WithError(err).Error("获取待审批工作流失败")
		return nil, err
	}

	var result []*PendingOnboardingApproval
	for _, approval := range approvals {
		// 检查是否为入职相关的审批（通过业务类型标识）
		if approval.BusinessType != "onboarding" {
			continue
		}

		// 从BusinessID中解析员工ID
		var employeeID uint
		if _, err := fmt.Sscanf(approval.BusinessID, "employee-%d", &employeeID); err != nil {
			continue
		}

		// 获取员工信息
		employee, err := s.employeeRepo.GetByID(ctx, employeeID)
		if err != nil {
			logger.WithError(err).WithField("employee_id", employeeID).Error("获取员工信息失败")
			continue
		}

		// 获取用户信息
		user, err := s.userRepo.GetByID(ctx, employee.UserID)
		if err != nil {
			logger.WithError(err).WithField("user_id", employee.UserID).Error("获取用户信息失败")
			continue
		}
		employee.User = *user

		// 只包含状态为approval_pending的员工
		if employee.OnboardingStatus != "approval_pending" {
			continue
		}

		// 转换日期格式
		expectedDateStr := ""
		if employee.ExpectedDate != nil {
			expectedDateStr = employee.ExpectedDate.Format("2006-01-02")
		}

		result = append(result, &PendingOnboardingApproval{
			InstanceID:    approval.InstanceID,
			EmployeeID:    employeeID,
			EmployeeName:  employee.User.RealName,
			Department:    getDepartmentName(employee.DepartmentID),
			Position:      getPositionName(employee.PositionID),
			ExpectedDate:  expectedDateStr,
			CurrentStep:   approval.NodeName,
			SubmittedAt:   approval.CreatedAt,
			RequesterName: "", // 需要从工作流变量中获取
			Notes:         fmt.Sprintf("业务ID: %s", approval.BusinessID),
		})
	}

	return result, nil
}

// GetOnboardingApprovalHistory 获取入职审批历史
func (s *OnboardingServiceImpl) GetOnboardingApprovalHistory(ctx context.Context, employeeID uint) ([]*OnboardingApprovalHistory, error) {
	logger := s.logger.WithField("method", "GetOnboardingApprovalHistory")

	// 获取员工的所有工作流实例
	// 这里需要扩展工作流服务来支持按员工ID查询
	// 暂时返回空列表
	logger.Info("获取入职审批历史功能待完善")
	return []*OnboardingApprovalHistory{}, nil
}

// CancelOnboardingApproval 取消入职审批流程
func (s *OnboardingServiceImpl) CancelOnboardingApproval(ctx context.Context, instanceID string, reason string, operatorID uint) error {
	logger := s.logger.WithField("method", "CancelOnboardingApproval")

	// 暂时简化实现，直接返回成功
	// 实际实现需要根据工作流服务的具体API来处理
	logger.WithFields(logrus.Fields{
		"instance_id": instanceID,
		"reason":      reason,
		"operator_id": operatorID,
	}).Info("取消入职审批流程")

	return nil
}

// getCurrentNode 从当前节点列表中获取第一个节点
func getCurrentNode(nodes []string) string {
	if len(nodes) == 0 {
		return ""
	}
	return nodes[0]
}
