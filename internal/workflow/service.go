package workflow

import (
	"context"
	"fmt"

	"taskmanage/internal/assignment"
	"taskmanage/pkg/logger"
)

// TaskServiceInterface 任务服务接口（避免循环依赖）
type TaskServiceInterface interface {
	CompleteTaskAssignmentWorkflow(ctx context.Context, workflowInstanceID string, approved bool, approverID uint) error
}

// WorkflowService 审批流程服务
type WorkflowService struct {
	engine            WorkflowEngine
	definitionManager *WorkflowDefinitionManager
	taskService       TaskServiceInterface
	selector          WorkflowSelector
}

// GetDefinitionManager 获取工作流定义管理器
func (s *WorkflowService) GetDefinitionManager() *WorkflowDefinitionManager {
	return s.definitionManager
}

// NewWorkflowService 创建审批流程服务
func NewWorkflowService(
	engine WorkflowEngine,
	definitionManager *WorkflowDefinitionManager,
) *WorkflowService {
	selector := NewWorkflowSelector(definitionManager)
	return &WorkflowService{
		engine:            engine,
		definitionManager: definitionManager,
		taskService:       nil, // 稍后通过SetTaskService注入
		selector:          selector,
	}
}

// SetTaskService 设置任务服务（解决循环依赖）
func (s *WorkflowService) SetTaskService(taskService TaskServiceInterface) {
	s.taskService = taskService
}

// StartTaskAssignmentApproval 启动任务分配审批流程
func (s *WorkflowService) StartTaskAssignmentApproval(ctx context.Context, req *TaskAssignmentApprovalRequest) (*WorkflowInstance, error) {
	logger.Infof("启动任务分配审批流程: 任务ID=%d", req.TaskID)

	// 使用工作流选择器动态选择合适的工作流
	selectionReq := &WorkflowSelectionRequest{
		BusinessType: "task_assignment",
		Priority:     req.Priority,
		UserID:       req.RequesterID,
		Context: map[string]interface{}{
			"task_id":         req.TaskID,
			"assignee_id":     req.AssigneeID,
			"assignment_type": string(req.AssignmentType),
			"reason":          req.Reason,
		},
	}

	workflowID, err := s.selector.SelectWorkflow(ctx, selectionReq)
	if err != nil {
		return nil, fmt.Errorf("选择工作流失败: %w", err)
	}

	logger.Infof("选择的工作流: %s", workflowID)

	// 构建流程启动请求
	startReq := &StartWorkflowRequest{
		WorkflowID:   "simple-onboarding-approval-v1",
		BusinessID:   fmt.Sprintf("task_%d", req.TaskID),
		BusinessType: "task_assignment",
		Variables: map[string]interface{}{
			"task_id":         req.TaskID,
			"assignee_id":     req.AssigneeID,
			"assignment_type": req.AssignmentType,
			"priority":        req.Priority,
			"requester_id":    req.RequesterID,
			"reason":          req.Reason,
		},
		StartedBy: req.RequesterID,
	}

	// 启动流程
	instance, err := s.engine.StartWorkflow(ctx, startReq)
	if err != nil {
		return nil, fmt.Errorf("启动任务分配审批流程失败: %w", err)
	}

	logger.Infof("任务分配审批流程启动成功: %s", instance.ID)
	return instance, nil
}

// StartOnboardingApproval 启动入职审批流程
func (s *WorkflowService) StartOnboardingApproval(ctx context.Context, req *OnboardingApprovalRequest) (*WorkflowInstance, error) {
	logger.Infof("启动入职审批流程: 员工ID=%d", req.EmployeeID)

	// 使用工作流选择器动态选择合适的工作流
	selectionReq := &WorkflowSelectionRequest{
		BusinessType: "onboarding",
		Priority:     req.Priority,
		UserID:       req.RequesterID,
		DepartmentID: req.DepartmentID,
		Context: map[string]interface{}{
			"employee_id":      req.EmployeeID,
			"employee_type":    req.EmployeeType,
			"department_id":    req.DepartmentID,
			"position_id":      req.PositionID,
			"expected_date":    req.ExpectedDate,
			"probation_period": req.ProbationPeriod,
		},
	}

	workflowID, err := s.selector.SelectWorkflow(ctx, selectionReq)
	if err != nil {
		return nil, fmt.Errorf("选择工作流失败: %w", err)
	}

	logger.Infof("选择的工作流: %s", workflowID)

	// 构建流程启动请求
	startReq := &StartWorkflowRequest{
		WorkflowID:   "simple-onboarding-approval-v1",
		BusinessID:   fmt.Sprintf("employee_%d", req.EmployeeID),
		BusinessType: "onboarding",
		Variables: map[string]interface{}{
			"employee_id":      req.EmployeeID,
			"employee_type":    req.EmployeeType,
			"department_id":    req.DepartmentID,
			"position_id":      req.PositionID,
			"expected_date":    req.ExpectedDate,
			"probation_period": req.ProbationPeriod,
		},
		StartedBy: req.RequesterID,
	}

	// 启动流程
	instance, err := s.engine.StartWorkflow(ctx, startReq)
	if err != nil {
		return nil, fmt.Errorf("启动入职审批流程失败: %w", err)
	}

	logger.Infof("入职审批流程启动成功: %s", instance.ID)
	return instance, nil
}

// ProcessTaskAssignmentApproval 处理任务分配审批
func (s *WorkflowService) ProcessTaskAssignmentApproval(ctx context.Context, req *ProcessApprovalRequest) (*ApprovalResult, error) {
	logger.Infof("处理任务分配审批: 实例=%s, 动作=%s", req.InstanceID, req.Action)

	// 构建审批请求
	approvalReq := &ApprovalRequest{
		InstanceID: req.InstanceID,
		NodeID:     req.NodeID,
		Action:     req.Action,
		Comment:    req.Comment,
		Variables:  req.Variables,
		ApprovedBy: req.ApprovedBy,
	}

	// 处理审批
	result, err := s.engine.ProcessApproval(ctx, approvalReq)
	if err != nil {
		return nil, fmt.Errorf("处理任务分配审批失败: %w", err)
	}

	// 如果审批完成，执行后续操作
	if result.IsCompleted {
		if err := s.handleApprovalCompletion(ctx, result); err != nil {
			logger.Errorf("处理审批完成后续操作失败: %v", err)
		}
	}

	return result, nil
}

// GetPendingTaskAssignmentApprovals 获取待审批的任务分配
func (s *WorkflowService) GetPendingTaskAssignmentApprovals(ctx context.Context, userID uint) ([]*PendingApproval, error) {
	approvals, err := s.engine.GetPendingApprovals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取待审批任务失败: %w", err)
	}

	// 过滤任务分配相关的审批
	var taskApprovals []*PendingApproval
	for _, approval := range approvals {
		if approval.BusinessType == "task_assignment" {
			taskApprovals = append(taskApprovals, approval)
		}
	}

	return taskApprovals, nil
}

// GetWorkflowInstance 获取流程实例
func (s *WorkflowService) GetWorkflowInstance(ctx context.Context, instanceID string) (*WorkflowInstance, error) {
	return s.engine.GetWorkflowInstance(ctx, instanceID)
}

// CancelTaskAssignmentApproval 取消任务分配审批
func (s *WorkflowService) CancelTaskAssignmentApproval(ctx context.Context, instanceID string, reason string) error {
	return s.engine.CancelWorkflow(ctx, instanceID, reason)
}

// CreateTaskAssignmentWorkflow 创建任务分配审批流程定义
func (s *WorkflowService) CreateTaskAssignmentWorkflow(ctx context.Context) error {
	logger.Info("创建任务分配审批流程定义")

	// 定义流程节点
	nodes := []WorkflowNode{
		{
			ID:       "start",
			Type:     NodeTypeStart,
			Name:     "开始",
			Position: NodePosition{X: 100, Y: 100},
		},
		{
			ID:   "manager_approval",
			Type: NodeTypeApproval,
			Name: "直属上级审批",
			Config: map[string]interface{}{
				"assignees": []ApprovalAssignee{
					{
						Type:  AssigneeTypeManager,
						Value: "requester",
					},
				},
				"approval_type": ApprovalTypeAny,
				"can_delegate":  true,
				"can_return":    true,
				"priority":      1,
			},
			Position: NodePosition{X: 300, Y: 100},
		},
		{
			ID:   "priority_check",
			Type: NodeTypeCondition,
			Name: "优先级检查",
			Config: map[string]interface{}{
				"conditions": []ConditionRule{
					{
						Expression: "priority == 'high'",
						Target:     "director_approval",
						Priority:   1,
					},
					{
						Expression: "priority != 'high'",
						Target:     "auto_approve",
						Priority:   2,
					},
				},
			},
			Position: NodePosition{X: 500, Y: 100},
		},
		{
			ID:   "director_approval",
			Type: NodeTypeApproval,
			Name: "总监审批",
			Config: map[string]interface{}{
				"assignees": []ApprovalAssignee{
					{
						Type:  AssigneeTypeRole,
						Value: "director",
					},
				},
				"approval_type": ApprovalTypeAny,
				"can_delegate":  false,
				"can_return":    true,
				"priority":      2,
			},
			Position: NodePosition{X: 500, Y: 300},
		},
		{
			ID:   "auto_approve",
			Type: NodeTypeScript,
			Name: "自动审批",
			Config: map[string]interface{}{
				"script": "auto_approve_task_assignment",
			},
			Position: NodePosition{X: 700, Y: 100},
		},
		{
			ID:   "notify_result",
			Type: NodeTypeNotify,
			Name: "通知结果",
			Config: map[string]interface{}{
				"type":       NotificationTypeSystem,
				"template":   "task_assignment_approved",
				"recipients": []string{"requester", "assignee"},
			},
			Position: NodePosition{X: 900, Y: 200},
		},
		{
			ID:       "end",
			Type:     NodeTypeEnd,
			Name:     "结束",
			Position: NodePosition{X: 1100, Y: 200},
		},
	}

	// 定义流程边
	edges := []WorkflowEdge{
		{
			ID:   "start_to_manager",
			From: "start",
			To:   "manager_approval",
		},
		{
			ID:        "manager_to_priority",
			From:      "manager_approval",
			To:        "priority_check",
			Condition: "action == 'approve'",
		},
		{
			ID:   "priority_to_director",
			From: "priority_check",
			To:   "director_approval",
		},
		{
			ID:   "priority_to_auto",
			From: "priority_check",
			To:   "auto_approve",
		},
		{
			ID:        "director_to_notify",
			From:      "director_approval",
			To:        "notify_result",
			Condition: "action == 'approve'",
		},
		{
			ID:   "auto_to_notify",
			From: "auto_approve",
			To:   "notify_result",
		},
		{
			ID:   "notify_to_end",
			From: "notify_result",
			To:   "end",
		},
	}

	// 创建流程定义请求
	createReq := &CreateWorkflowRequest{
		ID:          "task_assignment_approval",
		Name:        "任务分配审批流程",
		Description: "用于审批任务分配的标准流程",
		Version:     "1.0",
		Nodes:       nodes,
		Edges:       edges,
		Variables: map[string]interface{}{
			"auto_approve_threshold": "medium",
		},
	}

	// 创建流程定义
	_, err := s.definitionManager.CreateWorkflow(ctx, createReq)
	if err != nil {
		return fmt.Errorf("创建任务分配审批流程定义失败: %w", err)
	}

	logger.Info("任务分配审批流程定义创建成功")
	return nil
}

// handleApprovalCompletion 处理审批完成后续操作
func (s *WorkflowService) handleApprovalCompletion(ctx context.Context, result *ApprovalResult) error {
	// 获取流程实例
	instance, err := s.engine.GetWorkflowInstance(ctx, result.InstanceID)
	if err != nil {
		return fmt.Errorf("获取流程实例失败: %w", err)
	}

	// 只处理任务分配相关的工作流
	if instance.BusinessType != "task_assignment" {
		logger.Infof("跳过非任务分配工作流: %s", instance.BusinessType)
		return nil
	}

	// 提取业务数据
	taskID, ok := instance.Variables["task_id"].(float64)
	if !ok {
		return fmt.Errorf("无法获取任务ID")
	}

	assigneeID, ok := instance.Variables["assignee_id"].(float64)
	if !ok {
		return fmt.Errorf("无法获取分配人ID")
	}

	// 获取审批人ID (从执行历史中获取最后一个审批操作的执行人)
	approverID := uint(0)
	if len(instance.History) > 0 {
		lastHistory := instance.History[len(instance.History)-1]
		if executedBy, ok := lastHistory.Variables["executed_by"].(float64); ok {
			approverID = uint(executedBy)
		}
	}

	// 根据审批结果执行相应操作
	approved := result.Action == ActionApprove

	logger.Infof("工作流完成，调用任务服务处理结果: InstanceID=%s, Approved=%v, TaskID=%d, AssigneeID=%d, ApproverID=%d",
		result.InstanceID, approved, int(taskID), int(assigneeID), approverID)

	// 调用任务服务完成工作流
	if s.taskService != nil {
		err := s.taskService.CompleteTaskAssignmentWorkflow(ctx, result.InstanceID, approved, approverID)
		if err != nil {
			logger.Errorf("完成任务分配工作流失败: %v", err)
			return fmt.Errorf("完成任务分配工作流失败: %w", err)
		}

		if approved {
			logger.Infof("审批通过，任务分配已执行: 任务=%d, 分配给=%d", int(taskID), int(assigneeID))
		} else {
			logger.Infof("审批拒绝，任务分配已取消: 任务=%d", int(taskID))
		}
	} else {
		// 兜底处理，记录日志
		if approved {
			logger.Infof("审批通过，任务分配将被执行: 任务=%d, 分配给=%d", int(taskID), int(assigneeID))
		} else {
			logger.Infof("审批拒绝，任务分配被取消: 任务=%d", int(taskID))
		}
	}

	return nil
}

// TaskAssignmentApprovalRequest 任务分配审批请求
type TaskAssignmentApprovalRequest struct {
	TaskID         uint                          `json:"task_id"`
	AssigneeID     uint                          `json:"assignee_id"`
	AssignmentType assignment.AssignmentStrategy `json:"assignment_type"`
	Priority       string                        `json:"priority"`
	RequesterID    uint                          `json:"requester_id"`
	Reason         string                        `json:"reason,omitempty"`
}

// OnboardingApprovalRequest 入职审批请求
type OnboardingApprovalRequest struct {
	EmployeeID      uint   `json:"employee_id"`
	EmployeeType    string `json:"employee_type"` // permanent, intern, temporary
	DepartmentID    *uint  `json:"department_id"`
	PositionID      *uint  `json:"position_id"`
	ExpectedDate    string `json:"expected_date"`
	ProbationPeriod int    `json:"probation_period"`
	Priority        string `json:"priority"`
	RequesterID     uint   `json:"requester_id"`
}

// ProcessApprovalRequest 处理审批请求
type ProcessApprovalRequest struct {
	InstanceID string                 `json:"instance_id"`
	NodeID     string                 `json:"node_id"`
	Action     ApprovalAction         `json:"action"`
	Comment    string                 `json:"comment,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	ApprovedBy uint                   `json:"approved_by"`
}

// ProcessOnboardingApproval 处理入职审批
func (s *WorkflowService) ProcessOnboardingApproval(ctx context.Context, req *ApprovalRequest) (*ApprovalResult, error) {
	logger.Infof("处理入职审批: InstanceID=%s, Action=%s", req.InstanceID, req.Action)

	// 转换为ProcessApprovalRequest类型
	processReq := &ProcessApprovalRequest{
		InstanceID: req.InstanceID,
		NodeID:     req.NodeID,
		Action:     req.Action,
		Comment:    req.Comment,
		Variables:  req.Variables,
		ApprovedBy: req.ApprovedBy,
	}

	// 调用通用的审批处理方法
	result, err := s.ProcessTaskAssignmentApproval(ctx, processReq)
	if err != nil {
		logger.Errorf("处理入职审批失败: %v", err)
		return nil, fmt.Errorf("处理入职审批失败: %w", err)
	}

	logger.Infof("入职审批处理完成: InstanceID=%s, IsCompleted=%t", req.InstanceID, result.IsCompleted)
	return result, nil
}
