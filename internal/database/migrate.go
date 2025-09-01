package database

import (
	"fmt"
)

// Migrate 执行数据库迁移
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("数据库未连接")
	}

	// 自动迁移所有模型 (GORM会自动处理重复迁移)
	models := GetAllModels()
	if err := DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 创建索引 (已经有重复检查逻辑)
	if err := createIndexes(); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	// 插入初始数据 (只在数据为空时插入) - 权限角色初始化已移至bootstrap服务
	if err := seedData(); err != nil {
		return fmt.Errorf("插入初始数据失败: %w", err)
	}

	return nil
}

// createIndexes 创建额外的索引
func createIndexes() error {
	indexes := []struct {
		name  string
		table string
		sql   string
	}{
		{"idx_tasks_status_priority", "tasks", "CREATE INDEX idx_tasks_status_priority ON tasks(status, priority)"},
		{"idx_tasks_assignee_status", "tasks", "CREATE INDEX idx_tasks_assignee_status ON tasks(assignee_id, status)"},
		{"idx_tasks_creator_created", "tasks", "CREATE INDEX idx_tasks_creator_created ON tasks(creator_id, created_at)"},
		{"idx_tasks_due_date", "tasks", "CREATE INDEX idx_tasks_due_date ON tasks(due_date)"},
		{"idx_assignments_task_status", "assignments", "CREATE INDEX idx_assignments_task_status ON assignments(task_id, status)"},
		{"idx_assignments_assignee_status", "assignments", "CREATE INDEX idx_assignments_assignee_status ON assignments(assignee_id, status)"},
		{"idx_employees_status", "employees", "CREATE INDEX idx_employees_status ON employees(status)"},
		{"idx_notifications_user_read", "notifications", "CREATE INDEX idx_notifications_user_read ON notifications(user_id, is_read)"},
		{"idx_task_notifications_recipient_status", "task_notifications", "CREATE INDEX idx_task_notifications_recipient_status ON task_notifications(recipient_id, status)"},
		{"idx_task_notifications_task_type", "task_notifications", "CREATE INDEX idx_task_notifications_task_type ON task_notifications(task_id, type)"},
		{"idx_task_notifications_created_at", "task_notifications", "CREATE INDEX idx_task_notifications_created_at ON task_notifications(created_at)"},
	}

	for _, idx := range indexes {
		// 检查索引是否已存在
		var count int64
		err := DB.Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?", 
			idx.table, idx.name).Scan(&count).Error
		if err != nil {
			return fmt.Errorf("检查索引 %s 失败: %w", idx.name, err)
		}

		// 如果索引不存在，则创建
		if count == 0 {
			if err := DB.Exec(idx.sql).Error; err != nil {
				return fmt.Errorf("创建索引 %s 失败: %w", idx.name, err)
			}
		}
	}

	return nil
}

// seedData 插入初始数据
func seedData() error {
	// 注意：权限和角色的初始化现在由 bootstrap 服务处理
	// 这里只初始化其他基础数据

	// 创建默认技能 (使用FirstOrCreate避免重复)
	skillData := []Skill{
		{Name: "Java开发", Category: "编程语言", Description: "Java编程语言开发技能"},
		{Name: "Go开发", Category: "编程语言", Description: "Go编程语言开发技能"},
		{Name: "Python开发", Category: "编程语言", Description: "Python编程语言开发技能"},
		{Name: "前端开发", Category: "Web开发", Description: "前端Web开发技能"},
		{Name: "数据库设计", Category: "数据库", Description: "数据库设计和优化技能"},
		{Name: "项目管理", Category: "管理", Description: "项目管理和协调技能"},
		{Name: "测试", Category: "质量保证", Description: "软件测试技能"},
		{Name: "运维", Category: "系统运维", Description: "系统运维和部署技能"},
	}

	for _, skillInfo := range skillData {
		var skill Skill
		if err := DB.FirstOrCreate(&skill, Skill{Name: skillInfo.Name}, skillInfo).Error; err != nil {
			return fmt.Errorf("创建技能 %s 失败: %w", skillInfo.Name, err)
		}
	}

	// 创建默认系统配置 (使用FirstOrCreate避免重复)
	configData := []SystemConfig{
		{Key: "system.name", Value: "任务分配管理系统", Type: "string", Category: "基础", Description: "系统名称", IsPublic: true},
		{Key: "system.version", Value: "1.0.0", Type: "string", Category: "基础", Description: "系统版本", IsPublic: true},
		{Key: "task.auto_assign", Value: "true", Type: "bool", Category: "任务", Description: "是否启用自动分配", IsPublic: false},
		{Key: "task.default_priority", Value: "medium", Type: "string", Category: "任务", Description: "默认任务优先级", IsPublic: false},
		{Key: "employee.max_tasks", Value: "5", Type: "int", Category: "员工", Description: "员工最大任务数", IsPublic: false},
		{Key: "notification.email_enabled", Value: "false", Type: "bool", Category: "通知", Description: "是否启用邮件通知", IsPublic: false},
	}

	for _, configInfo := range configData {
		var config SystemConfig
		if err := DB.FirstOrCreate(&config, SystemConfig{Key: configInfo.Key}, configInfo).Error; err != nil {
			return fmt.Errorf("创建配置 %s 失败: %w", configInfo.Key, err)
		}
	}

	return nil
}

// DropTables 删除所有表 (谨慎使用)
func DropTables() error {
	if DB == nil {
		return fmt.Errorf("数据库未连接")
	}

	models := GetAllModels()
	for i := len(models) - 1; i >= 0; i-- {
		if err := DB.Migrator().DropTable(models[i]); err != nil {
			return fmt.Errorf("删除表失败: %w", err)
		}
	}

	return nil
}

// ResetDatabase 重置数据库 (删除所有表并重新创建)
func ResetDatabase() error {
	if err := DropTables(); err != nil {
		return err
	}

	return Migrate()
}
