package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// LoaderOptions 配置加载选项
type LoaderOptions struct {
	ConfigPath string
	ConfigName string
	EnvPrefix  string
}

// ConfigLoader 配置加载器
type ConfigLoader struct {
	options *LoaderOptions
	viper   *viper.Viper
}

// NewLoader 创建新的配置加载器
func NewLoader(opts *LoaderOptions) *ConfigLoader {
	if opts == nil {
		opts = &LoaderOptions{
			ConfigName: "config",
			EnvPrefix:  "TASKMANAGE",
		}
	}

	v := viper.New()
	return &ConfigLoader{
		options: opts,
		viper:   v,
	}
}

// LoadConfig 加载配置
func (l *ConfigLoader) LoadConfig() (*Config, error) {
	// 设置配置文件名和类型
	l.viper.SetConfigName(l.options.ConfigName)
	l.viper.SetConfigType("yaml")

	// 添加配置文件搜索路径
	if l.options.ConfigPath != "" {
		l.viper.AddConfigPath(l.options.ConfigPath)
	} else {
		l.addDefaultConfigPaths()
	}

	// 设置环境变量
	l.setupEnvironmentVariables()

	// 设置默认值
	l.setDefaults()

	// 读取配置文件
	if err := l.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("配置文件未找到: %w", err)
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	config := &Config{}
	if err := l.viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证配置
	if err := l.validateConfig(config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return config, nil
}

// addDefaultConfigPaths 添加默认配置文件搜索路径
func (l *ConfigLoader) addDefaultConfigPaths() {
	// 当前目录
	l.viper.AddConfigPath(".")
	l.viper.AddConfigPath("./configs")
	
	// 上级目录
	l.viper.AddConfigPath("../configs")
	l.viper.AddConfigPath("../../configs")
	
	// 系统配置目录
	l.viper.AddConfigPath("/etc/taskmanage")
	l.viper.AddConfigPath("$HOME/.taskmanage")
	
	// 可执行文件目录
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		l.viper.AddConfigPath(execDir)
		l.viper.AddConfigPath(filepath.Join(execDir, "configs"))
	}
}

// setupEnvironmentVariables 设置环境变量
func (l *ConfigLoader) setupEnvironmentVariables() {
	l.viper.SetEnvPrefix(l.options.EnvPrefix)
	l.viper.AutomaticEnv()
	
	// 支持嵌套配置的环境变量
	l.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// 手动绑定关键环境变量
	envBindings := map[string]string{
		"app.environment":     "APP_ENVIRONMENT",
		"app.debug":          "APP_DEBUG",
		"server.port":        "SERVER_PORT",
		"database.host":      "DB_HOST",
		"database.port":      "DB_PORT",
		"database.username":  "DB_USERNAME",
		"database.password":  "DB_PASSWORD",
		"database.database":  "DB_NAME",
		"redis.host":         "REDIS_HOST",
		"redis.port":         "REDIS_PORT",
		"redis.password":     "REDIS_PASSWORD",
		"jwt.secret":         "JWT_SECRET",
	}
	
	for key, env := range envBindings {
		l.viper.BindEnv(key, env)
	}
}

// setDefaults 设置默认值
func (l *ConfigLoader) setDefaults() {
	// 应用默认值
	l.viper.SetDefault("app.name", "TaskManage")
	l.viper.SetDefault("app.version", "1.0.0")
	l.viper.SetDefault("app.environment", "development")
	l.viper.SetDefault("app.debug", true)
	
	// 服务器默认值
	l.viper.SetDefault("server.host", "0.0.0.0")
	l.viper.SetDefault("server.port", 8080)
	l.viper.SetDefault("server.read_timeout", 60)
	l.viper.SetDefault("server.write_timeout", 60)
	l.viper.SetDefault("server.idle_timeout", 120)
	
	// 数据库默认值
	l.viper.SetDefault("database.driver", "mysql")
	l.viper.SetDefault("database.charset", "utf8mb4")
	l.viper.SetDefault("database.max_idle_conns", 10)
	l.viper.SetDefault("database.max_open_conns", 100)
	l.viper.SetDefault("database.conn_max_lifetime", 3600)
	
	// Redis默认值
	l.viper.SetDefault("redis.database", 0)
	l.viper.SetDefault("redis.max_retries", 3)
	l.viper.SetDefault("redis.pool_size", 10)
	l.viper.SetDefault("redis.min_idle_conn", 5)
	
	// JWT默认值
	l.viper.SetDefault("jwt.access_token_ttl", 3600)
	l.viper.SetDefault("jwt.refresh_token_ttl", 604800)
	l.viper.SetDefault("jwt.issuer", "taskmanage")
	l.viper.SetDefault("jwt.refresh_threshold", 1800)
	
	// Asynq默认值
	l.viper.SetDefault("asynq.concurrency", 10)
	l.viper.SetDefault("asynq.redis_db", 1)
	
	// 日志默认值
	l.viper.SetDefault("log.level", "info")
	l.viper.SetDefault("log.format", "json")
	l.viper.SetDefault("log.output", "stdout")
	l.viper.SetDefault("log.max_size", 100)
	l.viper.SetDefault("log.max_backups", 5)
	l.viper.SetDefault("log.max_age", 30)
	l.viper.SetDefault("log.compress", true)
}

// validateConfig 验证配置
func (l *ConfigLoader) validateConfig(config *Config) error {
	v := validator.New()
	return v.Struct(config)
}

// GetConfigFilePath 获取当前使用的配置文件路径
func (l *ConfigLoader) GetConfigFilePath() string {
	return l.viper.ConfigFileUsed()
}

// WatchConfig 监听配置文件变化
func (l *ConfigLoader) WatchConfig(callback func(*Config)) error {
	l.viper.WatchConfig()
	l.viper.OnConfigChange(func(e fsnotify.Event) {
		config := &Config{}
		if err := l.viper.Unmarshal(config); err == nil {
			if err := l.validateConfig(config); err == nil {
				callback(config)
			}
		}
	})
	return nil
}

// LoadByEnvironment 根据环境加载配置
func LoadByEnvironment(env string) (*Config, error) {
	configName := "config"
	
	// 环境名称映射
	envMapping := map[string]string{
		"development": "config",
		"dev":         "config.dev",
		"testing":     "config.test",
		"test":        "config.test",
		"production":  "config.prod",
		"prod":        "config.prod",
		"prod-test":   "config.prod.test", // 临时测试配置
	}
	
	if mappedName, exists := envMapping[env]; exists {
		configName = mappedName
	} else if env != "" {
		configName = fmt.Sprintf("config.%s", env)
	}
	
	loader := NewLoader(&LoaderOptions{
		ConfigName: configName,
		EnvPrefix:  "TASKMANAGE",
	})
	
	return loader.LoadConfig()
}

// MustLoad 必须成功加载配置，否则panic
func MustLoad(configPath string) *Config {
	config, err := Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}
	return config
}
