# 开发规范文档

## 代码规范

### Go 代码规范

#### 项目结构
```
taskmanage/
├── cmd/                    # 应用入口
│   ├── api/               # API 服务入口
│   │   └── main.go
│   ├── worker/            # Worker 服务入口
│   │   └── main.go
│   └── migrate/           # 数据库迁移工具
│       └── main.go
├── internal/              # 内部包 (不对外暴露)
│   ├── api/              # API 层
│   │   ├── handler/      # HTTP 处理器
│   │   ├── middleware/   # 中间件
│   │   └── router/       # 路由配置
│   ├── service/          # 业务逻辑层
│   │   ├── task/         # 任务服务
│   │   ├── staff/        # 员工服务
│   │   └── assignment/   # 分配服务
│   ├── repository/       # 数据访问层
│   │   ├── postgres/     # PostgreSQL 实现
│   │   └── redis/        # Redis 实现
│   ├── model/            # 数据模型
│   │   ├── entity/       # 实体模型
│   │   ├── dto/          # 数据传输对象
│   │   └── vo/           # 值对象
│   ├── config/           # 配置管理
│   └── worker/           # Worker 实现
├── pkg/                   # 公共包 (可对外暴露)
│   ├── database/         # 数据库连接
│   ├── queue/            # 队列封装
│   ├── logger/           # 日志
│   ├── validator/        # 验证器
│   └── utils/            # 工具函数
├── tests/                # 测试文件
│   ├── unit/             # 单元测试
│   ├── integration/      # 集成测试
│   └── e2e/              # 端到端测试
├── docs/                 # 文档
├── scripts/              # 脚本
├── deployments/          # 部署配置
└── migrations/           # 数据库迁移文件
```

#### 命名规范

**包名**:
- 使用小写字母
- 简短且有意义
- 避免下划线和驼峰

```go
// 正确
package task
package assignment

// 错误
package taskService
package task_service
```

**变量和函数名**:
- 使用驼峰命名法
- 私有成员小写开头，公有成员大写开头
- 布尔变量使用 is/has/can 前缀

```go
// 正确
var userName string
var isActive bool
func GetUserByID(id string) (*User, error)
func (s *service) createTask(task *Task) error

// 错误
var user_name string
var active bool
func getUserById(id string) (*User, error)
```

**常量**:
- 使用全大写字母和下划线
- 或使用驼峰命名法分组

```go
// 单个常量
const MAX_RETRY_COUNT = 5

// 常量组
const (
    StatusDraft      = "draft"
    StatusApproved   = "approved"
    StatusInProgress = "in_progress"
)
```

#### 错误处理

**自定义错误类型**:
```go
// 定义错误类型
type TaskError struct {
    Code    string
    Message string
    Cause   error
}

func (e *TaskError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// 错误包装
func (s *service) GetTask(id string) (*Task, error) {
    task, err := s.repo.GetByID(id)
    if err != nil {
        return nil, &TaskError{
            Code:    "TASK_NOT_FOUND",
            Message: "任务不存在",
            Cause:   err,
        }
    }
    return task, nil
}
```

**错误处理模式**:
```go
// 立即返回错误
func processTask(task *Task) error {
    if err := validateTask(task); err != nil {
        return fmt.Errorf("validate task failed: %w", err)
    }
    
    if err := saveTask(task); err != nil {
        return fmt.Errorf("save task failed: %w", err)
    }
    
    return nil
}

// 收集多个错误
func validateTaskBatch(tasks []*Task) error {
    var errs []error
    for i, task := range tasks {
        if err := validateTask(task); err != nil {
            errs = append(errs, fmt.Errorf("task[%d]: %w", i, err))
        }
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("validation failed: %v", errs)
    }
    return nil
}
```

#### 日志规范

```go
import (
    "github.com/sirupsen/logrus"
)

// 结构化日志
logger.WithFields(logrus.Fields{
    "task_id":    task.ID,
    "user_id":    userID,
    "action":     "create_task",
    "duration":   time.Since(start),
}).Info("任务创建成功")

// 错误日志
logger.WithFields(logrus.Fields{
    "task_id": task.ID,
    "error":   err.Error(),
}).Error("任务创建失败")

// 性能日志
defer func(start time.Time) {
    logger.WithFields(logrus.Fields{
        "method":   "CreateTask",
        "duration": time.Since(start),
    }).Debug("方法执行时间")
}(time.Now())
```

#### 测试规范

**单元测试**:
```go
func TestTaskService_CreateTask(t *testing.T) {
    tests := []struct {
        name    string
        task    *Task
        want    *Task
        wantErr bool
    }{
        {
            name: "成功创建任务",
            task: &Task{
                Title:       "测试任务",
                Description: "测试描述",
                Priority:    PriorityHigh,
            },
            want: &Task{
                ID:          "task_001",
                Title:       "测试任务",
                Description: "测试描述",
                Priority:    PriorityHigh,
                Status:      StatusDraft,
            },
            wantErr: false,
        },
        {
            name: "标题为空应该失败",
            task: &Task{
                Title:    "",
                Priority: PriorityHigh,
            },
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := NewTaskService(mockRepo, mockQueue)
            got, err := s.CreateTask(tt.task)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("CreateTask() got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**集成测试**:
```go
func TestTaskAPI_Integration(t *testing.T) {
    // 设置测试数据库
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // 创建测试服务器
    server := setupTestServer(t, db)
    defer server.Close()
    
    // 测试创建任务
    task := &Task{
        Title:    "集成测试任务",
        Priority: PriorityHigh,
    }
    
    resp, err := http.Post(server.URL+"/api/v1/tasks", "application/json", 
        strings.NewReader(toJSON(task)))
    require.NoError(t, err)
    require.Equal(t, http.StatusCreated, resp.StatusCode)
    
    // 验证响应
    var result TaskResponse
    err = json.NewDecoder(resp.Body).Decode(&result)
    require.NoError(t, err)
    assert.Equal(t, task.Title, result.Data.Title)
}
```

## API 设计规范

### RESTful API 设计

**URL 设计**:
```
# 资源命名使用复数名词
GET    /api/v1/tasks           # 获取任务列表
POST   /api/v1/tasks           # 创建任务
GET    /api/v1/tasks/{id}      # 获取特定任务
PUT    /api/v1/tasks/{id}      # 更新任务
DELETE /api/v1/tasks/{id}      # 删除任务

# 子资源
GET    /api/v1/tasks/{id}/assignments    # 获取任务的分配记录
POST   /api/v1/tasks/{id}/assignments    # 创建任务分配

# 动作使用动词
POST   /api/v1/tasks/{id}/assign         # 分配任务
POST   /api/v1/tasks/{id}/complete       # 完成任务
POST   /api/v1/tasks/{id}/cancel         # 取消任务
```

**HTTP 状态码**:
```
200 OK           - 成功获取资源
201 Created      - 成功创建资源
204 No Content   - 成功删除资源
400 Bad Request  - 请求参数错误
401 Unauthorized - 未认证
403 Forbidden    - 权限不足
404 Not Found    - 资源不存在
409 Conflict     - 资源冲突
422 Unprocessable Entity - 业务逻辑错误
500 Internal Server Error - 服务器内部错误
```

**请求/响应格式**:
```go
// 统一响应格式
type APIResponse struct {
    Code      int         `json:"code"`
    Message   string      `json:"message"`
    Data      interface{} `json:"data,omitempty"`
    Error     string      `json:"error,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}

// 分页响应
type PaginatedResponse struct {
    Items      interface{} `json:"items"`
    Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
    Page  int `json:"page"`
    Size  int `json:"size"`
    Total int `json:"total"`
    Pages int `json:"pages"`
}
```

### 参数验证

```go
import "github.com/go-playground/validator/v10"

type CreateTaskRequest struct {
    Title          string   `json:"title" validate:"required,min=1,max=500"`
    Description    string   `json:"description" validate:"max=2000"`
    Priority       string   `json:"priority" validate:"required,oneof=high medium low"`
    EstimatedHours int      `json:"estimated_hours" validate:"min=1,max=1000"`
    RequiredSkills []string `json:"required_skills" validate:"dive,min=1,max=50"`
    Deadline       string   `json:"deadline" validate:"datetime=2006-01-02T15:04:05Z07:00"`
}

// 自定义验证器
func validateDeadline(fl validator.FieldLevel) bool {
    deadline, err := time.Parse(time.RFC3339, fl.Field().String())
    if err != nil {
        return false
    }
    return deadline.After(time.Now())
}
```

## 数据库规范

### 迁移文件命名
```
migrations/
├── 001_create_users_table.up.sql
├── 001_create_users_table.down.sql
├── 002_create_tasks_table.up.sql
├── 002_create_tasks_table.down.sql
└── 003_add_task_indexes.up.sql
```

### SQL 编写规范
```sql
-- 使用大写关键字
SELECT t.id, t.title, s.name as assignee_name
FROM tasks t
LEFT JOIN staff s ON t.assignee_id = s.id
WHERE t.status = 'in_progress'
  AND t.deadline > NOW()
ORDER BY t.priority DESC, t.created_at ASC;

-- 复杂查询使用 CTE
WITH task_stats AS (
    SELECT 
        assignee_id,
        COUNT(*) as total_tasks,
        COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tasks
    FROM tasks
    WHERE created_at >= DATE_TRUNC('month', NOW())
    GROUP BY assignee_id
)
SELECT 
    s.name,
    ts.total_tasks,
    ts.completed_tasks,
    ROUND(ts.completed_tasks::DECIMAL / ts.total_tasks * 100, 2) as completion_rate
FROM staff s
JOIN task_stats ts ON s.id = ts.assignee_id
ORDER BY completion_rate DESC;
```

## Git 工作流

### 分支策略
```
main          # 主分支，生产环境代码
├── develop   # 开发分支
├── feature/  # 功能分支
│   ├── feature/task-assignment
│   └── feature/notification-system
├── hotfix/   # 热修复分支
│   └── hotfix/fix-assignment-bug
└── release/  # 发布分支
    └── release/v1.0.0
```

### 提交信息规范
```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型 (type)**:
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式化
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

**示例**:
```
feat(task): 添加任务自动分配功能

- 实现基于技能匹配的分配算法
- 支持负载均衡分配策略
- 添加分配历史记录

Closes #123
```

### Pull Request 规范
```markdown
## 变更描述
简要描述本次变更的内容和目的

## 变更类型
- [ ] 新功能
- [ ] Bug修复
- [ ] 文档更新
- [ ] 重构
- [ ] 性能优化

## 测试
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 手动测试完成

## 检查清单
- [ ] 代码符合规范
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] 无破坏性变更

## 相关Issue
Closes #123
```

## 性能优化规范

### 数据库优化
```go
// 使用批量操作
func (r *repository) CreateTasksBatch(tasks []*Task) error {
    tx := r.db.Begin()
    defer tx.Rollback()
    
    for _, task := range tasks {
        if err := tx.Create(task).Error; err != nil {
            return err
        }
    }
    
    return tx.Commit().Error
}

// 使用预加载避免 N+1 查询
func (r *repository) GetTasksWithAssignee() ([]*Task, error) {
    var tasks []*Task
    err := r.db.Preload("Assignee").Find(&tasks).Error
    return tasks, err
}

// 使用索引优化查询
func (r *repository) GetTasksByStatus(status string, limit, offset int) ([]*Task, error) {
    var tasks []*Task
    err := r.db.Where("status = ?", status).
        Order("priority DESC, created_at ASC").
        Limit(limit).Offset(offset).
        Find(&tasks).Error
    return tasks, err
}
```

### 缓存策略
```go
// Redis 缓存
func (s *service) GetTask(id string) (*Task, error) {
    // 先从缓存获取
    cacheKey := fmt.Sprintf("task:%s", id)
    if cached, err := s.cache.Get(cacheKey); err == nil {
        var task Task
        if err := json.Unmarshal([]byte(cached), &task); err == nil {
            return &task, nil
        }
    }
    
    // 缓存未命中，从数据库获取
    task, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    if data, err := json.Marshal(task); err == nil {
        s.cache.Set(cacheKey, string(data), 5*time.Minute)
    }
    
    return task, nil
}
```

## 安全规范

### 输入验证
```go
import "html"

// 防止 XSS
func sanitizeInput(input string) string {
    return html.EscapeString(strings.TrimSpace(input))
}

// SQL 注入防护 (使用参数化查询)
func (r *repository) GetTasksByTitle(title string) ([]*Task, error) {
    var tasks []*Task
    // 正确：使用参数化查询
    err := r.db.Where("title LIKE ?", "%"+title+"%").Find(&tasks).Error
    return tasks, err
}
```

### 认证授权
```go
// JWT 中间件
func JWTMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c.GetHeader("Authorization"))
        if token == "" {
            c.JSON(401, gin.H{"error": "Missing token"})
            c.Abort()
            return
        }
        
        claims, err := validateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Set("role", claims.Role)
        c.Next()
    }
}

// 权限检查
func RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("role")
        if userRole != role && userRole != "admin" {
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 监控和日志

### 指标收集
```go
import "github.com/prometheus/client_golang/prometheus"

var (
    taskCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "tasks_total",
            Help: "Total number of tasks",
        },
        []string{"status", "priority"},
    )
    
    taskDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "task_duration_seconds",
            Help: "Task processing duration",
        },
        []string{"operation"},
    )
)

func (s *service) CreateTask(task *Task) (*Task, error) {
    timer := prometheus.NewTimer(taskDuration.WithLabelValues("create"))
    defer timer.ObserveDuration()
    
    result, err := s.repo.Create(task)
    if err != nil {
        return nil, err
    }
    
    taskCounter.WithLabelValues(result.Status, result.Priority).Inc()
    return result, nil
}
```

### 链路追踪
```go
import "go.opentelemetry.io/otel"

func (s *service) ProcessTask(ctx context.Context, taskID string) error {
    ctx, span := otel.Tracer("task-service").Start(ctx, "ProcessTask")
    defer span.End()
    
    span.SetAttributes(attribute.String("task.id", taskID))
    
    task, err := s.GetTask(ctx, taskID)
    if err != nil {
        span.RecordError(err)
        return err
    }
    
    // 处理任务逻辑...
    
    return nil
}
```
