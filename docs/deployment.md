# 部署和配置文档

## 环境要求

### 系统要求
- **操作系统**: Linux (Ubuntu 20.04+, CentOS 8+) 或 macOS
- **CPU**: 最小 2 核，推荐 4 核以上
- **内存**: 最小 4GB，推荐 8GB 以上
- **存储**: 最小 20GB，推荐 SSD

### 软件依赖
- **Go**: 1.21+
- **PostgreSQL**: 13+
- **Redis**: 6+
- **Docker**: 20.10+
- **Docker Compose**: 2.0+

## 配置文件

### 环境变量配置 (.env)
```bash
# 服务配置
APP_ENV=production
APP_PORT=8080
APP_DEBUG=false

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_NAME=taskmanage
DB_USER=taskmanage
DB_PASSWORD=your_secure_password
DB_CHARSET=utf8mb4
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0
REDIS_MAX_RETRIES=3

# JWT 配置
JWT_SECRET=your_jwt_secret_key_here
JWT_EXPIRES_IN=3600

# Asynq 配置
ASYNQ_REDIS_ADDR=localhost:6379
ASYNQ_REDIS_PASSWORD=your_redis_password
ASYNQ_CONCURRENCY=10

# 监控配置
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json

# 文件上传配置
UPLOAD_MAX_SIZE=10485760  # 10MB
UPLOAD_PATH=/var/uploads

# 邮件配置 (可选)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@company.com
SMTP_PASSWORD=smtp_password
SMTP_FROM=noreply@company.com
```

### 应用配置 (config/config.yaml)
```yaml
server:
  port: ${APP_PORT:8080}
  mode: ${APP_ENV:development}
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 10s

database:
  host: ${DB_HOST:localhost}
  port: ${DB_PORT:3306}
  name: ${DB_NAME:taskmanage}
  user: ${DB_USER:taskmanage}
  password: ${DB_PASSWORD:password}
  charset: ${DB_CHARSET:utf8mb4}
  max_open_conns: ${DB_MAX_OPEN_CONNS:25}
  max_idle_conns: ${DB_MAX_IDLE_CONNS:5}
  conn_max_lifetime: 300s

redis:
  host: ${REDIS_HOST:localhost}
  port: ${REDIS_PORT:6379}
  password: ${REDIS_PASSWORD:}
  db: ${REDIS_DB:0}
  max_retries: ${REDIS_MAX_RETRIES:3}
  pool_size: 10

queue:
  concurrency: ${ASYNQ_CONCURRENCY:10}
  queues:
    critical: 6
    default: 3
    low: 1
  retry:
    max_retry: 5
    initial_interval: 30s
    max_interval: 300s

auth:
  jwt_secret: ${JWT_SECRET:default_secret}
  jwt_expires_in: ${JWT_EXPIRES_IN:3600}
  bcrypt_cost: 12

logging:
  level: ${LOG_LEVEL:info}
  format: ${LOG_FORMAT:text}
  output: stdout

monitoring:
  prometheus:
    enabled: ${PROMETHEUS_ENABLED:true}
    port: ${PROMETHEUS_PORT:9090}
    path: /metrics

upload:
  max_size: ${UPLOAD_MAX_SIZE:10485760}
  allowed_types: [".jpg", ".jpeg", ".png", ".pdf", ".doc", ".docx"]
  path: ${UPLOAD_PATH:/tmp/uploads}

notification:
  smtp:
    host: ${SMTP_HOST:}
    port: ${SMTP_PORT:587}
    user: ${SMTP_USER:}
    password: ${SMTP_PASSWORD:}
    from: ${SMTP_FROM:noreply@company.com}
```

## Docker 部署

### Dockerfile
```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

# 创建上传目录
RUN mkdir -p /var/uploads

EXPOSE 8080

CMD ["./main"]
```

### docker-compose.yml
```yaml
version: '3.8'

services:
  # MySQL 数据库
  mysql:
    image: mysql:8.0
    container_name: taskmanage_mysql
    environment:
      MYSQL_DATABASE: taskmanage
      MYSQL_USER: taskmanage
      MYSQL_PASSWORD: ${DB_PASSWORD}
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "3306:3306"
    networks:
      - taskmanage_network
    restart: unless-stopped
    command: --default-authentication-plugin=mysql_native_password

  # Redis 缓存和队列
  redis:
    image: redis:6-alpine
    container_name: taskmanage_redis
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    networks:
      - taskmanage_network
    restart: unless-stopped

  # API 服务
  api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: taskmanage_api
    environment:
      - APP_ENV=production
      - DB_HOST=mysql
      - REDIS_HOST=redis
    env_file:
      - .env
    ports:
      - "8080:8080"
    depends_on:
      - mysql
      - redis
    volumes:
      - upload_data:/var/uploads
    networks:
      - taskmanage_network
    restart: unless-stopped

  # Worker 服务
  worker:
    build:
      context: .
      dockerfile: Dockerfile.worker
    container_name: taskmanage_worker
    environment:
      - APP_ENV=production
      - DB_HOST=mysql
      - REDIS_HOST=redis
    env_file:
      - .env
    depends_on:
      - mysql
      - redis
    networks:
      - taskmanage_network
    restart: unless-stopped

  # Prometheus 监控
  prometheus:
    image: prom/prometheus:latest
    container_name: taskmanage_prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    ports:
      - "9090:9090"
    networks:
      - taskmanage_network
    restart: unless-stopped

  # Grafana 可视化
  grafana:
    image: grafana/grafana:latest
    container_name: taskmanage_grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources
    ports:
      - "3000:3000"
    networks:
      - taskmanage_network
    restart: unless-stopped

  # Nginx 反向代理
  nginx:
    image: nginx:alpine
    container_name: taskmanage_nginx
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - api
    networks:
      - taskmanage_network
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
  prometheus_data:
  grafana_data:
  upload_data:

networks:
  taskmanage_network:
    driver: bridge
```

### Worker Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker cmd/worker/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/worker .
COPY --from=builder /app/config ./config

CMD ["./worker"]
```

## Kubernetes 部署

### namespace.yaml
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: taskmanage
```

### configmap.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: taskmanage-config
  namespace: taskmanage
data:
  config.yaml: |
    server:
      port: 8080
      mode: production
    database:
      host: postgres-service
      port: 5432
      name: taskmanage
    redis:
      host: redis-service
      port: 6379
    # ... 其他配置
```

### secret.yaml
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: taskmanage-secrets
  namespace: taskmanage
type: Opaque
data:
  db-password: <base64-encoded-password>
  redis-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-secret>
```

### postgres-deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: taskmanage
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:13
        env:
        - name: POSTGRES_DB
          value: taskmanage
        - name: POSTGRES_USER
          value: taskmanage
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: taskmanage-secrets
              key: db-password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: taskmanage
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
```

### api-deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: taskmanage-api
  namespace: taskmanage
spec:
  replicas: 3
  selector:
    matchLabels:
      app: taskmanage-api
  template:
    metadata:
      labels:
        app: taskmanage-api
    spec:
      containers:
      - name: api
        image: taskmanage:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: taskmanage-secrets
              key: db-password
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: taskmanage-secrets
              key: redis-password
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: taskmanage-secrets
              key: jwt-secret
        volumeMounts:
        - name: config-volume
          mountPath: /root/config
        - name: upload-volume
          mountPath: /var/uploads
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config-volume
        configMap:
          name: taskmanage-config
      - name: upload-volume
        persistentVolumeClaim:
          claimName: upload-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: taskmanage-api-service
  namespace: taskmanage
spec:
  selector:
    app: taskmanage-api
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

## 监控配置

### Prometheus 配置 (monitoring/prometheus.yml)
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "rules/*.yml"

scrape_configs:
  - job_name: 'taskmanage-api'
    static_configs:
      - targets: ['api:8080']
    metrics_path: /metrics
    scrape_interval: 10s

  - job_name: 'taskmanage-worker'
    static_configs:
      - targets: ['worker:8080']
    metrics_path: /metrics
    scrape_interval: 10s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

### Grafana 数据源配置
```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
```

## Nginx 配置

### nginx.conf
```nginx
events {
    worker_connections 1024;
}

http {
    upstream api_backend {
        server api:8080;
    }

    upstream grafana_backend {
        server grafana:3000;
    }

    # API 服务
    server {
        listen 80;
        server_name api.taskmanage.com;

        location / {
            proxy_pass http://api_backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # Grafana 监控
    server {
        listen 80;
        server_name monitor.taskmanage.com;

        location / {
            proxy_pass http://grafana_backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # HTTPS 配置 (可选)
    server {
        listen 443 ssl;
        server_name api.taskmanage.com;

        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;

        location / {
            proxy_pass http://api_backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

## 部署脚本

### 部署脚本 (scripts/deploy.sh)
```bash
#!/bin/bash

set -e

# 配置变量
APP_NAME="taskmanage"
DOCKER_REGISTRY="your-registry.com"
VERSION=${1:-latest}

echo "开始部署 $APP_NAME 版本 $VERSION"

# 构建镜像
echo "构建 Docker 镜像..."
docker build -t $DOCKER_REGISTRY/$APP_NAME:$VERSION .
docker build -f Dockerfile.worker -t $DOCKER_REGISTRY/$APP_NAME-worker:$VERSION .

# 推送镜像
echo "推送镜像到仓库..."
docker push $DOCKER_REGISTRY/$APP_NAME:$VERSION
docker push $DOCKER_REGISTRY/$APP_NAME-worker:$VERSION

# 更新 docker-compose
echo "更新服务..."
export IMAGE_TAG=$VERSION
docker-compose down
docker-compose up -d

# 等待服务启动
echo "等待服务启动..."
sleep 30

# 健康检查
echo "执行健康检查..."
if curl -f http://localhost:8080/health; then
    echo "部署成功！"
else
    echo "部署失败，回滚..."
    docker-compose down
    export IMAGE_TAG=latest
    docker-compose up -d
    exit 1
fi
```

### 数据库迁移脚本 (scripts/migrate.sh)
```bash
#!/bin/bash

set -e

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-taskmanage}
DB_USER=${DB_USER:-taskmanage}

echo "开始数据库迁移..."

# 等待数据库启动
echo "等待数据库连接..."
until pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; do
    echo "等待 PostgreSQL 启动..."
    sleep 2
done

# 执行迁移
echo "执行数据库迁移..."
go run cmd/migrate/main.go

echo "数据库迁移完成！"
```

## 生产环境优化

### 性能调优
```yaml
# PostgreSQL 优化
postgresql.conf:
  shared_buffers: 256MB
  effective_cache_size: 1GB
  work_mem: 4MB
  maintenance_work_mem: 64MB
  checkpoint_completion_target: 0.9
  wal_buffers: 16MB
  default_statistics_target: 100

# Redis 优化
redis.conf:
  maxmemory: 512mb
  maxmemory-policy: allkeys-lru
  save: 900 1 300 10 60 10000
  tcp-keepalive: 300
```

### 安全配置
```bash
# 防火墙配置
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw deny 5432/tcp  # 仅内网访问
ufw deny 6379/tcp  # 仅内网访问
ufw enable

# SSL 证书配置 (Let's Encrypt)
certbot --nginx -d api.taskmanage.com
```

### 备份策略
```bash
# 数据库备份脚本
#!/bin/bash
BACKUP_DIR="/var/backups/taskmanage"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份数据库
pg_dump -h localhost -U taskmanage taskmanage | gzip > $BACKUP_DIR/db_backup_$DATE.sql.gz

# 备份 Redis
redis-cli --rdb $BACKUP_DIR/redis_backup_$DATE.rdb

# 清理旧备份 (保留30天)
find $BACKUP_DIR -name "*.gz" -mtime +30 -delete
find $BACKUP_DIR -name "*.rdb" -mtime +30 -delete
```

## 故障排查

### 常见问题
1. **数据库连接失败**: 检查网络连接和认证信息
2. **Redis 连接超时**: 检查 Redis 配置和网络
3. **队列任务堆积**: 检查 Worker 状态和并发配置
4. **内存不足**: 监控内存使用，调整容器资源限制

### 日志查看
```bash
# Docker 日志
docker-compose logs -f api
docker-compose logs -f worker

# Kubernetes 日志
kubectl logs -f deployment/taskmanage-api -n taskmanage
kubectl logs -f deployment/taskmanage-worker -n taskmanage
```

### 性能监控
- **应用指标**: 通过 Prometheus + Grafana 监控
- **系统指标**: CPU、内存、磁盘、网络使用率
- **数据库指标**: 连接数、查询性能、锁等待
- **队列指标**: 任务处理速度、失败率、队列深度
