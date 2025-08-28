package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config 应用程序配置结构
type Config struct {
	App      AppConfig      `mapstructure:"app" validate:"required"`
	Server   ServerConfig   `mapstructure:"server" validate:"required"`
	Database DatabaseConfig `mapstructure:"database" validate:"required"`
	Redis    RedisConfig    `mapstructure:"redis" validate:"required"`
	JWT      JWTConfig      `mapstructure:"jwt" validate:"required"`
	Asynq    AsynqConfig    `mapstructure:"asynq" validate:"required"`
	Log      LogConfig      `mapstructure:"log" validate:"required"`
}

// AppConfig 应用程序基础配置
type AppConfig struct {
	Name        string `mapstructure:"name" validate:"required"`
	Version     string `mapstructure:"version" validate:"required"`
	Environment string `mapstructure:"environment" validate:"required,oneof=development testing production"`
	Debug       bool   `mapstructure:"debug"`
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	Host         string `mapstructure:"host" validate:"required"`
	Port         int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout  int    `mapstructure:"read_timeout" validate:"min=1"`
	WriteTimeout int    `mapstructure:"write_timeout" validate:"min=1"`
	IdleTimeout  int    `mapstructure:"idle_timeout" validate:"min=1"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver" validate:"required,oneof=mysql postgres sqlite"`
	Host            string `mapstructure:"host" validate:"required"`
	Port            int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Username        string `mapstructure:"username" validate:"required"`
	Password        string `mapstructure:"password" validate:"required"`
	Database        string `mapstructure:"database" validate:"required"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns" validate:"min=1"`
	MaxOpenConns    int    `mapstructure:"max_open_conns" validate:"min=1"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime" validate:"min=1"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host        string `mapstructure:"host" validate:"required"`
	Port        int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Password    string `mapstructure:"password"`
	Database    int    `mapstructure:"database" validate:"min=0,max=15"`
	MaxRetries  int    `mapstructure:"max_retries" validate:"min=0"`
	PoolSize    int    `mapstructure:"pool_size" validate:"min=1"`
	MinIdleConn int    `mapstructure:"min_idle_conn" validate:"min=0"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret           string `mapstructure:"secret" validate:"required,min=32"`
	AccessTokenTTL   int    `mapstructure:"access_token_ttl" validate:"required,min=1"`
	RefreshTokenTTL  int    `mapstructure:"refresh_token_ttl" validate:"required,min=1"`
	Issuer           string `mapstructure:"issuer" validate:"required"`
	RefreshThreshold int    `mapstructure:"refresh_threshold" validate:"min=1"`
}

// AsynqConfig Asynq队列配置
type AsynqConfig struct {
	RedisAddr     string `mapstructure:"redis_addr" validate:"required"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db" validate:"min=0,max=15"`
	Concurrency   int    `mapstructure:"concurrency" validate:"min=1"`
	Queues        map[string]int `mapstructure:"queues" validate:"required"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level" validate:"required,oneof=debug info warn error fatal panic"`
	Format     string `mapstructure:"format" validate:"required,oneof=json text"`
	Output     string `mapstructure:"output" validate:"required,oneof=stdout stderr file"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size" validate:"min=1"`
	MaxBackups int    `mapstructure:"max_backups" validate:"min=0"`
	MaxAge     int    `mapstructure:"max_age" validate:"min=1"`
	Compress   bool   `mapstructure:"compress"`
}

var (
	cfg *Config
)

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := validator.New()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	
	if configPath != "" {
		viper.AddConfigPath(configPath)
	} else {
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("../configs")
		viper.AddConfigPath("../../configs")
	}

	// 设置环境变量前缀
	viper.SetEnvPrefix("TASKMANAGE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置到结构体
	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证配置
	if err := v.Struct(config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	cfg = config
	return config, nil
}

// Get 获取全局配置实例
func Get() *Config {
	return cfg
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	switch c.Database.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			c.Database.Username,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.Database,
			c.Database.Charset,
		)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Database.Host,
			c.Database.Port,
			c.Database.Username,
			c.Database.Password,
			c.Database.Database,
		)
	default:
		return ""
	}
}

// GetRedisAddr 获取Redis连接地址
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetServerAddr 获取服务器监听地址
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsProduction 判断是否为生产环境
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDevelopment 判断是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsDebug 判断是否开启调试模式
func (c *Config) IsDebug() bool {
	return c.App.Debug
}
