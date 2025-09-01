package workflow

import (
	"context"
	"fmt"
	"strings"

	"taskmanage/pkg/logger"
)

// WorkflowSelector 工作流选择器接口
type WorkflowSelector interface {
	// SelectWorkflow 根据业务类型和条件选择合适的工作流
	SelectWorkflow(ctx context.Context, req *WorkflowSelectionRequest) (string, error)
}

// WorkflowSelectionRequest 工作流选择请求
type WorkflowSelectionRequest struct {
	BusinessType string                 `json:"business_type"` // 业务类型
	Context      map[string]interface{} `json:"context"`       // 业务上下文
	Priority     string                 `json:"priority"`      // 优先级
	UserID       uint                   `json:"user_id"`       // 发起用户ID
	DepartmentID *uint                  `json:"department_id"` // 部门ID
	RoleNames    []string               `json:"role_names"`    // 用户角色
}

// WorkflowSelectorImpl 工作流选择器实现
type WorkflowSelectorImpl struct {
	definitionManager *WorkflowDefinitionManager
	rules             []WorkflowSelectionRule
}

// WorkflowSelectionRule 工作流选择规则
type WorkflowSelectionRule struct {
	BusinessType string                                                                    `json:"business_type"`
	Conditions   []WorkflowCondition                                                       `json:"conditions"`
	WorkflowID   string                                                                    `json:"workflow_id"`
	Priority     int                                                                       `json:"priority"` // 规则优先级，数字越小优先级越高
	Evaluator    func(ctx context.Context, req *WorkflowSelectionRequest) (bool, error)   `json:"-"`
}

// WorkflowCondition 工作流选择条件
type WorkflowCondition struct {
	Field    string      `json:"field"`    // 字段名
	Operator string      `json:"operator"` // 操作符: eq, ne, in, not_in, gt, lt, contains
	Value    interface{} `json:"value"`    // 期望值
}

// NewWorkflowSelector 创建工作流选择器
func NewWorkflowSelector(definitionManager *WorkflowDefinitionManager) *WorkflowSelectorImpl {
	selector := &WorkflowSelectorImpl{
		definitionManager: definitionManager,
		rules:             make([]WorkflowSelectionRule, 0),
	}
	
	// 初始化默认规则
	selector.initDefaultRules()
	
	return selector
}

// SelectWorkflow 选择工作流
func (s *WorkflowSelectorImpl) SelectWorkflow(ctx context.Context, req *WorkflowSelectionRequest) (string, error) {
	logger.Infof("选择工作流: BusinessType=%s, Priority=%s, UserID=%d", 
		req.BusinessType, req.Priority, req.UserID)

	// 按优先级排序规则
	matchedRules := s.getMatchingRules(req.BusinessType)
	
	// 评估每个规则
	for _, rule := range matchedRules {
		matched, err := s.evaluateRule(ctx, &rule, req)
		if err != nil {
			logger.Errorf("评估规则失败: %v", err)
			continue
		}
		
		if matched {
			logger.Infof("匹配到工作流规则: WorkflowID=%s, Priority=%d", rule.WorkflowID, rule.Priority)
			
			// 验证工作流定义是否存在且激活
			if s.isWorkflowAvailable(ctx, rule.WorkflowID) {
				return rule.WorkflowID, nil
			}
			
			logger.Warnf("工作流不可用: %s", rule.WorkflowID)
		}
	}
	
	// 如果没有匹配的规则，返回默认工作流
	defaultWorkflowID := s.getDefaultWorkflow(req.BusinessType)
	if defaultWorkflowID != "" {
		logger.Infof("使用默认工作流: %s", defaultWorkflowID)
		return defaultWorkflowID, nil
	}
	
	return "", fmt.Errorf("未找到适合的工作流: BusinessType=%s", req.BusinessType)
}

// initDefaultRules 初始化默认规则
func (s *WorkflowSelectorImpl) initDefaultRules() {
	// 入职审批规则
	s.rules = append(s.rules, []WorkflowSelectionRule{
		// 简化入职流程 - 针对实习生或临时员工
		{
			BusinessType: "onboarding",
			Conditions: []WorkflowCondition{
				{Field: "employee_type", Operator: "in", Value: []string{"intern", "temporary"}},
			},
			WorkflowID: "onboarding-simple-approval-v1",
			Priority:   1,
		},
		// 标准入职流程 - 针对正式员工
		{
			BusinessType: "onboarding",
			Conditions: []WorkflowCondition{
				{Field: "employee_type", Operator: "eq", Value: "permanent"},
			},
			WorkflowID: "onboarding-approval-v1",
			Priority:   2,
		},
		// 高优先级任务分配 - 需要额外审批
		{
			BusinessType: "task_assignment",
			Conditions: []WorkflowCondition{
				{Field: "priority", Operator: "eq", Value: "high"},
			},
			WorkflowID: "task-assignment-approval-v1",
			Priority:   1,
		},
		// 普通任务分配 - 简化流程
		{
			BusinessType: "task_assignment",
			Conditions: []WorkflowCondition{
				{Field: "priority", Operator: "in", Value: []string{"medium", "low"}},
			},
			WorkflowID: "task-assignment-simple-approval-v1",
			Priority:   2,
		},
	}...)
}

// getMatchingRules 获取匹配的规则
func (s *WorkflowSelectorImpl) getMatchingRules(businessType string) []WorkflowSelectionRule {
	var matchedRules []WorkflowSelectionRule
	
	for _, rule := range s.rules {
		if rule.BusinessType == businessType {
			matchedRules = append(matchedRules, rule)
		}
	}
	
	// 按优先级排序（优先级数字越小越优先）
	for i := 0; i < len(matchedRules)-1; i++ {
		for j := i + 1; j < len(matchedRules); j++ {
			if matchedRules[i].Priority > matchedRules[j].Priority {
				matchedRules[i], matchedRules[j] = matchedRules[j], matchedRules[i]
			}
		}
	}
	
	return matchedRules
}

// evaluateRule 评估规则
func (s *WorkflowSelectorImpl) evaluateRule(ctx context.Context, rule *WorkflowSelectionRule, req *WorkflowSelectionRequest) (bool, error) {
	// 如果有自定义评估器，优先使用
	if rule.Evaluator != nil {
		return rule.Evaluator(ctx, req)
	}
	
	// 评估所有条件（AND逻辑）
	for _, condition := range rule.Conditions {
		matched, err := s.evaluateCondition(&condition, req)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}
	
	return true, nil
}

// evaluateCondition 评估单个条件
func (s *WorkflowSelectorImpl) evaluateCondition(condition *WorkflowCondition, req *WorkflowSelectionRequest) (bool, error) {
	// 从请求中获取字段值
	var fieldValue interface{}
	
	switch condition.Field {
	case "priority":
		fieldValue = req.Priority
	case "user_id":
		fieldValue = req.UserID
	case "department_id":
		if req.DepartmentID != nil {
			fieldValue = *req.DepartmentID
		}
	case "role_names":
		fieldValue = req.RoleNames
	default:
		// 从上下文中获取
		if req.Context != nil {
			fieldValue = req.Context[condition.Field]
		}
	}
	
	// 根据操作符进行比较
	switch condition.Operator {
	case "eq":
		return fieldValue == condition.Value, nil
	case "ne":
		return fieldValue != condition.Value, nil
	case "in":
		return s.isValueInSlice(fieldValue, condition.Value), nil
	case "not_in":
		return !s.isValueInSlice(fieldValue, condition.Value), nil
	case "contains":
		if str, ok := fieldValue.(string); ok {
			if substr, ok := condition.Value.(string); ok {
				return strings.Contains(str, substr), nil
			}
		}
		return false, nil
	case "gt":
		return s.compareNumbers(fieldValue, condition.Value, ">"), nil
	case "lt":
		return s.compareNumbers(fieldValue, condition.Value, "<"), nil
	default:
		return false, fmt.Errorf("不支持的操作符: %s", condition.Operator)
	}
}

// isValueInSlice 检查值是否在切片中
func (s *WorkflowSelectorImpl) isValueInSlice(value interface{}, slice interface{}) bool {
	switch s := slice.(type) {
	case []string:
		if str, ok := value.(string); ok {
			for _, item := range s {
				if item == str {
					return true
				}
			}
		}
	case []interface{}:
		for _, item := range s {
			if item == value {
				return true
			}
		}
	}
	return false
}

// compareNumbers 比较数字
func (s *WorkflowSelectorImpl) compareNumbers(a, b interface{}, operator string) bool {
	// 简化实现，仅支持基本数字类型
	var aFloat, bFloat float64
	var ok bool
	
	if aFloat, ok = a.(float64); !ok {
		if aInt, ok := a.(int); ok {
			aFloat = float64(aInt)
		} else if aUint, ok := a.(uint); ok {
			aFloat = float64(aUint)
		} else {
			return false
		}
	}
	
	if bFloat, ok = b.(float64); !ok {
		if bInt, ok := b.(int); ok {
			bFloat = float64(bInt)
		} else if bUint, ok := b.(uint); ok {
			bFloat = float64(bUint)
		} else {
			return false
		}
	}
	
	switch operator {
	case ">":
		return aFloat > bFloat
	case "<":
		return aFloat < bFloat
	default:
		return false
	}
}

// isWorkflowAvailable 检查工作流是否可用
func (s *WorkflowSelectorImpl) isWorkflowAvailable(ctx context.Context, workflowID string) bool {
	// 这里应该调用定义管理器检查工作流是否存在且激活
	// 简化实现，假设所有工作流都可用
	return true
}

// getDefaultWorkflow 获取默认工作流
func (s *WorkflowSelectorImpl) getDefaultWorkflow(businessType string) string {
	defaults := map[string]string{
		"onboarding":      "onboarding-simple-approval-v1",
		"task_assignment": "task-assignment-approval-v1",
	}
	
	return defaults[businessType]
}

// AddRule 添加自定义规则
func (s *WorkflowSelectorImpl) AddRule(rule WorkflowSelectionRule) {
	s.rules = append(s.rules, rule)
}

// RemoveRule 移除规则
func (s *WorkflowSelectorImpl) RemoveRule(businessType, workflowID string) {
	for i, rule := range s.rules {
		if rule.BusinessType == businessType && rule.WorkflowID == workflowID {
			s.rules = append(s.rules[:i], s.rules[i+1:]...)
			break
		}
	}
}
