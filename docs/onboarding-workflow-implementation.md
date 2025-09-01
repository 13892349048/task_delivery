# 员工入职工作流实现文档

## 概述

本文档详细描述了TaskManage项目中员工入职工作流系统的完整实现，包括数据模型、服务层、API接口和业务流程。

## 系统架构

### 1. 数据模型层

#### Employee模型扩展
```go
type Employee struct {
    // 原有字段...
    
    // 入职工作流相关字段
    OnboardingStatus   string     `gorm:"type:varchar(50);default:'pending_onboard'" json:"onboarding_status"`
    ExpectedDate       *time.Time `gorm:"type:date" json:"expected_date"`
    HireDate          *time.Time `gorm:"type:date" json:"hire_date"`
    ProbationEndDate  *time.Time `gorm:"type:date" json:"probation_end_date"`
    ConfirmDate       *time.Time `gorm:"type:date" json:"confirm_date"`
    OnboardingNotes   string     `gorm:"type:text" json:"onboarding_notes"`
    
    // 部门和职位改为可选（入职前可能未分配）
    DepartmentID      *uint      `gorm:"index" json:"department_id"`
    PositionID        *uint      `gorm:"index" json:"position_id"`
}
```

#### OnboardingHistory模型
```go
type OnboardingHistory struct {
    BaseModel
    EmployeeID   uint      `gorm:"not null;index" json:"employee_id"`
    FromStatus   string    `gorm:"type:varchar(50)" json:"from_status"`
    ToStatus     string    `gorm:"type:varchar(50);not null" json:"to_status"`
    OperatorID   uint      `gorm:"not null" json:"operator_id"`
    Reason       string    `gorm:"type:varchar(255)" json:"reason"`
    Notes        string    `gorm:"type:text" json:"notes"`
    
    // 关联关系
    Employee database.Employee `gorm:"foreignKey:EmployeeID" json:"employee"`
    Operator database.User     `gorm:"foreignKey:OperatorID" json:"operator"`
}
```

### 2. 入职状态定义

```go
const (
    OnboardingStatusPendingOnboard = "pending_onboard"  // 待入职
    OnboardingStatusOnboarding     = "onboarding"       // 入职中
    OnboardingStatusProbation      = "probation"        // 试用期
    OnboardingStatusActive         = "active"           // 正式员工
    OnboardingStatusTerminated     = "terminated"       // 已离职
)
```

## 业务流程

### 入职工作流状态转换图

```
pending_onboard → onboarding → probation → active
       ↓              ↓           ↓          ↓
   terminated    terminated  terminated  terminated
```

### 状态转换规则

1. **待入职 → 入职中**: HR创建员工档案后，部门经理确认入职
2. **入职中 → 试用期**: 完成入职手续，进入试用期
3. **试用期 → 正式员工**: 试用期结束，转为正式员工
4. **任意状态 → 已离职**: 管理员可以将员工状态设为离职

## API接口

### 1. 创建待入职员工
- **端点**: `POST /api/v1/onboarding/pending`
- **权限**: `employee:create`
- **请求体**:
```json
{
    "real_name": "张三",
    "email": "zhangsan@example.com",
    "phone": "13800138001",
    "expected_date": "2024-01-15",
    "notes": "新员工入职"
}
```

### 2. 确认入职
- **端点**: `POST /api/v1/onboarding/confirm`
- **权限**: `employee:update`
- **请求体**:
```json
{
    "employee_id": 1,
    "department_id": 1,
    "position_id": 1,
    "hire_date": "2024-01-15",
    "probation_end_date": "2024-04-15",
    "notes": "确认入职"
}
```

### 3. 完成试用期
- **端点**: `POST /api/v1/onboarding/{employee_id}/probation`
- **权限**: `employee:update`

### 4. 确认员工转正
- **端点**: `POST /api/v1/onboarding/confirm-employee`
- **权限**: `employee:update`
- **请求体**:
```json
{
    "employee_id": 1,
    "confirm_date": "2024-04-15",
    "notes": "转正确认"
}
```

### 5. 更改员工状态
- **端点**: `POST /api/v1/onboarding/change-status`
- **权限**: `employee:update`
- **请求体**:
```json
{
    "employee_id": 1,
    "new_status": "active",
    "reason": "管理员调整",
    "notes": "状态调整说明"
}
```

### 6. 获取入职工作流列表
- **端点**: `GET /api/v1/onboarding/workflows`
- **权限**: `employee:read`
- **查询参数**:
  - `page`: 页码
  - `page_size`: 每页数量
  - `status`: 状态过滤
  - `department`: 部门过滤
  - `date_from`: 开始日期
  - `date_to`: 结束日期

### 7. 获取入职历史记录
- **端点**: `GET /api/v1/onboarding/{employee_id}/history`
- **权限**: `employee:read`

## 服务层实现

### OnboardingService接口
```go
type OnboardingService interface {
    CreatePendingEmployee(ctx context.Context, req *CreatePendingEmployeeRequest) (*OnboardingWorkflowResponse, error)
    ConfirmOnboarding(ctx context.Context, req *OnboardConfirmRequest) (*OnboardingWorkflowResponse, error)
    CompleteProbation(ctx context.Context, employeeID uint, operatorID uint) (*OnboardingWorkflowResponse, error)
    ConfirmEmployee(ctx context.Context, req *ProbationToActiveRequest, operatorID uint) (*OnboardingWorkflowResponse, error)
    ChangeEmployeeStatus(ctx context.Context, req *EmployeeStatusChangeRequest, operatorID uint) (*OnboardingWorkflowResponse, error)
    GetOnboardingWorkflows(ctx context.Context, filter *OnboardingWorkflowFilter) ([]*OnboardingWorkflowResponse, error)
    GetOnboardingHistory(ctx context.Context, employeeID uint) ([]*OnboardingHistoryResponse, error)
}
```

### 核心业务逻辑

#### 1. 状态转换验证
每个状态转换都有严格的验证规则，确保业务流程的正确性。

#### 2. 历史记录追踪
所有状态变更都会记录到`OnboardingHistory`表中，包括：
- 变更前后状态
- 操作人员
- 变更原因和备注
- 变更时间

#### 3. 事务处理
所有涉及多表操作的业务都使用数据库事务，确保数据一致性。

## 数据库设计

### 表结构

#### employees表新增字段
```sql
ALTER TABLE employees 
ADD COLUMN onboarding_status VARCHAR(50) DEFAULT 'pending_onboard',
ADD COLUMN expected_date DATE,
ADD COLUMN hire_date DATE,
ADD COLUMN probation_end_date DATE,
ADD COLUMN confirm_date DATE,
ADD COLUMN onboarding_notes TEXT,
MODIFY COLUMN department_id INT UNSIGNED NULL,
MODIFY COLUMN position_id INT UNSIGNED NULL;
```

#### onboarding_histories表
```sql
CREATE TABLE onboarding_histories (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    employee_id INT UNSIGNED NOT NULL,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    operator_id INT UNSIGNED NOT NULL,
    reason VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_employee_id (employee_id),
    INDEX idx_operator_id (operator_id),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (operator_id) REFERENCES users(id) ON DELETE RESTRICT
);
```

## 权限控制

### RBAC权限配置
- `employee:create`: 创建员工（HR）
- `employee:read`: 查看员工信息（HR、部门经理）
- `employee:update`: 更新员工信息（HR、部门经理）
- `employee:delete`: 删除员工（仅管理员）

### 业务权限逻辑
1. **HR角色**: 可以创建待入职员工、查看所有入职流程
2. **部门经理**: 可以确认本部门员工入职、查看本部门入职流程
3. **管理员**: 拥有所有权限，可以执行任意状态变更

## 测试

### 单元测试
- Repository层测试
- Service层业务逻辑测试
- Handler层API测试

### 集成测试
使用提供的PowerShell脚本进行完整的API集成测试：
```bash
.\scripts\test_onboarding_workflow.ps1
```

## 部署说明

### 数据库迁移
1. 运行数据库迁移脚本更新表结构
2. 确保现有员工数据的`onboarding_status`字段有默认值

### 配置更新
无需额外配置更新，使用现有的数据库和服务配置。

### 监控和日志
- 所有API操作都有详细的日志记录
- 状态转换操作会记录操作人员和时间
- 建议监控入职流程的完成率和平均时长

## 扩展功能

### 未来可扩展的功能
1. **入职任务清单**: 为每个入职阶段定义具体任务
2. **自动化通知**: 状态变更时自动发送邮件通知
3. **入职报表**: 统计入职效率和成功率
4. **批量操作**: 支持批量处理入职流程
5. **工作流引擎集成**: 与现有的审批工作流集成

## 故障排除

### 常见问题
1. **状态转换失败**: 检查当前状态是否允许转换
2. **权限不足**: 确认用户角色和权限配置
3. **数据不一致**: 检查事务处理和外键约束

### 日志分析
查看应用日志中的入职工作流相关操作，关键字：
- `onboarding`
- `employee status`
- `status transition`
