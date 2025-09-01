package service

import (
	"context"
	"fmt"
	"time"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// PermissionAssignmentService 权限分配服务接口
type PermissionAssignmentService interface {
	// 权限模板管理
	CreatePermissionTemplate(ctx context.Context, req *CreatePermissionTemplateRequest) (*PermissionTemplateResponse, error)
	GetPermissionTemplate(ctx context.Context, id uint) (*PermissionTemplateResponse, error)
	ListPermissionTemplates(ctx context.Context, req *ListPermissionTemplatesRequest) (*ListPermissionTemplatesResponse, error)
	UpdatePermissionTemplate(ctx context.Context, id uint, req *UpdatePermissionTemplateRequest) (*PermissionTemplateResponse, error)
	DeletePermissionTemplate(ctx context.Context, id uint) error
	
	// 权限模板初始化
	InitializePermissionTemplates(ctx context.Context) error
	InitializeDepartmentSpecificTemplates(ctx context.Context) error
	InitializeOnboardingPermissionConfigs(ctx context.Context) error
	
	// 权限规则管理
	CreatePermissionRule(ctx context.Context, req *CreatePermissionRuleRequest) (*PermissionRuleResponse, error)
	GetPermissionRule(ctx context.Context, id uint) (*PermissionRuleResponse, error)
	ListPermissionRules(ctx context.Context, req *ListPermissionRulesRequest) (*ListPermissionRulesResponse, error)
	UpdatePermissionRule(ctx context.Context, id uint, req *UpdatePermissionRuleRequest) (*PermissionRuleResponse, error)
	DeletePermissionRule(ctx context.Context, id uint) error
	
	// 权限分配管理
	AssignPermissions(ctx context.Context, req *AssignPermissionsRequest) (*PermissionAssignmentResponse, error)
	GetPermissionAssignment(ctx context.Context, id uint) (*PermissionAssignmentResponse, error)
	ListPermissionAssignments(ctx context.Context, req *ListPermissionAssignmentsRequest) (*ListPermissionAssignmentsResponse, error)
	UpdatePermissionAssignment(ctx context.Context, id uint, req *UpdatePermissionAssignmentRequest) (*PermissionAssignmentResponse, error)
	RevokePermissionAssignment(ctx context.Context, id uint, reason string, operatorID uint) error
	
	// 权限分配历史
	GetPermissionAssignmentHistory(ctx context.Context, req *GetPermissionAssignmentHistoryRequest) (*ListPermissionAssignmentHistoryResponse, error)
	
	// 入职权限配置
	CreateOnboardingPermissionConfig(ctx context.Context, req *CreateOnboardingPermissionConfigRequest) (*OnboardingPermissionConfigResponse, error)
	GetOnboardingPermissionConfig(ctx context.Context, id uint) (*OnboardingPermissionConfigResponse, error)
	ListOnboardingPermissionConfigs(ctx context.Context, req *ListOnboardingPermissionConfigsRequest) (*ListOnboardingPermissionConfigsResponse, error)
	UpdateOnboardingPermissionConfig(ctx context.Context, id uint, req *UpdateOnboardingPermissionConfigRequest) (*OnboardingPermissionConfigResponse, error)
	DeleteOnboardingPermissionConfig(ctx context.Context, id uint) error
	
	// 自动权限分配
	ProcessOnboardingPermissionAssignment(ctx context.Context, userID uint, onboardingStatus string, departmentID, positionID *uint) error
	EvaluatePermissionRules(ctx context.Context, userID uint, triggerCondition, value string) ([]*database.PermissionRule, error)
	ApplyPermissionTemplate(ctx context.Context, userID uint, templateID uint, operatorID uint, reason string) (*PermissionAssignmentResponse, error)
	
	// 权限审批
	ProcessPermissionApproval(ctx context.Context, assignmentID uint, approved bool, approverID uint, comments string) error
	GetPendingPermissionApprovals(ctx context.Context, approverID uint) ([]*PermissionAssignmentResponse, error)
}

// PermissionAssignmentServiceImpl 权限分配服务实现
type PermissionAssignmentServiceImpl struct {
	repos repository.RepositoryManager
}

// NewPermissionAssignmentService 创建权限分配服务
func NewPermissionAssignmentService(repos repository.RepositoryManager) PermissionAssignmentService {
	return &PermissionAssignmentServiceImpl{
		repos: repos,
	}
}

// CreatePermissionTemplate 创建权限模板
func (s *PermissionAssignmentServiceImpl) CreatePermissionTemplate(ctx context.Context, req *CreatePermissionTemplateRequest) (*PermissionTemplateResponse, error) {
	// 设置默认值
	projectScope := req.ProjectScope
	if projectScope == "" {
		projectScope = "assigned"
	}
	taskScope := req.TaskScope
	if taskScope == "" {
		taskScope = "assigned"
	}

	template := &database.PermissionTemplate{
		Name:             req.Name,
		Code:             req.Code,
		Description:      req.Description,
		Category:         req.Category,
		Level:            req.Level,
		DepartmentID:     req.DepartmentID,
		PositionID:       req.PositionID,
		ProjectScope:     projectScope,
		TaskScope:        taskScope,
		CanAssignToLevel: req.CanAssignToLevel,
		CrossDepartment:  req.CrossDepartment,
		MaxTasksPerDay:   req.MaxTasksPerDay,
		Permissions:      req.Permissions,
		IsActive:         true,
	}

	if err := s.repos.PermissionTemplateRepository().Create(ctx, template); err != nil {
		logger.Errorf("创建权限模板失败: %v", err)
		return nil, fmt.Errorf("创建权限模板失败: %w", err)
	}

	logger.Infof("成功创建权限模板: %s (ID: %d)", template.Name, template.ID)
	return s.buildPermissionTemplateResponse(template), nil
}

// GetPermissionTemplate 获取权限模板
func (s *PermissionAssignmentServiceImpl) GetPermissionTemplate(ctx context.Context, id uint) (*PermissionTemplateResponse, error) {
	template, err := s.repos.PermissionTemplateRepository().GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取权限模板失败: %w", err)
	}

	return s.buildPermissionTemplateResponse(template), nil
}

// ListPermissionTemplates 获取权限模板列表
func (s *PermissionAssignmentServiceImpl) ListPermissionTemplates(ctx context.Context, req *ListPermissionTemplatesRequest) (*ListPermissionTemplatesResponse, error) {
	filter := &repository.PermissionTemplateFilter{
		Page:         req.Page,
		PageSize:     req.PageSize,
		Category:     req.Category,
		Level:        req.Level,
		DepartmentID: req.DepartmentID,
		PositionID:   req.PositionID,
		IsActive:     req.IsActive,
		Search:       req.Search,
	}

	templates, err := s.repos.PermissionTemplateRepository().List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("获取权限模板列表失败: %w", err)
	}

	var items []*PermissionTemplateResponse
	for _, template := range templates {
		items = append(items, s.buildPermissionTemplateResponse(template))
	}

	return &ListPermissionTemplatesResponse{
		Items: items,
		Total: len(items),
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

// ProcessOnboardingPermissionAssignment 处理入职权限分配
func (s *PermissionAssignmentServiceImpl) ProcessOnboardingPermissionAssignment(ctx context.Context, userID uint, onboardingStatus string, departmentID, positionID *uint) error {
	// 获取入职权限配置
	config, err := s.repos.OnboardingPermissionConfigRepository().GetByStatusAndDepartment(ctx, onboardingStatus, departmentID, positionID)
	if err != nil {
		logger.Warnf("未找到入职权限配置: status=%s, dept=%v, pos=%v", onboardingStatus, departmentID, positionID)
		return nil // 没有配置不算错误
	}

	// 如果不是自动分配，跳过
	if !config.AutoAssign {
		logger.Infof("入职权限配置未启用自动分配: %d", config.ID)
		return nil
	}

	// 应用默认权限模板
	if config.DefaultTemplateID != nil {
		_, err := s.ApplyPermissionTemplate(ctx, userID, *config.DefaultTemplateID, 0, fmt.Sprintf("入职自动分配权限: %s", onboardingStatus))
		if err != nil {
			logger.Errorf("应用默认权限模板失败: %v", err)
			return fmt.Errorf("应用默认权限模板失败: %w", err)
		}
	}

	// 如果有延迟分配的下一级权限模板，创建定时任务（这里简化处理）
	if config.NextLevelTemplateID != nil && config.UpgradeAfterDays > 0 {
		logger.Infof("需要在 %d 天后分配下一级权限模板: %d", config.UpgradeAfterDays, *config.NextLevelTemplateID)
		// TODO: 实现定时任务或延迟分配机制
	}

	return nil
}

// EvaluatePermissionRules 评估权限规则
func (s *PermissionAssignmentServiceImpl) EvaluatePermissionRules(ctx context.Context, userID uint, triggerCondition, value string) ([]*database.PermissionRule, error) {
	rules, err := s.repos.PermissionRuleRepository().GetByTriggerCondition(ctx, triggerCondition, value)
	if err != nil {
		return nil, fmt.Errorf("获取权限规则失败: %w", err)
	}

	var applicableRules []*database.PermissionRule
	for _, rule := range rules {
		// 这里可以添加更复杂的规则评估逻辑
		// 比如检查用户是否已经有该权限、是否满足其他条件等
		applicableRules = append(applicableRules, rule)
	}

	return applicableRules, nil
}

// ApplyPermissionTemplate 应用权限模板
func (s *PermissionAssignmentServiceImpl) ApplyPermissionTemplate(ctx context.Context, userID uint, templateID uint, operatorID uint, reason string) (*PermissionAssignmentResponse, error) {
	_, err := s.repos.PermissionTemplateRepository().GetByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("获取权限模板失败: %w", err)
	}

	assignment := &database.PermissionAssignment{
		UserID:         userID,
		TemplateID:     &templateID,
		Status:         database.PermissionStatusActive,
		ApprovalStatus: database.ApprovalStatusApproved, // 自动分配默认已审批
		AssignedBy:     operatorID,
		AssignedAt:     time.Now(),
	}

	// 如果模板需要审批，设置为待审批状态
	// TODO: 添加RequireApproval字段到PermissionTemplate模型
	// if template.RequireApproval {
	// 	assignment.ApprovalStatus = database.ApprovalStatusPending
	// 	assignment.Status = database.PermissionStatusPending
	// }

	if err := s.repos.PermissionAssignmentRepository().Create(ctx, assignment); err != nil {
		logger.Errorf("创建权限分配失败: %v", err)
		return nil, fmt.Errorf("创建权限分配失败: %w", err)
	}

	// 记录分配历史
	history := &database.PermissionAssignmentHistory{
		AssignmentID: assignment.ID,
		Action:       "assign",
		Reason:       reason,
		OperatorID:   operatorID,
		OperatedAt:   time.Now(),
	}

	if err := s.repos.PermissionAssignmentHistoryRepository().Create(ctx, history); err != nil {
		logger.Warnf("记录权限分配历史失败: %v", err)
	}

	logger.Infof("成功应用权限模板: user=%d, template=%d, assignment=%d", userID, templateID, assignment.ID)
	return s.buildPermissionAssignmentResponse(assignment), nil
}

// 辅助方法：构建权限模板响应
func (s *PermissionAssignmentServiceImpl) buildPermissionTemplateResponse(template *database.PermissionTemplate) *PermissionTemplateResponse {
	return &PermissionTemplateResponse{
		ID:               template.ID,
		Name:             template.Name,
		Code:             template.Code,
		Description:      template.Description,
		Category:         template.Category,
		Level:            template.Level,
		DepartmentID:     template.DepartmentID,
		PositionID:       template.PositionID,
		ProjectScope:     template.ProjectScope,
		TaskScope:        template.TaskScope,
		CanAssignToLevel: template.CanAssignToLevel,
		CrossDepartment:  template.CrossDepartment,
		MaxTasksPerDay:   template.MaxTasksPerDay,
		Permissions:      template.Permissions,
		IsActive:         template.IsActive,
		CreatedAt:        template.CreatedAt,
		UpdatedAt:        template.UpdatedAt,
	}
}

// 辅助方法：构建权限分配响应
func (s *PermissionAssignmentServiceImpl) buildPermissionAssignmentResponse(assignment *database.PermissionAssignment) *PermissionAssignmentResponse {
	resp := &PermissionAssignmentResponse{
		ID:             assignment.ID,
		UserID:         assignment.UserID,
		Status:         assignment.Status,
		ApprovalStatus: assignment.ApprovalStatus,
		AssignedAt:     assignment.AssignedAt,
		CreatedAt:      assignment.CreatedAt,
		UpdatedAt:      assignment.UpdatedAt,
	}

	if assignment.TemplateID != nil {
		resp.TemplateID = *assignment.TemplateID
	}
	if assignment.PermissionID != nil {
		resp.PermissionID = *assignment.PermissionID
	}
	if assignment.ExpiresAt != nil {
		resp.ExpiresAt = assignment.ExpiresAt
	}
	if assignment.ApprovedAt != nil {
		resp.ApprovedAt = assignment.ApprovedAt
	}

	return resp
}

// 占位符实现，需要根据具体需求完善
func (s *PermissionAssignmentServiceImpl) UpdatePermissionTemplate(ctx context.Context, id uint, req *UpdatePermissionTemplateRequest) (*PermissionTemplateResponse, error) {
	// TODO: 实现更新权限模板
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) DeletePermissionTemplate(ctx context.Context, id uint) error {
	// TODO: 实现删除权限模板
	return fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) CreatePermissionRule(ctx context.Context, req *CreatePermissionRuleRequest) (*PermissionRuleResponse, error) {
	// TODO: 实现创建权限规则
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) GetPermissionRule(ctx context.Context, id uint) (*PermissionRuleResponse, error) {
	// TODO: 实现获取权限规则
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) ListPermissionRules(ctx context.Context, req *ListPermissionRulesRequest) (*ListPermissionRulesResponse, error) {
	// TODO: 实现获取权限规则列表
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) UpdatePermissionRule(ctx context.Context, id uint, req *UpdatePermissionRuleRequest) (*PermissionRuleResponse, error) {
	// TODO: 实现更新权限规则
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) DeletePermissionRule(ctx context.Context, id uint) error {
	// TODO: 实现删除权限规则
	return fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) AssignPermissions(ctx context.Context, req *AssignPermissionsRequest) (*PermissionAssignmentResponse, error) {
	// TODO: 实现权限分配
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) GetPermissionAssignment(ctx context.Context, id uint) (*PermissionAssignmentResponse, error) {
	// TODO: 实现获取权限分配
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) ListPermissionAssignments(ctx context.Context, req *ListPermissionAssignmentsRequest) (*ListPermissionAssignmentsResponse, error) {
	// TODO: 实现获取权限分配列表
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) UpdatePermissionAssignment(ctx context.Context, id uint, req *UpdatePermissionAssignmentRequest) (*PermissionAssignmentResponse, error) {
	// TODO: 实现更新权限分配
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) RevokePermissionAssignment(ctx context.Context, id uint, reason string, operatorID uint) error {
	// TODO: 实现撤销权限分配
	return fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) GetPermissionAssignmentHistory(ctx context.Context, req *GetPermissionAssignmentHistoryRequest) (*ListPermissionAssignmentHistoryResponse, error) {
	// TODO: 实现获取权限分配历史
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) CreateOnboardingPermissionConfig(ctx context.Context, req *CreateOnboardingPermissionConfigRequest) (*OnboardingPermissionConfigResponse, error) {
	// TODO: 实现创建入职权限配置
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) GetOnboardingPermissionConfig(ctx context.Context, id uint) (*OnboardingPermissionConfigResponse, error) {
	// TODO: 实现获取入职权限配置
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) ListOnboardingPermissionConfigs(ctx context.Context, req *ListOnboardingPermissionConfigsRequest) (*ListOnboardingPermissionConfigsResponse, error) {
	// TODO: 实现获取入职权限配置列表
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) UpdateOnboardingPermissionConfig(ctx context.Context, id uint, req *UpdateOnboardingPermissionConfigRequest) (*OnboardingPermissionConfigResponse, error) {
	// TODO: 实现更新入职权限配置
	return nil, fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) DeleteOnboardingPermissionConfig(ctx context.Context, id uint) error {
	// TODO: 实现删除入职权限配置
	return fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) ProcessPermissionApproval(ctx context.Context, assignmentID uint, approved bool, approverID uint, comments string) error {
	// TODO: 实现权限审批
	return fmt.Errorf("功能待实现")
}

func (s *PermissionAssignmentServiceImpl) GetPendingPermissionApprovals(ctx context.Context, approverID uint) ([]*PermissionAssignmentResponse, error) {
	// TODO: 实现获取待审批权限
	return nil, fmt.Errorf("功能待实现")
}
