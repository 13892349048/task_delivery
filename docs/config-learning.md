# 配置系统学习指南

## 问题分析

### 1. 配置文件映射问题
- **问题**：`--env production` 寻找 `config.production.yaml`，但实际文件名是 `config.prod.yaml`
- **解决**：在 `loader.go` 中添加环境名称映射表

### 2. JWT密钥验证问题
- **问题**：生产配置使用 `"${JWT_SECRET}"` 模板语法，但环境变量未设置
- **原因**：Viper的环境变量有两种处理方式

## Viper环境变量处理机制

### 方式1：直接绑定（推荐）
```go
// 在代码中绑定
viper.BindEnv("jwt.secret", "JWT_SECRET")
```
配置文件中直接使用键名：
```yaml
jwt:
  secret: ""  # 将被环境变量 JWT_SECRET 覆盖
```

### 方式2：模板替换（需要额外处理）
配置文件中使用模板语法：
```yaml
jwt:
  secret: "${JWT_SECRET}"  # 需要手动展开
```

## 当前系统使用的是方式1

我们的系统在 `loader.go` 第127行已经正确绑定：
```go
"jwt.secret": "JWT_SECRET",
```

但生产配置文件使用了方式2的语法，导致冲突。

## 解决方案

### 临时解决（学习用）
创建 `config.prod.test.yaml`，使用硬编码密钥进行测试

### 生产环境解决
1. 设置环境变量：`$env:JWT_SECRET="your-32-char-secret"`
2. 修改生产配置文件，移除模板语法

## 最佳实践

1. **开发环境**：使用硬编码配置便于开发
2. **测试环境**：使用环境变量覆盖关键配置
3. **生产环境**：完全依赖环境变量，不在配置文件中暴露敏感信息
