# TaskManage API 测试计划

## 测试概述

### 测试目标
- 验证所有API接口功能正确性
- 确保数据完整性和一致性
- 验证错误处理和边界条件
- 性能基准测试
- 安全性验证

### 测试环境
- **开发环境**: localhost:8080
- **测试数据库**: MySQL (test database)
- **认证方式**: JWT Token
- **测试工具**: Postman, Go test, curl

## 测试分类

### 1. 单元测试 (Unit Tests)
- Service层业务逻辑测试
- Repository层数据访问测试
- 工具函数测试

### 2. 集成测试 (Integration Tests)
- API端点功能测试
- 数据库集成测试
- 中间件测试

### 3. 端到端测试 (E2E Tests)
- 完整业务流程测试
- 用户场景测试

### 4. 性能测试 (Performance Tests)
- 并发请求测试
- 响应时间测试
- 吞吐量测试

## API测试用例

### 认证模块测试

#### 1. 用户注册 `POST /api/v1/auth/register`
**测试用例**:
- ✅ 正常注册流程
- ✅ 重复用户名注册
- ✅ 无效邮箱格式
- ✅ 密码强度验证
- ✅ 必填字段验证

#### 2. 用户登录 `POST /api/v1/auth/login`
**测试用例**:
- ✅ 正确凭据登录
- ✅ 错误密码登录
- ✅ 不存在用户登录
- ✅ JWT Token生成验证

### 员工管理模块测试

#### 1. 员工列表 `GET /api/v1/employees`
**测试用例**:
- ✅ 获取所有员工
- ✅ 分页参数测试
- ✅ 过滤条件测试
- ✅ 排序功能测试

#### 2. 创建员工 `POST /api/v1/employees`
**测试用例**:
- ✅ 正常创建员工
- ✅ 必填字段验证
- ✅ 重复员工编号
- ✅ 关联用户验证

#### 3. 员工状态管理 `PUT /api/v1/employees/:id/status`
**测试用例**:
- ✅ 更新为有效状态
- ✅ 无效状态值
- ✅ 不存在的员工ID
- ✅ 权限验证

#### 4. 工作负载统计 `GET /api/v1/employees/workload/stats`
**测试用例**:
- ✅ 获取全部统计
- ✅ 按部门过滤
- ✅ 日期范围过滤
- ✅ 数据准确性验证

### 技能管理模块测试

#### 1. 技能CRUD操作
**测试用例**:
- ✅ 创建技能 `POST /api/v1/skills`
- ✅ 获取技能列表 `GET /api/v1/skills`
- ✅ 获取单个技能 `GET /api/v1/skills/:id`
- ✅ 更新技能 `PUT /api/v1/skills/:id`
- ✅ 删除技能 `DELETE /api/v1/skills/:id`

#### 2. 员工技能分配
**测试用例**:
- ✅ 分配技能给员工 `POST /api/v1/skills/assign`
- ✅ 移除员工技能 `DELETE /api/v1/employees/:id/skills/:skill_id`
- ✅ 获取员工技能 `GET /api/v1/skills/employees/:id`

## 测试数据准备

### 基础测试数据
```json
{
  "test_users": [
    {
      "username": "admin",
      "email": "admin@test.com",
      "password": "Admin123!",
      "role": "admin"
    },
    {
      "username": "manager",
      "email": "manager@test.com", 
      "password": "Manager123!",
      "role": "manager"
    },
    {
      "username": "employee1",
      "email": "emp1@test.com",
      "password": "Emp123!",
      "role": "employee"
    }
  ],
  "test_employees": [
    {
      "employee_no": "EMP001",
      "department": "开发部",
      "position": "高级工程师",
      "status": "active",
      "max_tasks": 5
    },
    {
      "employee_no": "EMP002", 
      "department": "测试部",
      "position": "测试工程师",
      "status": "active",
      "max_tasks": 3
    }
  ],
  "test_skills": [
    {
      "name": "Go语言",
      "category": "编程语言",
      "description": "Go编程语言技能",
      "tags": ["backend", "programming"]
    },
    {
      "name": "MySQL",
      "category": "数据库",
      "description": "MySQL数据库技能", 
      "tags": ["database", "sql"]
    }
  ]
}
```

## 测试执行计划

### Phase 1: 环境准备 (30分钟)
1. 启动测试数据库
2. 运行数据库迁移
3. 准备测试数据
4. 启动API服务

### Phase 2: 单元测试 (1小时)
1. Service层测试
2. Repository层测试
3. 工具函数测试

### Phase 3: 集成测试 (2小时)
1. 认证模块测试
2. 员工管理模块测试
3. 技能管理模块测试
4. 错误处理测试

### Phase 4: 性能测试 (1小时)
1. 并发用户测试
2. 大数据量测试
3. 响应时间测试

### Phase 5: 安全测试 (30分钟)
1. SQL注入测试
2. XSS攻击测试
3. 权限绕过测试

## 测试通过标准

### 功能测试
- ✅ 所有API返回正确的HTTP状态码
- ✅ 响应数据格式符合API文档
- ✅ 业务逻辑正确执行
- ✅ 错误处理符合预期

### 性能测试
- ✅ 95%请求响应时间 < 200ms
- ✅ 支持100并发用户
- ✅ 无内存泄漏
- ✅ 数据库连接池正常

### 安全测试
- ✅ JWT认证正常工作
- ✅ 权限控制有效
- ✅ 输入验证防止注入攻击
- ✅ 敏感信息不泄露

## 测试报告模板

### 测试执行摘要
- 测试开始时间: 
- 测试结束时间:
- 测试用例总数:
- 通过用例数:
- 失败用例数:
- 跳过用例数:

### 缺陷统计
- 严重缺陷: 0
- 一般缺陷: 0
- 轻微缺陷: 0

### 性能指标
- 平均响应时间:
- 95%响应时间:
- 最大并发用户数:
- 吞吐量:

### 建议和改进
- 性能优化建议
- 功能改进建议
- 安全加固建议
