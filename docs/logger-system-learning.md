# TaskManage 日志系统学习指南

## 日志系统架构概览

TaskManage 采用基于 Logrus 的分层日志系统，支持多种输出格式、日志轮转和钩子扩展。

### 核心组件
- **日志核心**：`pkg/logger/logger.go` - 日志系统初始化和基础功能
- **钩子系统**：`pkg/logger/hooks.go` - 上下文信息和错误堆栈增强
- **中间件**：`internal/api/middleware/logger.go` - HTTP请求日志中间件
- **应用入口**：`cmd/taskmanage/main.go` - 日志系统启动和配置

## 日志初始化流程详解

### 1. 配置加载阶段
```go
// main.go 中的初始化顺序
cfg, err := config.LoadByEnvironment(*env)  // 加载配置
logger.Init(cfg)                           // 初始化日志系统
logger.AddHooks(logger.GetLogger(), ...)   // 添加钩子
```

### 2. 日志系统初始化过程

#### 步骤1：创建Logger实例
```go
func Init(cfg *config.Config) error {
    Logger = logrus.New()  // 创建新的logrus实例
    // ...
}
```

#### 步骤2：设置日志级别
```go
level, err := logrus.ParseLevel(cfg.Log.Level)
Logger.SetLevel(level)
```
支持的级别：`debug`, `info`, `warn`, `error`, `fatal`, `panic`

#### 步骤3：配置日志格式
```go
switch strings.ToLower(cfg.Log.Format) {
case "json":
    Logger.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: "2006-01-02 15:04:05",
        FieldMap: logrus.FieldMap{
            logrus.FieldKeyTime:  "timestamp",
            logrus.FieldKeyLevel: "level",
            logrus.FieldKeyMsg:   "message",
            logrus.FieldKeyFunc:  "caller",
        },
    })
case "text":
    Logger.SetFormatter(&logrus.TextFormatter{
        FullTimestamp:   true,
        TimestampFormat: "2006-01-02 15:04:05",
        ForceColors:     true,
    })
}
```

#### 步骤4：设置输出目标
```go
func getLogOutput(cfg *config.Config) (io.Writer, error) {
    switch strings.ToLower(cfg.Log.Output) {
    case "stdout":
        return os.Stdout, nil
    case "stderr":
        return os.Stderr, nil
    case "file":
        // 使用lumberjack进行日志轮转
        return &lumberjack.Logger{
            Filename:   cfg.Log.Filename,
            MaxSize:    cfg.Log.MaxSize,    // MB
            MaxBackups: cfg.Log.MaxBackups,
            MaxAge:     cfg.Log.MaxAge,     // days
            Compress:   cfg.Log.Compress,
        }, nil
    }
}
```

#### 步骤5：启用调用者信息
```go
Logger.SetReportCaller(true)  // 显示调用文件和行号
```

### 3. 钩子系统初始化
```go
func AddHooks(logger *logrus.Logger, appName, version, environment string) {
    logger.AddHook(NewContextHook(appName, version, environment))  // 上下文信息
    logger.AddHook(NewErrorStackHook())                           // 错误堆栈
    logger.AddHook(NewRequestIDHook())                            // 请求ID
}
```

## 钩子系统详解

### 1. 上下文钩子（ContextHook）
**作用**：为所有日志条目添加应用基础信息
```go
func (hook *ContextHook) Fire(entry *logrus.Entry) error {
    entry.Data["app"] = hook.AppName           // 应用名称
    entry.Data["version"] = hook.Version       // 版本号
    entry.Data["environment"] = hook.Environment // 环境
    return nil
}
```

### 2. 调用者钩子（CallerHook）
**作用**：添加代码调用位置信息
```go
func (hook *CallerHook) Fire(entry *logrus.Entry) error {
    if pc, file, line, ok := runtime.Caller(hook.Skip); ok {
        funcName := runtime.FuncForPC(pc).Name()
        entry.Data[hook.Field] = fmt.Sprintf("%s:%d %s()", file, line, funcName)
    }
    return nil
}
```

### 3. 错误堆栈钩子（ErrorStackHook）
**作用**：为错误级别日志添加详细堆栈信息
```go
func (hook *ErrorStackHook) Fire(entry *logrus.Entry) error {
    if err, ok := entry.Data[logrus.ErrorKey]; ok {
        if e, ok := err.(error); ok {
            entry.Data["error_stack"] = fmt.Sprintf("%+v", e)
        }
    }
    return nil
}
```

### 4. 请求ID钩子（RequestIDHook）
**作用**：为HTTP请求相关日志添加请求追踪ID
```go
func (hook *RequestIDHook) Fire(entry *logrus.Entry) error {
    if _, exists := entry.Data["request_id"]; !exists {
        // 从上下文中获取request_id
    }
    return nil
}
```

## 日志配置规则

### 配置文件结构
```yaml
log:
  level: "debug"              # 日志级别
  format: "json"              # 输出格式：json/text
  output: "stdout"            # 输出目标：stdout/stderr/file
  filename: "logs/app.log"    # 文件输出时的文件名
  max_size: 100               # 单文件最大大小(MB)
  max_backups: 5              # 保留的备份文件数
  max_age: 30                 # 文件保留天数
  compress: true              # 是否压缩备份文件
```

### 日志级别说明
- **debug**：调试信息，开发环境使用
- **info**：一般信息，记录程序运行状态
- **warn**：警告信息，需要注意但不影响运行
- **error**：错误信息，程序出现错误但可以继续运行
- **fatal**：致命错误，程序无法继续运行
- **panic**：恐慌级别，程序崩溃

### 输出格式对比

#### JSON格式（推荐生产环境）
```json
{
  "timestamp": "2024-08-28 19:30:15",
  "level": "info",
  "message": "HTTP Request Completed",
  "app": "TaskManage",
  "version": "1.0.0",
  "environment": "production",
  "status_code": 200,
  "method": "GET",
  "path": "/api/users",
  "latency": "15.2ms"
}
```

#### Text格式（适合开发环境）
```
INFO[2024-08-28 19:30:15] HTTP Request Completed app=TaskManage environment=development method=GET path=/api/users status_code=200
```

## 中间件系统分析

### 当前中间件状况

**问题发现**：存在中间件重复定义

1. **`pkg/logger/middleware.go`** - 功能完整但**未被使用**
   - `GinLogger()` - 基于Gin原生日志格式化器
   - `GinRecovery()` - Panic恢复中间件
   - `RequestLogger()` - 详细请求日志中间件
   - `DatabaseLogger` - 数据库日志记录器

2. **`internal/api/middleware/logger.go`** - **实际在使用**的简化版本
   - `Logger()` - 简化的HTTP请求日志中间件

### 中间件功能对比

#### pkg/logger/middleware.go（未使用）
```go
// 功能更完整，包含：
- 请求体记录（小于1KB的POST/PUT/PATCH请求）
- 响应大小记录
- 错误信息收集
- 慢请求检测（>1秒）
- 多级别日志（根据状态码和延迟）
```

#### internal/api/middleware/logger.go（在使用）
```go
// 功能简化，包含：
- 基础请求信息记录
- 状态码分级日志（>=400为错误，其他为信息）
- 延迟时间记录
```

### 建议的中间件整合方案

**方案1：使用pkg版本（推荐）**
- 删除 `internal/api/middleware/logger.go`
- 在路由中使用 `logger.RequestLogger()`
- 获得更完整的日志功能

**方案2：保持现状**
- 保留简化版本，满足基本需求
- 删除未使用的 `pkg/logger/middleware.go`

## 日志使用方法

### 1. 基础日志记录
```go
import "taskmanage/pkg/logger"

// 直接使用全局方法
logger.Info("用户登录成功")
logger.Errorf("数据库连接失败: %v", err)
logger.WithField("user_id", 123).Info("用户操作")
```

### 2. 结构化日志
```go
logger.WithFields(logrus.Fields{
    "user_id": 123,
    "action": "login",
    "ip": "192.168.1.1",
}).Info("用户操作记录")
```

### 3. 业务日志记录器
```go
// 创建业务模块专用日志记录器
businessLogger := logger.NewBusinessLogger("user_service")
businessLogger.Info("用户注册成功")

// 创建系统组件日志记录器
systemLogger := logger.NewSystemLogger("database")
systemLogger.Error("连接池耗尽")
```

### 4. HTTP请求日志记录器
```go
// 创建带请求ID的日志记录器
requestLogger := logger.NewHTTPRequestLogger("req-123456")
requestLogger.Info("处理用户请求")
```

## 自己搭建日志系统步骤

### 步骤1：安装依赖
```bash
go get github.com/sirupsen/logrus
go get gopkg.in/natefinch/lumberjack.v2
```

### 步骤2：创建日志配置结构
```go
type LogConfig struct {
    Level      string `yaml:"level"`
    Format     string `yaml:"format"`
    Output     string `yaml:"output"`
    Filename   string `yaml:"filename"`
    MaxSize    int    `yaml:"max_size"`
    MaxBackups int    `yaml:"max_backups"`
    MaxAge     int    `yaml:"max_age"`
    Compress   bool   `yaml:"compress"`
}
```

### 步骤3：实现日志初始化
```go
package logger

import (
    "github.com/sirupsen/logrus"
    "gopkg.in/natefinch/lumberjack.v2"
)

var Logger *logrus.Logger

func Init(cfg *LogConfig) error {
    Logger = logrus.New()
    
    // 设置级别
    level, err := logrus.ParseLevel(cfg.Level)
    if err != nil {
        return err
    }
    Logger.SetLevel(level)
    
    // 设置格式
    if cfg.Format == "json" {
        Logger.SetFormatter(&logrus.JSONFormatter{})
    } else {
        Logger.SetFormatter(&logrus.TextFormatter{})
    }
    
    // 设置输出
    if cfg.Output == "file" {
        Logger.SetOutput(&lumberjack.Logger{
            Filename:   cfg.Filename,
            MaxSize:    cfg.MaxSize,
            MaxBackups: cfg.MaxBackups,
            MaxAge:     cfg.MaxAge,
            Compress:   cfg.Compress,
        })
    }
    
    Logger.SetReportCaller(true)
    return nil
}

func GetLogger() *logrus.Logger {
    return Logger
}
```

### 步骤4：创建钩子
```go
type ContextHook struct {
    AppName string
}

func (h *ContextHook) Levels() []logrus.Level {
    return logrus.AllLevels
}

func (h *ContextHook) Fire(entry *logrus.Entry) error {
    entry.Data["app"] = h.AppName
    return nil
}

func AddHooks(logger *logrus.Logger, appName string) {
    logger.AddHook(&ContextHook{AppName: appName})
}
```

### 步骤5：创建中间件
```go
func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        
        GetLogger().WithFields(logrus.Fields{
            "method":     c.Request.Method,
            "path":       c.Request.URL.Path,
            "status":     c.Writer.Status(),
            "latency":    time.Since(start),
            "client_ip":  c.ClientIP(),
        }).Info("HTTP Request")
    }
}
```

### 步骤6：在main中使用
```go
func main() {
    // 初始化日志
    cfg := &LogConfig{
        Level:  "info",
        Format: "json",
        Output: "stdout",
    }
    
    if err := logger.Init(cfg); err != nil {
        log.Fatal("日志初始化失败:", err)
    }
    
    // 添加钩子
    logger.AddHooks(logger.GetLogger(), "MyApp")
    
    // 使用日志
    logger.GetLogger().Info("应用启动")
    
    // 在Gin中使用中间件
    r := gin.New()
    r.Use(LoggerMiddleware())
}
```

## 最佳实践总结

### 1. 日志级别使用规范
- **Debug**：详细的调试信息，仅开发环境
- **Info**：重要的业务流程节点
- **Warn**：需要注意的异常情况
- **Error**：错误但程序可继续运行
- **Fatal**：致命错误，程序退出

### 2. 结构化日志设计
- 使用一致的字段名称
- 避免敏感信息记录
- 合理控制日志大小
- 使用请求ID进行链路追踪

### 3. 性能优化
- 生产环境使用Info级别以上
- 合理设置日志轮转参数
- 避免在高频路径记录Debug日志
- 使用异步日志写入（可选）

### 4. 监控和告警
- 错误日志实时告警
- 日志量异常监控
- 慢请求日志分析
- 日志文件大小监控

## 常见问题解决

### 问题1：日志文件权限错误
**解决**：确保应用有写入日志目录的权限

### 问题2：日志轮转不生效
**解决**：检查MaxSize、MaxAge等配置参数

### 问题3：JSON格式日志不易读
**解决**：开发环境使用text格式，生产环境使用json格式

### 问题4：日志性能影响
**解决**：调整日志级别，使用异步写入，避免频繁的磁盘IO

### 问题5：中间件重复问题
**解决**：统一使用一套中间件，删除冗余代码

这个日志系统设计完整、功能强大，你可以根据这个指南理解其工作原理，并在其他项目中应用类似的设计模式。
