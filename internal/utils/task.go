package utils

import (
	"context"
	"fmt"
	"strings"
	"taskmanage/internal/container"
	"taskmanage/internal/service"
	"taskmanage/internal/workflow"
	"taskmanage/pkg/logger"
)

// InitializeDefaultWorkflows 初始化默认工作流定义
func InitializeDefaultWorkflows(appContainer *container.ApplicationContainer) error {
	ctx := context.Background()

	// 获取服务管理器
	serviceManager := appContainer.GetServiceManager()
	if serviceManager == nil {
		return fmt.Errorf("无法获取服务管理器")
	}

	// 获取工作流服务
	workflowService := serviceManager.WorkflowService()
	if workflowService == nil {
		return fmt.Errorf("无法获取工作流服务")
	}

	// 初始化任务分配审批工作流
	if err := initializeWorkflow(ctx, workflowService, "task_assignment_approval", "任务分配审批工作流", createTaskAssignmentWorkflowDefinition); err != nil {
		return fmt.Errorf("初始化任务分配审批工作流失败: %w", err)
	}

	// 初始化任务完成审批工作流
	if err := initializeWorkflow(ctx, workflowService, "task_completion_approval", "任务完成审批工作流", createTaskCompletionWorkflowDefinition); err != nil {
		return fmt.Errorf("初始化任务完成审批工作流失败: %w", err)
	}

	logger.Info("所有默认工作流定义初始化完成")
	return nil
}

// initializeWorkflow 初始化单个工作流
func initializeWorkflow(ctx context.Context, workflowService service.WorkflowService, workflowID, workflowName string, createFunc func(context.Context, service.WorkflowService) error) error {
	logger.Infof("检查%s是否已存在...", workflowName)
	_, err := workflowService.GetWorkflowDefinition(ctx, workflowID)
	if err == nil {
		logger.Infof("%s已存在，跳过初始化", workflowName)
		return nil
	}

	logger.Infof("%s不存在，开始创建", workflowName)
	if err := createFunc(ctx, workflowService); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "duplicate key") {
			logger.Infof("%s已被并发创建，跳过", workflowName)
			return nil
		}
		return fmt.Errorf("创建%s失败: %w", workflowName, err)
	}

	logger.Infof("%s创建成功", workflowName)
	return nil
}

// createTaskAssignmentWorkflowDefinition 创建任务分配审批工作流定义
// 业务逻辑：pending → assigned（直属领导审批）
func createTaskAssignmentWorkflowDefinition(ctx context.Context, workflowService service.WorkflowService) error {
	// 定义流程节点
	nodes := []workflow.WorkflowNode{
		{
			ID:       "start",
			Type:     workflow.NodeTypeStart,
			Name:     "开始",
			Position: workflow.NodePosition{X: 100, Y: 100},
		},
		{
			ID:   "manager_approval",
			Type: workflow.NodeTypeApproval,
			Name: "直属领导审批",
			Config: map[string]interface{}{
				"assignees": []workflow.ApprovalAssignee{
					{
						Type:  workflow.AssigneeTypeManager,
						Value: "requester",
					},
				},
				"approval_type": workflow.ApprovalTypeAny,
				"can_delegate":  false,
				"can_return":    true,
				"priority":      1,
				"description":   "审批任务分配，通过后任务状态变更为assigned",
			},
			Position: workflow.NodePosition{X: 300, Y: 100},
		},
		{
			ID:   "update_task_status",
			Type: workflow.NodeTypeScript,
			Name: "更新任务状态",
			Config: map[string]interface{}{
				"script": "update_task_status_to_assigned",
				"description": "将任务状态从pending更新为assigned",
			},
			Position: workflow.NodePosition{X: 500, Y: 100},
		},
		{
			ID:   "notify_assignee",
			Type: workflow.NodeTypeNotify,
			Name: "通知被分配人",
			Config: map[string]interface{}{
				"type":       workflow.NotificationTypeSystem,
				"template":   "task_assigned",
				"recipients": []string{"assignee"},
				"description": "通知被分配人任务已分配",
			},
			Position: workflow.NodePosition{X: 700, Y: 100},
		},
		{
			ID:       "end",
			Type:     workflow.NodeTypeEnd,
			Name:     "结束",
			Position: workflow.NodePosition{X: 900, Y: 100},
		},
	}

	// 定义流程边
	edges := []workflow.WorkflowEdge{
		{
			ID:   "start_to_manager",
			From: "start",
			To:   "manager_approval",
		},
		{
			ID:        "manager_to_update",
			From:      "manager_approval",
			To:        "update_task_status",
			Condition: "action == 'approve'",
		},
		{
			ID:   "update_to_notify",
			From: "update_task_status",
			To:   "notify_assignee",
		},
		{
			ID:   "notify_to_end",
			From: "notify_assignee",
			To:   "end",
		},
	}

	// 创建流程定义请求
	createReq := &workflow.CreateWorkflowRequest{
		ID:          "task_assignment_approval",
		Name:        "任务分配审批流程",
		Description: "直属领导审批任务分配，通过后任务状态从pending变更为assigned",
		Version:     "1.0",
		Nodes:       nodes,
		Edges:       edges,
		Variables: map[string]interface{}{
			"task_status_from": "pending",
			"task_status_to":   "assigned",
		},
	}

	// 创建流程定义
	_, err := workflowService.CreateWorkflowDefinition(ctx, createReq)
	if err != nil {
		return fmt.Errorf("创建任务分配审批流程定义失败: %w", err)
	}

	return nil
}

// createTaskCompletionWorkflowDefinition 创建任务完成审批工作流定义
// 业务逻辑：in_progress → done（直属领导审批）
func createTaskCompletionWorkflowDefinition(ctx context.Context, workflowService service.WorkflowService) error {
	// 定义流程节点
	nodes := []workflow.WorkflowNode{
		{
			ID:       "start",
			Type:     workflow.NodeTypeStart,
			Name:     "开始",
			Position: workflow.NodePosition{X: 100, Y: 100},
		},
		{
			ID:   "manager_approval",
			Type: workflow.NodeTypeApproval,
			Name: "直属领导审批",
			Config: map[string]interface{}{
				"assignees": []workflow.ApprovalAssignee{
					{
						Type:  workflow.AssigneeTypeManager,
						Value: "assignee",
					},
				},
				"approval_type": workflow.ApprovalTypeAny,
				"can_delegate":  false,
				"can_return":    true,
				"priority":      1,
				"description":   "审批任务完成，通过后任务状态变更为done",
			},
			Position: workflow.NodePosition{X: 300, Y: 100},
		},
		{
			ID:   "update_task_status",
			Type: workflow.NodeTypeScript,
			Name: "更新任务状态",
			Config: map[string]interface{}{
				"script": "update_task_status_to_done",
				"description": "将任务状态从in_progress更新为done",
			},
			Position: workflow.NodePosition{X: 500, Y: 100},
		},
		{
			ID:   "notify_completion",
			Type: workflow.NodeTypeNotify,
			Name: "通知任务完成",
			Config: map[string]interface{}{
				"type":       workflow.NotificationTypeSystem,
				"template":   "task_completed",
				"recipients": []string{"assignee", "requester"},
				"description": "通知相关人员任务已完成",
			},
			Position: workflow.NodePosition{X: 700, Y: 100},
		},
		{
			ID:       "end",
			Type:     workflow.NodeTypeEnd,
			Name:     "结束",
			Position: workflow.NodePosition{X: 900, Y: 100},
		},
	}

	// 定义流程边
	edges := []workflow.WorkflowEdge{
		{
			ID:   "start_to_manager",
			From: "start",
			To:   "manager_approval",
		},
		{
			ID:        "manager_to_update",
			From:      "manager_approval",
			To:        "update_task_status",
			Condition: "action == 'approve'",
		},
		{
			ID:   "update_to_notify",
			From: "update_task_status",
			To:   "notify_completion",
		},
		{
			ID:   "notify_to_end",
			From: "notify_completion",
			To:   "end",
		},
	}

	// 创建流程定义请求
	createReq := &workflow.CreateWorkflowRequest{
		ID:          "task_completion_approval",
		Name:        "任务完成审批流程",
		Description: "直属领导审批任务完成，通过后任务状态从in_progress变更为done",
		Version:     "1.0",
		Nodes:       nodes,
		Edges:       edges,
		Variables: map[string]interface{}{
			"task_status_from": "in_progress",
			"task_status_to":   "done",
		},
	}

	// 创建流程定义
	_, err := workflowService.CreateWorkflowDefinition(ctx, createReq)
	if err != nil {
		return fmt.Errorf("创建任务完成审批流程定义失败: %w", err)
	}

	return nil
}
