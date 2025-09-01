-- 入职审批工作流定义初始化脚本（修正版）

-- 插入入职审批工作流定义
INSERT INTO workflow_definitions (
    workflow_id, name, description, version, nodes, edges, variables, is_active, created_at, updated_at
) VALUES (
    'onboarding-approval-v1',
    '员工入职审批流程',
    '新员工入职审批工作流，包含HR初审、部门经理审批、试用期确认等步骤',
    '1.0.0',
    '[
        {
            "id": "start",
            "type": "start",
            "name": "开始",
            "description": "HR提交入职申请",
            "properties": {}
        },
        {
            "id": "hr_review",
            "type": "approval",
            "name": "HR初审",
            "description": "HR对入职申请进行初步审核",
            "properties": {
                "assignee_type": "role",
                "assignee_value": "hr",
                "timeout": 86400,
                "timeout_action": "auto_reject"
            }
        },
        {
            "id": "department_manager_approval",
            "type": "approval", 
            "name": "部门经理审批",
            "description": "部门经理审批员工入职",
            "properties": {
                "assignee_type": "multiple",
                "assignees": [
                    {
                        "type": "department_manager",
                        "value": "${department_id}"
                    },
                    {
                        "type": "role",
                        "value": "admin"
                    },
                    {
                        "type": "role",
                        "value": "superadmin"
                    }
                ],
                "timeout": 172800,
                "timeout_action": "escalate"
            }
        },
        {
            "id": "hr_confirm",
            "type": "approval",
            "name": "HR确认入职",
            "description": "HR确认员工正式入职",
            "properties": {
                "assignee_type": "role",
                "assignee_value": "hr",
                "timeout": 86400,
                "timeout_action": "auto_approve"
            }
        },
        {
            "id": "probation_review",
            "type": "approval",
            "name": "试用期评估",
            "description": "试用期结束后的评估审批",
            "properties": {
                "assignee_type": "multiple",
                "assignees": [
                    {
                        "type": "department_manager",
                        "value": "${department_id}"
                    },
                    {
                        "type": "role",
                        "value": "admin"
                    },
                    {
                        "type": "role",
                        "value": "superadmin"
                    }
                ],
                "timeout": 259200,
                "timeout_action": "auto_approve",
                "delay": "${probation_period}"
            }
        },
        {
            "id": "approved",
            "type": "end",
            "name": "审批通过",
            "description": "入职审批完成，员工转为正式员工",
            "properties": {
                "result": "approved"
            }
        },
        {
            "id": "rejected",
            "type": "end", 
            "name": "审批拒绝",
            "description": "入职审批被拒绝",
            "properties": {
                "result": "rejected"
            }
        }
    ]',
    '[
        {
            "id": "start_to_hr_review",
            "from": "start",
            "to": "hr_review",
            "condition": null
        },
        {
            "id": "hr_review_approved",
            "from": "hr_review",
            "to": "department_manager_approval",
            "condition": "decision == \"approved\""
        },
        {
            "id": "hr_review_rejected",
            "from": "hr_review", 
            "to": "rejected",
            "condition": "decision == \"rejected\""
        },
        {
            "id": "dept_manager_approved",
            "from": "department_manager_approval",
            "to": "hr_confirm",
            "condition": "decision == \"approved\""
        },
        {
            "id": "dept_manager_rejected",
            "from": "department_manager_approval",
            "to": "rejected",
            "condition": "decision == \"rejected\""
        },
        {
            "id": "hr_confirm_approved",
            "from": "hr_confirm",
            "to": "probation_review",
            "condition": "decision == \"approved\""
        },
        {
            "id": "hr_confirm_rejected",
            "from": "hr_confirm",
            "to": "rejected",
            "condition": "decision == \"rejected\""
        },
        {
            "id": "probation_approved",
            "from": "probation_review",
            "to": "approved",
            "condition": "decision == \"approved\""
        },
        {
            "id": "probation_rejected",
            "from": "probation_review",
            "to": "rejected",
            "condition": "decision == \"rejected\""
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
            "required": true,
            "description": "部门ID"
        },
        "position_id": {
            "type": "integer",
            "required": true,
            "description": "职位ID"
        },
        "expected_date": {
            "type": "date",
            "required": true,
            "description": "预期入职日期"
        },
        "probation_period": {
            "type": "integer",
            "required": false,
            "default": 90,
            "description": "试用期天数"
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

-- 插入简化版入职审批流程（仅部门经理审批）
INSERT INTO workflow_definitions (
    workflow_id, name, description, version, nodes, edges, variables, is_active, created_at, updated_at
) VALUES (
    'onboarding-simple-approval-v1',
    '简化入职审批流程',
    '简化的入职审批流程，仅需部门经理审批',
    '1.0.0',
    '[
        {
            "id": "start",
            "type": "start",
            "name": "开始",
            "description": "提交入职申请",
            "properties": {}
        },
        {
            "id": "department_manager_approval",
            "type": "approval",
            "name": "部门经理审批",
            "description": "部门经理审批员工入职",
            "properties": {
                "assignee_type": "multiple",
                "assignees": [
                    {
                        "type": "department_manager",
                        "value": "${department_id}"
                    },
                    {
                        "type": "role",
                        "value": "admin"
                    },
                    {
                        "type": "role",
                        "value": "superadmin"
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
            "properties": {
                "result": "approved"
            }
        },
        {
            "id": "rejected",
            "type": "end",
            "name": "审批拒绝", 
            "description": "入职审批被拒绝",
            "properties": {
                "result": "rejected"
            }
        }
    ]',
    '[
        {
            "id": "start_to_approval",
            "from": "start",
            "to": "department_manager_approval",
            "condition": null
        },
        {
            "id": "approved_path",
            "from": "department_manager_approval",
            "to": "approved",
            "condition": "decision == \"approved\""
        },
        {
            "id": "rejected_path",
            "from": "department_manager_approval",
            "to": "rejected",
            "condition": "decision == \"rejected\""
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
            "required": true,
            "description": "部门ID"
        },
        "position_id": {
            "type": "integer",
            "required": false,
            "description": "职位ID"
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
WHERE workflow_id IN ('onboarding-approval-v1', 'onboarding-simple-approval-v1');
