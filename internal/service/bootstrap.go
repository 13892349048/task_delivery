package service

import (
	"context"
	"fmt"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// BootstrapService 系统初始化服务
type BootstrapService struct {
	repoManager repository.RepositoryManager
}

// NewBootstrapService 创建初始化服务
func NewBootstrapService(repoManager repository.RepositoryManager) *BootstrapService {
	return &BootstrapService{
		repoManager: repoManager,
	}
}

// InitializeSystem 初始化系统默认数据
func (s *BootstrapService) InitializeSystem(ctx context.Context) error {
	logger.Info("开始初始化系统默认数据...")

	// 1. 创建默认权限
	if err := s.createDefaultPermissions(ctx); err != nil {
		return fmt.Errorf("创建默认权限失败: %w", err)
	}

	// 2. 创建默认角色
	if err := s.createDefaultRoles(ctx); err != nil {
		return fmt.Errorf("创建默认角色失败: %w", err)
	}

	// 3. 创建超级管理员用户
	if err := s.createSuperAdmin(ctx); err != nil {
		return fmt.Errorf("创建超级管理员失败: %w", err)
	}

	// 4. 验证角色权限关系
	logger.Info("验证角色权限关系...")
	if err := s.VerifyRolePermissions(ctx); err != nil {
		logger.Warnf("验证角色权限关系时出现警告: %v", err)
	}

	logger.Info("系统默认数据初始化完成")
	return nil
}

// createDefaultPermissions 创建默认权限
func (s *BootstrapService) createDefaultPermissions(ctx context.Context) error {
	permRepo := s.repoManager.PermissionRepository()

	defaultPermissions := []database.Permission{
		// 系统级权限
		{Name: "system:admin", DisplayName: "系统管理权限", Description: "拥有系统最高管理权限，可以管理所有系统功能和配置", Resource: "system", Action: "admin"},
		{Name: "system:config", DisplayName: "系统配置权限", Description: "可以修改系统配置参数，包括系统设置和环境配置", Resource: "system", Action: "config"},
		
		// 用户管理权限
		{Name: "user:create", DisplayName: "创建用户", Description: "可以在系统中创建新的用户账户", Resource: "user", Action: "create"},
		{Name: "user:read", DisplayName: "查看用户", Description: "可以查看用户信息和用户列表", Resource: "user", Action: "read"},
		{Name: "user:update", DisplayName: "更新用户", Description: "可以修改用户的基本信息、状态等属性", Resource: "user", Action: "update"},
		{Name: "user:delete", DisplayName: "删除用户", Description: "可以删除系统中的用户账户（软删除）", Resource: "user", Action: "delete"},
		{Name: "user:assign_role", DisplayName: "分配用户角色", Description: "可以为用户分配或移除角色权限", Resource: "user", Action: "assign_role"},
		
		// 角色管理权限
		{Name: "role:create", DisplayName: "创建角色", Description: "可以创建新的系统角色", Resource: "role", Action: "create"},
		{Name: "role:read", DisplayName: "查看角色", Description: "可以查看系统中的角色信息和权限配置", Resource: "role", Action: "read"},
		{Name: "role:update", DisplayName: "更新角色", Description: "可以修改角色的权限配置和基本信息", Resource: "role", Action: "update"},
		{Name: "role:delete", DisplayName: "删除角色", Description: "可以删除系统中的自定义角色", Resource: "role", Action: "delete"},
		
		// 任务管理权限
		{Name: "task:create", DisplayName: "创建任务", Description: "可以在系统中创建新的工作任务", Resource: "task", Action: "create"},
		{Name: "task:read", DisplayName: "查看任务", Description: "可以查看任务详情、任务列表和任务状态", Resource: "task", Action: "read"},
		{Name: "task:update", DisplayName: "更新任务", Description: "可以修改任务的基本信息、状态和优先级", Resource: "task", Action: "update"},
		{Name: "task:delete", DisplayName: "删除任务", Description: "可以删除系统中的任务（软删除）", Resource: "task", Action: "delete"},
		{Name: "task:assign", DisplayName: "分配任务", Description: "可以将任务分配给员工或重新分配任务", Resource: "task", Action: "assign"},
		{Name: "task:approve", DisplayName: "审批任务", Description: "可以审批任务分配请求和任务完成状态", Resource: "task", Action: "approve"},
		
		// 员工管理权限
		{Name: "employee:create", DisplayName: "创建员工", Description: "可以在系统中添加新的员工信息", Resource: "employee", Action: "create"},
		{Name: "employee:read", DisplayName: "查看员工", Description: "可以查看员工信息、技能和工作状态", Resource: "employee", Action: "read"},
		{Name: "employee:update", DisplayName: "更新员工", Description: "可以修改员工的基本信息、技能和工作状态", Resource: "employee", Action: "update"},
		{Name: "employee:delete", DisplayName: "删除员工", Description: "可以删除员工信息（软删除）", Resource: "employee", Action: "delete"},
		
		// 技能管理权限
		{Name: "skill:create", DisplayName: "创建技能", Description: "可以在系统中定义新的技能类型", Resource: "skill", Action: "create"},
		{Name: "skill:read", DisplayName: "查看技能", Description: "可以查看系统中的技能定义和员工技能分布", Resource: "skill", Action: "read"},
		{Name: "skill:update", DisplayName: "更新技能", Description: "可以修改技能的定义和描述信息", Resource: "skill", Action: "update"},
		{Name: "skill:delete", DisplayName: "删除技能", Description: "可以删除不再使用的技能类型", Resource: "skill", Action: "delete"},
		
		// 通知权限
		{Name: "notification:send", DisplayName: "发送通知", Description: "可以向用户发送系统通知和消息", Resource: "notification", Action: "send"},
		{Name: "notification:read", DisplayName: "查看通知", Description: "可以查看和管理系统通知记录", Resource: "notification", Action: "read"},
		
		// 部门管理权限
		{Name: "department:create", DisplayName: "创建部门", Description: "可以在系统中创建新的部门组织架构", Resource: "department", Action: "create"},
		{Name: "department:read", DisplayName: "查看部门", Description: "可以查看部门信息、组织架构和部门成员", Resource: "department", Action: "read"},
		{Name: "department:update", DisplayName: "更新部门", Description: "可以修改部门信息、层级关系和管理者", Resource: "department", Action: "update"},
		{Name: "department:delete", DisplayName: "删除部门", Description: "可以删除部门（软删除，需确保无关联数据）", Resource: "department", Action: "delete"},
		
		// 职位管理权限
		{Name: "position:create", DisplayName: "创建职位", Description: "可以在系统中定义新的职位类型和级别", Resource: "position", Action: "create"},
		{Name: "position:read", DisplayName: "查看职位", Description: "可以查看职位信息、级别分类和职位要求", Resource: "position", Action: "read"},
		{Name: "position:update", DisplayName: "更新职位", Description: "可以修改职位的级别、分类和职责描述", Resource: "position", Action: "update"},
		{Name: "position:delete", DisplayName: "删除职位", Description: "可以删除不再使用的职位类型", Resource: "position", Action: "delete"},
		
		// 项目管理权限
		{Name: "project:create", DisplayName: "创建项目", Description: "可以创建新的项目并设置项目基本信息", Resource: "project", Action: "create"},
		{Name: "project:read", DisplayName: "查看项目", Description: "可以查看项目详情、进度和项目成员信息", Resource: "project", Action: "read"},
		{Name: "project:update", DisplayName: "更新项目", Description: "可以修改项目信息、状态和项目配置", Resource: "project", Action: "update"},
		{Name: "project:delete", DisplayName: "删除项目", Description: "可以删除项目（软删除，需确保项目已完成）", Resource: "project", Action: "delete"},
		
		// 权限分配管理权限
		{Name: "permission:create", DisplayName: "创建权限分配", Description: "可以创建权限模板和分配权限给用户", Resource: "permission", Action: "create"},
		{Name: "permission:read", DisplayName: "查看权限分配", Description: "可以查看权限模板、权限分配记录和历史", Resource: "permission", Action: "read"},
		{Name: "permission:update", DisplayName: "更新权限分配", Description: "可以修改权限模板、处理权限审批和撤销权限", Resource: "permission", Action: "update"},
		{Name: "permission:delete", DisplayName: "删除权限分配", Description: "可以删除权限模板和权限配置", Resource: "permission", Action: "delete"},
	}

	for _, perm := range defaultPermissions {
		// 检查权限是否已存在
		existing, err := permRepo.GetByName(ctx, perm.Name)
		if err == nil && existing != nil {
			logger.Infof("权限 %s 已存在，跳过创建", perm.Name)
			continue
		}

		if err := permRepo.Create(ctx, &perm); err != nil {
			return fmt.Errorf("创建权限 %s 失败: %w", perm.Name, err)
		}
		logger.Infof("创建权限: %s", perm.Name)
	}

	return nil
}

// createDefaultRoles 创建默认角色
func (s *BootstrapService) createDefaultRoles(ctx context.Context) error {
	roleRepo := s.repoManager.RoleRepository()
	permRepo := s.repoManager.PermissionRepository()

	// 定义默认角色及其权限
	rolePermissions := map[string]struct {
		displayName string
		description string
		permissions []string
	}{
		"super_admin": {
			displayName: "超级管理员",
			description: "拥有系统最高权限，可以管理所有功能模块，包括用户管理、角色权限、系统配置等",
			permissions: []string{
				"system:admin", "system:config",
				"user:create", "user:read", "user:update", "user:delete", "user:assign_role",
				"role:create", "role:read", "role:update", "role:delete",
				"task:create", "task:read", "task:update", "task:delete", "task:assign", "task:approve",
				"employee:create", "employee:read", "employee:update", "employee:delete",
				"skill:create", "skill:read", "skill:update", "skill:delete",
				"department:create", "department:read", "department:update", "department:delete",
				"position:create", "position:read", "position:update", "position:delete",
				"project:create", "project:read", "project:update", "project:delete",
				"permission:create", "permission:read", "permission:update", "permission:delete",
				"notification:send", "notification:read",
			},
		},
		"admin": {
			displayName: "管理员",
			description: "系统管理员角色，负责日常管理工作，包括用户管理、任务管理、员工管理等，但不能修改系统核心配置",
			permissions: []string{
				"user:create", "user:read", "user:update", "user:assign_role",
				"role:read",
				"task:create", "task:read", "task:update", "task:delete", "task:assign", "task:approve",
				"employee:create", "employee:read", "employee:update", "employee:delete",
				"skill:create", "skill:read", "skill:update", "skill:delete",
				"department:create", "department:read", "department:update", "department:delete",
				"position:create", "position:read", "position:update", "position:delete",
				"project:create", "project:read", "project:update", "project:delete",
				"permission:create", "permission:read", "permission:update", "permission:delete",
				"notification:send", "notification:read",
			},
		},
		"manager": {
			displayName: "经理",
			description: "部门经理角色，负责团队管理和任务分配，可以创建和分配任务，管理下属员工信息",
			permissions: []string{
				"user:read",
				"task:create", "task:read", "task:update", "task:assign", "task:approve",
				"employee:read", "employee:update",
				"skill:read",
				"department:read", "department:update",
				"position:read",
				"project:create", "project:read", "project:update",
				"permission:read", "permission:update",
				"notification:send", "notification:read",
			},
		},
		"employee": {
			displayName: "员工",
			description: "普通员工角色，可以查看和更新自己的任务，查看员工信息和技能要求",
			permissions: []string{
				"task:read", "task:update",
				"employee:read",
				"skill:read",
				"notification:read",
			},
		},
	}

	for roleName, roleInfo := range rolePermissions {
		var role *database.Role
		
		// 检查角色是否已存在
		existing, err := roleRepo.GetByName(ctx, roleName)
		if err == nil && existing != nil {
			logger.Infof("角色 %s 已存在，使用现有角色", roleName)
			role = existing
		} else {
			// 创建新角色
			role = &database.Role{
				Name:        roleName,
				DisplayName: roleInfo.displayName,
				Description: roleInfo.description,
			}

			if err := roleRepo.Create(ctx, role); err != nil {
				return fmt.Errorf("创建角色 %s 失败: %w", roleName, err)
			}
			logger.Infof("创建角色: %s (%s)", roleName, roleInfo.displayName)
		}

		// 获取权限ID并分配给角色（无论角色是新建还是已存在）
		var permissionIDs []uint
		for _, permName := range roleInfo.permissions {
			perm, err := permRepo.GetByName(ctx, permName)
			if err != nil {
				logger.Warnf("权限 %s 不存在，跳过分配", permName)
				continue
			}
			permissionIDs = append(permissionIDs, perm.ID)
		}

		if len(permissionIDs) > 0 {
			// 直接分配权限（GORM的Association会自动处理重复）
			if err := roleRepo.AssignPermissions(ctx, role.ID, permissionIDs); err != nil {
				return fmt.Errorf("为角色 %s 分配权限失败: %w", roleName, err)
			}
			logger.Infof("为角色 %s 分配了 %d 个权限", roleName, len(permissionIDs))
		}
	}

	return nil
}

// createSuperAdmin 创建超级管理员用户
func (s *BootstrapService) createSuperAdmin(ctx context.Context) error {
	userRepo := s.repoManager.UserRepository()
	roleRepo := s.repoManager.RoleRepository()

	// 检查是否已存在超级管理员
	existing, err := userRepo.GetByUsername(ctx, "admin")
	if err == nil && existing != nil {
		logger.Info("超级管理员已存在，跳过创建")
		return nil
	}

	// 生成密码哈希
	password := "admin123" // 默认密码，生产环境应该使用更安全的密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("生成密码哈希失败: %w", err)
	}

	// 创建超级管理员用户
	admin := &database.User{
		Username:     "admin",
		Email:        "admin@taskmanage.com",
		PasswordHash: string(hashedPassword),
		RealName:     "系统管理员",
		Role:         "super_admin",
		Status:       "active",
	}

	if err := userRepo.Create(ctx, admin); err != nil {
		return fmt.Errorf("创建超级管理员用户失败: %w", err)
	}
	logger.Info("创建超级管理员用户成功")

	// 获取超级管理员角色并分配给用户
	superAdminRole, err := roleRepo.GetByName(ctx, "super_admin")
	if err != nil {
		logger.Warnf("获取超级管理员角色失败: %v", err)
		return nil
	}

	if err := userRepo.AssignRoles(ctx, admin.ID, []uint{superAdminRole.ID}); err != nil {
		logger.Warnf("为超级管理员分配角色失败: %v", err)
	} else {
		logger.Info("为超级管理员分配角色成功")
	}

	logger.Infof("超级管理员创建完成 - 用户名: %s, 密码: %s", admin.Username, password)
	return nil
}

// VerifyRolePermissions 验证角色权限关系是否正确初始化
func (s *BootstrapService) VerifyRolePermissions(ctx context.Context) error {
	roleRepo := s.repoManager.RoleRepository()
	
	// 检查所有默认角色的权限
	roleNames := []string{"super_admin", "admin", "manager", "employee"}
	
	for _, roleName := range roleNames {
		role, err := roleRepo.GetByName(ctx, roleName)
		if err != nil {
			logger.Errorf("角色 %s 不存在: %v", roleName, err)
			continue
		}
		
		// 获取角色及其权限
		roleWithPerms, err := roleRepo.GetRoleWithPermissions(ctx, role.ID)
		if err != nil {
			logger.Errorf("获取角色 %s 的权限失败: %v", roleName, err)
			continue
		}
		
		logger.Infof("角色 %s (ID: %d) 拥有 %d 个权限:", roleName, role.ID, len(roleWithPerms.Permissions))
		for _, perm := range roleWithPerms.Permissions {
			logger.Infof("  - %s: %s", perm.Name, perm.Description)
		}
	}
	
	return nil
}
