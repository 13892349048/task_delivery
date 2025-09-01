package workflow

import (
	"context"
	"time"
)

// AuditService 审计服务
type AuditService struct {
	instanceRepo WorkflowInstanceRepository
}

// NewAuditService 创建审计服务
func NewAuditService(instanceRepo WorkflowInstanceRepository) *AuditService {
	return &AuditService{
		instanceRepo: instanceRepo,
	}
}

// GetWorkflowHistory 获取工作流执行历史
func (s *AuditService) GetWorkflowHistory(ctx context.Context, instanceID string) (*WorkflowAuditReport, error) {
	instance, err := s.instanceRepo.GetInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	report := &WorkflowAuditReport{
		InstanceID:   instance.ID,
		WorkflowID:   instance.WorkflowID,
		BusinessID:   instance.BusinessID,
		BusinessType: instance.BusinessType,
		Status:       instance.Status,
		StartedBy:    instance.StartedBy,
		StartedAt:    instance.StartedAt,
		CompletedAt:  instance.CompletedAt,
		Duration:     s.calculateDuration(instance),
		History:      instance.History,
		Statistics:   s.calculateStatistics(instance.History),
	}

	return report, nil
}

// GetUserApprovalHistory 获取用户审批历史
func (s *AuditService) GetUserApprovalHistory(ctx context.Context, userID uint, filter *ApprovalHistoryFilter) ([]*UserApprovalRecord, error) {
	// 这里应该从数据库查询用户的审批历史
	// 简化实现，返回模拟数据
	records := []*UserApprovalRecord{
		{
			InstanceID:   "instance-1",
			WorkflowName: "任务分配审批",
			NodeName:     "直属上级审批",
			BusinessID:   "task_1",
			Action:       "approve",
			Comment:      "同意分配",
			ExecutedAt:   time.Now().Add(-24 * time.Hour),
			Duration:     30 * time.Minute,
		},
		{
			InstanceID:   "instance-2",
			WorkflowName: "任务分配审批",
			NodeName:     "总监审批",
			BusinessID:   "task_2",
			Action:       "reject",
			Comment:      "人员不合适",
			ExecutedAt:   time.Now().Add(-48 * time.Hour),
			Duration:     15 * time.Minute,
		},
	}

	// 应用过滤条件
	if filter != nil {
		records = s.applyFilter(records, filter)
	}

	return records, nil
}

// GetWorkflowStatistics 获取工作流统计信息
func (s *AuditService) GetWorkflowStatistics(ctx context.Context, filter *StatisticsFilter) (*WorkflowStatistics, error) {
	// 这里应该从数据库聚合统计数据
	// 简化实现，返回模拟数据
	stats := &WorkflowStatistics{
		TotalInstances:    100,
		CompletedInstances: 85,
		CancelledInstances: 10,
		RunningInstances:   5,
		AverageCompletionTime: 2 * time.Hour,
		ApprovalRates: map[string]float64{
			"manager_approval":  0.95,
			"director_approval": 0.80,
		},
		NodeStatistics: []NodeStatistic{
			{
				NodeID:              "manager_approval",
				NodeName:            "直属上级审批",
				TotalExecutions:     95,
				AverageExecutionTime: 45 * time.Minute,
				ApprovalRate:        0.95,
			},
			{
				NodeID:              "director_approval",
				NodeName:            "总监审批",
				TotalExecutions:     20,
				AverageExecutionTime: 2 * time.Hour,
				ApprovalRate:        0.80,
			},
		},
	}

	return stats, nil
}

// GetApprovalTrends 获取审批趋势分析
func (s *AuditService) GetApprovalTrends(ctx context.Context, period string) (*ApprovalTrends, error) {
	// 这里应该从数据库分析趋势数据
	// 简化实现，返回模拟数据
	trends := &ApprovalTrends{
		Period: period,
		DataPoints: []TrendDataPoint{
			{
				Date:            time.Now().AddDate(0, 0, -7),
				TotalApprovals:  15,
				ApprovedCount:   12,
				RejectedCount:   2,
				ReturnedCount:   1,
				AverageTime:     45 * time.Minute,
			},
			{
				Date:            time.Now().AddDate(0, 0, -6),
				TotalApprovals:  18,
				ApprovedCount:   16,
				RejectedCount:   1,
				ReturnedCount:   1,
				AverageTime:     38 * time.Minute,
			},
			{
				Date:            time.Now().AddDate(0, 0, -5),
				TotalApprovals:  22,
				ApprovedCount:   20,
				RejectedCount:   2,
				ReturnedCount:   0,
				AverageTime:     52 * time.Minute,
			},
		},
	}

	return trends, nil
}

// calculateDuration 计算工作流持续时间
func (s *AuditService) calculateDuration(instance *WorkflowInstance) time.Duration {
	if instance.CompletedAt != nil {
		return instance.CompletedAt.Sub(instance.StartedAt)
	}
	return time.Since(instance.StartedAt)
}

// calculateStatistics 计算执行统计
func (s *AuditService) calculateStatistics(history []ExecutionHistory) ExecutionStatistics {
	stats := ExecutionStatistics{
		TotalSteps:    len(history),
		TotalDuration: 0,
		NodeCounts:    make(map[string]int),
		ActionCounts:  make(map[string]int),
	}

	for _, record := range history {
		stats.TotalDuration += record.Duration
		stats.NodeCounts[record.NodeID]++
		stats.ActionCounts[record.Action]++
	}

	if len(history) > 0 {
		stats.AverageDuration = stats.TotalDuration / time.Duration(len(history))
	}

	return stats
}

// applyFilter 应用过滤条件
func (s *AuditService) applyFilter(records []*UserApprovalRecord, filter *ApprovalHistoryFilter) []*UserApprovalRecord {
	var filtered []*UserApprovalRecord

	for _, record := range records {
		if filter.StartDate != nil && record.ExecutedAt.Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && record.ExecutedAt.After(*filter.EndDate) {
			continue
		}
		if filter.Action != "" && record.Action != filter.Action {
			continue
		}
		if filter.WorkflowName != "" && record.WorkflowName != filter.WorkflowName {
			continue
		}
		filtered = append(filtered, record)
	}

	return filtered
}

// WorkflowAuditReport 工作流审计报告
type WorkflowAuditReport struct {
	InstanceID   string               `json:"instance_id"`
	WorkflowID   string               `json:"workflow_id"`
	BusinessID   string               `json:"business_id"`
	BusinessType string               `json:"business_type"`
	Status       InstanceStatus       `json:"status"`
	StartedBy    uint                 `json:"started_by"`
	StartedAt    time.Time            `json:"started_at"`
	CompletedAt  *time.Time           `json:"completed_at,omitempty"`
	Duration     time.Duration        `json:"duration"`
	History      []ExecutionHistory   `json:"history"`
	Statistics   ExecutionStatistics  `json:"statistics"`
}

// ExecutionStatistics 执行统计
type ExecutionStatistics struct {
	TotalSteps      int                    `json:"total_steps"`
	TotalDuration   time.Duration          `json:"total_duration"`
	AverageDuration time.Duration          `json:"average_duration"`
	NodeCounts      map[string]int         `json:"node_counts"`
	ActionCounts    map[string]int         `json:"action_counts"`
}

// UserApprovalRecord 用户审批记录
type UserApprovalRecord struct {
	InstanceID   string        `json:"instance_id"`
	WorkflowName string        `json:"workflow_name"`
	NodeName     string        `json:"node_name"`
	BusinessID   string        `json:"business_id"`
	Action       string        `json:"action"`
	Comment      string        `json:"comment"`
	ExecutedAt   time.Time     `json:"executed_at"`
	Duration     time.Duration `json:"duration"`
}

// ApprovalHistoryFilter 审批历史过滤条件
type ApprovalHistoryFilter struct {
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Action       string     `json:"action,omitempty"`
	WorkflowName string     `json:"workflow_name,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// WorkflowStatistics 工作流统计
type WorkflowStatistics struct {
	TotalInstances        int                    `json:"total_instances"`
	CompletedInstances    int                    `json:"completed_instances"`
	CancelledInstances    int                    `json:"cancelled_instances"`
	RunningInstances      int                    `json:"running_instances"`
	AverageCompletionTime time.Duration          `json:"average_completion_time"`
	ApprovalRates         map[string]float64     `json:"approval_rates"`
	NodeStatistics        []NodeStatistic        `json:"node_statistics"`
}

// NodeStatistic 节点统计
type NodeStatistic struct {
	NodeID               string        `json:"node_id"`
	NodeName             string        `json:"node_name"`
	TotalExecutions      int           `json:"total_executions"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	ApprovalRate         float64       `json:"approval_rate"`
}

// StatisticsFilter 统计过滤条件
type StatisticsFilter struct {
	WorkflowID string     `json:"workflow_id,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
}

// ApprovalTrends 审批趋势
type ApprovalTrends struct {
	Period     string            `json:"period"`
	DataPoints []TrendDataPoint  `json:"data_points"`
}

// TrendDataPoint 趋势数据点
type TrendDataPoint struct {
	Date           time.Time     `json:"date"`
	TotalApprovals int           `json:"total_approvals"`
	ApprovedCount  int           `json:"approved_count"`
	RejectedCount  int           `json:"rejected_count"`
	ReturnedCount  int           `json:"returned_count"`
	AverageTime    time.Duration `json:"average_time"`
}
