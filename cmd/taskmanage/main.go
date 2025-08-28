package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"taskmanage/internal/config"
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
		fmt.Printf("Redis配置:\n")
		fmt.Printf("  地址: %s\n", cfg.GetRedisAddr())
		fmt.Printf("  数据库: %d\n", cfg.Redis.Database)
		fmt.Println("jwt", cfg.JWT.Secret)
		os.Exit(0)
	}

	// 使用日志系统记录启动信息
	logger.Infof("TaskManage %s 启动中...", cfg.App.Version)
	logger.Infof("环境: %s", cfg.App.Environment)
	logger.Infof("服务器将在 %s 上启动", cfg.GetServerAddr())
	logger.Infof("日志级别: %s", cfg.Log.Level)
	logger.Infof("日志格式: %s", cfg.Log.Format)

	// TODO: 这里将来会启动HTTP服务器和其他组件
	logger.Info("配置管理和日志系统已完成，等待其他模块...")
}
