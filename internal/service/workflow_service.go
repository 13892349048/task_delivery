package service

import (
	"context"

	"taskmanage/internal/workflow"
)

// WorkflowServiceWrapper 工作流服务包装器
type WorkflowServiceWrapper struct {
	workflowService *workflow.WorkflowService
}

// NewWorkflowServiceWrapper 创建工作流服务包装器
func NewWorkflowServiceWrapper(workflowService *workflow.WorkflowService) WorkflowService {
	return &WorkflowServiceWrapper{
		workflowService: workflowService,
	}
}

// StartTaskAssignmentApproval 启动任务分配审批流程
func (w *WorkflowServiceWrapper) StartTaskAssignmentApproval(ctx context.Context, req *workflow.TaskAssignmentApprovalRequest) (*workflow.WorkflowInstance, error) {
	if w.workflowService == nil {
		return nil, workflow.ErrWorkflowServiceNotReady
	}
	return w.workflowService.StartTaskAssignmentApproval(ctx, req)
}

// ProcessTaskAssignmentApproval 处理任务分配审批
func (w *WorkflowServiceWrapper) ProcessTaskAssignmentApproval(ctx context.Context, req *workflow.ApprovalRequest) (*workflow.ApprovalResult, error) {
	if w.workflowService == nil {
		return nil, workflow.ErrWorkflowServiceNotReady
	}
	// 转换为ProcessApprovalRequest
	processReq := &workflow.ProcessApprovalRequest{
		InstanceID: req.InstanceID,
		NodeID:     req.NodeID,
		Action:     req.Action,
		Comment:    req.Comment,
		Variables:  req.Variables,
		ApprovedBy: req.ApprovedBy,
	}
	return w.workflowService.ProcessTaskAssignmentApproval(ctx, processReq)
}

// GetPendingTaskAssignmentApprovals 获取待审批任务分配
func (w *WorkflowServiceWrapper) GetPendingTaskAssignmentApprovals(ctx context.Context, userID uint) ([]*workflow.PendingApproval, error) {
	if w.workflowService == nil {
		return nil, workflow.ErrWorkflowServiceNotReady
	}
	return w.workflowService.GetPendingTaskAssignmentApprovals(ctx, userID)
}

// GetWorkflowInstance 获取流程实例
func (w *WorkflowServiceWrapper) GetWorkflowInstance(ctx context.Context, instanceID string) (*workflow.WorkflowInstance, error) {
	if w.workflowService == nil {
		return nil, workflow.ErrWorkflowServiceNotReady
	}
	return w.workflowService.GetWorkflowInstance(ctx, instanceID)
}

// GetPendingApprovals 获取待审批任务
func (w *WorkflowServiceWrapper) GetPendingApprovals(ctx context.Context, userID uint) ([]*workflow.PendingApproval, error) {
	if w.workflowService == nil {
		return nil, workflow.ErrWorkflowServiceNotReady
	}
	return w.workflowService.GetPendingTaskAssignmentApprovals(ctx, userID)
}

// CancelWorkflow 取消流程
func (w *WorkflowServiceWrapper) CancelWorkflow(ctx context.Context, instanceID string, reason string) error {
	if w.workflowService == nil {
		return workflow.ErrWorkflowServiceNotReady
	}
	return w.workflowService.CancelTaskAssignmentApproval(ctx, instanceID, reason)
}

// GetWorkflowHistory 获取流程历史
func (w *WorkflowServiceWrapper) GetWorkflowHistory(ctx context.Context, instanceID string) ([]workflow.ExecutionHistory, error) {
	if w.workflowService == nil {
		return nil, workflow.ErrWorkflowServiceNotReady
	}
	// 获取实例并返回历史
	instance, err := w.workflowService.GetWorkflowInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	return instance.History, nil
}

// CreateWorkflowDefinition 创建工作流定义
func (w *WorkflowServiceWrapper) CreateWorkflowDefinition(ctx context.Context, req *workflow.CreateWorkflowRequest) (*workflow.WorkflowDefinition, error) {
	if w.workflowService == nil {
		return nil, workflow.ErrWorkflowServiceNotReady
	}
	return w.workflowService.GetDefinitionManager().CreateWorkflow(ctx, req)
}

// GetWorkflowDefinitions 获取工作流定义列表
func (w *WorkflowServiceWrapper) GetWorkflowDefinitions(ctx context.Context) ([]*workflow.WorkflowDefinition, error) {
	if w.workflowService == nil {
		return nil, workflow.ErrWorkflowServiceNotReady
	}
	filter := workflow.WorkflowFilter{} // 空过滤器获取所有定义
	return w.workflowService.GetDefinitionManager().ListWorkflows(ctx, filter)
}

// GetWorkflowDefinition 获取工作流定义详情
func (w *WorkflowServiceWrapper) GetWorkflowDefinition(ctx context.Context, id string) (*workflow.WorkflowDefinition, error) {
	if w.workflowService == nil {
		return nil, workflow.ErrWorkflowServiceNotReady
	}
	return w.workflowService.GetDefinitionManager().GetWorkflow(ctx, id)
}

// UpdateWorkflowDefinition 更新工作流定义
func (w *WorkflowServiceWrapper) UpdateWorkflowDefinition(ctx context.Context, id string, req *workflow.UpdateWorkflowRequest) (*workflow.WorkflowDefinition, error) {
	if w.workflowService == nil {
		return nil, workflow.ErrWorkflowServiceNotReady
	}
	return w.workflowService.GetDefinitionManager().UpdateWorkflow(ctx, id, req)
}

// DeleteWorkflowDefinition 删除工作流定义
func (w *WorkflowServiceWrapper) DeleteWorkflowDefinition(ctx context.Context, id string) error {
	if w.workflowService == nil {
		return workflow.ErrWorkflowServiceNotReady
	}
	return w.workflowService.GetDefinitionManager().DeactivateWorkflow(ctx, id)
}

// ValidateWorkflowDefinition 验证工作流定义
func (w *WorkflowServiceWrapper) ValidateWorkflowDefinition(ctx context.Context, req *workflow.CreateWorkflowRequest) error {
	if w.workflowService == nil {
		return workflow.ErrWorkflowServiceNotReady
	}
	return w.workflowService.GetDefinitionManager().ValidateWorkflow(ctx, req)
}

// StartOnboardingApproval 启动入职审批流程
func (w *WorkflowServiceWrapper) StartOnboardingApproval(ctx context.Context, req *workflow.OnboardingApprovalRequest) (*workflow.WorkflowInstance, error) {
	if w.workflowService == nil {
		return nil, workflow.ErrWorkflowServiceNotReady
	}
	return w.workflowService.StartOnboardingApproval(ctx, req)
}

// ProcessOnboardingApproval 处理入职审批
func (w *WorkflowServiceWrapper) ProcessOnboardingApproval(ctx context.Context, req *workflow.ApprovalRequest) (*workflow.ApprovalResult, error) {
	if w.workflowService == nil {
		return nil, workflow.ErrWorkflowServiceNotReady
	}
	return w.workflowService.ProcessOnboardingApproval(ctx, req)
}
