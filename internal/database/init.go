package database

import (
	"fmt"
	"log"

	"taskmanage/internal/config"
	"taskmanage/pkg/logger"
)

// Initialize 初始化数据库连接和表结构
func Initialize(cfg *config.Config) error {
	// 连接数据库
	if err := Connect(cfg); err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	logger.Info("MySQL数据库连接成功")

	// 获取数据库版本信息
	version, err := GetVersion()
	if err != nil {
		logger.Warnf("获取MySQL版本失败: %v", err)
	} else {
		logger.Infof("MySQL版本: %s", version)
	}

	// 执行数据库迁移
	if err := Migrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	logger.Info("数据库迁移完成")

	// 输出连接池状态
	if stats, err := GetStats(); err == nil {
		logger.Infof("数据库连接池状态: %+v", stats)
	}

	return nil
}

// TestConnection 测试数据库连接
func TestConnection() error {
	if err := HealthCheck(); err != nil {
		return fmt.Errorf("数据库健康检查失败: %w", err)
	}

	// 测试基本CRUD操作
	if err := testBasicOperations(); err != nil {
		return fmt.Errorf("基本操作测试失败: %w", err)
	}

	log.Println("数据库连接测试通过")
	return nil
}

// testBasicOperations 测试基本的数据库操作
func testBasicOperations() error {
	// 简化测试：只做基本查询测试，避免插入操作
	var count int64
	if err := DB.Model(&SystemConfig{}).Count(&count).Error; err != nil {
		return fmt.Errorf("查询测试失败: %w", err)
	}

	return nil
}

// GetConnectionInfo 获取数据库连接信息
func GetConnectionInfo() map[string]interface{} {
	info := make(map[string]interface{})

	if DB == nil {
		info["connected"] = false
		return info
	}

	info["connected"] = IsConnected()

	if version, err := GetVersion(); err == nil {
		info["version"] = version
	}

	if stats, err := GetStats(); err == nil {
		info["stats"] = stats
	}

	return info
}
