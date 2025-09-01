package service

import (
	"context"
	"encoding/json"
	"fmt"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/internal/workflow"
)

// WorkflowRepositoryAdapter 工作流仓库适配器
type WorkflowRepositoryAdapter struct {
	repo repository.WorkflowRepository
}

// NewWorkflowRepositoryAdapter 创建工作流仓库适配器
func NewWorkflowRepositoryAdapter(repo repository.WorkflowRepository) workflow.WorkflowRepository {
	return &WorkflowRepositoryAdapter{repo: repo}
}

// GetWorkflowDefinition 获取流程定义
func (a *WorkflowRepositoryAdapter) GetWorkflowDefinition(ctx context.Context, workflowID string) (*workflow.WorkflowDefinition, error) {
	dbDef, err := a.repo.GetWorkflowDefinition(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	return convertToWorkflowDefinition(dbDef)
}

// SaveWorkflowDefinition 保存流程定义
func (a *WorkflowRepositoryAdapter) SaveWorkflowDefinition(ctx context.Context, definition *workflow.WorkflowDefinition) error {
	dbDef, err := convertFromWorkflowDefinition(definition)
	if err != nil {
		return err
	}
	return a.repo.SaveWorkflowDefinition(ctx, dbDef)
}

// ListWorkflowDefinitions 列出流程定义
func (a *WorkflowRepositoryAdapter) ListWorkflowDefinitions(ctx context.Context, filter workflow.WorkflowFilter) ([]*workflow.WorkflowDefinition, error) {
	dbDefs, err := a.repo.ListWorkflowDefinitions(ctx, filter.IsActive, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	
	var definitions []*workflow.WorkflowDefinition
	for _, dbDef := range dbDefs {
		def, err := convertToWorkflowDefinition(dbDef)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, def)
	}
	return definitions, nil
}

// WorkflowInstanceRepositoryAdapter 工作流实例仓库适配器
type WorkflowInstanceRepositoryAdapter struct {
	repo repository.WorkflowInstanceRepository
}

// NewWorkflowInstanceRepositoryAdapter 创建工作流实例仓库适配器
func NewWorkflowInstanceRepositoryAdapter(repo repository.WorkflowInstanceRepository) workflow.WorkflowInstanceRepository {
	return &WorkflowInstanceRepositoryAdapter{repo: repo}
}

// SaveInstance 保存流程实例
func (a *WorkflowInstanceRepositoryAdapter) SaveInstance(ctx context.Context, instance *workflow.WorkflowInstance) error {
	dbInstance, err := convertFromWorkflowInstance(instance)
	if err != nil {
		return err
	}
	return a.repo.SaveInstance(ctx, dbInstance)
}

// GetInstance 获取流程实例
func (a *WorkflowInstanceRepositoryAdapter) GetInstance(ctx context.Context, instanceID string) (*workflow.WorkflowInstance, error) {
	dbInstance, err := a.repo.GetInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	return convertToWorkflowInstance(dbInstance)
}

// UpdateInstanceStatus 更新实例状态
func (a *WorkflowInstanceRepositoryAdapter) UpdateInstanceStatus(ctx context.Context, instanceID string, status workflow.InstanceStatus) error {
	return a.repo.UpdateInstanceStatus(ctx, instanceID, string(status))
}

// UpdateInstance 更新流程实例
func (a *WorkflowInstanceRepositoryAdapter) UpdateInstance(ctx context.Context, instance *workflow.WorkflowInstance) error {
	dbInstance, err := convertFromWorkflowInstance(instance)
	if err != nil {
		return err
	}
	return a.repo.UpdateInstance(ctx, dbInstance)
}

// AddExecutionHistory 添加执行历史
func (a *WorkflowInstanceRepositoryAdapter) AddExecutionHistory(ctx context.Context, instanceID string, history workflow.ExecutionHistory) error {
	dbHistory := &database.WorkflowExecutionHistory{
		HistoryID:  history.ID,
		InstanceID: instanceID,
		NodeID:     history.NodeID,
		NodeName:   history.NodeName,
		Action:     history.Action,
		Result:     history.Result,
		Comment:    history.Comment,
		Variables:  database.JSONField{Data: history.Variables},
		ExecutedBy: history.ExecutedBy,
		ExecutedAt: history.ExecutedAt,
		Duration:   int64(history.Duration.Milliseconds()),
	}
	return a.repo.AddExecutionHistory(ctx, dbHistory)
}

// GetPendingApprovals 获取待审批任务
func (a *WorkflowInstanceRepositoryAdapter) GetPendingApprovals(ctx context.Context, userID uint) ([]*workflow.PendingApproval, error) {
	dbApprovals, err := a.repo.GetPendingApprovals(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	var approvals []*workflow.PendingApproval
	for _, dbApproval := range dbApprovals {
		approval := &workflow.PendingApproval{
			InstanceID:     dbApproval.InstanceID,
			WorkflowName:   dbApproval.WorkflowName,
			NodeID:         dbApproval.NodeID,
			NodeName:       dbApproval.NodeName,
			BusinessID:     dbApproval.BusinessID,
			BusinessType:   dbApproval.BusinessType,
			BusinessData:   getMapFromJSONField(dbApproval.BusinessData),
			Priority:       dbApproval.Priority,
			AssignedTo:     dbApproval.AssignedTo,
			CreatedAt:      dbApproval.CreatedAt,
			Deadline:       dbApproval.Deadline,
			CanDelegate:    dbApproval.CanDelegate,
		}
		
		// 转换RequiredActions
		if dbApproval.RequiredActions.Data != nil {
			var requiredActions []workflow.ApprovalAction
			if actionList, ok := dbApproval.RequiredActions.Data.([]interface{}); ok {
				for _, action := range actionList {
					if actionStr, ok := action.(string); ok {
						requiredActions = append(requiredActions, workflow.ApprovalAction(actionStr))
					}
				}
			}
			approval.RequiredAction = requiredActions
		}
		
		approvals = append(approvals, approval)
	}
	return approvals, nil
}

// SavePendingApproval 保存待审批记录
func (a *WorkflowInstanceRepositoryAdapter) SavePendingApproval(ctx context.Context, approval *workflow.PendingApproval) error {
	dbApproval := &database.WorkflowPendingApproval{
		BaseModel: database.BaseModel{
			CreatedAt: approval.CreatedAt,
		},
		InstanceID:      approval.InstanceID,
		WorkflowName:    approval.WorkflowName,
		NodeID:          approval.NodeID,
		NodeName:        approval.NodeName,
		BusinessID:      approval.BusinessID,
		BusinessType:    approval.BusinessType,
		BusinessData:    database.JSONField{Data: approval.BusinessData},
		Priority:        approval.Priority,
		AssignedTo:      approval.AssignedTo,
		Deadline:        approval.Deadline,
		CanDelegate:     approval.CanDelegate,
		RequiredActions: database.JSONField{Data: approval.RequiredAction},
	}
	return a.repo.SavePendingApproval(ctx, dbApproval)
}

// DeletePendingApproval 删除待审批记录
func (a *WorkflowInstanceRepositoryAdapter) DeletePendingApproval(ctx context.Context, instanceID, nodeID string, userID uint) error {
	return a.repo.DeletePendingApproval(ctx, instanceID, nodeID, userID)
}

// 转换函数
func convertToWorkflowDefinition(dbDef *database.WorkflowDefinition) (*workflow.WorkflowDefinition, error) {
	var nodes []workflow.WorkflowNode
	if dbDef.Nodes.Data != nil {
		// 解析节点数据
		if nodeData, err := json.Marshal(dbDef.Nodes.Data); err == nil {
			if err := json.Unmarshal(nodeData, &nodes); err != nil {
				return nil, fmt.Errorf("解析工作流节点失败: %w", err)
			}
		}
	}
	
	var edges []workflow.WorkflowEdge
	if dbDef.Edges.Data != nil {
		// 解析边数据
		if edgeData, err := json.Marshal(dbDef.Edges.Data); err == nil {
			if err := json.Unmarshal(edgeData, &edges); err != nil {
				return nil, fmt.Errorf("解析工作流边失败: %w", err)
			}
		}
	}
	
	return &workflow.WorkflowDefinition{
		ID:          dbDef.WorkflowID,
		Name:        dbDef.Name,
		Description: dbDef.Description,
		Version:     dbDef.Version,
		Nodes:       nodes,
		Edges:       edges,
		Variables:   getMapFromJSONField(dbDef.Variables),
		CreatedAt:   dbDef.CreatedAt,
		UpdatedAt:   dbDef.UpdatedAt,
		IsActive:    dbDef.IsActive,
	}, nil
}

func convertFromWorkflowDefinition(def *workflow.WorkflowDefinition) (*database.WorkflowDefinition, error) {
	nodesJSON := database.JSONField{Data: def.Nodes}
	
	edgesJSON := database.JSONField{Data: def.Edges}
	
	return &database.WorkflowDefinition{
		WorkflowID:  def.ID,
		Name:        def.Name,
		Description: def.Description,
		Version:     def.Version,
		Nodes:       nodesJSON,
		Edges:       edgesJSON,
		Variables:   database.JSONField{Data: def.Variables},
		IsActive:    def.IsActive,
	}, nil
}

func convertToWorkflowInstance(dbInstance *database.WorkflowInstance) (*workflow.WorkflowInstance, error) {
	var currentNodes []string
	if dbInstance.CurrentNodes.Data != nil {
		if nodeList, ok := dbInstance.CurrentNodes.Data.([]interface{}); ok {
			for _, node := range nodeList {
				if nodeStr, ok := node.(string); ok {
					currentNodes = append(currentNodes, nodeStr)
				}
			}
		}
	}
	
	return &workflow.WorkflowInstance{
		ID:           dbInstance.InstanceID,
		WorkflowID:   dbInstance.WorkflowID,
		BusinessID:   dbInstance.BusinessID,
		BusinessType: dbInstance.BusinessType,
		Status:       workflow.InstanceStatus(dbInstance.Status),
		CurrentNodes: currentNodes,
		Variables:    getMapFromJSONField(dbInstance.Variables),
		StartedBy:    dbInstance.StartedBy,
		StartedAt:    dbInstance.StartedAt,
		CompletedAt:  dbInstance.CompletedAt,
		History:      []workflow.ExecutionHistory{}, // 需要单独查询
	}, nil
}

func convertFromWorkflowInstance(instance *workflow.WorkflowInstance) (*database.WorkflowInstance, error) {
	currentNodesJSON := database.JSONField{Data: instance.CurrentNodes}
	
	return &database.WorkflowInstance{
		InstanceID:   instance.ID,
		WorkflowID:   instance.WorkflowID,
		BusinessID:   instance.BusinessID,
		BusinessType: instance.BusinessType,
		Status:       string(instance.Status),
		CurrentNodes: currentNodesJSON,
		Variables:    database.JSONField{Data: instance.Variables},
		StartedBy:    instance.StartedBy,
		StartedAt:    instance.StartedAt,
		CompletedAt:  instance.CompletedAt,
	}, nil
}

// getMapFromJSONField 从JSONField中提取map[string]interface{}
func getMapFromJSONField(field database.JSONField) map[string]interface{} {
	if field.Data == nil {
		return nil
	}
	if mapData, ok := field.Data.(map[string]interface{}); ok {
		return mapData
	}
	return nil
}
