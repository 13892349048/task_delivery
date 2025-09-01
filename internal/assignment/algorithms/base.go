package algorithms

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

// ScoredCandidate 带评分的候选人
type ScoredCandidate struct {
	Candidate AssignmentCandidate
	Score     float64
	Reasons   []string
}
