# 任务分配管理系统 (Task Assignment Management System)

基于 Go + Asynq + Redis 的企业级任务分配管理系统，支持智能分配、负载均衡、审批流程和实时监控。

## 🚀 核心功能

- **智能任务分配**: 支持轮询、负载均衡、技能匹配、优先级匹配等多种分配算法
- **完整审批流程**: 任务创建审批 → 分配确认 → 完成审批的完整流程
- **实时状态管理**: 员工状态、任务进度、系统负载的实时监控
- **灵活重分配**: 支持手动和自动重分配，处理员工请假、任务超时等场景
- **可观测性**: 基于 Prometheus + Grafana 的完整监控体系

## 📋 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Frontend  │    │   Admin Panel   │    │   Mobile App    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
┌─────────────────────────────────────────────────────────────────┐
│                        API Gateway                             │
│              (Authentication, Rate Limiting)                   │
└─────────────────────────────────────────────────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Task Service   │    │  Staff Service  │    │ Assignment Svc  │
│                 │    │                 │    │                 │
│ • CRUD          │    │ • Profile Mgmt  │    │ • Algorithms    │
│ • Status Mgmt   │    │ • Skill Mgmt    │    │ • Load Balance  │
│ • Approval      │    │ • Availability  │    │ • Reassignment  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
┌─────────────────────────────────────────────────────────────────┐
│                     Message Queue (Asynq)                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌──────────┐  │
│  │ Critical Q  │ │ Default Q   │ │   Low Q     │ │  DLQ     │  │
│  └─────────────┘ └─────────────┘ └─────────────┘ └──────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                 │
┌─────────────────────────────────────────────────────────────────┐
│                    Worker Cluster                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │  Worker 1   │ │  Worker 2   │ │  Worker N   │               │
│  └─────────────┘ └─────────────┘ └─────────────┘               │
└─────────────────────────────────────────────────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │     Redis       │    │   Monitoring    │
│                 │    │                 │    │                 │
│ • Business Data │    │ • Queue Storage │    │ • Prometheus    │
│ • User Data     │    │ • Cache         │    │ • Grafana       │
│ • Audit Logs    │    │ • Session       │    │ • AlertManager  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🛠️ 技术栈

- **后端**: Go 1.21+, Gin, GORM
- **队列**: Asynq (Redis-based)
- **数据库**: MySQL 8.0+, Redis
- **监控**: Prometheus, Grafana
- **部署**: Docker, Docker Compose, Kubernetes
- **前端**: React/Vue.js (可选)

## 📁 项目结构

```
taskmanage/
├── cmd/                    # 应用入口
│   ├── api/               # API 服务
│   ├── worker/            # Worker 服务
│   └── admin/             # 管理后台
├── internal/              # 内部包
│   ├── api/              # API 层
│   ├── service/          # 业务逻辑层
│   ├── repository/       # 数据访问层
│   ├── model/            # 数据模型
│   ├── config/           # 配置管理
│   └── middleware/       # 中间件
├── pkg/                   # 公共包
│   ├── queue/            # 队列封装
│   ├── database/         # 数据库连接
│   ├── logger/           # 日志
│   └── utils/            # 工具函数
├── docs/                  # 文档
├── scripts/              # 脚本
├── deployments/          # 部署配置
└── tests/                # 测试
```

## 🚀 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 6+
- Docker & Docker Compose

### 本地开发

1. **克隆项目**
```bash
git clone git@github.com:13892349048/task_magage.git
cd taskmanage
```

2. **启动依赖服务**
```bash
docker-compose up -d mysql redis
```

3. **配置环境变量**
```bash
cp .env.example .env
# 编辑 .env 文件配置数据库连接等
```

4. **运行数据库迁移**
```bash
go run cmd/migrate/main.go
```

5. **启动服务**
```bash
# 启动 API 服务
go run cmd/api/main.go

# 启动 Worker 服务
go run cmd/worker/main.go
```

### Docker 部署

```bash
docker-compose up -d
```

## 📚 文档

- [API 接口文档](docs/api.md)
- [数据库设计](docs/database.md)
- [部署指南](docs/deployment.md)
- [开发规范](docs/development.md)
- [监控配置](docs/monitoring.md)

## 🔧 配置

主要配置项在 `config/config.yaml`:

```yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  name: taskmanage
  user: postgres
  password: password

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

queue:
  concurrency: 10
  queues:
    critical: 6
    default: 3
    low: 1
```

## 📊 监控

系统提供完整的监控指标:

- **任务指标**: 完成率、处理时长、失败率
- **员工指标**: 负载分布、技能匹配率、绩效
- **系统指标**: QPS、延迟、错误率、队列深度

访问 Grafana Dashboard: http://localhost:3000

## 🤝 贡献

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。
