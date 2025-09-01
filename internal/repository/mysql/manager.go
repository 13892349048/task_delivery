package mysql

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// RepositoryManagerImpl 仓储管理器实现
type RepositoryManagerImpl struct {
	db                    *gorm.DB
	userRepo              repository.UserRepository
	roleRepo              repository.RoleRepository
	permissionRepo        repository.PermissionRepository
	employeeRepo          repository.EmployeeRepository
	skillRepo             repository.SkillRepository
	taskRepo              repository.TaskRepository
	assignmentRepo        repository.AssignmentRepository
	notificationRepo      repository.NotificationRepository
	auditLogRepo          repository.AuditLogRepository
	systemConfigRepo      repository.SystemConfigRepository
	workflowRepo          repository.WorkflowRepository
	workflowInstanceRepo  repository.WorkflowInstanceRepository
	departmentRepo        repository.DepartmentRepository
	positionRepo          repository.PositionRepository
	projectRepo           repository.ProjectRepository
	onboardingHistoryRepo repository.OnboardingHistoryRepository
	
	// 权限分配相关仓储
	permissionTemplateRepo        repository.PermissionTemplateRepository
	permissionRuleRepo            repository.PermissionRuleRepository
	permissionAssignmentRepo      repository.PermissionAssignmentRepository
	permissionAssignmentHistoryRepo repository.PermissionAssignmentHistoryRepository
	onboardingPermissionConfigRepo repository.OnboardingPermissionConfigRepository
}

// NewRepositoryManager 创建Repository管理器
func NewRepositoryManager(db *gorm.DB) repository.RepositoryManager {
	return &RepositoryManagerImpl{
		db:                   db,
		userRepo:             NewUserRepository(db),
		roleRepo:             NewRoleRepository(db),
		permissionRepo:       NewPermissionRepository(db),
		employeeRepo:         NewEmployeeRepository(db),
		skillRepo:            NewSkillRepository(db),
		taskRepo:             NewTaskRepository(db),
		assignmentRepo:       NewAssignmentRepository(db),
		notificationRepo:     NewNotificationRepository(db),
		auditLogRepo:         NewAuditLogRepository(db),
		systemConfigRepo:     NewSystemConfigRepository(db),
		workflowRepo:         NewWorkflowRepository(db),
		workflowInstanceRepo: NewWorkflowInstanceRepository(db),
		departmentRepo:       NewDepartmentRepository(db),
		positionRepo:         NewPositionRepository(db),
		projectRepo:          NewProjectRepository(db),
		onboardingHistoryRepo: NewOnboardingHistoryRepository(db),
		
		// 权限分配相关仓储
		permissionTemplateRepo:        NewPermissionTemplateRepository(db),
		permissionRuleRepo:            NewPermissionRuleRepository(db),
		permissionAssignmentRepo:      NewPermissionAssignmentRepository(db),
		permissionAssignmentHistoryRepo: NewPermissionAssignmentHistoryRepository(db),
		onboardingPermissionConfigRepo: NewOnboardingPermissionConfigRepository(db),
	}
}

// UserRepository 获取用户仓储
func (m *RepositoryManagerImpl) UserRepository() repository.UserRepository {
	return m.userRepo
}

// RoleRepository 获取角色仓储
func (m *RepositoryManagerImpl) RoleRepository() repository.RoleRepository {
	return m.roleRepo
}

// PermissionRepository 获取权限仓储
func (m *RepositoryManagerImpl) PermissionRepository() repository.PermissionRepository {
	return m.permissionRepo
}

// EmployeeRepository 获取员工仓储
func (m *RepositoryManagerImpl) EmployeeRepository() repository.EmployeeRepository {
	return m.employeeRepo
}

// SkillRepository 获取技能仓储
func (m *RepositoryManagerImpl) SkillRepository() repository.SkillRepository {
	return m.skillRepo
}

// TaskRepository 获取任务仓储
func (m *RepositoryManagerImpl) TaskRepository() repository.TaskRepository {
	return m.taskRepo
}

// AssignmentRepository 获取分配仓储
func (m *RepositoryManagerImpl) AssignmentRepository() repository.AssignmentRepository {
	return m.assignmentRepo
}

// NotificationRepository 获取通知仓储
func (m *RepositoryManagerImpl) NotificationRepository() repository.NotificationRepository {
	return m.notificationRepo
}

// AuditLogRepository 获取审计日志仓储
func (m *RepositoryManagerImpl) AuditLogRepository() repository.AuditLogRepository {
	return m.auditLogRepo
}

// SystemConfigRepository 获取系统配置仓储
func (m *RepositoryManagerImpl) SystemConfigRepository() repository.SystemConfigRepository {
	return m.systemConfigRepo
}

// WorkflowRepository 获取工作流定义仓储
func (m *RepositoryManagerImpl) WorkflowRepository() repository.WorkflowRepository {
	return m.workflowRepo
}

// WorkflowInstanceRepository 获取工作流实例仓储
func (m *RepositoryManagerImpl) WorkflowInstanceRepository() repository.WorkflowInstanceRepository {
	return m.workflowInstanceRepo
}

// DepartmentRepository 获取部门仓储
func (m *RepositoryManagerImpl) DepartmentRepository() repository.DepartmentRepository {
	return m.departmentRepo
}

// PositionRepository 获取职位仓储
func (m *RepositoryManagerImpl) PositionRepository() repository.PositionRepository {
	return m.positionRepo
}

// ProjectRepository 获取项目仓储
func (m *RepositoryManagerImpl) ProjectRepository() repository.ProjectRepository {
	return m.projectRepo
}

// OnboardingHistoryRepository 获取入职历史仓储
func (m *RepositoryManagerImpl) OnboardingHistoryRepository() repository.OnboardingHistoryRepository {
	return m.onboardingHistoryRepo
}

// PermissionTemplateRepository 获取权限模板仓储
func (m *RepositoryManagerImpl) PermissionTemplateRepository() repository.PermissionTemplateRepository {
	return m.permissionTemplateRepo
}

// PermissionRuleRepository 获取权限规则仓储
func (m *RepositoryManagerImpl) PermissionRuleRepository() repository.PermissionRuleRepository {
	return m.permissionRuleRepo
}

// PermissionAssignmentRepository 获取权限分配仓储
func (m *RepositoryManagerImpl) PermissionAssignmentRepository() repository.PermissionAssignmentRepository {
	return m.permissionAssignmentRepo
}

// PermissionAssignmentHistoryRepository 获取权限分配历史仓储
func (m *RepositoryManagerImpl) PermissionAssignmentHistoryRepository() repository.PermissionAssignmentHistoryRepository {
	return m.permissionAssignmentHistoryRepo
}

// OnboardingPermissionConfigRepository 获取入职权限配置仓储
func (m *RepositoryManagerImpl) OnboardingPermissionConfigRepository() repository.OnboardingPermissionConfigRepository {
	return m.onboardingPermissionConfigRepo
}

// WithTx 在事务中执行操作
func (m *RepositoryManagerImpl) WithTx(ctx context.Context, fn func(ctx context.Context, repos repository.RepositoryManager) error) error {
	// 设置事务超时
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建事务版本的Repository管理器
		txManager := &RepositoryManagerImpl{
			db:                   tx,
			userRepo:             NewUserRepository(tx),
			roleRepo:             NewRoleRepository(tx),
			permissionRepo:       NewPermissionRepository(tx),
			employeeRepo:         NewEmployeeRepository(tx),
			skillRepo:            NewSkillRepository(tx),
			taskRepo:             NewTaskRepository(tx),
			assignmentRepo:       NewAssignmentRepository(tx),
			notificationRepo:     NewNotificationRepository(tx),
			auditLogRepo:         NewAuditLogRepository(tx),
			systemConfigRepo:     NewSystemConfigRepository(tx),
			workflowRepo:         NewWorkflowRepository(tx),
			workflowInstanceRepo: NewWorkflowInstanceRepository(tx),
			departmentRepo:       NewDepartmentRepository(tx),
			positionRepo:         NewPositionRepository(tx),
			projectRepo:          NewProjectRepository(tx),
			onboardingHistoryRepo: NewOnboardingHistoryRepository(tx),
			
			// 权限分配相关仓储
			permissionTemplateRepo:        NewPermissionTemplateRepository(tx),
			permissionRuleRepo:            NewPermissionRuleRepository(tx),
			permissionAssignmentRepo:      NewPermissionAssignmentRepository(tx),
			permissionAssignmentHistoryRepo: NewPermissionAssignmentHistoryRepository(tx),
			onboardingPermissionConfigRepo: NewOnboardingPermissionConfigRepository(tx),
		}

		// 执行业务逻辑
		if err := fn(ctx, txManager); err != nil {
			logger.Errorf("事务执行失败: %v", err)
			return fmt.Errorf("事务执行失败: %w", err)
		}

		return nil
	})
}

// HealthCheck 健康检查
func (m *RepositoryManagerImpl) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 检查数据库连接
	sqlDB, err := m.db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("数据库连接检查失败: %w", err)
	}

	// 执行简单查询测试
	var result int
	if err := m.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("数据库查询测试失败: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("数据库查询结果异常")
	}

	return nil
}

// GetDB 获取数据库连接
func (m *RepositoryManagerImpl) GetDB() *gorm.DB {
	return m.db
}

// 占位符实现，后续需要完整实现
func NewRoleRepository(db *gorm.DB) repository.RoleRepository {
	return &RoleRepositoryImpl{BaseRepositoryImpl: NewBaseRepository[database.Role](db)}
}

func NewPermissionRepository(db *gorm.DB) repository.PermissionRepository {
	return &PermissionRepositoryImpl{BaseRepositoryImpl: NewBaseRepository[database.Permission](db)}
}

// NewEmployeeRepository 和 NewSkillRepository 已移动到独立文件

// func NewAssignmentRepository(db *gorm.DB) repository.AssignmentRepository {
// 	return &AssignmentRepositoryImpl{BaseRepositoryImpl: NewBaseRepository[database.Assignment](db)}
// }

func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &NotificationRepositoryImpl{BaseRepositoryImpl: NewBaseRepository[database.TaskNotification](db)}
}

func NewAuditLogRepository(db *gorm.DB) repository.AuditLogRepository {
	return &AuditLogRepositoryImpl{BaseRepositoryImpl: NewBaseRepository[database.AuditLog](db)}
}

func NewSystemConfigRepository(db *gorm.DB) repository.SystemConfigRepository {
	return &SystemConfigRepositoryImpl{BaseRepositoryImpl: NewBaseRepository[database.SystemConfig](db)}
}
