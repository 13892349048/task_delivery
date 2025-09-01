package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// WorkflowDefinition 工作流定义数据库模型
type WorkflowDefinition struct {
	BaseModel
	WorkflowID  string    `gorm:"column:workflow_id;uniqueIndex;size:100;not null" json:"workflow_id"`
	Name        string    `gorm:"column:name;size:200;not null" json:"name"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	Version     string    `gorm:"column:version;size:50;not null" json:"version"`
	Nodes       JSONField `gorm:"column:nodes;type:json" json:"nodes"`
	Edges       JSONField `gorm:"column:edges;type:json" json:"edges"`
	Variables   JSONField `gorm:"column:variables;type:json" json:"variables"`
	IsActive    bool      `gorm:"column:is_active;default:true" json:"is_active"`
}

// TableName 指定表名
func (WorkflowDefinition) TableName() string {
	return "workflow_definitions"
}

// WorkflowInstance 工作流实例数据库模型
type WorkflowInstance struct {
	BaseModel
	InstanceID   string    `gorm:"column:instance_id;uniqueIndex;size:100;not null" json:"instance_id"`
	WorkflowID   string    `gorm:"column:workflow_id;size:100;not null;index" json:"workflow_id"`
	BusinessID   string    `gorm:"column:business_id;size:100;not null;index" json:"business_id"`
	BusinessType string    `gorm:"column:business_type;size:50;not null;index" json:"business_type"`
	Status       string    `gorm:"column:status;size:20;not null;index" json:"status"`
	CurrentNodes JSONField `gorm:"column:current_nodes;type:json" json:"current_nodes"`
	Variables    JSONField `gorm:"column:variables;type:json" json:"variables"`
	StartedBy    uint      `gorm:"column:started_by;not null;index" json:"started_by"`
	StartedAt    time.Time `gorm:"column:started_at;not null" json:"started_at"`
	CompletedAt  *time.Time `gorm:"column:completed_at" json:"completed_at"`
}

// TableName 指定表名
func (WorkflowInstance) TableName() string {
	return "workflow_instances"
}

// WorkflowExecutionHistory 工作流执行历史数据库模型
type WorkflowExecutionHistory struct {
	BaseModel
	HistoryID  string    `gorm:"column:history_id;uniqueIndex;size:100;not null" json:"history_id"`
	InstanceID string    `gorm:"column:instance_id;size:100;not null;index" json:"instance_id"`
	NodeID     string    `gorm:"column:node_id;size:100;not null" json:"node_id"`
	NodeName   string    `gorm:"column:node_name;size:200;not null" json:"node_name"`
	Action     string    `gorm:"column:action;size:50;not null" json:"action"`
	Result     string    `gorm:"column:result;size:50;not null" json:"result"`
	Comment    string    `gorm:"column:comment;type:text" json:"comment"`
	Variables  JSONField `gorm:"column:variables;type:json" json:"variables"`
	ExecutedBy uint      `gorm:"column:executed_by;not null;index" json:"executed_by"`
	ExecutedAt time.Time `gorm:"column:executed_at;not null" json:"executed_at"`
	Duration   int64     `gorm:"column:duration;not null" json:"duration"` // 毫秒
}

// TableName 指定表名
func (WorkflowExecutionHistory) TableName() string {
	return "workflow_execution_histories"
}

// WorkflowPendingApproval 待审批任务数据库模型
type WorkflowPendingApproval struct {
	BaseModel
	InstanceID     string    `gorm:"column:instance_id;size:100;not null;index" json:"instance_id"`
	WorkflowName   string    `gorm:"column:workflow_name;size:200;not null" json:"workflow_name"`
	NodeID         string    `gorm:"column:node_id;size:100;not null" json:"node_id"`
	NodeName       string    `gorm:"column:node_name;size:200;not null" json:"node_name"`
	BusinessID     string    `gorm:"column:business_id;size:100;not null;index" json:"business_id"`
	BusinessType   string    `gorm:"column:business_type;size:50;not null;index" json:"business_type"`
	BusinessData   JSONField `gorm:"column:business_data;type:json" json:"business_data"`
	Priority       int       `gorm:"column:priority;not null;default:1" json:"priority"`
	AssignedTo     uint      `gorm:"column:assigned_to;not null;index" json:"assigned_to"`
	Deadline       *time.Time `gorm:"column:deadline" json:"deadline"`
	CanDelegate    bool      `gorm:"column:can_delegate;default:false" json:"can_delegate"`
	RequiredActions JSONField `gorm:"column:required_actions;type:json" json:"required_actions"`
	IsCompleted    bool      `gorm:"column:is_completed;default:false;index" json:"is_completed"`
}

// TableName 指定表名
func (WorkflowPendingApproval) TableName() string {
	return "workflow_pending_approvals"
}

// JSONField 自定义JSON字段类型，可以存储任意JSON数据
type JSONField struct {
	Data interface{}
}

// Value 实现driver.Valuer接口
func (j JSONField) Value() (driver.Value, error) {
	if j.Data == nil {
		return nil, nil
	}
	return json.Marshal(j.Data)
}

// Scan 实现sql.Scanner接口
func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		j.Data = nil
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan into JSONField")
	}
	
	// 先尝试解析为通用interface{}
	var data interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return err
	}
	j.Data = data
	return nil
}

// MarshalJSON 实现json.Marshaler接口
func (j JSONField) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Data)
}

// UnmarshalJSON 实现json.Unmarshaler接口
func (j *JSONField) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &j.Data)
}
