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
