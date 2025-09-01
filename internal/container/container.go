package container

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"taskmanage/internal/assignment"
	"taskmanage/internal/config"
	"taskmanage/internal/repository"
	"taskmanage/internal/repository/mysql"
	"taskmanage/internal/service"
	"taskmanage/pkg/jwt"
	"taskmanage/pkg/logger"
)

// Container 依赖注入容器
type Container struct {
	mu        sync.RWMutex
	instances map[string]interface{}
	factories map[string]func() (interface{}, error)
}

// NewContainer 创建新的容器
func NewContainer() *Container {
	return &Container{
		instances: make(map[string]interface{}),
		factories: make(map[string]func() (interface{}, error)),
	}
}

// Register 注册工厂函数
func (c *Container) Register(name string, factory func() (interface{}, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[name] = factory
}

// RegisterSingleton 注册单例
func (c *Container) RegisterSingleton(name string, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.instances[name] = instance
}

// Get 获取实例
func (c *Container) Get(name string) (interface{}, error) {
	logger.Infof("开始获取实例: %s", name)

	// 先检查是否已存在
	c.mu.RLock()
	if instance, exists := c.instances[name]; exists {
		c.mu.RUnlock()
		logger.Infof("实例 %s 已存在，直接返回", name)
		return instance, nil
	}

	// 检查工厂函数是否存在
	factory, exists := c.factories[name]
	c.mu.RUnlock()

	if !exists {
		logger.Errorf("未找到名为 '%s' 的工厂函数", name)
		return nil, fmt.Errorf("未找到名为 '%s' 的工厂函数", name)
	}

	logger.Infof("实例 %s 不存在，开始创建", name)

	// 执行工厂函数（不持有锁，避免死锁）
	logger.Infof("开始执行工厂函数创建实例: %s", name)
	instance, err := factory()
	if err != nil {
		logger.Errorf("创建实例 '%s' 失败: %v", name, err)
		return nil, fmt.Errorf("创建实例 '%s' 失败: %w", name, err)
	}

	// 缓存实例
	c.mu.Lock()
	// 双重检查，防止并发创建
	if existingInstance, exists := c.instances[name]; exists {
		c.mu.Unlock()
		logger.Infof("双重检查：实例 %s 已存在，使用已有实例", name)
		return existingInstance, nil
	}

	c.instances[name] = instance
	c.mu.Unlock()

	logger.Infof("实例 %s 创建成功并缓存", name)
	return instance, nil
}

// MustGet 获取实例，失败时panic
func (c *Container) MustGet(name string) interface{} {
	instance, err := c.Get(name)
	if err != nil {
		panic(err)
	}
	return instance
}

// GetTyped 获取指定类型的实例
func GetTyped[T any](c *Container, name string) (T, error) {
	var zero T
	instance, err := c.Get(name)
	if err != nil {
		return zero, err
	}

	typed, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("类型断言失败: 期望 %T, 实际 %T", zero, instance)
	}

	return typed, nil
}

// ApplicationContainer 应用程序容器
type ApplicationContainer struct {
	*Container
	config *config.Config
	db     *gorm.DB
}

// NewApplicationContainer 创建应用程序容器
func NewApplicationContainer(cfg *config.Config, db *gorm.DB) *ApplicationContainer {
	container := &ApplicationContainer{
		Container: NewContainer(),
		config:    cfg,
		db:        db,
	}

	container.registerDefaults()
	return container
}

// registerDefaults 注册默认依赖
func (c *ApplicationContainer) registerDefaults() {
	// 注册配置
	c.RegisterSingleton("config", c.config)

	// 注册数据库
	c.RegisterSingleton("db", c.db)

	// 注册Logger
	c.RegisterSingleton("logger", logger.GetLogger())

	// 注册JWT管理器
	c.Register("jwt.manager", func() (interface{}, error) {
		return jwt.NewJWTManager(
			c.config.JWT.Secret,
			time.Duration(c.config.JWT.AccessTokenTTL)*time.Hour,
			time.Duration(c.config.JWT.RefreshTokenTTL)*time.Hour*24,
			c.config.JWT.Issuer,
		), nil
	})

	// 注册Repository管理器
	c.Register("repository.manager", func() (interface{}, error) {
		return mysql.NewRepositoryManager(c.db), nil
	})

	// 注册各个Repository
	c.Register("repository.user", func() (interface{}, error) {
		return mysql.NewUserRepository(c.db), nil
	})

	// 注册Service管理器
	c.Register("service.manager", func() (interface{}, error) {
		repoManager, err := GetTyped[repository.RepositoryManager](c.Container, "repository.manager")
		if err != nil {
			return nil, err
		}
		logger := c.GetLogger()
		return NewServiceManager(repoManager, c.config, logger), nil
	})
	// 注册各个Service
	c.Register("service.user", func() (interface{}, error) {
		repoManager, err := GetTyped[repository.RepositoryManager](c.Container, "repository.manager")
		if err != nil {
			return nil, err
		}
		return NewUserService(repoManager, c.config), nil
	})

	c.Register("service.task", func() (interface{}, error) {
		repoManager, err := GetTyped[repository.RepositoryManager](c.Container, "repository.manager")
		if err != nil {
			return nil, err
		}
		serviceManager, err := GetTyped[service.ServiceManager](c.Container, "service.manager")
		if err != nil {
			return nil, err
		}
		// 创建分配服务
		assignmentService := assignment.NewAssignmentService(repoManager)
		// 获取workflow服务
		workflowService := serviceManager.WorkflowService()
		return service.NewTaskService(repoManager.TaskRepository(), repoManager.EmployeeRepository(), repoManager.UserRepository(), repoManager.AssignmentRepository(), assignmentService, workflowService), nil
	})

	// 注册分配管理服务
	c.Register("service.assignment_management", func() (interface{}, error) {
		repoManager, err := GetTyped[repository.RepositoryManager](c.Container, "repository.manager")
		if err != nil {
			return nil, err
		}
		serviceManager, err := GetTyped[service.ServiceManager](c.Container, "service.manager")
		if err != nil {
			return nil, err
		}
		// 创建分配服务
		assignmentService := assignment.NewAssignmentService(repoManager)
		// 获取workflow服务
		workflowService := serviceManager.WorkflowService()
		notificationService := serviceManager.NotificationService()
		return service.NewAssignmentManagementService(assignmentService, workflowService, repoManager, notificationService), nil
	})
}

// GetConfig 获取配置
func (c *ApplicationContainer) GetConfig() *config.Config {
	return c.config
}

// GetDB 获取数据库连接
func (c *ApplicationContainer) GetDB() *gorm.DB {
	return c.db
}

// GetRepositoryManager 获取Repository管理器
func (c *ApplicationContainer) GetRepositoryManager() repository.RepositoryManager {
	return c.MustGet("repository.manager").(repository.RepositoryManager)
}

// GetServiceManager 获取服务管理器
func (c *ApplicationContainer) GetServiceManager() service.ServiceManager {
	logger.Info("开始获取ServiceManager...")
	serviceManager, err := GetTyped[service.ServiceManager](c.Container, "service.manager")
	if err != nil {
		logger.Errorf("获取服务管理器失败: %v", err)
		panic(fmt.Sprintf("获取服务管理器失败: %v", err))
	}
	logger.Info("ServiceManager获取成功")
	return serviceManager
}

// GetLogger 获取日志器
func (c *ApplicationContainer) GetLogger() *logrus.Logger {
	logger, err := GetTyped[*logrus.Logger](c.Container, "logger")
	if err != nil {
		panic(fmt.Sprintf("获取日志器失败: %v", err))
	}
	return logger
}

// GetJWTManager 获取JWT管理器
func (c *ApplicationContainer) GetJWTManager() (*jwt.JWTManager, error) {
	return GetTyped[*jwt.JWTManager](c.Container, "jwt.manager")
}

// HealthCheck 健康检查
func (c *ApplicationContainer) HealthCheck(ctx context.Context) error {
	// 检查Repository层
	repoManager := c.GetRepositoryManager()
	if err := repoManager.HealthCheck(ctx); err != nil {
		return fmt.Errorf("Repository层健康检查失败: %w", err)
	}

	// 检查Service层
	serviceManager := c.GetServiceManager()
	if err := serviceManager.HealthCheck(ctx); err != nil {
		return fmt.Errorf("Service层健康检查失败: %w", err)
	}

	return nil
}

// Shutdown 关闭容器
func (c *ApplicationContainer) Shutdown(ctx context.Context) error {
	logger.Info("正在关闭应用程序容器...")

	// 这里可以添加清理逻辑
	// 例如关闭连接池、清理缓存等

	logger.Info("应用程序容器已关闭")
	return nil
}

// 全局容器实例
var globalContainer *ApplicationContainer
var containerOnce sync.Once

// InitGlobalContainer 初始化全局容器
func InitGlobalContainer(cfg *config.Config, db *gorm.DB) {
	containerOnce.Do(func() {
		globalContainer = NewApplicationContainer(cfg, db)
	})
}

// GetGlobalContainer 获取全局容器
func GetGlobalContainer() *ApplicationContainer {
	if globalContainer == nil {
		panic("全局容器未初始化")
	}
	return globalContainer
}

// 便捷函数
func GetConfig() *config.Config {
	return GetGlobalContainer().GetConfig()
}

func GetDB() *gorm.DB {
	return GetGlobalContainer().GetDB()
}

func GetRepositoryManager() repository.RepositoryManager {
	return GetGlobalContainer().GetRepositoryManager()
}

func GetServiceManager() service.ServiceManager {
	return GetGlobalContainer().GetServiceManager()
}

// NewServiceManager 创建ServiceManager - 现在使用实际实现
func NewServiceManager(repoManager repository.RepositoryManager, cfg *config.Config, logger *logrus.Logger) service.ServiceManager {
	return service.NewServiceManager(repoManager, cfg, logger)
}

func NewUserService(repoManager repository.RepositoryManager, cfg *config.Config) service.UserService {
	return service.NewUserService(repoManager, cfg)
}

func NewTaskService(repoManager repository.RepositoryManager, cfg *config.Config) service.TaskService {
	// 创建分配服务
	assignmentService := assignment.NewAssignmentService(repoManager)
	// 获取workflow服务
	logger := logrus.New() // TODO: Get from container
	serviceManager := service.NewServiceManager(repoManager, cfg, logger)
	workflowService := serviceManager.WorkflowService()
	return service.NewTaskService(repoManager.TaskRepository(), repoManager.EmployeeRepository(), repoManager.UserRepository(), repoManager.AssignmentRepository(), assignmentService, workflowService)
}

// GetEmployeeService 获取员工服务
func (c *ApplicationContainer) GetEmployeeService() service.EmployeeService {
	return c.GetServiceManager().EmployeeService()
}

// GetSkillService 获取技能服务
func (c *ApplicationContainer) GetSkillService() service.SkillService {
	return c.GetServiceManager().SkillService()
}

// GetAssignmentManagementService 获取分配管理服务
func (c *ApplicationContainer) GetAssignmentManagementService() service.AssignmentService {
	assignmentMgmtService, err := GetTyped[*service.AssignmentManagementService](c.Container, "service.assignment_management")
	if err != nil {
		panic(fmt.Sprintf("获取分配管理服务失败: %v", err))
	}
	return assignmentMgmtService
}
