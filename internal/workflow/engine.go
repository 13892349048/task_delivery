package workflow

import (
	"context"
	"fmt"
	"time"

	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"

	"github.com/google/uuid"
)

// WorkflowEngineImpl 审批流程引擎实现
type WorkflowEngineImpl struct {
	definitionManager          *WorkflowDefinitionManager
	instanceRepo               WorkflowInstanceRepository
	taskExecutorRegistry       *ExecutorRegistry
	onboardingExecutorRegistry *ExecutorRegistry
}

// NewWorkflowEngine 创建流程引擎
func NewWorkflowEngine(
	definitionManager *WorkflowDefinitionManager,
	instanceRepo WorkflowInstanceRepository,
	employeeRepo repository.EmployeeRepository,
	userRepo repository.UserRepository,
) *WorkflowEngineImpl {
	return &WorkflowEngineImpl{
		definitionManager:          definitionManager,
		instanceRepo:               instanceRepo,
		taskExecutorRegistry:       NewExecutorRegistry(instanceRepo, employeeRepo, userRepo),
		onboardingExecutorRegistry: NewOnboardingExecutorRegistry(instanceRepo, employeeRepo, userRepo),
	}
}

// StartWorkflow 启动审批流程
func (e *WorkflowEngineImpl) StartWorkflow(ctx context.Context, req *StartWorkflowRequest) (*WorkflowInstance, error) {
	logger.Infof("启动流程: %s, 业务ID: %s", req.WorkflowID, req.BusinessID)

	// 获取流程定义
	definition, err := e.definitionManager.GetWorkflow(ctx, req.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("获取流程定义失败: %w", err)
	}

	if !definition.IsActive {
		return nil, fmt.Errorf("流程定义已停用")
	}

	// 创建流程实例
	instance := &WorkflowInstance{
		ID:           uuid.New().String(),
		WorkflowID:   req.WorkflowID,
		BusinessID:   req.BusinessID,
		BusinessType: req.BusinessType,
		Status:       StatusRunning,
		CurrentNodes: []string{},
		Variables:    req.Variables,
		StartedBy:    req.StartedBy,
		StartedAt:    time.Now(),
		History:      []ExecutionHistory{},
	}

	if instance.Variables == nil {
		instance.Variables = make(map[string]interface{})
	}

	// 保存实例
	if err := e.instanceRepo.SaveInstance(ctx, instance); err != nil {
		return nil, fmt.Errorf("保存流程实例失败: %w", err)
	}

	// 执行开始节点
	startNode := definition.GetStartNode()
	if startNode == nil {
		return nil, fmt.Errorf("流程定义中未找到开始节点")
	}

	if err := e.executeNode(ctx, instance, definition, startNode); err != nil {
		logger.Errorf("执行开始节点失败: %v", err)
		// 更新实例状态为失败
		e.instanceRepo.UpdateInstanceStatus(ctx, instance.ID, StatusFailed)
		return nil, fmt.Errorf("执行开始节点失败: %w", err)
	}

	// 重新获取更新后的实例
	updatedInstance, err := e.instanceRepo.GetInstance(ctx, instance.ID)
	if err != nil {
		return instance, nil // 返回原实例，避免启动失败
	}

	logger.Infof("流程启动成功: %s", instance.ID)
	return updatedInstance, nil
}

// ProcessApproval 处理审批决策
func (e *WorkflowEngineImpl) ProcessApproval(ctx context.Context, req *ApprovalRequest) (*ApprovalResult, error) {
	logger.Infof("处理审批: 实例=%s, 节点=%s, 动作=%s", req.InstanceID, req.NodeID, req.Action)

	// 获取流程实例
	instance, err := e.instanceRepo.GetInstance(ctx, req.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("获取流程实例失败: %w", err)
	}

	if instance.Status != StatusRunning {
		return nil, fmt.Errorf("流程实例状态不正确: %s", instance.Status)
	}

	// 获取流程定义
	definition, err := e.definitionManager.GetWorkflow(ctx, instance.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("获取流程定义失败: %w", err)
	}

	// 查找当前节点
	currentNode := e.findNodeByID(definition, req.NodeID)
	if currentNode == nil {
		return nil, fmt.Errorf("未找到节点: %s", req.NodeID)
	}

	// 验证节点是否在当前活跃节点中
	if !e.isNodeActive(instance, req.NodeID) {
		return nil, fmt.Errorf("节点不在活跃状态: %s", req.NodeID)
	}

	// 记录审批历史
	history := ExecutionHistory{
		ID:         uuid.New().String(),
		NodeID:     req.NodeID,
		NodeName:   currentNode.Name,
		Action:     string(req.Action),
		Result:     e.getApprovalResultString(req.Action),
		Comment:    req.Comment,
		Variables:  req.Variables,
		ExecutedBy: req.ApprovedBy,
		ExecutedAt: time.Now(),
		Duration:   0, // 审批节点持续时间需要单独计算
	}

	if err := e.instanceRepo.AddExecutionHistory(ctx, req.InstanceID, history); err != nil {
		logger.Errorf("添加执行历史失败: %v", err)
	}

	// 更新实例变量
	if req.Variables != nil {
		for k, v := range req.Variables {
			instance.Variables[k] = v
		}
	}

	// 处理审批结果
	var nextNodes []string
	var isCompleted bool
	var message string

	switch req.Action {
	case ActionApprove:
		// 审批通过，选择 approved 分支
		nextNodes = e.getNextNodesByCondition(definition, req.NodeID, "approved")
		message = "审批通过"
	case ActionReject:
		// 审批拒绝，选择 rejected 分支
		nextNodes = e.getNextNodesByCondition(definition, req.NodeID, "rejected")
		message = "审批拒绝"
	case ActionReturn:
		// 退回到上一个节点
		previousNodes := e.getPreviousNodes(definition, req.NodeID)
		nextNodes = previousNodes
		message = "审批退回"
	case ActionDelegate:
		// 委托给其他人，节点状态不变
		nextNodes = []string{req.NodeID}
		message = "审批已委托"
	default:
		return nil, fmt.Errorf("不支持的审批动作: %s", req.Action)
	}

	// 更新当前活跃节点
	instance.CurrentNodes = e.removeNode(instance.CurrentNodes, req.NodeID)
	instance.CurrentNodes = append(instance.CurrentNodes, nextNodes...)

	// 执行下一个节点
	if len(nextNodes) > 0 && !isCompleted {
		for _, nodeID := range nextNodes {
			nextNode := e.findNodeByID(definition, nodeID)
			if nextNode != nil {
				if err := e.executeNode(ctx, instance, definition, nextNode); err != nil {
					logger.Errorf("执行下一个节点失败: %s, error: %v", nodeID, err)
				}
			}
		}
	}

	// 检查流程是否完成
	if len(instance.CurrentNodes) == 0 || e.isWorkflowCompleted(instance, definition) {
		instance.Status = StatusCompleted
		completedAt := time.Now()
		instance.CompletedAt = &completedAt
		isCompleted = true
		message += "，流程已完成"
	}

	// 保存更新后的实例
	if err := e.instanceRepo.UpdateInstance(ctx, instance); err != nil {
		logger.Errorf("更新流程实例失败: %v", err)
	}

	result := &ApprovalResult{
		InstanceID:  req.InstanceID,
		NodeID:      req.NodeID,
		Action:      req.Action,
		NextNodes:   nextNodes,
		IsCompleted: isCompleted,
		Message:     message,
		ExecutedAt:  history.ExecutedAt,
	}

	logger.Infof("审批处理完成: %s", message)
	return result, nil
}

// GetWorkflowInstance 获取流程实例
func (e *WorkflowEngineImpl) GetWorkflowInstance(ctx context.Context, instanceID string) (*WorkflowInstance, error) {
	return e.instanceRepo.GetInstance(ctx, instanceID)
}

// GetPendingApprovals 获取待审批任务
func (e *WorkflowEngineImpl) GetPendingApprovals(ctx context.Context, userID uint) ([]*PendingApproval, error) {
	return e.instanceRepo.GetPendingApprovals(ctx, userID)
}

// CancelWorkflow 取消流程
func (e *WorkflowEngineImpl) CancelWorkflow(ctx context.Context, instanceID string, reason string) error {
	logger.Infof("取消流程: %s, 原因: %s", instanceID, reason)

	instance, err := e.instanceRepo.GetInstance(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("获取流程实例失败: %w", err)
	}

	if instance.Status != StatusRunning {
		return fmt.Errorf("只能取消运行中的流程")
	}

	// 添加取消历史记录
	history := ExecutionHistory{
		ID:         uuid.New().String(),
		NodeID:     "",
		NodeName:   "系统",
		Action:     "cancel",
		Result:     "cancelled",
		Comment:    reason,
		ExecutedBy: 0, // 系统操作
		ExecutedAt: time.Now(),
		Duration:   0,
	}

	if err := e.instanceRepo.AddExecutionHistory(ctx, instanceID, history); err != nil {
		logger.Errorf("添加取消历史失败: %v", err)
	}

	// 更新实例状态
	if err := e.instanceRepo.UpdateInstanceStatus(ctx, instanceID, StatusCancelled); err != nil {
		return fmt.Errorf("更新实例状态失败: %w", err)
	}

	logger.Infof("流程取消成功: %s", instanceID)
	return nil
}

// executeNode 执行节点
func (e *WorkflowEngineImpl) executeNode(ctx context.Context, instance *WorkflowInstance, definition *WorkflowDefinition, node *WorkflowNode) error {
	logger.Infof("执行节点: %s (%s)", node.ID, node.Type)

	startTime := time.Now()

	// 根据业务类型获取正确的执行器注册表
	registry := e.getExecutorRegistry(instance.BusinessType)

	// 获取节点执行器
	executor, err := registry.GetExecutor(node.Type)
	if err != nil {
		return fmt.Errorf("获取节点执行器失败: %w", err)
	}

	// 执行节点
	result, err := executor.ExecuteWithDefinition(ctx, instance, node, definition)
	if err != nil {
		return fmt.Errorf("节点执行失败: %w", err)
	}

	duration := time.Since(startTime)

	// 记录执行历史
	history := ExecutionHistory{
		ID:         uuid.New().String(),
		NodeID:     node.ID,
		NodeName:   node.Name,
		Action:     "execute",
		Result:     e.getExecutionResultString(result.Success),
		Comment:    result.Message,
		Variables:  result.Variables,
		ExecutedBy: instance.StartedBy,
		ExecutedAt: time.Now(),
		Duration:   duration,
	}

	if err := e.instanceRepo.AddExecutionHistory(ctx, instance.ID, history); err != nil {
		logger.Errorf("添加执行历史失败: %v", err)
	}

	// 更新实例变量
	if result.Variables != nil {
		for k, v := range result.Variables {
			instance.Variables[k] = v
		}
	}

	// 更新当前活跃节点
	if !result.WaitForUser {
		// 移除当前节点，添加下一个节点
		instance.CurrentNodes = e.removeNode(instance.CurrentNodes, node.ID)
		instance.CurrentNodes = append(instance.CurrentNodes, result.NextNodes...)

		// 继续执行下一个节点
		for _, nextNodeID := range result.NextNodes {
			nextNode := e.findNodeByID(definition, nextNodeID)
			if nextNode != nil {
				if err := e.executeNode(ctx, instance, definition, nextNode); err != nil {
					logger.Errorf("执行下一个节点失败: %s, error: %v", nextNodeID, err)
				}
			}
		}
	} else {
		// 需要等待用户操作，添加到活跃节点
		if !e.containsNode(instance.CurrentNodes, node.ID) {
			instance.CurrentNodes = append(instance.CurrentNodes, node.ID)
		}
	}

	// 更新实例的CurrentNodes和Variables到数据库
	if err := e.instanceRepo.UpdateInstance(ctx, instance); err != nil {
		logger.Errorf("更新实例失败: %v", err)
		return fmt.Errorf("更新实例失败: %w", err)
	}

	return nil
}

// 辅助方法
func (e *WorkflowEngineImpl) findStartNode(definition *WorkflowDefinition) *WorkflowNode {
	for i, node := range definition.Nodes {
		if node.Type == NodeTypeStart {
			return &definition.Nodes[i]
		}
	}
	return nil
}

func (e *WorkflowEngineImpl) findNodeByID(definition *WorkflowDefinition, nodeID string) *WorkflowNode {
	for i, node := range definition.Nodes {
		if node.ID == nodeID {
			return &definition.Nodes[i]
		}
	}
	return nil
}

func (e *WorkflowEngineImpl) isNodeActive(instance *WorkflowInstance, nodeID string) bool {
	for _, activeNode := range instance.CurrentNodes {
		if activeNode == nodeID {
			return true
		}
	}
	return false
}

func (e *WorkflowEngineImpl) getNextNodes(definition *WorkflowDefinition, nodeID string) []string {
	var nextNodes []string
	for _, edge := range definition.Edges {
		if edge.From == nodeID {
			nextNodes = append(nextNodes, edge.To)
		}
	}
	return nextNodes
}

// getNextNodesByCondition 根据条件获取下一个节点
func (e *WorkflowEngineImpl) getNextNodesByCondition(definition *WorkflowDefinition, nodeID string, condition string) []string {
	var nextNodes []string
	for _, edge := range definition.Edges {
		if edge.From == nodeID {
			// 如果边有条件，检查条件是否匹配
			if edge.Condition != "" && edge.Condition == condition {
				nextNodes = append(nextNodes, edge.To)
			} else if edge.Condition == "" && condition == "approved" {
				// 如果没有条件且是 approved，默认选择第一个节点
				nextNodes = append(nextNodes, edge.To)
				break
			}
		}
	}
	return nextNodes
}

func (e *WorkflowEngineImpl) getPreviousNodes(definition *WorkflowDefinition, nodeID string) []string {
	var previousNodes []string
	for _, edge := range definition.Edges {
		if edge.To == nodeID {
			previousNodes = append(previousNodes, edge.From)
		}
	}
	return previousNodes
}

func (e *WorkflowEngineImpl) removeNode(nodes []string, nodeID string) []string {
	var result []string
	for _, node := range nodes {
		if node != nodeID {
			result = append(result, node)
		}
	}
	return result
}

func (e *WorkflowEngineImpl) containsNode(nodes []string, nodeID string) bool {
	for _, node := range nodes {
		if node == nodeID {
			return true
		}
	}
	return false
}

func (e *WorkflowEngineImpl) isWorkflowCompleted(instance *WorkflowInstance, definition *WorkflowDefinition) bool {
	// 检查是否有结束节点在活跃节点中
	for _, nodeID := range instance.CurrentNodes {
		node := e.findNodeByID(definition, nodeID)
		if node != nil && node.Type == NodeTypeEnd {
			return true
		}
	}
	return false
}

func (e *WorkflowEngineImpl) getApprovalResultString(action ApprovalAction) string {
	switch action {
	case ActionApprove:
		return "approved"
	case ActionReject:
		return "rejected"
	case ActionReturn:
		return "returned"
	case ActionDelegate:
		return "delegated"
	default:
		return "unknown"
	}
}

func (e *WorkflowEngineImpl) getExecutionResultString(success bool) string {
	if success {
		return "success"
	}
	return "failed"
}

// getExecutorRegistry 根据业务类型获取执行器注册表
func (e *WorkflowEngineImpl) getExecutorRegistry(businessType string) *ExecutorRegistry {
	switch businessType {
	case "onboarding":
		return e.onboardingExecutorRegistry
	case "task_assignment":
		return e.taskExecutorRegistry
	default:
		// 默认使用任务分配执行器注册表
		return e.taskExecutorRegistry
	}
}
