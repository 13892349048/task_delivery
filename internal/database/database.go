package database

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"taskmanage/internal/config"
)

// DB 全局数据库实例
var DB *gorm.DB

// Connect 连接数据库并配置连接池
func Connect(cfg *config.Config) error {
	// 创建MySQL连接
	dsn := cfg.GetDSN()
	if dsn == "" {
		return fmt.Errorf("不支持的数据库驱动: %s", cfg.Database.Driver)
	}

	// GORM配置
	gormConfig := &gorm.Config{
		Logger: getLogger(cfg),
		// 禁用外键约束检查 (MySQL特定优化)
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("连接MySQL数据库失败: %w", err)
	}

	// 获取底层sql.DB实例进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 配置MySQL连接池参数
	configureConnectionPool(sqlDB, cfg)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("MySQL连接测试失败: %w", err)
	}

	// 设置MySQL特定的会话参数
	if err := setMySQLSessionParams(db); err != nil {
		return fmt.Errorf("设置MySQL会话参数失败: %w", err)
	}

	DB = db
	return nil
}

// configureConnectionPool 配置MySQL连接池
func configureConnectionPool(sqlDB *sql.DB, cfg *config.Config) {
	// 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	
	// 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	
	// 设置了连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)
	
	// 设置空闲连接的最大生存时间 (MySQL 8.0+ 推荐)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
}

// setMySQLSessionParams 设置MySQL会话参数
func setMySQLSessionParams(db *gorm.DB) error {
	// 设置MySQL时区
	if err := db.Exec("SET time_zone = '+08:00'").Error; err != nil {
		return fmt.Errorf("设置时区失败: %w", err)
	}

	// 设置字符集
	if err := db.Exec("SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		return fmt.Errorf("设置字符集失败: %w", err)
	}

	// 设置SQL模式 (严格模式)
	if err := db.Exec("SET sql_mode = 'STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO'").Error; err != nil {
		return fmt.Errorf("设置SQL模式失败: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// getLogger 获取GORM日志配置
func getLogger(cfg *config.Config) logger.Interface {
	logLevel := logger.Silent
	
	if cfg.IsDebug() {
		switch cfg.Log.Level {
		case "debug":
			logLevel = logger.Info
		case "info":
			logLevel = logger.Warn
		case "warn", "error":
			logLevel = logger.Error
		}
	}

	return logger.Default.LogMode(logLevel)
}

// Transaction 执行事务
func Transaction(fn func(tx *gorm.DB) error) error {
	return DB.Transaction(fn)
}

// IsConnected 检查数据库是否已连接
func IsConnected() bool {
	if DB == nil {
		return false
	}
	
	sqlDB, err := DB.DB()
	if err != nil {
		return false
	}
	
	return sqlDB.Ping() == nil
}

// GetStats 获取数据库连接池统计信息
func GetStats() (map[string]interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	
	sqlDB, err := DB.DB()
	if err != nil {
		return nil, err
	}
	
	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                  stats.InUse,
		"idle":                    stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration":           stats.WaitDuration.String(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
	}, nil
}

// HealthCheck 数据库健康检查
func HealthCheck() error {
	if !IsConnected() {
		return fmt.Errorf("数据库未连接")
	}

	// 执行简单查询测试
	var result int
	if err := DB.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("数据库查询测试失败: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("数据库查询结果异常")
	}

	return nil
}

// GetVersion 获取MySQL版本信息
func GetVersion() (string, error) {
	if DB == nil {
		return "", fmt.Errorf("数据库未连接")
	}

	var version string
	if err := DB.Raw("SELECT VERSION()").Scan(&version).Error; err != nil {
		return "", fmt.Errorf("获取MySQL版本失败: %w", err)
	}

	return version, nil
}
