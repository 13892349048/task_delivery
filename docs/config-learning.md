# TaskManage 配置系统学习指南

## 配置系统架构概览

TaskManage 采用分层配置管理系统，支持多环境配置和环境变量覆盖。

### 核心组件
- **配置文件**：`configs/config.yaml` 及环境特定配置
- **配置结构**：`internal/config/config.go` - 定义配置数据结构
- **配置加载器**：`internal/config/loader.go` - 高级配置加载逻辑
- **应用入口**：`cmd/taskmanage/main.go` - 配置初始化和使用

## 配置加载流程详解

### 1. 启动参数解析
```go
// main.go 中的命令行参数
configPath = flag.String("config", "", "配置文件路径")
env        = flag.String("env", "development", "运行环境")
showConfig = flag.Bool("show-config", false, "显示配置信息")
```

### 2. 配置加载策略
```go
// 优先级：指定路径 > 环境配置
if *configPath != "" {
    cfg, err = config.Load(*configPath)
} else {
    cfg, err = config.LoadByEnvironment(*env)
}
```

### 3. 环境配置映射
在 `loader.go` 中定义了环境名称到配置文件的映射：
```go
envMapping := map[string]string{
    "development": "config",
    "dev":         "config.dev",
    "testing":     "config.test", 
    "test":        "config.test",
    "production":  "config.prod",
    "prod":        "config.prod",
    "prod-test":   "config.prod.test",
}
```

### 4. 配置文件搜索路径
系统按以下顺序搜索配置文件：
```go
// 当前目录
"."
"./configs"

// 上级目录
"../configs"
"../../configs"

// 系统目录
"/etc/taskmanage"
"$HOME/.taskmanage"

// 可执行文件目录
execDir
execDir/configs
```

### 5. 环境变量处理
支持两种环境变量覆盖方式：

#### 自动环境变量（推荐）
```go
// 设置前缀和替换规则
viper.SetEnvPrefix("TASKMANAGE")
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
viper.AutomaticEnv()
```

#### 手动绑定关键变量
```go
envBindings := map[string]string{
    "app.environment":    "APP_ENVIRONMENT",
    "server.port":        "SERVER_PORT", 
    "database.host":      "DB_HOST",
    "database.password":  "DB_PASSWORD",
    "jwt.secret":         "JWT_SECRET",
    // ... 更多绑定
}
```

### 6. 默认值设置
系统为所有配置项设置了合理的默认值：
```go
l.viper.SetDefault("app.name", "TaskManage")
l.viper.SetDefault("server.port", 8080)
l.viper.SetDefault("database.driver", "mysql")
// ... 更多默认值
```

### 7. 配置验证
使用 `validator` 包进行配置验证：
```go
type JWTConfig struct {
    Secret          string `validate:"required,min=32"`
    AccessTokenTTL  int    `validate:"required,min=1"`
    // ... 更多验证规则
}
```

## 配置结构详解

### 主配置结构
```go
type Config struct {
    App      AppConfig      `mapstructure:"app"`
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    Asynq    AsynqConfig    `mapstructure:"asynq"`
    Log      LogConfig      `mapstructure:"log"`
}
```

### 配置分类说明
- **App**：应用基础信息（名称、版本、环境）
- **Server**：HTTP服务器配置（端口、超时）
- **Database**：数据库连接配置
- **Redis**：缓存配置
- **JWT**：认证令牌配置
- **Asynq**：异步任务队列配置
- **Log**：日志系统配置

## 实用工具方法

### 配置访问方法
```go
// 获取全局配置
cfg := config.Get()

// 获取连接字符串
dsn := cfg.GetDSN()
redisAddr := cfg.GetRedisAddr()
serverAddr := cfg.GetServerAddr()

// 环境判断
if cfg.IsProduction() { /* 生产环境逻辑 */ }
if cfg.IsDevelopment() { /* 开发环境逻辑 */ }
```

### 配置监听
```go
// 监听配置文件变化
loader.WatchConfig(func(newConfig *Config) {
    // 配置更新回调
})
```

## 自己搭建配置系统步骤

### 步骤1：创建配置结构
```go
// 定义你的配置结构
type MyConfig struct {
    App    MyAppConfig    `mapstructure:"app"`
    Server MyServerConfig `mapstructure:"server"`
}

type MyAppConfig struct {
    Name string `mapstructure:"name" validate:"required"`
    Port int    `mapstructure:"port" validate:"required,min=1,max=65535"`
}
```

### 步骤2：创建配置文件
```yaml
# config.yaml
app:
  name: "MyApp"
  port: 8080

server:
  host: "localhost"
  timeout: 30
```

### 步骤3：实现加载器
```go
func LoadConfig(configPath string) (*MyConfig, error) {
    v := viper.New()
    
    // 设置配置文件
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath(configPath)
    
    // 环境变量支持
    v.SetEnvPrefix("MYAPP")
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    // 读取配置
    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }
    
    // 解析到结构体
    config := &MyConfig{}
    if err := v.Unmarshal(config); err != nil {
        return nil, err
    }
    
    // 验证配置
    validator := validator.New()
    if err := validator.Struct(config); err != nil {
        return nil, err
    }
    
    return config, nil
}
```

### 步骤4：在main中使用
```go
func main() {
    // 加载配置
    cfg, err := LoadConfig("./configs")
    if err != nil {
        log.Fatal("配置加载失败:", err)
    }
    
    // 使用配置
    fmt.Printf("应用名称: %s\n", cfg.App.Name)
    fmt.Printf("监听端口: %d\n", cfg.App.Port)
}
```

## 环境变量使用示例

### Windows PowerShell
```powershell
# 设置环境变量
$env:TASKMANAGE_APP_ENVIRONMENT="production"
$env:TASKMANAGE_SERVER_PORT="9090"
$env:TASKMANAGE_DATABASE_PASSWORD="secure_password"
$env:TASKMANAGE_JWT_SECRET="your-super-secret-jwt-key-32-chars"

# 运行应用
.\taskmanage.exe --env production
```

### Linux/macOS
```bash
# 设置环境变量
export TASKMANAGE_APP_ENVIRONMENT=production
export TASKMANAGE_SERVER_PORT=9090
export TASKMANAGE_DATABASE_PASSWORD=secure_password
export TASKMANAGE_JWT_SECRET=your-super-secret-jwt-key-32-chars

# 运行应用
./taskmanage --env production
```

## 最佳实践总结

### 1. 配置分层
- **默认值**：代码中设置合理默认值
- **配置文件**：环境特定的配置文件
- **环境变量**：敏感信息和部署特定配置

### 2. 安全考虑
- 敏感信息（密码、密钥）使用环境变量
- 配置文件不包含生产环境敏感数据
- 使用配置验证确保安全性

### 3. 环境管理
- **开发环境**：使用默认配置文件，便于开发调试
- **测试环境**：使用测试专用配置，模拟生产环境
- **生产环境**：依赖环境变量，确保安全性

### 4. 配置验证
- 使用结构体标签进行配置验证
- 启动时验证所有必需配置
- 提供清晰的错误信息

### 5. 配置监听
- 支持配置热重载（可选）
- 配置变更通知机制
- 优雅处理配置更新

## 常见问题解决

### 问题1：配置文件未找到
**解决**：检查配置文件路径和搜索路径设置

### 问题2：环境变量未生效
**解决**：确认环境变量前缀和键名映射规则

### 问题3：配置验证失败
**解决**：检查配置值是否符合验证规则

### 问题4：JWT密钥长度不足
**解决**：确保JWT密钥至少32字符长度
