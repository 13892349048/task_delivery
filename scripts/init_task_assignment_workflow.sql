-- 初始化任务分配审批工作流定义
INSERT INTO workflow_definitions (
    workflow_id, 
    name, 
    description, 
    version, 
    nodes, 
    edges, 
    variables, 
    is_active, 
    created_at, 
    updated_at
) VALUES (
    'task_assignment_approval',
    '任务分配审批',
    '任务分配需要经过审批流程',
    '1.0',
    JSON_ARRAY(
        JSON_OBJECT(
            'id', 'start',
            'type', 'start',
            'name', '开始',
            'description', '流程开始',
            'config', JSON_OBJECT(),
            'position', JSON_OBJECT('x', 100, 'y', 100)
        ),
        JSON_OBJECT(
            'id', 'approval_node',
            'type', 'approval',
            'name', '任务分配审批',
            'description', '需要主管审批任务分配',
            'config', JSON_OBJECT(
                'assignees', JSON_ARRAY(
                    JSON_OBJECT(
                        'type', 'manager',
                        'value', '',
                        'backup', JSON_ARRAY()
                    ),
                    JSON_OBJECT(
                        'type', 'role',
                        'value', 'admin',
                        'backup', JSON_ARRAY()
                    ),
                    JSON_OBJECT(
                        'type', 'role',
                        'value', 'superadmin',
                        'backup', JSON_ARRAY()
                    )
                ),
                'approval_type', 'any',
                'deadline', NULL,
                'auto_approve', false,
                'can_delegate', true,
                'can_return', true,
                'priority', 2
            ),
            'position', JSON_OBJECT('x', 300, 'y', 100)
        ),
        JSON_OBJECT(
            'id', 'end',
            'type', 'end',
            'name', '结束',
            'description', '流程结束',
            'config', JSON_OBJECT(),
            'position', JSON_OBJECT('x', 500, 'y', 100)
        )
    ),
    JSON_ARRAY(
        JSON_OBJECT(
            'id', 'edge1',
            'from', 'start',
            'to', 'approval_node',
            'condition', '',
            'config', JSON_OBJECT()
        ),
        JSON_OBJECT(
            'id', 'edge2',
            'from', 'approval_node',
            'to', 'end',
            'condition', '',
            'config', JSON_OBJECT()
        )
    ),
    JSON_OBJECT(
        'default_priority', 'medium',
        'auto_complete', true
    ),
    true,
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE
    nodes = VALUES(nodes),
    edges = VALUES(edges),
    variables = VALUES(variables),
    updated_at = NOW();

-- 确保有测试用户数据用于审批
INSERT IGNORE INTO users (id, username, email, password_hash, real_name, phone, status, created_at, updated_at) VALUES
(1, 'admin', 'admin@example.com', '$2a$10$hash', '管理员', '13800000001', 'active', NOW(), NOW()),
(2, 'manager', 'manager@example.com', '$2a$10$hash', '经理', '13800000002', 'active', NOW(), NOW()),
(3, 'employee', 'employee@example.com', '$2a$10$hash', '员工', '13800000003', 'active', NOW(), NOW()),
(4, 'requester', 'requester@example.com', '$2a$10$hash', '申请人', '13800000004', 'active', NOW(), NOW());

-- 确保有员工数据
INSERT IGNORE INTO employees (id, user_id, department, position, status, max_tasks, current_tasks, created_at, updated_at) VALUES
(1, 1, 'IT', '管理员', 'active', 10, 0, NOW(), NOW()),
(2, 2, 'IT', '经理', 'active', 8, 0, NOW(), NOW()),
(3, 3, 'IT', '开发工程师', 'active', 5, 0, NOW(), NOW()),
(4, 4, 'IT', '产品经理', 'active', 6, 0, NOW(), NOW());

-- 创建测试任务
INSERT IGNORE INTO tasks (id, title, description, priority, status, due_date, created_by, created_at, updated_at) VALUES
(7, '测试任务分配审批', '这是一个需要审批的任务分配测试', 'high', 'pending', DATE_ADD(NOW(), INTERVAL 7 DAY), 4, NOW(), NOW());

SELECT 'Task assignment approval workflow initialized successfully' as message;
