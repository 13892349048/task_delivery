package service

import (
	"context"
	"fmt"

	"taskmanage/internal/database"
	"taskmanage/pkg/logger"
)

// InitializePermissionTemplates 初始化权限模板
func (s *PermissionAssignmentServiceImpl) InitializePermissionTemplates(ctx context.Context) error {
	logger.Info("开始初始化权限模板...")

	templates := getDefaultPermissionTemplates()
	
	for _, template := range templates {
		// 检查模板是否已存在
		existing, err := s.repos.PermissionTemplateRepository().GetByCode(ctx, template.Code)
		if err == nil && existing != nil {
			logger.Infof("权限模板已存在，跳过: %s", template.Code)
			continue
		}

		// 创建新模板
		if err := s.repos.PermissionTemplateRepository().Create(ctx, template); err != nil {
			logger.Errorf("创建权限模板失败: %s, error: %v", template.Code, err)
			return fmt.Errorf("创建权限模板失败: %s, %w", template.Code, err)
		}

		logger.Infof("成功创建权限模板: %s (Level: %d)", template.Name, template.Level)
	}

	logger.Info("权限模板初始化完成")
	return nil
}

// getDefaultPermissionTemplates 获取默认权限模板
func getDefaultPermissionTemplates() []*database.PermissionTemplate {
	return []*database.PermissionTemplate{
		// 1. 初级员工权限模板
		{
			Name:             "初级员工基础权限",
			Code:             "junior_employee",
			Description:      "新入职员工的基础权限，只能查看和更新分配给自己的任务",
			Category:         "basic",
			Level:            1,
			ProjectScope:     "assigned",
			TaskScope:        "assigned",
			CanAssignToLevel: 0,
			CrossDepartment:  false,
			MaxTasksPerDay:   0,
			IsActive:         true,
		},

		// 2. 中级员工权限模板
		{
			Name:             "中级员工权限",
			Code:             "mid_employee",
			Description:      "中级员工权限，可以查看团队任务并创建新任务",
			Category:         "intermediate",
			Level:            3,
			ProjectScope:     "team",
			TaskScope:        "team",
			CanAssignToLevel: 1,
			CrossDepartment:  false,
			MaxTasksPerDay:   5,
			IsActive:         true,
		},

		// 3. 高级员工权限模板
		{
			Name:             "高级员工权限",
			Code:             "senior_employee",
			Description:      "高级员工权限，可以分配任务给同级和下级员工",
			Category:         "advanced",
			Level:            5,
			ProjectScope:     "department",
			TaskScope:        "team",
			CanAssignToLevel: 3,
			CrossDepartment:  false,
			MaxTasksPerDay:   10,
			IsActive:         true,
		},

		// 4. 团队负责人权限模板
		{
			Name:             "团队负责人权限",
			Code:             "team_lead",
			Description:      "团队负责人权限，可以管理团队内所有任务和人员分配",
			Category:         "manager",
			Level:            6,
			ProjectScope:     "department",
			TaskScope:        "all",
			CanAssignToLevel: 5,
			CrossDepartment:  false,
			MaxTasksPerDay:   20,
			IsActive:         true,
		},

		// 5. 项目经理权限模板
		{
			Name:             "项目经理权限",
			Code:             "project_manager",
			Description:      "项目经理权限，可以跨部门分配任务和管理项目",
			Category:         "manager",
			Level:            7,
			ProjectScope:     "all",
			TaskScope:        "all",
			CanAssignToLevel: 6,
			CrossDepartment:  true,
			MaxTasksPerDay:   50,
			IsActive:         true,
		},

		// 6. 部门主管权限模板
		{
			Name:             "部门主管权限",
			Code:             "department_head",
			Description:      "部门主管权限，拥有部门内最高权限",
			Category:         "manager",
			Level:            8,
			ProjectScope:     "all",
			TaskScope:        "all",
			CanAssignToLevel: 7,
			CrossDepartment:  true,
			MaxTasksPerDay:   0, // 无限制
			IsActive:         true,
		},

		// 7. 试用期员工权限模板
		{
			Name:             "试用期员工权限",
			Code:             "probation_employee",
			Description:      "试用期员工权限，受限的基础权限",
			Category:         "basic",
			Level:            0,
			ProjectScope:     "assigned",
			TaskScope:        "assigned",
			CanAssignToLevel: 0,
			CrossDepartment:  false,
			MaxTasksPerDay:   3,
			IsActive:         true,
		},

		// 8. 实习生权限模板
		{
			Name:             "实习生权限",
			Code:             "intern",
			Description:      "实习生权限，最基础的只读权限",
			Category:         "basic",
			Level:            0,
			ProjectScope:     "assigned",
			TaskScope:        "assigned",
			CanAssignToLevel: 0,
			CrossDepartment:  false,
			MaxTasksPerDay:   2,
			IsActive:         true,
		},
	}
}

// InitializeDepartmentSpecificTemplates 初始化部门特定权限模板
func (s *PermissionAssignmentServiceImpl) InitializeDepartmentSpecificTemplates(ctx context.Context) error {
	logger.Info("开始初始化部门特定权限模板...")

	// 这里可以根据实际的部门ID创建部门特定模板
	// 需要先查询部门信息，然后创建对应的模板
	
	// 示例：开发部门特定模板
	devTemplates := []*database.PermissionTemplate{
		{
			Name:             "开发部高级工程师",
			Code:             "dev_senior_engineer",
			Description:      "开发部高级工程师权限，包含代码审查和部署权限",
			Category:         "advanced",
			Level:            4,
			DepartmentID:     nil, // 需要根据实际部门ID设置
			ProjectScope:     "department",
			TaskScope:        "team",
			CanAssignToLevel: 3,
			CrossDepartment:  false,
			MaxTasksPerDay:   8,
			IsActive:         true,
		},
		{
			Name:             "测试部负责人",
			Code:             "test_team_lead",
			Description:      "测试部门负责人权限，管理测试任务和质量控制",
			Category:         "manager",
			Level:            6,
			DepartmentID:     nil, // 需要根据实际部门ID设置
			ProjectScope:     "department",
			TaskScope:        "all",
			CanAssignToLevel: 5,
			CrossDepartment:  false,
			MaxTasksPerDay:   15,
			IsActive:         true,
		},
	}

	for _, template := range devTemplates {
		// 检查模板是否已存在
		existing, err := s.repos.PermissionTemplateRepository().GetByCode(ctx, template.Code)
		if err == nil && existing != nil {
			logger.Infof("部门权限模板已存在，跳过: %s", template.Code)
			continue
		}

		// 创建新模板
		if err := s.repos.PermissionTemplateRepository().Create(ctx, template); err != nil {
			logger.Errorf("创建部门权限模板失败: %s, error: %v", template.Code, err)
			continue // 继续创建其他模板
		}

		logger.Infof("成功创建部门权限模板: %s", template.Name)
	}

	logger.Info("部门特定权限模板初始化完成")
	return nil
}

// InitializeOnboardingPermissionConfigs 初始化入职权限配置
func (s *PermissionAssignmentServiceImpl) InitializeOnboardingPermissionConfigs(ctx context.Context) error {
	logger.Info("开始初始化入职权限配置...")

	configs := getDefaultOnboardingPermissionConfigs()
	
	for _, config := range configs {
		// 检查配置是否已存在
		existing, err := s.repos.OnboardingPermissionConfigRepository().GetByStatus(ctx, config.OnboardingStatus)
		if err == nil && len(existing) > 0 {
			logger.Infof("入职权限配置已存在，跳过: %s", config.OnboardingStatus)
			continue
		}

		// 创建新配置
		if err := s.repos.OnboardingPermissionConfigRepository().Create(ctx, config); err != nil {
			logger.Errorf("创建入职权限配置失败: %s, error: %v", config.OnboardingStatus, err)
			continue
		}

		logger.Infof("成功创建入职权限配置: %s", config.OnboardingStatus)
	}

	logger.Info("入职权限配置初始化完成")
	return nil
}

// getDefaultOnboardingPermissionConfigs 获取默认入职权限配置
func getDefaultOnboardingPermissionConfigs() []*database.OnboardingPermissionConfig {
	return []*database.OnboardingPermissionConfig{
		// 试用期配置
		{
			OnboardingStatus:    "probation",
			AutoAssign:          true,
			RequireApproval:     false,
			EffectiveAfterDays:  0,
			ExpirationDays:      90, // 90天后自动升级
			UpgradeAfterDays:    90,
			// DefaultTemplateID 和 NextLevelTemplateID 需要在模板创建后设置
		},
		// 正式员工配置
		{
			OnboardingStatus:    "active",
			AutoAssign:          true,
			RequireApproval:     false,
			EffectiveAfterDays:  0,
			ExpirationDays:      0, // 永不过期
			UpgradeAfterDays:    180, // 180天后可升级
		},
		// 入职中配置
		{
			OnboardingStatus:    "onboarding",
			AutoAssign:          true,
			RequireApproval:     false,
			EffectiveAfterDays:  0,
			ExpirationDays:      30, // 30天内完成入职
		},
	}
}
