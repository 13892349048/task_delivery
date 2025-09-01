package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"taskmanage/pkg/logger"
	"time"
)

// Logger is imported from pkg/logger

// WorkflowDefinitionManager 流程定义管理器
type WorkflowDefinitionManager struct {
	repository WorkflowRepository
}

// NewWorkflowDefinitionManager 创建流程定义管理器
func NewWorkflowDefinitionManager(repository WorkflowRepository) *WorkflowDefinitionManager {
	return &WorkflowDefinitionManager{
		repository: repository,
	}
}

// CreateWorkflow 创建流程定义
func (m *WorkflowDefinitionManager) CreateWorkflow(ctx context.Context, req *CreateWorkflowRequest) (*WorkflowDefinition, error) {
	// 验证流程定义
	if err := m.validateWorkflowDefinition(req); err != nil {
		return nil, fmt.Errorf("流程定义验证失败: %w", err)
	}

	definition := &WorkflowDefinition{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Version:     req.Version,
		Nodes:       req.Nodes,
		Edges:       req.Edges,
		Variables:   req.Variables,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}

	if err := m.repository.SaveWorkflowDefinition(ctx, definition); err != nil {
		return nil, fmt.Errorf("保存流程定义失败: %w", err)
	}

	logger.Infof("创建流程定义成功: %s", definition.ID)
	return definition, nil
}

// UpdateWorkflow 更新流程定义
func (m *WorkflowDefinitionManager) UpdateWorkflow(ctx context.Context, workflowID string, req *UpdateWorkflowRequest) (*WorkflowDefinition, error) {
	// 获取现有定义
	existing, err := m.repository.GetWorkflowDefinition(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("获取流程定义失败: %w", err)
	}

	// 更新字段
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Version != "" {
		existing.Version = req.Version
	}
	if len(req.Nodes) > 0 {
		existing.Nodes = req.Nodes
	}
	if len(req.Edges) > 0 {
		existing.Edges = req.Edges
	}
	if req.Variables != nil {
		existing.Variables = req.Variables
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	existing.UpdatedAt = time.Now()

	// 验证更新后的定义
	validateReq := &CreateWorkflowRequest{
		ID:          existing.ID,
		Name:        existing.Name,
		Description: existing.Description,
		Version:     existing.Version,
		Nodes:       existing.Nodes,
		Edges:       existing.Edges,
		Variables:   existing.Variables,
	}
	if err := m.validateWorkflowDefinition(validateReq); err != nil {
		return nil, fmt.Errorf("更新后的流程定义验证失败: %w", err)
	}

	if err := m.repository.SaveWorkflowDefinition(ctx, existing); err != nil {
		return nil, fmt.Errorf("保存更新的流程定义失败: %w", err)
	}

	logger.Infof("更新流程定义成功: %s", workflowID)
	return existing, nil
}

// GetWorkflow 获取流程定义
func (m *WorkflowDefinitionManager) GetWorkflow(ctx context.Context, workflowID string) (*WorkflowDefinition, error) {
	definition, err := m.repository.GetWorkflowDefinition(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("获取流程定义失败: %w", err)
	}
	return definition, nil
}

// ListWorkflows 列出流程定义
func (m *WorkflowDefinitionManager) ListWorkflows(ctx context.Context, filter WorkflowFilter) ([]*WorkflowDefinition, error) {
	definitions, err := m.repository.ListWorkflowDefinitions(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("列出流程定义失败: %w", err)
	}
	return definitions, nil
}

// DeactivateWorkflow 停用流程定义
func (m *WorkflowDefinitionManager) DeactivateWorkflow(ctx context.Context, workflowID string) error {
	definition, err := m.repository.GetWorkflowDefinition(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("获取流程定义失败: %w", err)
	}

	definition.IsActive = false
	definition.UpdatedAt = time.Now()

	if err := m.repository.SaveWorkflowDefinition(ctx, definition); err != nil {
		return fmt.Errorf("停用流程定义失败: %w", err)
	}

	logger.Infof("停用流程定义成功: %s", workflowID)
	return nil
}

// ValidateWorkflow 验证流程定义
func (m *WorkflowDefinitionManager) ValidateWorkflow(ctx context.Context, req *CreateWorkflowRequest) error {
	return m.validateWorkflowDefinition(req)
}

// validateWorkflowDefinition 验证流程定义
func (m *WorkflowDefinitionManager) validateWorkflowDefinition(req *CreateWorkflowRequest) error {
	if req.ID == "" {
		return fmt.Errorf("流程ID不能为空")
	}
	if req.Name == "" {
		return fmt.Errorf("流程名称不能为空")
	}
	if len(req.Nodes) == 0 {
		return fmt.Errorf("流程必须包含至少一个节点")
	}

	// 验证节点
	nodeMap := make(map[string]*WorkflowNode)
	hasStart := false
	hasEnd := false

	for i, node := range req.Nodes {
		if node.ID == "" {
			return fmt.Errorf("节点[%d]ID不能为空", i)
		}
		if node.Type == "" {
			return fmt.Errorf("节点[%s]类型不能为空", node.ID)
		}
		if node.Name == "" {
			return fmt.Errorf("节点[%s]名称不能为空", node.ID)
		}

		// 检查重复节点ID
		if _, exists := nodeMap[node.ID]; exists {
			return fmt.Errorf("节点ID[%s]重复", node.ID)
		}
		nodeMap[node.ID] = &req.Nodes[i]

		// 检查开始和结束节点
		if node.Type == NodeTypeStart {
			hasStart = true
		}
		if node.Type == NodeTypeEnd {
			hasEnd = true
		}

		// 验证节点配置
		if err := m.validateNodeConfig(&req.Nodes[i]); err != nil {
			return fmt.Errorf("节点[%s]配置验证失败: %w", node.ID, err)
		}
	}

	if !hasStart {
		return fmt.Errorf("流程必须包含开始节点")
	}
	if !hasEnd {
		return fmt.Errorf("流程必须包含结束节点")
	}

	// 验证边
	for i, edge := range req.Edges {
		if edge.ID == "" {
			return fmt.Errorf("边[%d]ID不能为空", i)
		}
		if edge.From == "" {
			return fmt.Errorf("边[%s]起始节点不能为空", edge.ID)
		}
		if edge.To == "" {
			return fmt.Errorf("边[%s]目标节点不能为空", edge.ID)
		}

		// 检查节点是否存在
		if _, exists := nodeMap[edge.From]; !exists {
			return fmt.Errorf("边[%s]起始节点[%s]不存在", edge.ID, edge.From)
		}
		if _, exists := nodeMap[edge.To]; !exists {
			return fmt.Errorf("边[%s]目标节点[%s]不存在", edge.ID, edge.To)
		}
	}

	// 验证流程连通性
	if err := m.validateWorkflowConnectivity(nodeMap, req.Edges); err != nil {
		return fmt.Errorf("流程连通性验证失败: %w", err)
	}

	return nil
}

// validateNodeConfig 验证节点配置
func (m *WorkflowDefinitionManager) validateNodeConfig(node *WorkflowNode) error {
	switch node.Type {
	case NodeTypeApproval:
		return m.validateApprovalNodeConfig(node)
	case NodeTypeCondition:
		return m.validateConditionNodeConfig(node)
	case NodeTypeNotify:
		return m.validateNotifyNodeConfig(node)
	}
	return nil
}

// validateApprovalNodeConfig 验证审批节点配置
func (m *WorkflowDefinitionManager) validateApprovalNodeConfig(node *WorkflowNode) error {
	if node.Config == nil {
		return fmt.Errorf("审批节点必须包含配置")
	}

	configBytes, err := json.Marshal(node.Config)
	if err != nil {
		return fmt.Errorf("节点配置序列化失败: %w", err)
	}

	var config ApprovalNodeConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return fmt.Errorf("审批节点配置解析失败: %w", err)
	}

	if len(config.Assignees) == 0 {
		return fmt.Errorf("审批节点必须配置审批人")
	}

	for i, assignee := range config.Assignees {
		if assignee.Type == "" {
			return fmt.Errorf("审批人[%d]类型不能为空", i)
		}
		if assignee.Value == "" {
			return fmt.Errorf("审批人[%d]值不能为空", i)
		}
	}

	return nil
}

// validateConditionNodeConfig 验证条件节点配置
func (m *WorkflowDefinitionManager) validateConditionNodeConfig(node *WorkflowNode) error {
	if node.Config == nil {
		return fmt.Errorf("条件节点必须包含配置")
	}

	configBytes, err := json.Marshal(node.Config)
	if err != nil {
		return fmt.Errorf("节点配置序列化失败: %w", err)
	}

	var config ConditionNodeConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return fmt.Errorf("条件节点配置解析失败: %w", err)
	}

	if len(config.Conditions) == 0 {
		return fmt.Errorf("条件节点必须配置条件规则")
	}

	for i, condition := range config.Conditions {
		if condition.Expression == "" {
			return fmt.Errorf("条件规则[%d]表达式不能为空", i)
		}
		if condition.Target == "" {
			return fmt.Errorf("条件规则[%d]目标节点不能为空", i)
		}
	}

	return nil
}

// validateNotifyNodeConfig 验证通知节点配置
func (m *WorkflowDefinitionManager) validateNotifyNodeConfig(node *WorkflowNode) error {
	if node.Config == nil {
		return nil // 通知节点配置可选
	}

	configBytes, err := json.Marshal(node.Config)
	if err != nil {
		return fmt.Errorf("节点配置序列化失败: %w", err)
	}

	var config NotificationConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return fmt.Errorf("通知节点配置解析失败: %w", err)
	}

	if config.Type == "" {
		return fmt.Errorf("通知类型不能为空")
	}

	return nil
}

// validateWorkflowConnectivity 验证流程连通性
func (m *WorkflowDefinitionManager) validateWorkflowConnectivity(nodes map[string]*WorkflowNode, edges []WorkflowEdge) error {
	// 构建邻接表
	graph := make(map[string][]string)
	for _, edge := range edges {
		graph[edge.From] = append(graph[edge.From], edge.To)
	}

	// 找到开始节点
	var startNode string
	for nodeID, node := range nodes {
		if node.Type == NodeTypeStart {
			startNode = nodeID
			break
		}
	}

	if startNode == "" {
		return fmt.Errorf("未找到开始节点")
	}

	// DFS检查所有节点是否可达
	visited := make(map[string]bool)
	m.dfsVisit(startNode, graph, visited)

	// 检查是否所有节点都被访问
	for nodeID := range nodes {
		if !visited[nodeID] {
			return fmt.Errorf("节点[%s]不可达", nodeID)
		}
	}

	return nil
}

// dfsVisit DFS访问节点
func (m *WorkflowDefinitionManager) dfsVisit(nodeID string, graph map[string][]string, visited map[string]bool) {
	visited[nodeID] = true
	for _, neighbor := range graph[nodeID] {
		if !visited[neighbor] {
			m.dfsVisit(neighbor, graph, visited)
		}
	}
}

// CreateWorkflowRequest 创建流程请求
type CreateWorkflowRequest struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Nodes       []WorkflowNode         `json:"nodes"`
	Edges       []WorkflowEdge         `json:"edges"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
}

// UpdateWorkflowRequest 更新流程请求
type UpdateWorkflowRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Version     string                 `json:"version,omitempty"`
	Nodes       []WorkflowNode         `json:"nodes,omitempty"`
	Edges       []WorkflowEdge         `json:"edges,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty"`
}
