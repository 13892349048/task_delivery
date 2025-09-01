package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// ExecutorRegistry 执行器注册表
type ExecutorRegistry struct {
	executors    map[NodeType]NodeExecutor
	instanceRepo WorkflowInstanceRepository
}

// NewExecutorRegistry 创建任务分配审批执行器注册表
func NewExecutorRegistry(instanceRepo WorkflowInstanceRepository, employeeRepo repository.EmployeeRepository, userRepo repository.UserRepository) *ExecutorRegistry {
	registry := &ExecutorRegistry{
		executors:    make(map[NodeType]NodeExecutor),
		instanceRepo: instanceRepo,
	}

	// 注册内置执行器 - 用于任务分配审批
	registry.RegisterExecutor(&StartNodeExecutor{registry: registry})
	registry.RegisterExecutor(&EndNodeExecutor{})
	registry.RegisterExecutor(&ApprovalNodeExecutor{instanceRepo: instanceRepo, employeeRepo: employeeRepo, userRepo: userRepo})
	registry.RegisterExecutor(&ConditionNodeExecutor{registry: registry})
	registry.RegisterExecutor(&ParallelNodeExecutor{registry: registry})
	registry.RegisterExecutor(&JoinNodeExecutor{registry: registry})
	registry.RegisterExecutor(&ScriptNodeExecutor{registry: registry})
	registry.RegisterExecutor(&NotifyNodeExecutor{registry: registry})

	return registry
}

// NewOnboardingExecutorRegistry 创建入职审批执行器注册表
func NewOnboardingExecutorRegistry(instanceRepo WorkflowInstanceRepository, employeeRepo repository.EmployeeRepository, userRepo repository.UserRepository) *ExecutorRegistry {
	registry := &ExecutorRegistry{
		executors:    make(map[NodeType]NodeExecutor),
		instanceRepo: instanceRepo,
	}

	// 注册内置执行器 - 用于入职审批
	registry.RegisterExecutor(&StartNodeExecutor{registry: registry})
	registry.RegisterExecutor(&EndNodeExecutor{})
	registry.RegisterExecutor(&OnboardingApprovalNodeExecutor{instanceRepo: instanceRepo, employeeRepo: employeeRepo, userRepo: userRepo})
	registry.RegisterExecutor(&ConditionNodeExecutor{registry: registry})
	registry.RegisterExecutor(&ParallelNodeExecutor{registry: registry})
	registry.RegisterExecutor(&JoinNodeExecutor{registry: registry})
	registry.RegisterExecutor(&ScriptNodeExecutor{registry: registry})
	registry.RegisterExecutor(&NotifyNodeExecutor{registry: registry})

	return registry
}

// RegisterExecutor 注册执行器
func (r *ExecutorRegistry) RegisterExecutor(executor NodeExecutor) {
	r.executors[executor.GetSupportedNodeType()] = executor
}

// GetExecutor 获取执行器
func (r *ExecutorRegistry) GetExecutor(nodeType NodeType) (NodeExecutor, error) {
	executor, exists := r.executors[nodeType]
	if !exists {
		return nil, fmt.Errorf("不支持的节点类型: %s", nodeType)
	}
	return executor, nil
}

// GetNextNodes 从工作流定义中获取下一个节点
func (r *ExecutorRegistry) GetNextNodes(definition *WorkflowDefinition, nodeID string) []string {
	var nextNodes []string
	for _, edge := range definition.Edges {
		if edge.From == nodeID {
			nextNodes = append(nextNodes, edge.To)
		}
	}
	return nextNodes
}

// StartNodeExecutor 开始节点执行器
type StartNodeExecutor struct {
	registry *ExecutorRegistry
}

func (e *StartNodeExecutor) Execute(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode) (*NodeExecutionResult, error) {
	return e.ExecuteWithDefinition(ctx, instance, node, nil)
}

func (e *StartNodeExecutor) ExecuteWithDefinition(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode, definition *WorkflowDefinition) (*NodeExecutionResult, error) {
	logger.Infof("执行开始节点: %s", node.ID)

	var nextNodes []string
	if definition != nil && e.registry != nil {
		nextNodes = e.registry.GetNextNodes(definition, node.ID)
	} else {
		nextNodes = e.getNextNodes(instance, node.ID)
	}

	// 开始节点直接执行成功，流向下一个节点
	return &NodeExecutionResult{
		Success:     true,
		NextNodes:   nextNodes,
		Variables:   nil,
		Message:     "流程已启动",
		WaitForUser: false,
	}, nil
}

func (e *StartNodeExecutor) GetSupportedNodeType() NodeType {
	return NodeTypeStart
}

func (e *StartNodeExecutor) getNextNodes(instance *WorkflowInstance, nodeID string) []string {
	// 这里简化处理，返回下一个节点ID
	// 实际应该从流程定义中获取
	return []string{"manager_approval"}
}

// EndNodeExecutor 结束节点执行器
type EndNodeExecutor struct{}

func (e *EndNodeExecutor) Execute(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode) (*NodeExecutionResult, error) {
	return e.ExecuteWithDefinition(ctx, instance, node, nil)
}

func (e *EndNodeExecutor) ExecuteWithDefinition(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode, definition *WorkflowDefinition) (*NodeExecutionResult, error) {
	logger.Infof("执行结束节点: %s", node.ID)

	return &NodeExecutionResult{
		Success:     true,
		NextNodes:   []string{}, // 等待审批完成
		Variables:   nil,
		Message:     "任务分配审批任务已创建，等待审批",
		WaitForUser: true,
	}, nil
}

func (e *EndNodeExecutor) GetSupportedNodeType() NodeType {
	return NodeTypeEnd
}

// ApprovalNodeExecutor 任务分配审批节点执行器
type ApprovalNodeExecutor struct {
	instanceRepo WorkflowInstanceRepository
	employeeRepo repository.EmployeeRepository
	userRepo     repository.UserRepository
}

// OnboardingApprovalNodeExecutor 入职审批节点执行器
type OnboardingApprovalNodeExecutor struct {
	instanceRepo WorkflowInstanceRepository
	employeeRepo repository.EmployeeRepository
	userRepo     repository.UserRepository
}

func (e *ApprovalNodeExecutor) Execute(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode) (*NodeExecutionResult, error) {
	return e.ExecuteWithDefinition(ctx, instance, node, nil)
}

func (e *ApprovalNodeExecutor) ExecuteWithDefinition(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode, definition *WorkflowDefinition) (*NodeExecutionResult, error) {
	logger.Infof("执行任务分配审批节点: %s", node.ID)

	// 解析审批节点配置
	config, err := e.parseApprovalConfig(node)
	if err != nil {
		logger.Errorf("解析审批节点配置失败: %v", err)
		return nil, fmt.Errorf("解析审批节点配置失败: %w", err)
	}

	// 解析审批人
	assignees, err := e.resolveAssignees(ctx, instance, config.Assignees)
	if err != nil {
		logger.Errorf("解析审批人失败: %v", err)
		return nil, fmt.Errorf("解析审批人失败: %w", err)
	}

	if len(assignees) == 0 {
		logger.Errorf("未找到有效的审批人")
		return nil, fmt.Errorf("未找到有效的审批人")
	}

	// 创建任务分配待审批记录
	for _, assigneeID := range assignees {
		pendingApproval := &PendingApproval{
			InstanceID:     instance.ID,
			WorkflowName:   "任务分配审批",
			NodeID:         node.ID,
			NodeName:       node.Name,
			BusinessID:     instance.BusinessID,
			BusinessType:   instance.BusinessType,
			BusinessData:   instance.Variables,
			Priority:       config.Priority,
			AssignedTo:     assigneeID,
			CreatedAt:      time.Now(),
			CanDelegate:    config.CanDelegate,
			RequiredAction: []ApprovalAction{ActionApprove, ActionReject},
		}

		if config.Deadline != nil {
			deadline := time.Now().Add(*config.Deadline)
			pendingApproval.Deadline = &deadline
		}

		if err := e.instanceRepo.SavePendingApproval(ctx, pendingApproval); err != nil {
			logger.Errorf("保存待审批记录失败: %v", err)
			return nil, fmt.Errorf("保存待审批记录失败: %w", err)
		}
		
		logger.Infof("成功创建任务分配待审批记录: InstanceID=%s, NodeID=%s, AssignedTo=%d", instance.ID, node.ID, assigneeID)
	}

	// 为相关用户创建查看记录（请求者和被分配者）
	stakeholders := []uint{}

	// 添加请求者
	if instance.StartedBy > 0 {
		stakeholders = append(stakeholders, instance.StartedBy)
	}

	// 添加被分配者（从业务数据中获取）
	if assigneeIDValue, exists := instance.Variables["assignee_id"]; exists {
		if assigneeID, ok := assigneeIDValue.(float64); ok {
			stakeholders = append(stakeholders, uint(assigneeID))
		}
	}

	// 为相关用户创建只读记录
	for _, stakeholderID := range stakeholders {
		// 避免重复创建（如果用户既是审批人又是相关人）
		isApprover := false
		for _, approverID := range assignees {
			if approverID == stakeholderID {
				isApprover = true
				break
			}
		}

		if !isApprover {
			logger.Infof("为相关用户 %d 创建查看记录", stakeholderID)

			viewRecord := &PendingApproval{
				InstanceID:     instance.ID,
				WorkflowName:   "任务分配审批",
				NodeID:         node.ID,
				NodeName:       node.Name,
				BusinessID:     instance.BusinessID,
				BusinessType:   instance.BusinessType,
				BusinessData:   instance.Variables,
				Priority:       config.Priority,
				AssignedTo:     stakeholderID,
				CreatedAt:      time.Now(),
				CanDelegate:    false,
				RequiredAction: []ApprovalAction{}, // 只能查看，不能操作
			}

			if err := e.instanceRepo.SavePendingApproval(ctx, viewRecord); err != nil {
				logger.Errorf("保存查看记录失败: %v", err)
			}
		}
	}

	return &NodeExecutionResult{
		Success:   true,
		NextNodes: []string{}, // 审批节点需要等待用户操作
		Variables: map[string]interface{}{
			"assignees":     assignees,
			"approval_type": config.ApprovalType,
		},
		Message:     fmt.Sprintf("已分配给 %d 个审批人", len(assignees)),
		WaitForUser: true, // 需要等待用户审批
	}, nil
}

func (e *ApprovalNodeExecutor) GetSupportedNodeType() NodeType {
	return NodeTypeApproval
}

func (e *ApprovalNodeExecutor) parseApprovalConfig(node *WorkflowNode) (*ApprovalNodeConfig, error) {
	if node.Config == nil {
		return nil, fmt.Errorf("审批节点配置为空")
	}

	// 检查是否是新格式（multiple assignees）
	if assigneeType, ok := node.Config["assignee_type"].(string); ok && assigneeType == "multiple" {
		return e.parseMultipleAssigneeConfig(node)
	}

	// 处理旧格式（单个assignee）
	return e.parseLegacyAssigneeConfig(node)
}

func (e *ApprovalNodeExecutor) parseMultipleAssigneeConfig(node *WorkflowNode) (*ApprovalNodeConfig, error) {
	config := &ApprovalNodeConfig{
		ApprovalType: ApprovalTypeAny, // 默认任意一人审批
		Priority:     1,
		CanDelegate:  true,
		CanReturn:    true,
	}

	// 解析assignees数组
	if assigneesData, ok := node.Config["assignees"]; ok {
		assigneesBytes, err := json.Marshal(assigneesData)
		if err != nil {
			return nil, fmt.Errorf("序列化assignees失败: %w", err)
		}

		var assignees []ApprovalAssignee
		if err := json.Unmarshal(assigneesBytes, &assignees); err != nil {
			return nil, fmt.Errorf("解析assignees失败: %w", err)
		}
		config.Assignees = assignees
	}

	// 解析其他配置项
	if timeout, ok := node.Config["timeout"].(float64); ok {
		duration := time.Duration(timeout) * time.Second
		config.Deadline = &duration
	}

	if autoApprove, ok := node.Config["auto_approve"].(bool); ok {
		config.AutoApprove = autoApprove
	}

	if priority, ok := node.Config["priority"].(float64); ok {
		config.Priority = int(priority)
	}

	return config, nil
}

func (e *ApprovalNodeExecutor) parseLegacyAssigneeConfig(node *WorkflowNode) (*ApprovalNodeConfig, error) {
	configBytes, err := json.Marshal(node.Config)
	if err != nil {
		return nil, fmt.Errorf("序列化节点配置失败: %w", err)
	}

	var config ApprovalNodeConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("解析审批节点配置失败: %w", err)
	}

	return &config, nil
}

func (e *ApprovalNodeExecutor) resolveAssignees(ctx context.Context, instance *WorkflowInstance, assigneeConfigs []ApprovalAssignee) ([]uint, error) {
	var assignees []uint
	logger.Infof("开始解析审批人，配置数量: %d", len(assigneeConfigs))

	for i, config := range assigneeConfigs {
		logger.Infof("处理审批人配置 %d: type=%s, value=%s", i+1, config.Type, config.Value)
		
		switch config.Type {
		case AssigneeTypeRole:
			// 根据角色查找用户
			logger.Infof("根据角色查找用户: %s", config.Value)
			roleUsers, err := e.getUsersByRole(ctx, config.Value)
			if err != nil {
				logger.Errorf("根据角色查找用户失败: role=%s, error=%v", config.Value, err)
				continue
			}
			logger.Infof("角色 %s 找到用户: %v", config.Value, roleUsers)
			assignees = append(assignees, roleUsers...)
		case AssigneeTypeDepartment:
			// 根据部门查找用户
			logger.Infof("根据部门查找用户: %s", config.Value)
			deptUsers, err := e.getUsersByDepartment(ctx, config.Value)
			if err != nil {
				logger.Errorf("根据部门查找用户失败: dept=%s, error=%v", config.Value, err)
				continue
			}
			logger.Infof("部门 %s 找到用户: %v", config.Value, deptUsers)
			assignees = append(assignees, deptUsers...)
		case AssigneeTypeManager:
			// 查找直属上级
			logger.Infof("查找用户 %d 的直属上级", instance.StartedBy)
			managerID, err := e.getManagerByUser(ctx, instance.StartedBy)
			if err != nil {
				logger.Errorf("查找直属上级失败: userID=%d, error=%v", instance.StartedBy, err)
				continue
			}
			logger.Infof("找到直属上级: %d", managerID)
			assignees = append(assignees, managerID)
		case "department_manager":
			// 查找部门经理 - 兼容新格式
			logger.Infof("查找用户 %d 的部门经理", instance.StartedBy)
			managerID, err := e.getManagerByUser(ctx, instance.StartedBy)
			if err != nil {
				logger.Errorf("查找部门经理失败: userID=%d, error=%v", instance.StartedBy, err)
				continue
			}
			logger.Infof("找到部门经理: %d", managerID)
			assignees = append(assignees, managerID)
		case AssigneeTypeStarter:
			// 流程发起人
			logger.Infof("添加流程发起人: %d", instance.StartedBy)
			assignees = append(assignees, instance.StartedBy)
		case AssigneeTypeVariable:
			// 从变量中获取
			logger.Infof("从变量获取审批人: %s", config.Value)
			if value, exists := instance.Variables[config.Value]; exists {
				if userID, ok := value.(uint); ok {
					logger.Infof("从变量找到用户: %d", userID)
					assignees = append(assignees, userID)
				} else {
					logger.Warnf("变量 %s 的值不是有效的用户ID: %v", config.Value, value)
				}
			} else {
				logger.Warnf("变量 %s 不存在", config.Value)
			}
		default:
			logger.Warnf("未知的审批人类型: %s", config.Type)
		}
	}

	// 去重
	uniqueAssignees := make([]uint, 0, len(assignees))
	seen := make(map[uint]bool)
	for _, assignee := range assignees {
		if !seen[assignee] {
			uniqueAssignees = append(uniqueAssignees, assignee)
			seen[assignee] = true
		}
	}

	return uniqueAssignees, nil
}

func (e *ApprovalNodeExecutor) getUsersByRole(ctx context.Context, role string) ([]uint, error) {
	// 根据角色查找用户
	users, err := e.userRepo.GetUsersByRole(ctx, role)
	if err != nil {
		logger.Errorf("根据角色查找用户失败: role=%s, error=%v", role, err)
		return nil, fmt.Errorf("根据角色查找用户失败: %w", err)
	}
	
	var userIDs []uint
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}
	
	logger.Infof("找到角色 %s 的用户: %v", role, userIDs)
	return userIDs, nil
}

func (e *ApprovalNodeExecutor) getUsersByDepartment(ctx context.Context, department string) ([]uint, error) {
	// 这里应该调用用户服务查找部门用户
	// 简化处理，返回模拟数据
	return []uint{3, 4}, nil
}

func (e *ApprovalNodeExecutor) getManagerByUser(ctx context.Context, userID uint) (uint, error) {
	// 根据用户ID查找员工信息
	employee, err := e.employeeRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Errorf("查找员工失败: userID=%d, error=%v", userID, err)
		return 0, fmt.Errorf("查找员工失败: %w", err)
	}

	// 检查是否有直属领导
	if employee.DirectManagerID == nil {
		logger.Warnf("员工 %d 没有设置直属领导", userID)
		return 0, fmt.Errorf("员工没有设置直属领导")
	}

	// 获取直属领导的员工信息
	manager, err := e.employeeRepo.GetByID(ctx, *employee.DirectManagerID)
	if err != nil {
		logger.Errorf("查找直属领导失败: managerID=%d, error=%v", *employee.DirectManagerID, err)
		return 0, fmt.Errorf("查找直属领导失败: %w", err)
	}

	logger.Infof("找到员工 %d 的直属领导: %d", userID, manager.UserID)
	return manager.UserID, nil
}

// ConditionNodeExecutor 条件节点执行器
type ConditionNodeExecutor struct {
	registry *ExecutorRegistry
}

func (e *ConditionNodeExecutor) Execute(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode) (*NodeExecutionResult, error) {
	return e.ExecuteWithDefinition(ctx, instance, node, nil)
}

func (e *ConditionNodeExecutor) ExecuteWithDefinition(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode, definition *WorkflowDefinition) (*NodeExecutionResult, error) {
	logger.Infof("执行条件节点: %s", node.ID)

	// 解析条件节点配置
	config, err := e.parseConditionConfig(node)
	if err != nil {
		return nil, fmt.Errorf("解析条件节点配置失败: %w", err)
	}

	// 评估条件并确定下一个节点
	nextNodes, err := e.evaluateConditions(instance, config.Conditions)
	if err != nil {
		return nil, fmt.Errorf("评估条件失败: %w", err)
	}

	return &NodeExecutionResult{
		Success:     true,
		NextNodes:   nextNodes,
		Variables:   nil,
		Message:     fmt.Sprintf("条件评估完成，流向 %d 个节点", len(nextNodes)),
		WaitForUser: false,
	}, nil
}

func (e *ConditionNodeExecutor) GetSupportedNodeType() NodeType {
	return NodeTypeCondition
}

func (e *ConditionNodeExecutor) parseConditionConfig(node *WorkflowNode) (*ConditionNodeConfig, error) {
	if node.Config == nil {
		return nil, fmt.Errorf("条件节点配置为空")
	}

	configBytes, err := json.Marshal(node.Config)
	if err != nil {
		return nil, fmt.Errorf("序列化节点配置失败: %w", err)
	}

	var config ConditionNodeConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("解析条件节点配置失败: %w", err)
	}

	return &config, nil
}

func (e *ConditionNodeExecutor) evaluateConditions(instance *WorkflowInstance, conditions []ConditionRule) ([]string, error) {
	var nextNodes []string

	// 按优先级排序条件
	sortedConditions := make([]ConditionRule, len(conditions))
	copy(sortedConditions, conditions)

	// 简单排序（实际应该使用sort包）
	for i := 0; i < len(sortedConditions)-1; i++ {
		for j := i + 1; j < len(sortedConditions); j++ {
			if sortedConditions[i].Priority < sortedConditions[j].Priority {
				sortedConditions[i], sortedConditions[j] = sortedConditions[j], sortedConditions[i]
			}
		}
	}

	// 评估每个条件
	for _, condition := range sortedConditions {
		result, err := e.evaluateExpression(condition.Expression, instance.Variables)
		if err != nil {
			logger.Errorf("评估条件表达式失败: %s, error: %v", condition.Expression, err)
			continue
		}

		if result {
			nextNodes = append(nextNodes, condition.Target)
			logger.Infof("条件 %s 评估为真，流向节点 %s", condition.Expression, condition.Target)
		}
	}

	return nextNodes, nil
}

func (e *ConditionNodeExecutor) evaluateExpression(expression string, variables map[string]interface{}) (bool, error) {
	// 简化的表达式评估器，支持基本的比较操作
	// 格式: "variable operator value"
	parts := strings.Fields(expression)
	if len(parts) != 3 {
		return false, fmt.Errorf("表达式格式错误，应为: variable operator value")
	}

	varName := parts[0]
	operator := parts[1]
	expectedValue := parts[2]

	// 获取变量值
	actualValue, exists := variables[varName]
	if !exists {
		return false, fmt.Errorf("变量不存在: %s", varName)
	}

	// 转换为字符串进行比较
	actualStr := fmt.Sprintf("%v", actualValue)
	expectedValue = strings.Trim(expectedValue, "'\"") // 移除引号

	// 根据操作符进行比较
	switch operator {
	case "==", "=":
		return actualStr == expectedValue, nil
	case "!=", "<>":
		return actualStr != expectedValue, nil
	case ">":
		return e.compareNumeric(actualStr, expectedValue, ">")
	case "<":
		return e.compareNumeric(actualStr, expectedValue, "<")
	case ">=":
		return e.compareNumeric(actualStr, expectedValue, ">=")
	case "<=":
		return e.compareNumeric(actualStr, expectedValue, "<=")
	default:
		return false, fmt.Errorf("不支持的操作符: %s", operator)
	}
}

func (e *ConditionNodeExecutor) compareNumeric(actual, expected, operator string) (bool, error) {
	actualNum, err1 := strconv.ParseFloat(actual, 64)
	expectedNum, err2 := strconv.ParseFloat(expected, 64)

	if err1 != nil || err2 != nil {
		// 如果不是数字，按字符串比较
		switch operator {
		case ">":
			return actual > expected, nil
		case "<":
			return actual < expected, nil
		case ">=":
			return actual >= expected, nil
		case "<=":
			return actual <= expected, nil
		}
	}

	switch operator {
	case ">":
		return actualNum > expectedNum, nil
	case "<":
		return actualNum < expectedNum, nil
	case ">=":
		return actualNum >= expectedNum, nil
	case "<=":
		return actualNum <= expectedNum, nil
	}

	return false, nil
}

// ParallelNodeExecutor 并行节点执行器
type ParallelNodeExecutor struct {
	registry *ExecutorRegistry
}

func (e *ParallelNodeExecutor) Execute(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode) (*NodeExecutionResult, error) {
	return e.ExecuteWithDefinition(ctx, instance, node, nil)
}

func (e *ParallelNodeExecutor) ExecuteWithDefinition(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode, definition *WorkflowDefinition) (*NodeExecutionResult, error) {
	logger.Infof("执行并行节点: %s", node.ID)

	// 并行节点将流程分发到多个分支
	nextNodes := e.getParallelBranches(instance, node.ID)

	return &NodeExecutionResult{
		Success:     true,
		NextNodes:   nextNodes,
		Variables:   nil,
		Message:     fmt.Sprintf("并行分发到 %d 个分支", len(nextNodes)),
		WaitForUser: false,
	}, nil
}

func (e *ParallelNodeExecutor) GetSupportedNodeType() NodeType {
	return NodeTypeParallel
}

func (e *ParallelNodeExecutor) getParallelBranches(instance *WorkflowInstance, nodeID string) []string {
	// 这里应该从流程定义中获取并行分支
	// 简化处理，返回模拟数据
	return []string{"branch1", "branch2"}
}

// JoinNodeExecutor 汇聚节点执行器
type JoinNodeExecutor struct {
	registry *ExecutorRegistry
}

func (e *JoinNodeExecutor) Execute(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode) (*NodeExecutionResult, error) {
	return e.ExecuteWithDefinition(ctx, instance, node, nil)
}

func (e *JoinNodeExecutor) ExecuteWithDefinition(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode, definition *WorkflowDefinition) (*NodeExecutionResult, error) {
	logger.Infof("执行汇聚节点: %s", node.ID)

	// 检查所有前置节点是否都已完成
	completed, err := e.checkPredecessors(instance, node)
	if err != nil {
		return nil, fmt.Errorf("检查前置节点失败: %w", err)
	}

	if !completed {
		// 还有前置节点未完成，等待
		return &NodeExecutionResult{
			Success:     true,
			NextNodes:   []string{},
			Variables:   nil,
			Message:     "等待前置节点完成",
			WaitForUser: false,
		}, nil
	}

	// 所有前置节点已完成，继续下一个节点
	nextNodes := e.getNextNodes(instance, node.ID)

	return &NodeExecutionResult{
		Success:     true,
		NextNodes:   nextNodes,
		Variables:   nil,
		Message:     "汇聚完成",
		WaitForUser: false,
	}, nil
}

func (e *JoinNodeExecutor) GetSupportedNodeType() NodeType {
	return NodeTypeJoin
}

func (e *JoinNodeExecutor) checkPredecessors(instance *WorkflowInstance, node *WorkflowNode) (bool, error) {
	// 这里应该检查所有前置节点是否都已完成
	// 简化处理，返回true
	return true, nil
}

func (e *JoinNodeExecutor) getNextNodes(instance *WorkflowInstance, nodeID string) []string {
	// 这里应该从流程定义中获取下一个节点
	// 简化处理，返回模拟数据
	return []string{"next_node"}
}

// ScriptNodeExecutor 脚本节点执行器
type ScriptNodeExecutor struct {
	registry *ExecutorRegistry
}

func (e *ScriptNodeExecutor) Execute(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode) (*NodeExecutionResult, error) {
	return e.ExecuteWithDefinition(ctx, instance, node, nil)
}

func (e *ScriptNodeExecutor) ExecuteWithDefinition(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode, definition *WorkflowDefinition) (*NodeExecutionResult, error) {
	logger.Infof("执行脚本节点: %s", node.ID)

	// 执行脚本逻辑
	variables, err := e.executeScript(instance, node)
	if err != nil {
		return nil, fmt.Errorf("执行脚本失败: %w", err)
	}

	nextNodes := e.getNextNodes(instance, node.ID)

	return &NodeExecutionResult{
		Success:     true,
		NextNodes:   nextNodes,
		Variables:   variables,
		Message:     "脚本执行完成",
		WaitForUser: false,
	}, nil
}

func (e *ScriptNodeExecutor) GetSupportedNodeType() NodeType {
	return NodeTypeScript
}

func (e *ScriptNodeExecutor) executeScript(instance *WorkflowInstance, node *WorkflowNode) (map[string]interface{}, error) {
	// 这里应该执行实际的脚本逻辑
	// 简化处理，返回模拟变量
	return map[string]interface{}{
		"script_result": "success",
		"timestamp":     time.Now(),
	}, nil
}

func (e *ScriptNodeExecutor) getNextNodes(instance *WorkflowInstance, nodeID string) []string {
	// 这里应该从流程定义中获取下一个节点
	// 简化处理，返回模拟数据
	return []string{"next_node"}
}

// NotifyNodeExecutor 通知节点执行器
type NotifyNodeExecutor struct {
	registry *ExecutorRegistry
}

func (e *NotifyNodeExecutor) Execute(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode) (*NodeExecutionResult, error) {
	return e.ExecuteWithDefinition(ctx, instance, node, nil)
}

func (e *NotifyNodeExecutor) ExecuteWithDefinition(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode, definition *WorkflowDefinition) (*NodeExecutionResult, error) {
	logger.Infof("执行通知节点: %s", node.ID)

	// 发送通知
	err := e.sendNotification(ctx, instance, node)
	if err != nil {
		logger.Errorf("发送通知失败: %v", err)
		// 通知失败不阻止流程继续
	}

	nextNodes := e.getNextNodes(instance, node.ID)

	return &NodeExecutionResult{
		Success:     true,
		NextNodes:   nextNodes,
		Variables:   nil,
		Message:     "通知已发送",
		WaitForUser: false,
	}, nil
}

func (e *NotifyNodeExecutor) GetSupportedNodeType() NodeType {
	return NodeTypeNotify
}

func (e *NotifyNodeExecutor) sendNotification(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode) error {
	// 这里应该发送实际的通知
	// 简化处理，只记录日志
	logger.Infof("发送通知: 流程 %s 在节点 %s", instance.ID, node.ID)
	return nil
}

func (e *NotifyNodeExecutor) getNextNodes(instance *WorkflowInstance, nodeID string) []string {
	// 这里应该从流程定义中获取下一个节点
	// 简化处理，返回模拟数据
	return []string{"next_node"}
}

// OnboardingApprovalNodeExecutor 入职审批执行器方法
func (e *OnboardingApprovalNodeExecutor) Execute(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode) (*NodeExecutionResult, error) {
	return e.ExecuteWithDefinition(ctx, instance, node, nil)
}

func (e *OnboardingApprovalNodeExecutor) ExecuteWithDefinition(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode, definition *WorkflowDefinition) (*NodeExecutionResult, error) {
	logger.Infof("执行入职审批节点: %s", node.ID)

	// 解析审批节点配置
	config, err := e.parseApprovalConfig(node)
	if err != nil {
		logger.Errorf("解析入职审批节点配置失败: %v", err)
		return nil, fmt.Errorf("解析入职审批节点配置失败: %w", err)
	}
	logger.Infof("入职审批节点配置解析成功，审批人配置数量: %d", len(config.Assignees))

	// 解析审批人
	assignees, err := e.resolveAssignees(ctx, instance, config.Assignees)
	if err != nil {
		logger.Errorf("解析入职审批人失败: %v", err)
		return nil, fmt.Errorf("解析入职审批人失败: %w", err)
	}
	logger.Infof("解析到入职审批人: %v", assignees)

	if len(assignees) == 0 {
		logger.Errorf("未找到有效的入职审批人")
		return nil, fmt.Errorf("未找到有效的入职审批人")
	}

	// 创建入职待审批记录
	for _, assigneeID := range assignees {
		logger.Infof("为用户 %d 创建入职审批任务", assigneeID)

		pendingApproval := &PendingApproval{
			InstanceID:     instance.ID,
			WorkflowName:   "入职审批",
			NodeID:         node.ID,
			NodeName:       node.Name,
			BusinessID:     instance.BusinessID,
			BusinessType:   instance.BusinessType,
			BusinessData:   instance.Variables,
			Priority:       config.Priority,
			AssignedTo:     assigneeID,
			CreatedAt:      time.Now(),
			CanDelegate:    config.CanDelegate,
			RequiredAction: []ApprovalAction{ActionApprove, ActionReject},
		}

		if config.Deadline != nil {
			deadline := time.Now().Add(*config.Deadline)
			pendingApproval.Deadline = &deadline
		}

		// 保存待审批记录
		if err := e.instanceRepo.SavePendingApproval(ctx, pendingApproval); err != nil {
			logger.Errorf("保存入职待审批记录失败: %v", err)
			return nil, fmt.Errorf("保存入职待审批记录失败: %w", err)
		}
		
		logger.Infof("成功创建入职待审批记录: InstanceID=%s, NodeID=%s, AssignedTo=%d", instance.ID, node.ID, assigneeID)
	}

	return &NodeExecutionResult{
		Success:     true,
		NextNodes:   []string{}, // 等待审批完成
		Variables:   nil,
		Message:     "入职审批任务已创建，等待审批",
		WaitForUser: true,
	}, nil
}

func (e *OnboardingApprovalNodeExecutor) GetSupportedNodeType() NodeType {
	return NodeTypeApproval
}

// 入职审批执行器需要复用任务审批执行器的解析方法
func (e *OnboardingApprovalNodeExecutor) parseApprovalConfig(node *WorkflowNode) (*ApprovalNodeConfig, error) {
	// 创建临时的任务审批执行器来复用解析逻辑
	tempExecutor := &ApprovalNodeExecutor{
		instanceRepo: e.instanceRepo,
		employeeRepo: e.employeeRepo,
		userRepo:     e.userRepo,
	}
	return tempExecutor.parseApprovalConfig(node)
}

func (e *OnboardingApprovalNodeExecutor) resolveAssignees(ctx context.Context, instance *WorkflowInstance, assignees []ApprovalAssignee) ([]uint, error) {
	// 创建临时的任务审批执行器来复用解析逻辑
	tempExecutor := &ApprovalNodeExecutor{
		instanceRepo: e.instanceRepo,
		employeeRepo: e.employeeRepo,
		userRepo:     e.userRepo,
	}
	return tempExecutor.resolveAssignees(ctx, instance, assignees)
}
