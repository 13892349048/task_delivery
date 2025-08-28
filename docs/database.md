# 数据库设计文档

## 数据库选型

- **主数据库**: MySQL 8.0+ (业务数据、用户数据、审计日志)
- **缓存数据库**: Redis 6+ (队列存储、缓存、会话)
- **时序数据库**: InfluxDB (可选，用于监控指标存储)

## 表结构设计

### 用户相关表

#### users (用户表)
```sql
CREATE TABLE users (
    id VARCHAR(50) PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user', -- admin, manager, user
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, inactive, locked
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
```

#### staff (员工表)
```sql
CREATE TABLE staff (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(50) REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    employee_id VARCHAR(50) UNIQUE,
    department VARCHAR(100) NOT NULL,
    position VARCHAR(100) NOT NULL,
    level INTEGER DEFAULT 1, -- 员工等级 1-5
    max_concurrent_tasks INTEGER DEFAULT 3,
    current_load INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active', -- active, busy, on_leave, offline, inactive
    timezone VARCHAR(50) DEFAULT 'Asia/Shanghai',
    work_start_time TIME DEFAULT '09:00:00',
    work_end_time TIME DEFAULT '18:00:00',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_staff_department ON staff(department);
CREATE INDEX idx_staff_status ON staff(status);
CREATE INDEX idx_staff_user_id ON staff(user_id);
```

#### staff_skills (员工技能表)
```sql
CREATE TABLE staff_skills (
    id INT AUTO_INCREMENT PRIMARY KEY,
    staff_id VARCHAR(50) REFERENCES staff(id) ON DELETE CASCADE,
    skill_name VARCHAR(100) NOT NULL,
    skill_level INTEGER NOT NULL CHECK (skill_level >= 1 AND skill_level <= 5),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(staff_id, skill_name)
);

CREATE INDEX idx_staff_skills_staff_id ON staff_skills(staff_id);
CREATE INDEX idx_staff_skills_skill_name ON staff_skills(skill_name);
```

#### staff_projects (员工项目关联表)
```sql
CREATE TABLE staff_projects (
    id INT AUTO_INCREMENT PRIMARY KEY,
    staff_id VARCHAR(50) REFERENCES staff(id) ON DELETE CASCADE,
    project_id VARCHAR(50) NOT NULL,
    project_name VARCHAR(200) NOT NULL,
    role VARCHAR(100), -- 在项目中的角色
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_staff_projects_staff_id ON staff_projects(staff_id);
CREATE INDEX idx_staff_projects_project_id ON staff_projects(project_id);
```

### 任务相关表

#### tasks (任务表)
```sql
CREATE TABLE tasks (
    id VARCHAR(50) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    priority VARCHAR(20) NOT NULL DEFAULT 'medium', -- high, medium, low
    status VARCHAR(30) NOT NULL DEFAULT 'draft', -- draft, pending_approval, approved, assigned, in_progress, paused, completed, verified, cancelled
    estimated_hours INTEGER,
    actual_hours INTEGER DEFAULT 0,
    progress_percentage INTEGER DEFAULT 0 CHECK (progress_percentage >= 0 AND progress_percentage <= 100),
    deadline TIMESTAMP NULL,
    assignee_id VARCHAR(50) REFERENCES staff(id),
    creator_id VARCHAR(50) REFERENCES users(id) NOT NULL,
    project_id VARCHAR(50),
    parent_task_id VARCHAR(50) REFERENCES tasks(id), -- 子任务关联
    metadata JSON, -- 扩展字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    assigned_at TIMESTAMP NULL,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX idx_tasks_creator_id ON tasks(creator_id);
CREATE INDEX idx_tasks_deadline ON tasks(deadline);
CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
```

#### task_required_skills (任务所需技能表)
```sql
CREATE TABLE task_required_skills (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_id VARCHAR(50) REFERENCES tasks(id) ON DELETE CASCADE,
    skill_name VARCHAR(100) NOT NULL,
    required_level INTEGER NOT NULL CHECK (required_level >= 1 AND required_level <= 5),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_required_skills_task_id ON task_required_skills(task_id);
CREATE INDEX idx_task_required_skills_skill_name ON task_required_skills(skill_name);
```

#### task_dependencies (任务依赖关系表)
```sql
CREATE TABLE task_dependencies (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_id VARCHAR(50) REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_task_id VARCHAR(50) REFERENCES tasks(id) ON DELETE CASCADE,
    dependency_type VARCHAR(20) DEFAULT 'finish_to_start', -- finish_to_start, start_to_start, finish_to_finish, start_to_finish
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(task_id, depends_on_task_id)
);

CREATE INDEX idx_task_dependencies_task_id ON task_dependencies(task_id);
CREATE INDEX idx_task_dependencies_depends_on ON task_dependencies(depends_on_task_id);
```

### 分配相关表

#### task_assignments (任务分配表)
```sql
CREATE TABLE task_assignments (
    id VARCHAR(50) PRIMARY KEY,
    task_id VARCHAR(50) REFERENCES tasks(id) ON DELETE CASCADE,
    assignee_id VARCHAR(50) REFERENCES staff(id),
    previous_assignee_id VARCHAR(50) REFERENCES staff(id), -- 重分配时的原分配人
    assignment_type VARCHAR(20) NOT NULL, -- manual, auto_round_robin, auto_load_balance, auto_skill_match, auto_hybrid
    algorithm_config JSON, -- 分配算法配置参数
    assignment_score DECIMAL(5,2), -- 分配匹配度评分
    status VARCHAR(20) DEFAULT 'active', -- active, reassigned, cancelled
    assigned_by VARCHAR(50) REFERENCES users(id) NOT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    comment TEXT
);

CREATE INDEX idx_task_assignments_task_id ON task_assignments(task_id);
CREATE INDEX idx_task_assignments_assignee_id ON task_assignments(assignee_id);
CREATE INDEX idx_task_assignments_status ON task_assignments(status);
CREATE INDEX idx_task_assignments_assigned_at ON task_assignments(assigned_at);
```

#### assignment_history (分配历史表)
```sql
CREATE TABLE assignment_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_id VARCHAR(50) REFERENCES tasks(id),
    assignment_id VARCHAR(50) REFERENCES task_assignments(id),
    action VARCHAR(30) NOT NULL, -- assigned, reassigned, cancelled, completed
    from_assignee_id VARCHAR(50) REFERENCES staff(id),
    to_assignee_id VARCHAR(50) REFERENCES staff(id),
    reason TEXT,
    operated_by VARCHAR(50) REFERENCES users(id) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_assignment_history_task_id ON assignment_history(task_id);
CREATE INDEX idx_assignment_history_assignment_id ON assignment_history(assignment_id);
CREATE INDEX idx_assignment_history_created_at ON assignment_history(created_at);
```

### 审批相关表

#### approvals (审批表)
```sql
CREATE TABLE approvals (
    id VARCHAR(50) PRIMARY KEY,
    type VARCHAR(30) NOT NULL, -- task_creation, task_completion, task_reassignment, priority_change
    target_type VARCHAR(20) NOT NULL, -- task, assignment, staff
    target_id VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected, cancelled
    submitter_id VARCHAR(50) REFERENCES users(id) NOT NULL,
    approver_id VARCHAR(50) REFERENCES users(id),
    submit_comment TEXT,
    approval_comment TEXT,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    metadata JSON -- 审批相关的额外数据
);

CREATE INDEX idx_approvals_type ON approvals(type);
CREATE INDEX idx_approvals_status ON approvals(status);
CREATE INDEX idx_approvals_target ON approvals(target_type, target_id);
CREATE INDEX idx_approvals_submitter_id ON approvals(submitter_id);
CREATE INDEX idx_approvals_approver_id ON approvals(approver_id);
CREATE INDEX idx_approvals_submitted_at ON approvals(submitted_at);
```

#### approval_workflows (审批流程配置表)
```sql
CREATE TABLE approval_workflows (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(30) NOT NULL,
    conditions JSON, -- 触发条件
    approvers JSON, -- 审批人配置
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_approval_workflows_type ON approval_workflows(type);
CREATE INDEX idx_approval_workflows_is_active ON approval_workflows(is_active);
```

### 通知相关表

#### notifications (通知表)
```sql
CREATE TABLE notifications (
    id VARCHAR(50) PRIMARY KEY,
    type VARCHAR(30) NOT NULL, -- task_assigned, task_completed, task_overdue, approval_required, etc.
    recipient_id VARCHAR(50) REFERENCES users(id) NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    data JSON, -- 通知相关数据
    status VARCHAR(20) DEFAULT 'unread', -- unread, read, archived
    priority VARCHAR(20) DEFAULT 'normal', -- high, normal, low
    channels VARCHAR(100) DEFAULT 'system', -- system, email, sms, push
    sent_at TIMESTAMP NULL,
    read_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_recipient_id ON notifications(recipient_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
```

### 系统配置表

#### system_configs (系统配置表)
```sql
CREATE TABLE system_configs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    category VARCHAR(50) DEFAULT 'general',
    is_encrypted BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_system_configs_key ON system_configs(key);
CREATE INDEX idx_system_configs_category ON system_configs(category);
```

### 审计日志表

#### audit_logs (审计日志表)
```sql
CREATE TABLE audit_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(50) REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(30) NOT NULL,
    resource_id VARCHAR(50) NOT NULL,
    old_values JSON,
    new_values JSON,
    ip_address VARCHAR(45), -- IPv4/IPv6 地址
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- MySQL 分区表，按月分区
-- ALTER TABLE audit_logs PARTITION BY RANGE (YEAR(created_at)*100 + MONTH(created_at)) (
--     PARTITION p202408 VALUES LESS THAN (202409),
--     PARTITION p202409 VALUES LESS THAN (202410)
-- );
```

### 文件附件表

#### attachments (附件表)
```sql
CREATE TABLE attachments (
    id VARCHAR(50) PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    related_type VARCHAR(30) NOT NULL, -- task, staff, approval
    related_id VARCHAR(50) NOT NULL,
    uploaded_by VARCHAR(50) REFERENCES users(id) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_attachments_related ON attachments(related_type, related_id);
CREATE INDEX idx_attachments_uploaded_by ON attachments(uploaded_by);
```

## 视图定义

### 任务统计视图
```sql
CREATE VIEW task_statistics AS
SELECT 
    DATE(created_at) as date,
    COUNT(*) as total_tasks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress_tasks,
    COUNT(CASE WHEN status = 'overdue' OR deadline < NOW() THEN 1 END) as overdue_tasks,
    AVG(CASE WHEN completed_at IS NOT NULL THEN 
        EXTRACT(EPOCH FROM (completed_at - created_at))/3600 
    END) as avg_completion_hours
FROM tasks 
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

### 员工工作负载视图
```sql
CREATE VIEW staff_workload AS
SELECT 
    s.id,
    s.name,
    s.department,
    s.max_concurrent_tasks,
    COUNT(t.id) as current_tasks,
    ROUND(COUNT(t.id)::DECIMAL / s.max_concurrent_tasks * 100, 2) as load_percentage,
    AVG(CASE WHEN t.completed_at IS NOT NULL THEN 
        EXTRACT(EPOCH FROM (t.completed_at - t.assigned_at))/3600 
    END) as avg_task_completion_hours
FROM staff s
LEFT JOIN tasks t ON s.id = t.assignee_id AND t.status IN ('assigned', 'in_progress')
WHERE s.status = 'active'
GROUP BY s.id, s.name, s.department, s.max_concurrent_tasks;
```

## 索引优化

### 复合索引
```sql
-- 任务查询优化
CREATE INDEX idx_tasks_status_priority_deadline ON tasks(status, priority, deadline);
CREATE INDEX idx_tasks_assignee_status_created ON tasks(assignee_id, status, created_at);

-- 分配查询优化
CREATE INDEX idx_assignments_task_status_assigned ON task_assignments(task_id, status, assigned_at);

-- 审批查询优化
CREATE INDEX idx_approvals_status_type_submitted ON approvals(status, type, submitted_at);

-- 通知查询优化
CREATE INDEX idx_notifications_recipient_status_created ON notifications(recipient_id, status, created_at);
```

### 部分索引
```sql
-- 只为活跃任务创建索引
CREATE INDEX idx_active_tasks_assignee ON tasks(assignee_id) WHERE status IN ('assigned', 'in_progress');

-- 只为未读通知创建索引
CREATE INDEX idx_unread_notifications ON notifications(recipient_id, created_at) WHERE status = 'unread';

-- 只为待审批项创建索引
CREATE INDEX idx_pending_approvals ON approvals(type, submitted_at) WHERE status = 'pending';
```

## 数据库函数

### 更新任务状态函数
```sql
CREATE OR REPLACE FUNCTION update_task_status(
    p_task_id VARCHAR(50),
    p_new_status VARCHAR(30),
    p_user_id VARCHAR(50)
) RETURNS BOOLEAN AS $$
DECLARE
    old_status VARCHAR(30);
BEGIN
    -- 获取当前状态
    SELECT status INTO old_status FROM tasks WHERE id = p_task_id;
    
    -- 验证状态转换是否合法
    IF NOT is_valid_status_transition(old_status, p_new_status) THEN
        RAISE EXCEPTION '无效的状态转换: % -> %', old_status, p_new_status;
    END IF;
    
    -- 更新任务状态
    UPDATE tasks 
    SET status = p_new_status,
        updated_at = NOW(),
        started_at = CASE WHEN p_new_status = 'in_progress' AND started_at IS NULL THEN NOW() ELSE started_at END,
        completed_at = CASE WHEN p_new_status = 'completed' THEN NOW() ELSE completed_at END
    WHERE id = p_task_id;
    
    -- 记录审计日志
    INSERT INTO audit_logs (user_id, action, resource_type, resource_id, old_values, new_values)
    VALUES (p_user_id, 'status_change', 'task', p_task_id, 
            jsonb_build_object('status', old_status),
            jsonb_build_object('status', p_new_status));
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;
```

### 计算员工负载函数
```sql
CREATE OR REPLACE FUNCTION calculate_staff_load(p_staff_id VARCHAR(50))
RETURNS INTEGER AS $$
DECLARE
    current_load INTEGER;
BEGIN
    SELECT COUNT(*) INTO current_load
    FROM tasks 
    WHERE assignee_id = p_staff_id 
    AND status IN ('assigned', 'in_progress');
    
    -- 更新员工当前负载
    UPDATE staff SET current_load = current_load WHERE id = p_staff_id;
    
    RETURN current_load;
END;
$$ LANGUAGE plpgsql;
```

## 触发器

### 任务状态变更触发器
```sql
CREATE OR REPLACE FUNCTION task_status_change_trigger()
RETURNS TRIGGER AS $$
BEGIN
    -- 更新分配人负载
    IF OLD.assignee_id IS NOT NULL THEN
        PERFORM calculate_staff_load(OLD.assignee_id);
    END IF;
    
    IF NEW.assignee_id IS NOT NULL AND NEW.assignee_id != OLD.assignee_id THEN
        PERFORM calculate_staff_load(NEW.assignee_id);
    END IF;
    
    -- 发送状态变更通知
    IF OLD.status != NEW.status THEN
        INSERT INTO notifications (id, type, recipient_id, title, content, data)
        VALUES (
            'notif_' || generate_random_string(20),
            'task_status_changed',
            (SELECT user_id FROM staff WHERE id = NEW.assignee_id),
            '任务状态更新',
            '任务 "' || NEW.title || '" 状态已更新为 ' || NEW.status,
            jsonb_build_object('task_id', NEW.id, 'old_status', OLD.status, 'new_status', NEW.status)
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER task_status_change_trigger
    AFTER UPDATE ON tasks
    FOR EACH ROW
    WHEN (OLD.status IS DISTINCT FROM NEW.status OR OLD.assignee_id IS DISTINCT FROM NEW.assignee_id)
    EXECUTE FUNCTION task_status_change_trigger();
```

## 数据迁移脚本

### 初始化基础数据
```sql
-- 插入默认管理员用户
INSERT INTO users (id, username, email, password_hash, role) VALUES
('admin_001', 'admin', 'admin@company.com', '$2a$10$...', 'admin');

-- 插入系统配置
INSERT INTO system_configs (key, value, description, category) VALUES
('max_concurrent_tasks_default', '3', '默认最大并发任务数', 'assignment'),
('assignment_algorithm_default', 'hybrid', '默认分配算法', 'assignment'),
('notification_enabled', 'true', '是否启用通知', 'notification'),
('approval_required_task_creation', 'true', '任务创建是否需要审批', 'approval');

-- 插入默认审批流程
INSERT INTO approval_workflows (name, type, conditions, approvers, is_active) VALUES
('任务创建审批', 'task_creation', '{"priority": ["high"]}', '{"roles": ["manager", "admin"]}', true),
('任务完成审批', 'task_completion', '{"estimated_hours": {"gt": 40}}', '{"roles": ["manager"]}', true);
```

## 性能优化建议

1. **分区策略**: 对大表如 `audit_logs` 按时间分区
2. **索引优化**: 根据查询模式创建合适的复合索引
3. **查询优化**: 使用 EXPLAIN ANALYZE 分析慢查询
4. **连接池**: 配置合适的数据库连接池大小
5. **读写分离**: 读多写少的场景可考虑主从分离
6. **缓存策略**: 热点数据使用 Redis 缓存

## 备份策略

1. **全量备份**: 每日凌晨进行全量备份
2. **增量备份**: 每小时进行 WAL 归档
3. **备份验证**: 定期验证备份文件完整性
4. **异地备份**: 重要数据异地存储
