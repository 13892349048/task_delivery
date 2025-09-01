package workflow

import (
	"context"
	"errors"
	"time"
)

// 错误定义
var (
	ErrWorkflowServiceNotReady = errors.New("workflow service not ready")
)

// WorkflowEngine 审批流程引擎接口
type WorkflowEngine interface {
	// StartWorkflow 启动审批流程
	StartWorkflow(ctx context.Context, req *StartWorkflowRequest) (*WorkflowInstance, error)

	// ProcessApproval 处理审批决策
	ProcessApproval(ctx context.Context, req *ApprovalRequest) (*ApprovalResult, error)

	// GetWorkflowInstance 获取流程实例
	GetWorkflowInstance(ctx context.Context, instanceID string) (*WorkflowInstance, error)

	// GetPendingApprovals 获取待审批任务
	GetPendingApprovals(ctx context.Context, userID uint) ([]*PendingApproval, error)

	// CancelWorkflow 取消流程
	CancelWorkflow(ctx context.Context, instanceID string, reason string) error
}

// WorkflowDefinition 流程定义
type WorkflowDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Nodes       []WorkflowNode         `json:"nodes"`
	Edges       []WorkflowEdge         `json:"edges"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	IsActive    bool                   `json:"is_active"`
}

// GetStartNode 获取开始节点
func (wd *WorkflowDefinition) GetStartNode() *WorkflowNode {
	for i := range wd.Nodes {
		if wd.Nodes[i].Type == NodeTypeStart {
			return &wd.Nodes[i]
		}
	}
	return nil
}

// WorkflowNode 流程节点
type WorkflowNode struct {
	ID          string                 `json:"id"`
	Type        NodeType               `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Position    NodePosition           `json:"position,omitempty"`
}

// WorkflowEdge 流程边（连接）
type WorkflowEdge struct {
	ID        string                 `json:"id"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Condition string                 `json:"condition,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

// NodeType 节点类型
type NodeType string

const (
	NodeTypeStart     NodeType = "start"     // 开始节点
	NodeTypeEnd       NodeType = "end"       // 结束节点
	NodeTypeApproval  NodeType = "approval"  // 审批节点
	NodeTypeCondition NodeType = "condition" // 条件节点
	NodeTypeParallel  NodeType = "parallel"  // 并行节点
	NodeTypeJoin      NodeType = "join"      // 汇聚节点
	NodeTypeScript    NodeType = "script"    // 脚本节点
	NodeTypeNotify    NodeType = "notify"    // 通知节点
)

// NodePosition 节点位置（用于流程图显示）
type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// WorkflowInstance 流程实例
type WorkflowInstance struct {
	ID           string                 `json:"id"`
	WorkflowID   string                 `json:"workflow_id"`
	BusinessID   string                 `json:"business_id"`   // 业务对象ID（如任务ID）
	BusinessType string                 `json:"business_type"` // 业务类型（如task_assignment）
	Status       InstanceStatus         `json:"status"`
	CurrentNodes []string               `json:"current_nodes"` // 当前活跃节点
	Variables    map[string]interface{} `json:"variables"`
	StartedBy    uint                   `json:"started_by"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	History      []ExecutionHistory     `json:"history"`
}

// InstanceStatus 实例状态
type InstanceStatus string

const (
	StatusRunning   InstanceStatus = "running"   // 运行中
	StatusCompleted InstanceStatus = "completed" // 已完成
	StatusCancelled InstanceStatus = "cancelled" // 已取消
	StatusFailed    InstanceStatus = "failed"    // 失败
	StatusSuspended InstanceStatus = "suspended" // 暂停
)

// ExecutionHistory 执行历史
type ExecutionHistory struct {
	ID          string                 `json:"id"`
	NodeID      string                 `json:"node_id"`
	NodeName    string                 `json:"node_name"`
	Action      string                 `json:"action"`
	Result      string                 `json:"result"`
	Comment     string                 `json:"comment,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	ExecutedBy  uint                   `json:"executed_by"`
	ExecutedAt  time.Time              `json:"executed_at"`
	Duration    time.Duration          `json:"duration"`
}

// StartWorkflowRequest 启动流程请求
type StartWorkflowRequest struct {
	WorkflowID   string                 `json:"workflow_id"`
	BusinessID   string                 `json:"business_id"`
	BusinessType string                 `json:"business_type"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	StartedBy    uint                   `json:"started_by"`
}

// ApprovalRequest 审批请求
type ApprovalRequest struct {
	InstanceID string                 `json:"instance_id"`
	NodeID     string                 `json:"node_id"`
	Action     ApprovalAction         `json:"action"`
	Comment    string                 `json:"comment,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	ApprovedBy uint                   `json:"approved_by"`
}

// ApprovalAction 审批动作
type ApprovalAction string

const (
	ActionApprove ApprovalAction = "approve" // 同意
	ActionReject  ApprovalAction = "reject"  // 拒绝
	ActionReturn  ApprovalAction = "return"  // 退回
	ActionDelegate ApprovalAction = "delegate" // 委托
)

// ApprovalResult 审批结果
type ApprovalResult struct {
	InstanceID   string         `json:"instance_id"`
	NodeID       string         `json:"node_id"`
	Action       ApprovalAction `json:"action"`
	NextNodes    []string       `json:"next_nodes"`
	IsCompleted  bool           `json:"is_completed"`
	Message      string         `json:"message"`
	ExecutedAt   time.Time      `json:"executed_at"`
}

// PendingApproval 待审批任务
type PendingApproval struct {
	InstanceID     string                 `json:"instance_id"`
	WorkflowName   string                 `json:"workflow_name"`
	NodeID         string                 `json:"node_id"`
	NodeName       string                 `json:"node_name"`
	BusinessID     string                 `json:"business_id"`
	BusinessType   string                 `json:"business_type"`
	BusinessData   map[string]interface{} `json:"business_data,omitempty"`
	Priority       int                    `json:"priority"`
	AssignedTo     uint                   `json:"assigned_to"`
	CreatedAt      time.Time              `json:"created_at"`
	Deadline       *time.Time             `json:"deadline,omitempty"`
	CanDelegate    bool                   `json:"can_delegate"`
	RequiredAction []ApprovalAction       `json:"required_actions"`
}

// ApprovalNodeConfig 审批节点配置
type ApprovalNodeConfig struct {
	Assignees    []ApprovalAssignee `json:"assignees"`              // 审批人配置
	ApprovalType ApprovalType       `json:"approval_type"`          // 审批类型
	Deadline     *time.Duration     `json:"deadline,omitempty"`     // 审批期限
	AutoApprove  bool               `json:"auto_approve,omitempty"` // 超时自动审批
	CanDelegate  bool               `json:"can_delegate,omitempty"` // 允许委托
	CanReturn    bool               `json:"can_return,omitempty"`   // 允许退回
	Priority     int                `json:"priority,omitempty"`     // 优先级
}

// ApprovalAssignee 审批人配置
type ApprovalAssignee struct {
	Type   AssigneeType `json:"type"`             // 分配类型
	Value  string       `json:"value"`            // 分配值
	Backup []string     `json:"backup,omitempty"` // 备用审批人
}

// AssigneeType 审批人分配类型
type AssigneeType string

const (
	AssigneeTypeUser       AssigneeType = "user"        // 指定用户
	AssigneeTypeRole       AssigneeType = "role"        // 指定角色
	AssigneeTypeDepartment AssigneeType = "department"  // 指定部门
	AssigneeTypeManager    AssigneeType = "manager"     // 直属上级
	AssigneeTypeStarter    AssigneeType = "starter"     // 流程发起人
	AssigneeTypeVariable   AssigneeType = "variable"    // 变量指定
	AssigneeTypeScript     AssigneeType = "script"      // 脚本计算
)

// ApprovalType 审批类型
type ApprovalType string

const (
	ApprovalTypeSequential ApprovalType = "sequential" // 串行审批（依次审批）
	ApprovalTypeParallel   ApprovalType = "parallel"   // 并行审批（同时审批）
	ApprovalTypeAny        ApprovalType = "any"        // 任意一人审批
	ApprovalTypeAll        ApprovalType = "all"        // 全部人员审批
	ApprovalTypeMajority   ApprovalType = "majority"   // 多数审批
)

// ConditionNodeConfig 条件节点配置
type ConditionNodeConfig struct {
	Conditions []ConditionRule `json:"conditions"`
}

// ConditionRule 条件规则
type ConditionRule struct {
	Expression string `json:"expression"` // 条件表达式
	Target     string `json:"target"`     // 目标节点
	Priority   int    `json:"priority"`   // 优先级
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	Type      NotificationType `json:"type"`
	Template  string           `json:"template"`
	Recipients []string        `json:"recipients"`
}

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeEmail  NotificationType = "email"
	NotificationTypeSMS    NotificationType = "sms"
	NotificationTypeSystem NotificationType = "system"
	NotificationTypeWebhook NotificationType = "webhook"
)

// WorkflowRepository 流程定义仓库接口
type WorkflowRepository interface {
	// GetWorkflowDefinition 获取流程定义
	GetWorkflowDefinition(ctx context.Context, workflowID string) (*WorkflowDefinition, error)

	// SaveWorkflowDefinition 保存流程定义
	SaveWorkflowDefinition(ctx context.Context, definition *WorkflowDefinition) error

	// ListWorkflowDefinitions 列出流程定义
	ListWorkflowDefinitions(ctx context.Context, filter WorkflowFilter) ([]*WorkflowDefinition, error)
}

// WorkflowInstanceRepository 流程实例仓库接口
type WorkflowInstanceRepository interface {
	// SaveInstance 保存流程实例
	SaveInstance(ctx context.Context, instance *WorkflowInstance) error

	// GetInstance 获取流程实例
	GetInstance(ctx context.Context, instanceID string) (*WorkflowInstance, error)

	// UpdateInstanceStatus 更新实例状态
	UpdateInstanceStatus(ctx context.Context, instanceID string, status InstanceStatus) error

	// UpdateInstance 更新流程实例
	UpdateInstance(ctx context.Context, instance *WorkflowInstance) error

	// AddExecutionHistory 添加执行历史
	AddExecutionHistory(ctx context.Context, instanceID string, history ExecutionHistory) error

	// GetPendingApprovals 获取待审批任务
	GetPendingApprovals(ctx context.Context, userID uint) ([]*PendingApproval, error)

	// SavePendingApproval 保存待审批记录
	SavePendingApproval(ctx context.Context, approval *PendingApproval) error

	// DeletePendingApproval 删除待审批记录
	DeletePendingApproval(ctx context.Context, instanceID, nodeID string, userID uint) error
}

// WorkflowFilter 流程过滤条件
type WorkflowFilter struct {
	IsActive *bool  `json:"is_active,omitempty"`
	Name     string `json:"name,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// NodeExecutor 节点执行器接口
type NodeExecutor interface {
	// Execute 执行节点
	Execute(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode) (*NodeExecutionResult, error)

	// ExecuteWithDefinition 带定义执行节点
	ExecuteWithDefinition(ctx context.Context, instance *WorkflowInstance, node *WorkflowNode, definition *WorkflowDefinition) (*NodeExecutionResult, error)

	// GetSupportedNodeType 获取支持的节点类型
	GetSupportedNodeType() NodeType
}

// NodeExecutionResult 节点执行结果
type NodeExecutionResult struct {
	Success     bool                   `json:"success"`
	NextNodes   []string               `json:"next_nodes"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Message     string                 `json:"message,omitempty"`
	WaitForUser bool                   `json:"wait_for_user"` // 是否等待用户操作
	Error       error                  `json:"error,omitempty"`
}
