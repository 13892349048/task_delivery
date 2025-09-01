package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	router "taskmanage/internal/api"
	"taskmanage/internal/config"
	"taskmanage/internal/container"
	"taskmanage/internal/database"
	"taskmanage/internal/service"
	"taskmanage/internal/utils"
	"taskmanage/pkg/logger"
)

func main() {
	// 设置全局错误恢复
	defer utils.Recovery()

	// 命令行参数
	var (
		configPath = flag.String("config", "", "配置文件路径")
		env        = flag.String("env", "development", "运行环境 (development, testing, production)")
		showConfig = flag.Bool("show-config", false, "显示配置信息")
	)
	flag.Parse()

	// 加载配置
	var cfg *config.Config
	var err error

	if *configPath != "" {
		cfg, err = config.Load(*configPath)
	} else {
		cfg, err = config.LoadByEnvironment(*env)
	}

	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志系统
	if err := logger.Init(cfg); err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}

	// 添加钩子
	logger.AddHooks(logger.GetLogger(), cfg.App.Name, cfg.App.Version, cfg.App.Environment)

	// 显示配置信息
	if *showConfig {
		fmt.Printf("应用程序配置:\n")
		fmt.Printf("  名称: %s\n", cfg.App.Name)
		fmt.Printf("  版本: %s\n", cfg.App.Version)
		fmt.Printf("  环境: %s\n", cfg.App.Environment)
		fmt.Printf("  调试模式: %t\n", cfg.App.Debug)
		fmt.Printf("服务器配置:\n")
		fmt.Printf("  监听地址: %s\n", cfg.GetServerAddr())
		fmt.Printf("数据库配置:\n")
		fmt.Printf("  驱动: %s\n", cfg.Database.Driver)
		fmt.Printf("  主机: %s:%d\n", cfg.Database.Host, cfg.Database.Port)
		fmt.Printf("  数据库: %s\n", cfg.Database.Database)
		fmt.Printf("  连接池: 最大连接数=%d, 空闲连接数=%d\n", cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns)
		fmt.Printf("Redis配置:\n")
		fmt.Printf("  地址: %s\n", cfg.GetRedisAddr())
		fmt.Printf("  数据库: %d\n", cfg.Redis.Database)
		os.Exit(0)
	}

	// 使用日志系统记录启动信息
	logger.Infof("TaskManage %s 启动中...", cfg.App.Version)
	logger.Infof("环境: %s", cfg.App.Environment)
	logger.Infof("服务器将在 %s 上启动", cfg.GetServerAddr())
	logger.Infof("日志级别: %s", cfg.Log.Level)
	logger.Infof("日志格式: %s", cfg.Log.Format)

	// 初始化数据库连接池
	logger.Info("正在初始化MySQL数据库连接...")
	if err := database.Initialize(cfg); err != nil {
		logger.Fatalf("数据库初始化失败: %v", err)
	}

	// 确保程序退出时关闭数据库连接
	defer func() {
		if err := database.Close(); err != nil {
			logger.Errorf("关闭数据库连接失败: %v", err)
		} else {
			logger.Info("数据库连接已关闭")
		}
	}()

	// 测试数据库连接
	if err := database.TestConnection(); err != nil {
		logger.Errorf("数据库连接测试失败: %v", err)
	} else {
		logger.Info("数据库连接测试通过")
	}

	// 输出数据库连接信息
	dbInfo := database.GetConnectionInfo()
	logger.Infof("数据库连接信息: %+v", dbInfo)

	logger.Info("MySQL连接池配置完成，系统准备就绪")

	// 启动HTTP服务器
	startHTTPServer(cfg)
}

// startHTTPServer 启动HTTP服务器
func startHTTPServer(cfg *config.Config) {
	logger.Info("正在启动HTTP服务器...")

	// 获取数据库连接
	db := database.GetDB()
	if db == nil {
		logger.Fatal("无法获取数据库连接")
	}

	// 初始化全局容器
	container.InitGlobalContainer(cfg, db)
	appContainer := container.GetGlobalContainer()

	// 初始化系统默认数据（角色、权限、超级管理员）
	if err := initializeSystemData(appContainer); err != nil {
		logger.Errorf("初始化系统默认数据失败: %v", err)
	} else {
		logger.Info("系统默认数据初始化完成")
	}

	// 初始化默认工作流定义
	if err := utils.InitializeDefaultWorkflows(appContainer); err != nil {
		logger.Errorf("初始化默认工作流定义失败: %v", err)
	} else {
		logger.Info("默认工作流定义初始化完成")
	}

	// 初始化权限模板
	if err := initializePermissionTemplates(appContainer); err != nil {
		logger.Errorf("初始化权限模板失败: %v", err)
	} else {
		logger.Info("权限模板初始化完成")
	}

	// 创建路由器
	logger.Info("正在创建路由器...")
	engine := router.NewRouter(appContainer, logger.GetLogger())
	logger.Info("路由器创建完成")

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    cfg.GetServerAddr(),
		Handler: engine,
	}
	
	// 启动服务器
	go func() {
		logger.Infof("HTTP服务器正在启动，监听地址: %s", cfg.GetServerAddr())
		logger.Info("服务器已准备就绪，等待请求...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("HTTP服务器启动失败: %v", err)
		}
	}()
	
	// 给服务器一点时间完全启动
	time.Sleep(100 * time.Millisecond)
	logger.Info("HTTP服务器启动完成，现在可以接收请求")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("服务器关闭失败: %v", err)
	} else {
		logger.Info("服务器已优雅关闭")
	}
}

// initializeSystemData 初始化系统默认数据
func initializeSystemData(appContainer *container.ApplicationContainer) error {
	// 获取仓储管理器
	repoManager := appContainer.GetRepositoryManager()
	
	// 创建bootstrap服务
	bootstrapService := service.NewBootstrapService(repoManager)
	
	// 初始化系统数据
	ctx := context.Background()
	return bootstrapService.InitializeSystem(ctx)
}

// initializePermissionTemplates 初始化权限模板
func initializePermissionTemplates(appContainer *container.ApplicationContainer) error {
	// 获取权限分配服务
	serviceManager := appContainer.GetServiceManager()
	permissionService := serviceManager.PermissionAssignmentService()
	
	ctx := context.Background()
	
	// 初始化基础权限模板
	if err := permissionService.InitializePermissionTemplates(ctx); err != nil {
		return fmt.Errorf("初始化基础权限模板失败: %w", err)
	}
	
	// 初始化部门特定权限模板
	if err := permissionService.InitializeDepartmentSpecificTemplates(ctx); err != nil {
		logger.Warnf("初始化部门特定权限模板失败: %v", err)
		// 不返回错误，因为部门特定模板不是必需的
	}
	
	// 初始化入职权限配置
	if err := permissionService.InitializeOnboardingPermissionConfigs(ctx); err != nil {
		logger.Warnf("初始化入职权限配置失败: %v", err)
		// 不返回错误，因为可以后续手动配置
	}
	
	return nil
}
