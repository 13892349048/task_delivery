# API 接口文档

## 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: Bearer Token (JWT)
- **Content-Type**: `application/json`

## 通用响应格式

### 成功响应
```json
{
  "code": 200,
  "message": "success",
  "data": {},
  "timestamp": "2024-08-28T14:15:56+08:00"
}
```

### 错误响应
```json
{
  "code": 400,
  "message": "参数错误",
  "error": "详细错误信息",
  "timestamp": "2024-08-28T14:15:56+08:00"
}
```

## 认证接口

### 用户登录
```http
POST /auth/login
```

**请求参数**:
```json
{
  "username": "admin",
  "password": "password123"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "user": {
      "id": "1",
      "username": "admin",
      "role": "admin"
    }
  }
}
```

### 刷新Token
```http
POST /auth/refresh
```

## 任务管理接口

### 创建任务
```http
POST /tasks
```

**请求参数**:
```json
{
  "title": "开发用户管理模块",
  "description": "实现用户的增删改查功能",
  "priority": "HIGH",
  "deadline": "2024-09-15T18:00:00Z",
  "estimated_hours": 40,
  "required_skills": ["Go", "PostgreSQL", "Redis"],
  "metadata": {
    "project_id": "proj_001",
    "module": "user_management"
  }
}
```

**响应**:
```json
{
  "code": 200,
  "message": "任务创建成功",
  "data": {
    "id": "task_001",
    "title": "开发用户管理模块",
    "status": "DRAFT",
    "created_at": "2024-08-28T14:15:56+08:00"
  }
}
```

### 获取任务列表
```http
GET /tasks?page=1&size=20&status=IN_PROGRESS&priority=HIGH&assignee_id=staff_001
```

**查询参数**:
- `page`: 页码 (默认: 1)
- `size`: 每页数量 (默认: 20)
- `status`: 任务状态
- `priority`: 优先级
- `assignee_id`: 分配员工ID
- `creator_id`: 创建者ID
- `deadline_start`: 截止时间开始
- `deadline_end`: 截止时间结束

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "task_001",
        "title": "开发用户管理模块",
        "description": "实现用户的增删改查功能",
        "priority": "HIGH",
        "status": "IN_PROGRESS",
        "deadline": "2024-09-15T18:00:00Z",
        "assignee": {
          "id": "staff_001",
          "name": "张三",
          "department": "技术部"
        },
        "creator": {
          "id": "admin_001",
          "name": "管理员"
        },
        "created_at": "2024-08-28T14:15:56+08:00",
        "updated_at": "2024-08-28T14:20:00+08:00"
      }
    ],
    "pagination": {
      "page": 1,
      "size": 20,
      "total": 100,
      "pages": 5
    }
  }
}
```

### 获取任务详情
```http
GET /tasks/{task_id}
```

### 更新任务
```http
PUT /tasks/{task_id}
```

### 删除任务
```http
DELETE /tasks/{task_id}
```

### 任务状态变更
```http
POST /tasks/{task_id}/status
```

**请求参数**:
```json
{
  "status": "IN_PROGRESS",
  "comment": "开始执行任务"
}
```

## 员工管理接口

### 创建员工
```http
POST /staff
```

**请求参数**:
```json
{
  "name": "张三",
  "email": "zhangsan@company.com",
  "department": "技术部",
  "position": "高级开发工程师",
  "projects": ["proj_001", "proj_002"],
  "skills": [
    {"name": "Go", "level": 4},
    {"name": "PostgreSQL", "level": 3},
    {"name": "Redis", "level": 3}
  ],
  "max_concurrent_tasks": 3,
  "work_hours": {
    "start": "09:00",
    "end": "18:00",
    "timezone": "Asia/Shanghai"
  }
}
```

### 获取员工列表
```http
GET /staff?page=1&size=20&department=技术部&status=ACTIVE&skill=Go
```

### 获取员工详情
```http
GET /staff/{staff_id}
```

### 更新员工信息
```http
PUT /staff/{staff_id}
```

### 员工状态变更
```http
POST /staff/{staff_id}/status
```

**请求参数**:
```json
{
  "status": "ON_LEAVE",
  "reason": "年假",
  "start_date": "2024-09-01",
  "end_date": "2024-09-07"
}
```

### 获取员工当前任务
```http
GET /staff/{staff_id}/tasks
```

## 任务分配接口

### 手动分配任务
```http
POST /assignments
```

**请求参数**:
```json
{
  "task_id": "task_001",
  "assignee_id": "staff_001",
  "assignment_type": "MANUAL",
  "comment": "根据技能匹配手动分配"
}
```

### 自动分配任务
```http
POST /assignments/auto
```

**请求参数**:
```json
{
  "task_id": "task_001",
  "algorithm": "HYBRID",
  "options": {
    "skill_weight": 0.5,
    "load_weight": 0.3,
    "priority_weight": 0.2
  }
}
```

### 重新分配任务
```http
POST /assignments/{assignment_id}/reassign
```

**请求参数**:
```json
{
  "new_assignee_id": "staff_002",
  "reason": "原分配员工请假",
  "transfer_progress": true
}
```

### 获取分配历史
```http
GET /assignments/history?task_id=task_001
```

### 批量分配任务
```http
POST /assignments/batch
```

**请求参数**:
```json
{
  "task_ids": ["task_001", "task_002", "task_003"],
  "algorithm": "LOAD_BALANCE",
  "auto_approve": false
}
```

## 审批管理接口

### 提交审批
```http
POST /approvals
```

**请求参数**:
```json
{
  "type": "TASK_CREATION",
  "target_id": "task_001",
  "comment": "请审批新任务创建"
}
```

### 审批处理
```http
POST /approvals/{approval_id}/process
```

**请求参数**:
```json
{
  "action": "APPROVE",
  "comment": "任务需求明确，同意创建"
}
```

### 获取待审批列表
```http
GET /approvals/pending?type=TASK_CREATION&assignee_id=admin_001
```

### 获取审批历史
```http
GET /approvals/history?target_id=task_001
```

## 监控统计接口

### 任务统计
```http
GET /statistics/tasks
```

**查询参数**:
- `start_date`: 开始日期
- `end_date`: 结束日期
- `department`: 部门
- `assignee_id`: 员工ID

**响应**:
```json
{
  "code": 200,
  "data": {
    "total_tasks": 150,
    "completed_tasks": 120,
    "in_progress_tasks": 25,
    "overdue_tasks": 5,
    "completion_rate": 0.8,
    "average_completion_time": 72.5,
    "by_priority": {
      "HIGH": 30,
      "MEDIUM": 80,
      "LOW": 40
    },
    "by_status": {
      "COMPLETED": 120,
      "IN_PROGRESS": 25,
      "PENDING_APPROVAL": 3,
      "OVERDUE": 2
    }
  }
}
```

### 员工统计
```http
GET /statistics/staff
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "total_staff": 50,
    "active_staff": 45,
    "on_leave_staff": 3,
    "offline_staff": 2,
    "average_load": 2.3,
    "load_distribution": {
      "0": 5,
      "1": 15,
      "2": 20,
      "3": 5
    },
    "skill_distribution": {
      "Go": 30,
      "Java": 25,
      "Python": 20,
      "JavaScript": 35
    }
  }
}
```

### 系统性能指标
```http
GET /statistics/system
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "queue_status": {
      "critical": {
        "pending": 5,
        "processing": 2,
        "completed": 1000,
        "failed": 3
      },
      "default": {
        "pending": 15,
        "processing": 8,
        "completed": 5000,
        "failed": 12
      }
    },
    "worker_status": {
      "total_workers": 10,
      "active_workers": 8,
      "idle_workers": 2
    },
    "response_times": {
      "avg": 150,
      "p95": 300,
      "p99": 500
    }
  }
}
```

## 通知接口

### 发送通知
```http
POST /notifications
```

**请求参数**:
```json
{
  "type": "TASK_ASSIGNED",
  "recipient_id": "staff_001",
  "title": "新任务分配",
  "content": "您有一个新的任务需要处理",
  "data": {
    "task_id": "task_001",
    "task_title": "开发用户管理模块"
  }
}
```

### 获取通知列表
```http
GET /notifications?recipient_id=staff_001&status=UNREAD
```

### 标记通知已读
```http
POST /notifications/{notification_id}/read
```

## 文件上传接口

### 上传任务附件
```http
POST /files/upload
Content-Type: multipart/form-data
```

**请求参数**:
- `file`: 文件内容
- `task_id`: 关联任务ID
- `description`: 文件描述

### 下载文件
```http
GET /files/{file_id}/download
```

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 422 | 业务逻辑错误 |
| 500 | 服务器内部错误 |

## 业务错误码

| 业务码 | 说明 |
|--------|------|
| 10001 | 任务不存在 |
| 10002 | 任务状态不允许此操作 |
| 10003 | 员工不存在 |
| 10004 | 员工状态不可用 |
| 10005 | 技能不匹配 |
| 10006 | 负载超限 |
| 10007 | 审批已处理 |
| 10008 | 分配冲突 |

## 限流规则

- **普通用户**: 100 请求/分钟
- **管理员**: 500 请求/分钟
- **系统API**: 1000 请求/分钟

## 示例代码

### Go 客户端示例
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type Client struct {
    BaseURL string
    Token   string
}

func (c *Client) CreateTask(task *Task) (*TaskResponse, error) {
    data, _ := json.Marshal(task)
    req, _ := http.NewRequest("POST", c.BaseURL+"/tasks", bytes.NewBuffer(data))
    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result TaskResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}
```

### JavaScript 客户端示例
```javascript
class TaskClient {
    constructor(baseURL, token) {
        this.baseURL = baseURL;
        this.token = token;
    }
    
    async createTask(task) {
        const response = await fetch(`${this.baseURL}/tasks`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(task)
        });
        
        return await response.json();
    }
    
    async getTasks(params = {}) {
        const url = new URL(`${this.baseURL}/tasks`);
        Object.keys(params).forEach(key => 
            url.searchParams.append(key, params[key])
        );
        
        const response = await fetch(url, {
            headers: {
                'Authorization': `Bearer ${this.token}`
            }
        });
        
        return await response.json();
    }
}
```
