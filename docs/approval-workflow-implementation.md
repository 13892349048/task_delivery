# 审批流程引擎实现文档

## 概述

本文档描述了任务管理系统中审批流程引擎的完整实现，包括核心组件、功能特性、集成方式和使用指南。

## 系统架构

### 核心组件

1. **工作流定义管理器 (WorkflowDefinitionManager)**
   - 创建、更新、验证工作流定义
   - 支持版本管理和激活状态控制
   - 提供工作流定义的CRUD操作

2. **工作流执行引擎 (WorkflowEngine)**
   - 启动工作流实例
   - 处理审批动作
   - 管理工作流状态转换
   - 执行节点逻辑

3. **节点执行器 (Node Executors)**
   - 开始节点执行器：初始化工作流
   - 结束节点执行器：完成工作流
   - 审批节点执行器：处理审批逻辑
   - 条件节点执行器：条件判断和分支
   - 并行节点执行器：并行分支处理
   - 汇聚节点执行器：并行分支合并
   - 脚本节点执行器：自定义脚本执行
   - 通知节点执行器：发送通知

4. **工作流服务 (WorkflowService)**
   - 任务分配审批流程管理
   - 与任务服务的集成接口
   - 审批流程的业务逻辑封装

5. **审计服务 (AuditService)**
   - 工作流执行历史记录
   - 用户审批历史统计
   - 审批趋势分析

6. **通知服务 (NotificationService)**
   - 审批通知发送
   - 提醒调度管理
   - 多渠道通知支持

## 功能特性

### 工作流定义

- **节点类型支持**：
  - 开始节点 (start)
  - 结束节点 (end)
  - 审批节点 (approval)
  - 条件节点 (condition)
  - 并行节点 (parallel)
  - 汇聚节点 (join)
  - 脚本节点 (script)
  - 通知节点 (notification)

- **审批配置**：
  - 指派人类型：用户、角色、部门、上级、发起人、变量、脚本
  - 审批类型：任意一人、所有人、多数人
  - 超时处理：自动通过、自动拒绝、升级处理

- **条件判断**：
  - 支持基本比较运算符 (==, !=, >, <, >=, <=)
  - 变量引用和表达式计算
  - 多条件组合和优先级

### 工作流执行

- **实例管理**：
  - 工作流实例创建和状态跟踪
  - 当前活动节点管理
  - 实例变量存储和更新

- **审批处理**：
  - 审批动作：同意、拒绝、退回、委托
  - 审批意见和附件支持
  - 审批历史记录

- **状态管理**：
  - 实例状态：运行中、已完成、已取消、已暂停
  - 节点状态：待处理、处理中、已完成、已跳过

### 集成特性

- **任务服务集成**：
  - 任务分配审批流程
  - 自动分配和手动审批的混合模式
  - 审批完成后的任务状态更新

- **通知集成**：
  - 审批通知自动发送
  - 提醒调度和超时处理
  - 多渠道通知支持（系统、邮件）

- **审计集成**：
  - 完整的执行历史记录
  - 用户操作审计
  - 统计分析和报告

## 数据结构

### 核心类型

```go
// 工作流定义
type WorkflowDefinition struct {
    ID          string
    Name        string
    Description string
    Version     string
    IsActive    bool
    Nodes       []WorkflowNode
    Edges       []WorkflowEdge
    Variables   map[string]VariableDefinition
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// 工作流实例
type WorkflowInstance struct {
    ID           string
    WorkflowID   string
    BusinessID   string
    BusinessType string
    Status       InstanceStatus
    CurrentNodes []string
    Variables    map[string]interface{}
    StartedBy    uint
    StartedAt    time.Time
    CompletedAt  *time.Time
    History      []ExecutionHistory
}

// 工作流节点
type WorkflowNode struct {
    ID          string
    Type        NodeType
    Name        string
    Description string
    Config      map[string]interface{}
    Position    NodePosition
}
```

### 审批相关

```go
// 审批请求
type ApprovalRequest struct {
    InstanceID string
    NodeID     string
    Action     ApprovalAction
    Comment    string
    ApprovedBy uint
    Variables  map[string]interface{}
}

// 待审批项
type PendingApproval struct {
    InstanceID   string
    NodeID       string
    NodeName     string
    WorkflowName string
    BusinessID   string
    BusinessType string
    AssignedTo   uint
    CreatedAt    time.Time
    Deadline     *time.Time
    Priority     string
}
```

## 使用示例

### 创建工作流定义

```go
// 创建任务分配审批流程
req := &CreateWorkflowRequest{
    ID:          "task_assignment_approval",
    Name:        "任务分配审批流程",
    Description: "用于任务分配的审批流程",
    Version:     "1.0",
    Nodes: []WorkflowNode{
        {
            ID:   "start",
            Type: NodeTypeStart,
            Name: "开始",
        },
        {
            ID:   "manager_approval",
            Type: NodeTypeApproval,
            Name: "经理审批",
            Config: map[string]interface{}{
                "assignees": []map[string]interface{}{
                    {"type": "manager", "value": "starter"},
                },
                "approval_type": "any",
                "timeout": 24 * 60, // 24小时
            },
        },
        {
            ID:   "end",
            Type: NodeTypeEnd,
            Name: "结束",
        },
    },
    Edges: []WorkflowEdge{
        {ID: "start_to_approval", From: "start", To: "manager_approval"},
        {ID: "approval_to_end", From: "manager_approval", To: "end"},
    },
}

definition, err := workflowManager.CreateWorkflow(ctx, req)
```

### 启动工作流

```go
// 启动任务分配审批
req := &StartWorkflowRequest{
    WorkflowID:   "task_assignment_approval",
    BusinessID:   "task_123",
    BusinessType: "task_assignment",
    StartedBy:    userID,
    Variables: map[string]interface{}{
        "task_id":     123,
        "assignee_id": 456,
        "priority":    "high",
    },
}

instance, err := workflowEngine.StartWorkflow(ctx, req)
```

### 处理审批

```go
// 处理审批请求
req := &ApprovalRequest{
    InstanceID: "instance_123",
    NodeID:     "manager_approval",
    Action:     ActionApprove,
    Comment:    "同意分配",
    ApprovedBy: managerID,
}

result, err := workflowEngine.ProcessApproval(ctx, req)
```

## 配置说明

### 审批节点配置

```go
config := map[string]interface{}{
    // 指派人配置
    "assignees": []map[string]interface{}{
        {
            "type":  "user",     // 指派类型
            "value": "123",      // 用户ID
        },
        {
            "type":  "role",     // 角色指派
            "value": "manager",  // 角色名称
        },
    },
    
    // 审批类型
    "approval_type": "any", // any: 任意一人, all: 所有人, majority: 多数人
    
    // 超时设置（分钟）
    "timeout": 1440, // 24小时
    
    // 超时处理
    "timeout_action": "auto_approve", // auto_approve, auto_reject, escalate
    
    // 升级处理
    "escalation": map[string]interface{}{
        "type":  "user",
        "value": "789", // 升级到的用户ID
    },
}
```

### 条件节点配置

```go
config := map[string]interface{}{
    "conditions": []map[string]interface{}{
        {
            "expression": "priority == high",
            "target":     "high_priority_approval",
            "priority":   1,
        },
        {
            "expression": "amount > 1000",
            "target":     "amount_approval",
            "priority":   2,
        },
    },
    "default_target": "default_approval", // 默认分支
}
```

## 测试覆盖

实现了全面的单元测试，覆盖以下组件：

1. **工作流定义管理器测试**
   - 创建工作流定义
   - 工作流验证逻辑
   - 错误处理

2. **工作流引擎测试**
   - 启动工作流
   - 处理审批
   - 状态管理

3. **节点执行器测试**
   - 审批节点执行
   - 条件节点执行
   - 表达式评估

4. **服务层测试**
   - 工作流服务
   - 通知服务
   - 审计服务

5. **集成测试**
   - 任务服务集成
   - 端到端流程测试

## 部署和运维

### 依赖要求

- Go 1.23+
- MySQL 5.7+
- Redis 6.0+
- 相关Go依赖包

### 配置项

```yaml
workflow:
  enabled: true
  default_timeout: 1440  # 默认超时时间（分钟）
  max_parallel_nodes: 10 # 最大并行节点数
  
notification:
  enabled: true
  channels:
    - system
    - email
  reminder_intervals:
    - 1440  # 24小时前
    - 240   # 4小时前
    - 60    # 1小时前
```

### 监控指标

- 工作流实例数量和状态分布
- 平均审批处理时间
- 超时实例统计
- 节点执行性能

## 扩展性

### 自定义节点类型

可以通过实现 `NodeExecutor` 接口来添加自定义节点类型：

```go
type CustomNodeExecutor struct{}

func (e *CustomNodeExecutor) Execute(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode) (*ExecutionResult, error) {
    // 自定义执行逻辑
    return &ExecutionResult{
        Success:     true,
        WaitForUser: false,
        NextNodes:   []string{"next_node"},
    }, nil
}

func (e *CustomNodeExecutor) GetNodeType() NodeType {
    return "custom"
}
```

### 自定义表达式函数

可以扩展条件节点的表达式评估器，添加自定义函数。

### 自定义通知渠道

可以通过实现通知接口来添加新的通知渠道（如短信、微信等）。

## 最佳实践

1. **工作流设计**
   - 保持流程简洁明了
   - 合理设置超时时间
   - 提供清晰的节点命名

2. **审批配置**
   - 明确指派人规则
   - 设置合适的升级机制
   - 考虑并行审批的使用场景

3. **性能优化**
   - 避免过深的嵌套流程
   - 合理使用并行节点
   - 定期清理历史数据

4. **错误处理**
   - 提供详细的错误信息
   - 实现重试机制
   - 监控异常情况

## 总结

审批流程引擎为任务管理系统提供了强大而灵活的审批能力，支持复杂的业务流程需求。通过模块化设计和丰富的扩展点，系统可以适应不同的业务场景和未来的功能扩展需求。
