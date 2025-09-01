package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"taskmanage/internal/assignment"
	"taskmanage/internal/config"
	"taskmanage/internal/repository"
	"taskmanage/internal/workflow"
)

// serviceManager 服务管理器实现
type serviceManager struct {
	repoManager repository.RepositoryManager
	config      *config.Config
	logger      *logrus.Logger

	userService         UserService
	taskService         TaskService
	employeeService     EmployeeService
	skillService        SkillService
	notificationService NotificationService
	assignmentService   *assignment.AssignmentService
	workflowService     *workflow.WorkflowService
	departmentService   DepartmentService
	positionService     PositionService
	projectService      ProjectService
	onboardingService   OnboardingService
	permissionAssignmentService PermissionAssignmentService
}

// NewServiceManager 创建服务管理器
func NewServiceManager(repoManager repository.RepositoryManager, cfg *config.Config, logger *logrus.Logger) ServiceManager {
	return &serviceManager{
		repoManager: repoManager,
		config:      cfg,
		logger:      logger,
	}
}

// UserService 获取用户服务
func (sm *serviceManager) UserService() UserService {
	if sm.userService == nil {
		sm.userService = NewUserService(sm.repoManager, sm.config)
	}
	return sm.userService
}

// TaskService 获取任务服务
func (sm *serviceManager) TaskService() TaskService {
	if sm.taskService == nil {
		assignmentService := assignment.NewAssignmentService(sm.repoManager)
		workflowService := sm.WorkflowService()
		sm.taskService = NewTaskService(
			sm.repoManager.TaskRepository(),
			sm.repoManager.EmployeeRepository(),
			sm.repoManager.UserRepository(),
			sm.repoManager.AssignmentRepository(),
			assignmentService,
			workflowService,
		)
		// 解决循环依赖：将TaskService注入到WorkflowService中
		if sm.workflowService != nil {
			sm.workflowService.SetTaskService(sm.taskService)
		}
	}
	return sm.taskService
}

// EmployeeService 获取员工服务
func (sm *serviceManager) EmployeeService() EmployeeService {
	if sm.employeeService == nil {
		sm.employeeService = NewEmployeeService(sm.repoManager.EmployeeRepository(), sm.repoManager.SkillRepository(), sm.repoManager.UserRepository())
	}
	return sm.employeeService
}

// SkillService 获取技能服务
func (sm *serviceManager) SkillService() SkillService {
	if sm.skillService == nil {
		sm.skillService = NewSkillService(sm.repoManager.SkillRepository(), sm.repoManager.EmployeeRepository())
	}
	return sm.skillService
}

// NotificationService 获取通知服务
func (sm *serviceManager) NotificationService() NotificationService {
	if sm.notificationService == nil {
		sm.notificationService = NewNotificationService(sm.repoManager)
	}
	return sm.notificationService
}

// WorkflowService 获取工作流服务
func (sm *serviceManager) WorkflowService() WorkflowService {
	if sm.workflowService == nil {
		// 创建workflow repository adapters
		workflowRepoAdapter := NewWorkflowRepositoryAdapter(sm.repoManager.WorkflowRepository())
		workflowInstanceRepoAdapter := NewWorkflowInstanceRepositoryAdapter(sm.repoManager.WorkflowInstanceRepository())
		
		// 创建workflow definition manager
		definitionManager := workflow.NewWorkflowDefinitionManager(workflowRepoAdapter)
		
		// 创建workflow engine
		engine := workflow.NewWorkflowEngine(definitionManager, workflowInstanceRepoAdapter, sm.repoManager.EmployeeRepository(), sm.repoManager.UserRepository())
		
		// 创建workflow service
		sm.workflowService = workflow.NewWorkflowService(engine, definitionManager)
	}
	return NewWorkflowServiceWrapper(sm.workflowService)
}

// DepartmentService 获取部门服务
func (sm *serviceManager) DepartmentService() DepartmentService {
	if sm.departmentService == nil {
		sm.departmentService = NewDepartmentService(sm.repoManager, sm.logger)
	}
	return sm.departmentService
}

// PositionService 获取职位服务
func (sm *serviceManager) PositionService() PositionService {
	if sm.positionService == nil {
		sm.positionService = NewPositionService(sm.repoManager, sm.logger)
	}
	return sm.positionService
}

// ProjectService 获取项目服务
func (sm *serviceManager) ProjectService() ProjectService {
	if sm.projectService == nil {
		sm.projectService = NewProjectService(sm.repoManager, sm.logger)
	}
	return sm.projectService
}

// OnboardingService 获取入职工作流服务
func (sm *serviceManager) OnboardingService() OnboardingService {
	if sm.onboardingService == nil {
		sm.onboardingService = NewOnboardingService(sm.repoManager, sm.WorkflowService(), sm.PermissionAssignmentService(), sm.logger)
	}
	return sm.onboardingService
}

// PermissionAssignmentService 获取权限分配服务
func (sm *serviceManager) PermissionAssignmentService() PermissionAssignmentService {
	if sm.permissionAssignmentService == nil {
		sm.permissionAssignmentService = NewPermissionAssignmentService(sm.repoManager)
	}
	return sm.permissionAssignmentService
}

// HealthCheck 健康检查
func (sm *serviceManager) HealthCheck(ctx context.Context) error {
	// 检查Repository管理器
	if err := sm.repoManager.HealthCheck(ctx); err != nil {
		return fmt.Errorf("Repository管理器健康检查失败: %w", err)
	}

	// 这里可以添加更多服务层的健康检查逻辑
	// 例如检查外部API连接、缓存服务等

	return nil
}

// 占位符服务实现函数，需要后续完善

// func NewEmployeeService(repoManager repository.RepositoryManager, cfg *config.Config) EmployeeService {
// 	// TODO: 实现EmployeeService
// 	return nil
// }

// func NewNotificationService(repoManager repository.RepositoryManager, cfg *config.Config) NotificationService {
// 	// TODO: 实现NotificationService
// 	return nil
// }
