package service

import (
	"context"
	"fmt"
	"time"

	"taskmanage/internal/assignment"
	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// AssignmentManagementService 任务分配管理服务
type AssignmentManagementService struct {
	assignmentService   *assignment.AssignmentService
	workflowService     WorkflowService
	notificationService NotificationService
	taskRepo            repository.TaskRepository
	employeeRepo        repository.EmployeeRepository
	assignmentRepo      repository.AssignmentRepository
}

// 确保实现了AssignmentService接口
var _ AssignmentService = (*AssignmentManagementService)(nil)

// NewAssignmentManagementService 创建任务分配管理服务
func NewAssignmentManagementService(
	assignmentService *assignment.AssignmentService,
	workflowService WorkflowService,
	repoManager repository.RepositoryManager,
	notificationService NotificationService,
) AssignmentService {
	return &AssignmentManagementService{
		assignmentService:   assignmentService,
		workflowService:     workflowService,
		notificationService: notificationService,
		taskRepo:            repoManager.TaskRepository(),
		employeeRepo:        repoManager.EmployeeRepository(),
		assignmentRepo:      repoManager.AssignmentRepository(),
	}
}

// ManualAssignmentRequest 手动分配请求
type ManualAssignmentRequest struct {
	TaskID          uint   `json:"task_id" binding:"required"`
	EmployeeID      uint   `json:"employee_id" binding:"required"`
	AssignedBy      uint   `json:"assigned_by"`
	Reason          string `json:"reason"`
	Task            *database.Task
	Priority        string `json:"priority"`
	RequireApproval bool   `json:"require_approval"`
}

// AssignmentSuggestionRequest 分配建议请求
type AssignmentSuggestionRequest struct {
	TaskID         uint     `json:"task_id" binding:"required"`
	Strategy       string   `json:"strategy"`
	RequiredSkills []string `json:"required_skills"`
	Department     string   `json:"department"`
	ExcludeUsers   []uint   `json:"exclude_users"`
	MaxSuggestions int      `json:"max_suggestions"`
}

// GetAssignmentSuggestionsRequest 获取分配建议请求
type GetAssignmentSuggestionsRequest struct {
	TaskID         uint                          `json:"task_id"`
	Task           *database.Task                `json:"task"`
	RequiredSkills []assignment.SkillRequirement `json:"required_skills"`
	Strategy       string                        `json:"strategy"`
	Limit          int                           `json:"limit"`
}

// AssignmentConflictCheck 分配冲突检查
type AssignmentConflictCheck struct {
	TaskID     uint `json:"task_id"`
	EmployeeID uint `json:"employee_id"`
}

// AssignmentConflict 分配冲突
type AssignmentConflict struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	TaskID      uint      `json:"task_id,omitempty"`
	TaskTitle   string    `json:"task_title,omitempty"`
	Deadline    time.Time `json:"deadline,omitempty"`
}

// AssignmentSuggestion 分配建议
type AssignmentSuggestion struct {
	Employee     *database.Employee `json:"employee"`
	Score        float64            `json:"score"`
	Reason       string             `json:"reason"`
	Confidence   float64            `json:"confidence"`
	Workload     WorkloadInfo       `json:"workload"`
	SkillMatch   float64            `json:"skill_match"`
	Availability float64            `json:"availability"`
}

// WorkloadInfo 工作负载信息
type WorkloadInfo struct {
	CurrentTasks    int     `json:"current_tasks"`
	MaxTasks        int     `json:"max_tasks"`
	UtilizationRate float64 `json:"utilization_rate"`
	AvgTaskDuration int64   `json:"avg_task_duration"`
}

// AssignmentHistory 分配历史
type AssignmentHistory struct {
	ID           uint                `json:"id"`
	TaskID       uint                `json:"task_id"`
	TaskTitle    string              `json:"task_title"`
	EmployeeID   uint                `json:"employee_id"`
	EmployeeName string              `json:"employee_name"`
	AssignedBy   uint                `json:"assigned_by"`
	AssignedAt   time.Time           `json:"assigned_at"`
	Strategy     string              `json:"strategy"`
	Status       string              `json:"status"`
	Reason       string              `json:"reason"`
	ApprovalInfo *AssignmentApproval `json:"approval_info,omitempty"`
}

// AssignmentApproval 分配审批信息
type AssignmentApproval struct {
	Required   bool      `json:"required"`
	Status     string    `json:"status"`
	ApprovedBy uint      `json:"approved_by,omitempty"`
	ApprovedAt time.Time `json:"approved_at,omitempty"`
	Comment    string    `json:"comment,omitempty"`
	InstanceID string    `json:"instance_id,omitempty"`
}

// ManualAssign 手动分配任务
func (s *AssignmentManagementService) ManualAssign(ctx context.Context, req *ManualAssignmentRequest) (*AssignmentHistory, error) {
	logger.Infof("开始手动分配任务: TaskID=%d, EmployeeID=%d", req.TaskID, req.EmployeeID)

	// 验证任务和员工存在性
	task, err := s.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("任务不存在: %w", err)
	}

	_, err = s.employeeRepo.GetByID(ctx, req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("员工不存在: %w", err)
	}

	// 检查任务状态
	if task.Status != "pending" {
		return nil, fmt.Errorf("只有待分配状态的任务才能进行分配")
	}

	// 简化冲突检查 - 检查员工是否已有该任务
	existingAssignment, err := s.assignmentRepo.GetActiveByTaskID(ctx, req.TaskID)
	if err == nil && existingAssignment != nil {
		return nil, fmt.Errorf("任务已分配给其他员工")
	}

	// 创建分配记录
	assignment := &database.Assignment{
		TaskID:     req.TaskID,
		AssigneeID: req.EmployeeID,
		AssignerID: req.AssignedBy,
		AssignedAt: time.Now(),
		Method:     "manual",
		Status:     "pending",
		Reason:     req.Reason,
	}

	// 根据是否需要审批决定处理方式
	if req.RequireApproval {
		// 设置为待审批状态
		assignment.Status = "pending"
	} else {
		// 直接批准分配
		assignment.Status = "approved"
		// 更新任务状态和分配人
		task.Status = "assigned"
		task.AssigneeID = &req.EmployeeID
		if err := s.taskRepo.Update(ctx, task); err != nil {
			return nil, fmt.Errorf("更新任务分配失败: %w", err)
		}

		// 更新员工当前任务数
		employee, err := s.employeeRepo.GetByID(ctx, req.EmployeeID)
		if err == nil && employee != nil {
			employee.CurrentTasks++
			s.employeeRepo.Update(ctx, employee)
		}
	}

	if err := s.assignmentRepo.Create(ctx, assignment); err != nil {
		return nil, fmt.Errorf("保存分配记录失败: %w", err)
	}

	// 创建任务分配通知
	if s.notificationService != nil {
		err := s.notificationService.CreateTaskAssignmentNotification(ctx, req.TaskID, req.EmployeeID, req.AssignedBy)
		if err != nil {
			logger.Errorf("创建任务分配通知失败: %v", err)
			// 不中断流程，只记录错误
		}
	}

	// 返回分配历史
	history, err := s.buildAssignmentHistory(ctx, assignment)
	if err != nil {
		return nil, fmt.Errorf("构建分配历史失败: %w", err)
	}

	logger.Infof("手动分配任务成功: TaskID=%d, EmployeeID=%d, Status=%s",
		req.TaskID, req.EmployeeID, assignment.Status)

	return history, nil
}

// GetAssignmentSuggestions 获取分配建议
func (s *AssignmentManagementService) GetAssignmentSuggestions(ctx context.Context, req *AssignmentSuggestionRequest) ([]*AssignmentSuggestion, error) {
	logger.Infof("获取分配建议: TaskID=%d, Strategy=%s", req.TaskID, req.Strategy)

	// 获取任务信息
	task, err := s.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	// 构建分配请求
	assignmentReq := &assignment.AssignmentRequest{
		TaskID:           req.TaskID,
		Strategy:         assignment.AssignmentStrategy(req.Strategy),
		RequiredSkills:   s.convertSkillRequirements(req.RequiredSkills),
		Department:       req.Department,
		Priority:         task.Priority,
		Deadline:         task.DueDate,
		ExcludeEmployees: req.ExcludeUsers,
	}

	// 获取候选人
	candidates, err := s.assignmentService.GetCandidates(ctx, assignmentReq)
	if err != nil {
		return nil, fmt.Errorf("获取候选人失败: %w", err)
	}

	// 限制建议数量
	maxSuggestions := req.MaxSuggestions
	if maxSuggestions == 0 || maxSuggestions > 10 {
		maxSuggestions = 5
	}
	if len(candidates) > maxSuggestions {
		candidates = candidates[:maxSuggestions]
	}

	// 构建建议列表
	suggestions := make([]*AssignmentSuggestion, len(candidates))
	for i, candidate := range candidates {
		suggestions[i] = &AssignmentSuggestion{
			Employee:   &candidate.Employee,
			Score:      candidate.Score,
			Reason:     fmt.Sprintf("匹配度: %.2f, 工作负载: %d/%d", candidate.Score, candidate.Workload.CurrentTasks, candidate.Workload.MaxTasks),
			Confidence: candidate.Score,
			Workload: WorkloadInfo{
				CurrentTasks:    candidate.Workload.CurrentTasks,
				MaxTasks:        candidate.Workload.MaxTasks,
				UtilizationRate: candidate.Workload.UtilizationRate,
				AvgTaskDuration: int64(candidate.Workload.AvgTaskDuration / time.Hour),
			},
			SkillMatch:   candidate.Score,                            // 技能匹配度
			Availability: 100.0 - candidate.Workload.UtilizationRate, // 可用性百分比
		}
	}

	logger.Infof("获取到 %d 个分配建议", len(suggestions))
	return suggestions, nil
}

// GetAssignmentHistory 获取分配历史
func (s *AssignmentManagementService) GetAssignmentHistory(ctx context.Context, taskID uint) ([]*AssignmentHistory, error) {
	logger.Infof("获取任务分配历史: TaskID=%d", taskID)

	// 获取所有分配记录
	assignments, err := s.assignmentRepo.GetAssignmentHistory(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取分配历史失败: %w", err)
	}

	// 转换为历史记录格式
	historyList := make([]*AssignmentHistory, len(assignments))
	for i, assignment := range assignments {
		history, err := s.buildAssignmentHistory(ctx, assignment)
		if err != nil {
			logger.Errorf("构建分配历史失败: %v", err)
			continue
		}
		historyList[i] = history
	}

	logger.Infof("获取到 %d 条分配历史记录", len(historyList))
	return historyList, nil
}

// ReassignTask 重新分配任务
func (s *AssignmentManagementService) ReassignTask(ctx context.Context, taskID uint, newEmployeeID uint, reason string, assignedBy uint) error {
	logger.Infof("重新分配任务: TaskID=%d, NewEmployeeID=%d", taskID, newEmployeeID)

	// 获取当前分配
	currentAssignment, err := s.assignmentRepo.GetActiveByTaskID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取当前分配失败: %w", err)
	}

	// 结束当前分配
	currentAssignment.Status = "reassigned"
	if currentAssignment.Status == "completed" {
		now := time.Now()
		currentAssignment.ApprovedAt = &now
	}
	if err := s.assignmentRepo.Update(ctx, currentAssignment); err != nil {
		return fmt.Errorf("更新当前分配状态失败: %w", err)
	}

	// 创建新分配
	req := &ManualAssignmentRequest{
		TaskID:     taskID,
		EmployeeID: newEmployeeID,
		AssignedBy: assignedBy,
		Reason:     fmt.Sprintf("重新分配: %s", reason),
		Priority:   "normal",
	}

	_, err = s.ManualAssign(ctx, req)
	if err != nil {
		return fmt.Errorf("创建新分配失败: %w", err)
	}

	logger.Infof("任务重新分配成功: TaskID=%d", taskID)
	return nil
}

// CancelAssignment 取消分配
func (s *AssignmentManagementService) CancelAssignment(ctx context.Context, taskID uint, reason string, cancelledBy uint) error {
	logger.Infof("取消任务分配: TaskID=%d", taskID)

	assignment, err := s.assignmentRepo.GetActiveByTaskID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取分配记录失败: %w", err)
	}

	// 如果有审批流程，取消审批 (暂时略过，待实现)
	// if s.workflowService != nil {
	//	if err := s.workflowService.CancelWorkflow(ctx, assignment.ApprovalInstanceID, reason); err != nil {
	//		logger.Errorf("取消审批流程失败: %v", err)
	//	}
	// }

	// 更新分配状态
	assignment.Status = "cancelled"
	now := time.Now()
	assignment.ApprovedAt = &now // 使用ApprovedAt字段代替CompletedAt
	assignment.Reason = fmt.Sprintf("%s (取消原因: %s)", assignment.Reason, reason)

	if err := s.assignmentRepo.Update(ctx, assignment); err != nil {
		return fmt.Errorf("更新分配状态失败: %w", err)
	}

	// 更新任务状态
	if err := s.taskRepo.UpdateStatus(ctx, taskID, "unassigned"); err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	logger.Infof("任务分配取消成功: TaskID=%d", taskID)
	return nil
}

// buildAssignmentHistory 构建分配历史
func (s *AssignmentManagementService) buildAssignmentHistory(ctx context.Context, assignment *database.Assignment) (*AssignmentHistory, error) {
	// 获取任务信息
	task, err := s.taskRepo.GetByID(ctx, assignment.TaskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务信息失败: %w", err)
	}

	// 获取员工信息
	employee, err := s.employeeRepo.GetByID(ctx, assignment.AssigneeID)
	if err != nil {
		return nil, fmt.Errorf("获取员工信息失败: %w", err)
	}

	history := &AssignmentHistory{
		ID:           assignment.ID,
		TaskID:       assignment.TaskID,
		TaskTitle:    task.Title,
		EmployeeID:   assignment.AssigneeID,
		EmployeeName: employee.User.RealName,
		AssignedBy:   assignment.AssignerID,
		AssignedAt:   assignment.AssignedAt,
		Strategy:     assignment.Method,
		Status:       assignment.Status,
		Reason:       assignment.Reason,
	}

	// 如果有审批信息 (暂时略过，待实现)
	// if assignment.ApprovalInstanceID != "" {
	//	approvalInfo, err := s.buildApprovalInfo(ctx, assignment.ApprovalInstanceID)
	//	if err != nil {
	//		logger.Errorf("构建审批信息失败: %v", err)
	//	} else {
	//		history.ApprovalInfo = approvalInfo
	//	}
	// }

	return history, nil
}

// hasTimeConflict 检查时间冲突
func (s *AssignmentManagementService) hasTimeConflict(task1, task2 *database.Task) bool {
	// 简化的时间冲突检查
	if task1.DueDate == nil || task2.DueDate == nil {
		return false // 如果没有截止时间，不认为有冲突
	}

	// 如果两个任务的截止时间相近（7天内），认为有冲突
	diff := task1.DueDate.Sub(*task2.DueDate)
	if diff < 0 {
		diff = -diff
	}

	return diff < 7*24*time.Hour
}

// hasRequiredSkills 检查是否具备所需技能
func (s *AssignmentManagementService) hasRequiredSkills(employee *database.Employee, task *database.Task) bool {
	// 简化的技能检查，实际应该查询技能表
	return true
}

// convertSkillRequirements 转换技能需求格式
func (s *AssignmentManagementService) convertSkillRequirements(skillNames []string) []assignment.SkillRequirement {
	// 简化实现，实际应该根据技能名称查询技能ID
	skillReqs := make([]assignment.SkillRequirement, len(skillNames))
	for i, skillName := range skillNames {
		// 这里使用简化的方式，实际应该查询数据库获取技能ID
		skillReqs[i] = assignment.SkillRequirement{
			SkillID:  uint(i + 1), // 使用索引+1作为临时ID
			MinLevel: 1,           // 默认最低等级
		}
		_ = skillName // 避免未使用变量警告
	}
	return skillReqs
}

// AutoAssign 自动分配任务
func (s *AssignmentManagementService) AutoAssign(ctx context.Context, taskID uint, strategy string) (*AssignmentResponse, error) {
	logger.Infof("开始自动分配任务: TaskID=%d, Strategy=%s", taskID, strategy)

	// 获取任务信息
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	// 验证任务状态
	if task.Status != "pending" {
		return nil, fmt.Errorf("只有待分配状态的任务才能进行自动分配")
	}

	// 构建分配请求
	req := &assignment.AssignmentRequest{
		TaskID:   taskID,
		Strategy: assignment.AssignmentStrategy(strategy),
		Priority: task.Priority,
	}

	if task.DueDate != nil {
		req.Deadline = task.DueDate
	}

	// 执行自动分配
	result, err := s.assignmentService.AssignTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("自动分配失败: %w", err)
	}

	// 创建分配记录
	assignmentRecord := &database.Assignment{
		TaskID:     taskID,
		AssigneeID: result.SelectedEmployee.ID,
		AssignerID: 1, // TODO: 从上下文获取当前用户ID
		AssignedAt: result.ExecutedAt,
		Method:     strategy,
		Status:     "approved", // 自动分配直接批准
		Reason:     result.Reason,
	}

	if err := s.assignmentRepo.Create(ctx, assignmentRecord); err != nil {
		return nil, fmt.Errorf("保存分配记录失败: %w", err)
	}

	// 更新任务状态和分配信息
	task.Status = "assigned"
	task.AssigneeID = &result.SelectedEmployee.UserID

	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Errorf("更新任务分配失败: %v", err)
		return nil, fmt.Errorf("更新任务分配失败: %w", err)
	}

	// 更新员工当前任务数
	employee, err := s.employeeRepo.GetByID(ctx, result.SelectedEmployee.ID)
	if err == nil && employee != nil {
		employee.CurrentTasks++
		s.employeeRepo.Update(ctx, employee)
	}

	logger.Infof("自动分配任务成功: TaskID=%d, EmployeeID=%d", taskID, result.SelectedEmployee.ID)

	// 返回分配响应
	return &AssignmentResponse{
		ID:         assignmentRecord.ID,
		TaskID:     taskID,
		EmployeeID: result.SelectedEmployee.ID,
		Status:     "approved",
		AssignedBy: 1, // TODO: 从上下文获取当前用户ID
		AssignedAt: result.ExecutedAt,
		Comment:    result.Reason,
	}, nil
}

// CheckAssignmentConflicts 检查分配冲突
func (s *AssignmentManagementService) CheckAssignmentConflicts(ctx context.Context, taskID uint, employeeID uint) ([]*AssignmentConflict, error) {
	logger.Infof("检查分配冲突: TaskID=%d, EmployeeID=%d", taskID, employeeID)

	conflicts := []*AssignmentConflict{}

	// 检查员工是否已有该任务的分配
	existingAssignment, err := s.assignmentRepo.GetActiveByTaskID(ctx, taskID)
	if err == nil && existingAssignment != nil {
		if existingAssignment.AssigneeID == employeeID && existingAssignment.Status == "approved" {
			conflicts = append(conflicts, &AssignmentConflict{
				Type:        "duplicate_assignment",
				Description: "员工已被分配此任务",
				Severity:    "high",
				TaskID:      taskID,
			})
		}
	}

	// 检查员工工作负载
	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err == nil && employee != nil {
		if employee.CurrentTasks >= employee.MaxTasks {
			conflicts = append(conflicts, &AssignmentConflict{
				Type:        "workload_exceeded",
				Description: fmt.Sprintf("员工工作负载已满 (%d/%d)", employee.CurrentTasks, employee.MaxTasks),
				Severity:    "medium",
				TaskID:      taskID,
			})
		}
	}

	logger.Infof("冲突检查完成: 发现 %d 个冲突", len(conflicts))
	return conflicts, nil
}
