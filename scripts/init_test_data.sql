-- TaskManage 测试数据初始化脚本
-- 用于企业级API测试的基础数据

-- 清理现有测试数据
DELETE FROM employee_skills WHERE 1=1;
DELETE FROM assignments WHERE 1=1;
DELETE FROM tasks WHERE 1=1;
DELETE FROM employees WHERE 1=1;
DELETE FROM skills WHERE 1=1;
DELETE FROM user_roles WHERE 1=1;
DELETE FROM role_permissions WHERE 1=1;
DELETE FROM permissions WHERE 1=1;
DELETE FROM roles WHERE 1=1;
DELETE FROM users WHERE 1=1;

-- 重置自增ID
ALTER TABLE users AUTO_INCREMENT = 1;
ALTER TABLE roles AUTO_INCREMENT = 1;
ALTER TABLE permissions AUTO_INCREMENT = 1;
ALTER TABLE employees AUTO_INCREMENT = 1;
ALTER TABLE skills AUTO_INCREMENT = 1;
ALTER TABLE tasks AUTO_INCREMENT = 1;
ALTER TABLE assignments AUTO_INCREMENT = 1;

-- 插入基础权限数据
INSERT INTO permissions (name, resource, action, description, created_at, updated_at) VALUES
('user.create', 'user', 'create', '创建用户', NOW(), NOW()),
('user.read', 'user', 'read', '查看用户', NOW(), NOW()),
('user.update', 'user', 'update', '更新用户', NOW(), NOW()),
('user.delete', 'user', 'delete', '删除用户', NOW(), NOW()),
('employee.create', 'employee', 'create', '创建员工', NOW(), NOW()),
('employee.read', 'employee', 'read', '查看员工', NOW(), NOW()),
('employee.update', 'employee', 'update', '更新员工', NOW(), NOW()),
('employee.delete', 'employee', 'delete', '删除员工', NOW(), NOW()),
('skill.create', 'skill', 'create', '创建技能', NOW(), NOW()),
('skill.read', 'skill', 'read', '查看技能', NOW(), NOW()),
('skill.update', 'skill', 'update', '更新技能', NOW(), NOW()),
('skill.delete', 'skill', 'delete', '删除技能', NOW(), NOW()),
('task.create', 'task', 'create', '创建任务', NOW(), NOW()),
('task.read', 'task', 'read', '查看任务', NOW(), NOW()),
('task.update', 'task', 'update', '更新任务', NOW(), NOW()),
('task.delete', 'task', 'delete', '删除任务', NOW(), NOW());

-- 插入角色数据
INSERT INTO roles (name, description, created_at, updated_at) VALUES
('admin', '系统管理员', NOW(), NOW()),
('manager', '部门经理', NOW(), NOW()),
('employee', '普通员工', NOW(), NOW());

-- 分配权限给角色
-- 管理员拥有所有权限
INSERT INTO role_permissions (role_id, permission_id) 
SELECT 1, id FROM permissions;

-- 经理拥有员工和任务相关权限
INSERT INTO role_permissions (role_id, permission_id) 
SELECT 2, id FROM permissions WHERE resource IN ('employee', 'task', 'skill');

-- 普通员工只有查看权限
INSERT INTO role_permissions (role_id, permission_id) 
SELECT 3, id FROM permissions WHERE action = 'read';

-- 插入测试用户数据
INSERT INTO users (username, email, password, real_name, phone, status, created_at, updated_at) VALUES
('admin', 'admin@taskmanage.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '系统管理员', '13800000001', 'active', NOW(), NOW()),
('manager1', 'manager1@taskmanage.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '张经理', '13800000002', 'active', NOW(), NOW()),
('manager2', 'manager2@taskmanage.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '李经理', '13800000003', 'active', NOW(), NOW()),
('emp1', 'emp1@taskmanage.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '王小明', '13800000004', 'active', NOW(), NOW()),
('emp2', 'emp2@taskmanage.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '赵小红', '13800000005', 'active', NOW(), NOW()),
('emp3', 'emp3@taskmanage.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '刘小强', '13800000006', 'active', NOW(), NOW());

-- 分配角色给用户
INSERT INTO user_roles (user_id, role_id) VALUES
(1, 1), -- admin -> admin role
(2, 2), -- manager1 -> manager role  
(3, 2), -- manager2 -> manager role
(4, 3), -- emp1 -> employee role
(5, 3), -- emp2 -> employee role
(6, 3); -- emp3 -> employee role

-- 插入员工数据
INSERT INTO employees (user_id, employee_no, department, position, level, status, max_tasks, current_tasks, created_at, updated_at) VALUES
(2, 'MGR001', '开发部', '部门经理', '高级', 'active', 8, 2, NOW(), NOW()),
(3, 'MGR002', '测试部', '部门经理', '高级', 'active', 6, 1, NOW(), NOW()),
(4, 'EMP001', '开发部', '高级工程师', '高级', 'active', 5, 3, NOW(), NOW()),
(5, 'EMP002', '开发部', '中级工程师', '中级', 'busy', 4, 4, NOW(), NOW()),
(6, 'EMP003', '测试部', '测试工程师', '中级', 'available', 3, 1, NOW(), NOW());

-- 插入技能数据
INSERT INTO skills (name, category, description, tags, created_at, updated_at) VALUES
('Go语言', '编程语言', 'Go语言开发技能', '["golang", "backend", "programming"]', NOW(), NOW()),
('JavaScript', '编程语言', 'JavaScript开发技能', '["javascript", "frontend", "programming"]', NOW(), NOW()),
('MySQL', '数据库', 'MySQL数据库管理', '["mysql", "database", "sql"]', NOW(), NOW()),
('Redis', '数据库', 'Redis缓存技术', '["redis", "cache", "nosql"]', NOW(), NOW()),
('Docker', '运维工具', 'Docker容器技术', '["docker", "container", "devops"]', NOW(), NOW()),
('项目管理', '管理技能', '项目管理和团队协调', '["management", "leadership", "planning"]', NOW(), NOW()),
('测试设计', '测试技能', '测试用例设计和执行', '["testing", "qa", "automation"]', NOW(), NOW()),
('API设计', '架构技能', 'RESTful API设计', '["api", "rest", "architecture"]', NOW(), NOW());

-- 分配技能给员工
INSERT INTO employee_skills (employee_id, skill_id, level, created_at, updated_at) VALUES
-- 开发部经理
(1, 1, 5, NOW(), NOW()), -- Go语言 - 专家级
(1, 3, 4, NOW(), NOW()), -- MySQL - 高级
(1, 6, 5, NOW(), NOW()), -- 项目管理 - 专家级
(1, 8, 4, NOW(), NOW()), -- API设计 - 高级

-- 测试部经理  
(2, 7, 5, NOW(), NOW()), -- 测试设计 - 专家级
(2, 3, 3, NOW(), NOW()), -- MySQL - 中级
(2, 6, 4, NOW(), NOW()), -- 项目管理 - 高级

-- 高级工程师
(3, 1, 4, NOW(), NOW()), -- Go语言 - 高级
(3, 2, 3, NOW(), NOW()), -- JavaScript - 中级
(3, 3, 4, NOW(), NOW()), -- MySQL - 高级
(3, 4, 3, NOW(), NOW()), -- Redis - 中级
(3, 5, 3, NOW(), NOW()), -- Docker - 中级

-- 中级工程师
(4, 1, 3, NOW(), NOW()), -- Go语言 - 中级
(4, 3, 3, NOW(), NOW()), -- MySQL - 中级
(4, 5, 2, NOW(), NOW()), -- Docker - 初级

-- 测试工程师
(5, 7, 4, NOW(), NOW()), -- 测试设计 - 高级
(5, 2, 3, NOW(), NOW()), -- JavaScript - 中级
(5, 3, 2, NOW(), NOW()); -- MySQL - 初级

-- 插入测试任务数据
INSERT INTO tasks (title, description, priority, status, estimated_hours, actual_hours, due_date, created_by, assigned_to, created_at, updated_at) VALUES
('用户认证模块开发', '实现JWT认证和权限控制系统', 'high', 'in_progress', 40, 25, DATE_ADD(NOW(), INTERVAL 7 DAY), 1, 4, NOW(), NOW()),
('API文档编写', '编写完整的API接口文档', 'medium', 'pending', 16, 0, DATE_ADD(NOW(), INTERVAL 5 DAY), 2, 4, NOW(), NOW()),
('数据库性能优化', '优化查询性能和索引设计', 'high', 'pending', 24, 0, DATE_ADD(NOW(), INTERVAL 10 DAY), 1, 4, NOW(), NOW()),
('前端界面开发', '开发用户管理界面', 'medium', 'in_progress', 32, 20, DATE_ADD(NOW(), INTERVAL 14 DAY), 2, 5, NOW(), NOW()),
('单元测试编写', '为核心模块编写单元测试', 'medium', 'pending', 20, 0, DATE_ADD(NOW(), INTERVAL 8 DAY), 3, 6, NOW(), NOW()),
('部署脚本开发', '编写自动化部署脚本', 'low', 'pending', 12, 0, DATE_ADD(NOW(), INTERVAL 12 DAY), 1, 5, NOW(), NOW());

-- 插入任务分配记录
INSERT INTO assignments (task_id, employee_id, assigned_by, assigned_at, status, notes, created_at, updated_at) VALUES
(1, 4, 1, NOW(), 'active', '负责用户认证核心功能开发', NOW(), NOW()),
(2, 4, 2, NOW(), 'active', '编写技术文档', NOW(), NOW()),
(3, 4, 1, NOW(), 'pending', '数据库优化任务', NOW(), NOW()),
(4, 5, 2, NOW(), 'active', '前端开发任务', NOW(), NOW()),
(5, 6, 3, NOW(), 'pending', '测试用例开发', NOW(), NOW()),
(6, 5, 1, NOW(), 'pending', '部署自动化', NOW(), NOW());

-- 验证数据插入
SELECT 'Users' as TableName, COUNT(*) as RecordCount FROM users
UNION ALL
SELECT 'Roles', COUNT(*) FROM roles  
UNION ALL
SELECT 'Permissions', COUNT(*) FROM permissions
UNION ALL
SELECT 'Employees', COUNT(*) FROM employees
UNION ALL
SELECT 'Skills', COUNT(*) FROM skills
UNION ALL
SELECT 'Employee Skills', COUNT(*) FROM employee_skills
UNION ALL
SELECT 'Tasks', COUNT(*) FROM tasks
UNION ALL
SELECT 'Assignments', COUNT(*) FROM assignments;

-- 创建任务通知表
CREATE TABLE IF NOT EXISTS task_notifications (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    type VARCHAR(50) NOT NULL COMMENT '通知类型',
    title VARCHAR(200) NOT NULL COMMENT '通知标题',
    content TEXT COMMENT '通知内容',
    recipient_id BIGINT UNSIGNED NOT NULL COMMENT '接收者ID',
    sender_id BIGINT UNSIGNED NULL COMMENT '发送者ID',
    task_id BIGINT UNSIGNED NULL COMMENT '关联任务ID',
    priority VARCHAR(20) DEFAULT 'medium' COMMENT '优先级',
    status VARCHAR(20) DEFAULT 'unread' COMMENT '状态',
    action_type VARCHAR(50) NULL COMMENT '操作类型',
    action_data JSON NULL COMMENT '操作数据',
    expires_at TIMESTAMP NULL COMMENT '过期时间',
    read_at TIMESTAMP NULL COMMENT '已读时间',
    action_at TIMESTAMP NULL COMMENT '操作时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    
    INDEX idx_recipient_id (recipient_id),
    INDEX idx_sender_id (sender_id),
    INDEX idx_task_id (task_id),
    INDEX idx_status (status),
    INDEX idx_type (type),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务通知表';

-- 创建任务通知操作记录表
CREATE TABLE IF NOT EXISTS task_notification_actions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    notification_id BIGINT UNSIGNED NOT NULL COMMENT '通知ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    action_type VARCHAR(50) NOT NULL COMMENT '操作类型',
    action_data JSON NULL COMMENT '操作数据',
    reason TEXT NULL COMMENT '操作理由',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX idx_notification_id (notification_id),
    INDEX idx_user_id (user_id),
    INDEX idx_action_type (action_type),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (notification_id) REFERENCES task_notifications(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务通知操作记录表';
