package mysql

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/internal/workflow"
)

// WorkflowRepositoryImpl 工作流定义仓库实现
type WorkflowRepositoryImpl struct {
	db *gorm.DB
}

// NewWorkflowRepository 创建工作流定义仓库
func NewWorkflowRepository(db *gorm.DB) repository.WorkflowRepository {
	return &WorkflowRepositoryImpl{db: db}
}

// GetWorkflowDefinition 获取流程定义
func (r *WorkflowRepositoryImpl) GetWorkflowDefinition(ctx context.Context, workflowID string) (*database.WorkflowDefinition, error) {
	var definition database.WorkflowDefinition
	err := r.db.WithContext(ctx).Where("workflow_id = ? AND is_active = ?", workflowID, true).First(&definition).Error
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

// SaveWorkflowDefinition 保存流程定义
func (r *WorkflowRepositoryImpl) SaveWorkflowDefinition(ctx context.Context, definition *database.WorkflowDefinition) error {
	return r.db.WithContext(ctx).Save(definition).Error
}

// ListWorkflowDefinitions 列出流程定义
func (r *WorkflowRepositoryImpl) ListWorkflowDefinitions(ctx context.Context, isActive *bool, limit, offset int) ([]*database.WorkflowDefinition, error) {
	var definitions []*database.WorkflowDefinition
	query := r.db.WithContext(ctx)
	
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&definitions).Error
	return definitions, err
}

// UpdateWorkflowStatus 更新流程状态
func (r *WorkflowRepositoryImpl) UpdateWorkflowStatus(ctx context.Context, workflowID string, isActive bool) error {
	return r.db.WithContext(ctx).Model(&database.WorkflowDefinition{}).
		Where("workflow_id = ?", workflowID).
		Update("is_active", isActive).Error
}

// WorkflowInstanceRepositoryImpl 工作流实例仓库实现
type WorkflowInstanceRepositoryImpl struct {
	db *gorm.DB
}

// NewWorkflowInstanceRepository 创建工作流实例仓库
func NewWorkflowInstanceRepository(db *gorm.DB) repository.WorkflowInstanceRepository {
	return &WorkflowInstanceRepositoryImpl{db: db}
}

// SaveInstance 保存流程实例
func (r *WorkflowInstanceRepositoryImpl) SaveInstance(ctx context.Context, instance *database.WorkflowInstance) error {
	return r.db.WithContext(ctx).Save(instance).Error
}

// GetInstance 获取流程实例
func (r *WorkflowInstanceRepositoryImpl) GetInstance(ctx context.Context, instanceID string) (*database.WorkflowInstance, error) {
	var instance database.WorkflowInstance
	err := r.db.WithContext(ctx).Where("instance_id = ?", instanceID).First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// UpdateInstanceStatus 更新实例状态
func (r *WorkflowInstanceRepositoryImpl) UpdateInstanceStatus(ctx context.Context, instanceID string, status string) error {
	return r.db.WithContext(ctx).Model(&database.WorkflowInstance{}).
		Where("instance_id = ?", instanceID).
		Update("status", status).Error
}

// UpdateInstanceNodes 更新实例当前节点
func (r *WorkflowInstanceRepositoryImpl) UpdateInstanceNodes(ctx context.Context, instanceID string, currentNodes []string) error {
	nodesJSON, err := json.Marshal(currentNodes)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&database.WorkflowInstance{}).
		Where("instance_id = ?", instanceID).
		Update("current_nodes", nodesJSON).Error
}

// UpdateInstance 更新流程实例
func (r *WorkflowInstanceRepositoryImpl) UpdateInstance(ctx context.Context, instance *database.WorkflowInstance) error {
	updates := map[string]interface{}{
		"current_nodes": instance.CurrentNodes,
		"variables":     instance.Variables,
		"status":        instance.Status,
	}
	
	// 只有当 UpdatedAt 不为零值时才更新
	if !instance.UpdatedAt.IsZero() {
		updates["updated_at"] = instance.UpdatedAt
	}
	
	return r.db.WithContext(ctx).Model(&database.WorkflowInstance{}).
		Where("instance_id = ?", instance.InstanceID).
		Updates(updates).Error
}

// AddExecutionHistory 添加执行历史
func (r *WorkflowInstanceRepositoryImpl) AddExecutionHistory(ctx context.Context, history *database.WorkflowExecutionHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetExecutionHistory 获取执行历史
func (r *WorkflowInstanceRepositoryImpl) GetExecutionHistory(ctx context.Context, instanceID string) ([]*database.WorkflowExecutionHistory, error) {
	var histories []*database.WorkflowExecutionHistory
	err := r.db.WithContext(ctx).Where("instance_id = ?", instanceID).
		Order("executed_at ASC").Find(&histories).Error
	return histories, err
}

// GetPendingApprovals 获取待审批任务
func (r *WorkflowInstanceRepositoryImpl) GetPendingApprovals(ctx context.Context, userID uint) ([]*database.WorkflowPendingApproval, error) {
	var approvals []*database.WorkflowPendingApproval
	err := r.db.WithContext(ctx).Where("assigned_to = ? AND is_completed = ?", userID, false).
		Order("priority DESC, created_at ASC").Find(&approvals).Error
	return approvals, err
}

// CreatePendingApproval 创建待审批任务
func (r *WorkflowInstanceRepositoryImpl) CreatePendingApproval(ctx context.Context, approval *database.WorkflowPendingApproval) error {
	return r.db.WithContext(ctx).Create(approval).Error
}

// CompletePendingApproval 完成待审批任务
func (r *WorkflowInstanceRepositoryImpl) CompletePendingApproval(ctx context.Context, instanceID, nodeID string, userID uint) error {
	return r.db.WithContext(ctx).Model(&database.WorkflowPendingApproval{}).
		Where("instance_id = ? AND node_id = ? AND assigned_to = ?", instanceID, nodeID, userID).
		Update("is_completed", true).Error
}

// SavePendingApproval 保存待审批记录
func (r *WorkflowInstanceRepositoryImpl) SavePendingApproval(ctx context.Context, approval *database.WorkflowPendingApproval) error {
	return r.db.WithContext(ctx).Save(approval).Error
}

// DeletePendingApproval 删除待审批记录
func (r *WorkflowInstanceRepositoryImpl) DeletePendingApproval(ctx context.Context, instanceID, nodeID string, userID uint) error {
	return r.db.WithContext(ctx).Where("instance_id = ? AND node_id = ? AND assigned_to = ?", instanceID, nodeID, userID).
		Delete(&database.WorkflowPendingApproval{}).Error
}

// GetInstancesByBusinessID 根据业务ID获取实例
func (r *WorkflowInstanceRepositoryImpl) GetInstancesByBusinessID(ctx context.Context, businessID, businessType string) ([]*database.WorkflowInstance, error) {
	var instances []*database.WorkflowInstance
	err := r.db.WithContext(ctx).Where("business_id = ? AND business_type = ?", businessID, businessType).
		Order("created_at DESC").Find(&instances).Error
	return instances, err
}

// ConvertToWorkflowInstance 转换数据库模型到workflow模型
func ConvertToWorkflowInstance(dbInstance *database.WorkflowInstance) (*workflow.WorkflowInstance, error) {
	var currentNodes []string
	if dbInstance.CurrentNodes.Data != nil {
		// 从JSONField中提取节点列表
		if nodeList, ok := dbInstance.CurrentNodes.Data.([]interface{}); ok {
			for _, node := range nodeList {
				if nodeStr, ok := node.(string); ok {
					currentNodes = append(currentNodes, nodeStr)
				}
			}
		}
	}

	var variables map[string]interface{}
	if dbInstance.Variables.Data != nil {
		if varMap, ok := dbInstance.Variables.Data.(map[string]interface{}); ok {
			variables = varMap
		}
	}

	// 获取执行历史需要额外查询，这里先返回空
	var history []workflow.ExecutionHistory

	return &workflow.WorkflowInstance{
		ID:           dbInstance.InstanceID,
		WorkflowID:   dbInstance.WorkflowID,
		BusinessID:   dbInstance.BusinessID,
		BusinessType: dbInstance.BusinessType,
		Status:       workflow.InstanceStatus(dbInstance.Status),
		CurrentNodes: currentNodes,
		Variables:    variables,
		StartedBy:    dbInstance.StartedBy,
		StartedAt:    dbInstance.StartedAt,
		CompletedAt:  dbInstance.CompletedAt,
		History:      history,
	}, nil
}

// ConvertFromWorkflowInstance 转换workflow模型到数据库模型
func ConvertFromWorkflowInstance(wfInstance *workflow.WorkflowInstance) (*database.WorkflowInstance, error) {
	currentNodesJSON := database.JSONField{Data: wfInstance.CurrentNodes}

	variablesJSON := database.JSONField{Data: wfInstance.Variables}

	return &database.WorkflowInstance{
		InstanceID:   wfInstance.ID,
		WorkflowID:   wfInstance.WorkflowID,
		BusinessID:   wfInstance.BusinessID,
		BusinessType: wfInstance.BusinessType,
		Status:       string(wfInstance.Status),
		CurrentNodes: currentNodesJSON,
		Variables:    variablesJSON,
		StartedBy:    wfInstance.StartedBy,
		StartedAt:    wfInstance.StartedAt,
		CompletedAt:  wfInstance.CompletedAt,
	}, nil
}
