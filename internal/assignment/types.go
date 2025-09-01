package assignment

import (
	"context"
	"time"

	"taskmanage/internal/database"
)

// AssignmentStrategy 分配策略类型
type AssignmentStrategy string

const (
	StrategyRoundRobin    AssignmentStrategy = "round_robin"
	StrategyLoadBalance   AssignmentStrategy = "load_balance"
	StrategySkillMatch    AssignmentStrategy = "skill_match"
	StrategyComprehensive AssignmentStrategy = "comprehensive"
	StrategyManual        AssignmentStrategy = "manual"
)

// AssignmentAlgorithm 分配算法接口
type AssignmentAlgorithm interface {
	// GetName 获取算法名称
	GetName() string

	// GetStrategy 获取算法策略
	GetStrategy() AssignmentStrategy

	// Assign 执行分配算法
	Assign(ctx context.Context, req *AssignmentRequest, candidates []AssignmentCandidate) (*AssignmentResult, error)
}

// AssignmentRequest 分配请求
type AssignmentRequest struct {
	TaskID           uint                    `json:"task_id"`
	Strategy         AssignmentStrategy      `json:"strategy"`
	RequiredSkills   []SkillRequirement      `json:"required_skills,omitempty"`
	Department       string                  `json:"department,omitempty"`
	Priority         string                  `json:"priority,omitempty"`
	Deadline         *time.Time              `json:"deadline,omitempty"`
	ExcludeEmployees []uint                  `json:"exclude_employees,omitempty"`
	Preferences      map[string]interface{}  `json:"preferences,omitempty"`
}

// SkillRequirement 技能要求
type SkillRequirement struct {
	SkillID  uint `json:"skill_id"`
	MinLevel int  `json:"min_level"`
}

// AssignmentCandidate 分配候选人
type AssignmentCandidate struct {
	Employee database.Employee `json:"employee"`
	Workload WorkloadInfo      `json:"workload"`
	Score    float64           `json:"score,omitempty"`
}

// WorkloadInfo 工作负载信息
type WorkloadInfo struct {
	CurrentTasks      int           `json:"current_tasks"`
	MaxTasks          int           `json:"max_tasks"`
	UtilizationRate   float64       `json:"utilization_rate"`
	AvgTaskDuration   time.Duration `json:"avg_task_duration"`
}

// AssignmentResult 分配结果
type AssignmentResult struct {
	TaskID           uint              `json:"task_id"`
	Strategy         AssignmentStrategy `json:"strategy"`
	SelectedEmployee database.Employee `json:"selected_employee"`
	Score            float64           `json:"score"`
	Reason           string            `json:"reason"`
	Alternatives     []AssignmentCandidate `json:"alternatives,omitempty"`
	ExecutedAt       time.Time         `json:"executed_at"`
}

// AssignmentEngine 分配引擎接口
type AssignmentEngine interface {
	// RegisterAlgorithm 注册分配算法
	RegisterAlgorithm(algorithm AssignmentAlgorithm) error

	// GetAlgorithm 获取分配算法
	GetAlgorithm(strategy AssignmentStrategy) (AssignmentAlgorithm, error)

	// ListAlgorithms 列出所有可用算法
	ListAlgorithms() []AssignmentAlgorithm

	// ExecuteAssignment 执行任务分配
	ExecuteAssignment(ctx context.Context, req *AssignmentRequest) (*AssignmentResult, error)

	// PreviewAssignment 预览分配结果（不实际分配）
	PreviewAssignment(ctx context.Context, req *AssignmentRequest) (*AssignmentResult, error)

	// GetCandidates 获取分配候选人
	GetCandidates(ctx context.Context, req *AssignmentRequest) ([]AssignmentCandidate, error)
}

// CandidateProvider 候选人提供者接口
type CandidateProvider interface {
	// GetAvailableEmployees 获取可用员工
	GetAvailableEmployees(ctx context.Context, req *AssignmentRequest) ([]*database.Employee, error)

	// GetEmployeeSkills 获取员工技能
	GetEmployeeSkills(ctx context.Context, employeeID uint) ([]database.Skill, error)

	// GetEmployeeWorkload 获取员工工作负载
	GetEmployeeWorkload(ctx context.Context, employeeID uint) (*WorkloadInfo, error)

	// CheckEmployeeAvailability 检查员工可用性
	CheckEmployeeAvailability(ctx context.Context, employeeID uint, deadline *time.Time) (bool, error)
}

// AssignmentHistory 分配历史记录接口
type AssignmentHistory interface {
	// RecordAssignment 记录分配结果
	RecordAssignment(ctx context.Context, result *AssignmentResult) error

	// GetAssignmentHistory 获取分配历史
	GetAssignmentHistory(ctx context.Context, taskID uint) ([]*AssignmentResult, error)

	// GetEmployeeAssignmentStats 获取员工分配统计
	GetEmployeeAssignmentStats(ctx context.Context, employeeID uint, period time.Duration) (*AssignmentStats, error)
}

// AssignmentStats 分配统计信息
type AssignmentStats struct {
	EmployeeID        uint          `json:"employee_id"`
	TotalAssignments  int           `json:"total_assignments"`
	CompletedTasks    int           `json:"completed_tasks"`
	AverageScore      float64       `json:"average_score"`
	SuccessRate       float64       `json:"success_rate"`
	AverageCompletion time.Duration `json:"average_completion"`
}
