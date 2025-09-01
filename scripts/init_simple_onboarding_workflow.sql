-- 简化入职审批工作流定义初始化脚本

-- 插入简化入职审批工作流定义
INSERT INTO workflow_definitions (
    workflow_id, name, description, version, nodes, edges, variables, is_active, created_at, updated_at
) VALUES (
    'simple-onboarding-approval-v1',
    '简化入职审批流程',
    '仅需管理员或超级管理员审批的简化入职流程',
    '1.0.0',
    '[
        {
            "id": "start",
            "type": "start",
            "name": "开始",
            "description": "提交入职申请",
            "config": {}
        },
        {
            "id": "manager_approval",
            "type": "approval",
            "name": "管理员审批",
            "description": "管理员或超级管理员审批员工入职",
            "config": {
                "assignee_type": "multiple",
                "assignees": [
                    {
                        "type": "role",
                        "value": "admin"
                    },
                    {
                        "type": "role",
                        "value": "super_admin"
                    }
                ],
                "timeout": 172800,
                "timeout_action": "auto_approve"
            }
        },
        {
            "id": "approved",
            "type": "end",
            "name": "审批通过",
            "description": "入职审批完成",
            "config": {
                "result": "approved"
            }
        },
        {
            "id": "rejected",
            "type": "end",
            "name": "审批拒绝",
            "description": "入职审批被拒绝",
            "config": {
                "result": "rejected"
            }
        }
    ]',
    '[
        {
            "id": "start_to_approval",
            "from": "start",
            "to": "manager_approval"
        },
        {
            "id": "approved_path",
            "from": "manager_approval",
            "to": "approved",
            "condition": "decision == \'approved\'"
        },
        {
            "id": "rejected_path",
            "from": "manager_approval",
            "to": "rejected",
            "condition": "decision == \'rejected\'"
        }
    ]',
    '{
        "employee_id": {
            "type": "integer",
            "required": true,
            "description": "员工ID"
        },
        "department_id": {
            "type": "integer",
            "required": false,
            "description": "部门ID"
        }
    }',
    true,
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE
    nodes = VALUES(nodes),
    edges = VALUES(edges),
    variables = VALUES(variables),
    updated_at = NOW();

-- 验证插入结果
SELECT workflow_id, name, version, is_active FROM workflow_definitions 
WHERE workflow_id = 'simple-onboarding-approval-v1';
