package algorithms

import (
	"context"
	"testing"
	"time"

	"taskmanage/internal/database"
)

// TestRoundRobinAlgorithm 测试轮询分配算法
func TestRoundRobinAlgorithm(t *testing.T) {
	algorithm := NewRoundRobinAlgorithm()

	// 验证算法基本信息
	if algorithm.GetName() != "Round Robin Assignment" {
		t.Errorf("Expected name 'Round Robin Assignment', got %s", algorithm.GetName())
	}

	if algorithm.GetStrategy() != StrategyRoundRobin {
		t.Errorf("Expected strategy %s, got %s", StrategyRoundRobin, algorithm.GetStrategy())
	}

	// 创建测试候选人
	candidates := createTestCandidates()

	// 创建测试请求
	req := &AssignmentRequest{
		TaskID:   1,
		Strategy: StrategyRoundRobin,
		Priority: "medium",
	}

	ctx := context.Background()

	// 执行分配算法
	result, err := algorithm.Assign(ctx, req, candidates)
	if err != nil {
		t.Fatalf("Assignment failed: %v", err)
	}

	// 验证结果
	if result.TaskID != req.TaskID {
		t.Errorf("Expected TaskID %d, got %d", req.TaskID, result.TaskID)
	}

	if result.Strategy != req.Strategy {
		t.Errorf("Expected strategy %s, got %s", req.Strategy, result.Strategy)
	}

	if result.Score != 80.0 {
		t.Errorf("Expected score 80.0, got %f", result.Score)
	}
}

// TestLoadBalanceAlgorithm 测试负载均衡分配算法
func TestLoadBalanceAlgorithm(t *testing.T) {
	algorithm := NewLoadBalanceAlgorithm()

	// 验证算法基本信息
	if algorithm.GetName() != "Load Balance Assignment" {
		t.Errorf("Expected name 'Load Balance Assignment', got %s", algorithm.GetName())
	}

	// 创建测试候选人（不同负载）
	candidates := []AssignmentCandidate{
		{
			Employee: database.Employee{
				BaseModel: database.BaseModel{
					ID: 1,
				},
				UserID:       1,
				CurrentTasks: 5,
				MaxTasks:     10,
				User: database.User{
					BaseModel: database.BaseModel{
						ID: 1,
					},
					Username: "employee1",
				},
			},
			Workload: WorkloadInfo{
				CurrentTasks:    5,
				MaxTasks:        10,
				UtilizationRate: 0.5,
			},
		},
		{
			Employee: database.Employee{
				BaseModel: database.BaseModel{
					ID: 2,
				},
				UserID:       2,
				CurrentTasks: 8,
				MaxTasks:     10,
				User: database.User{
					BaseModel: database.BaseModel{
						ID: 2,
					},
					Username: "employee2",
				},
			},
			Workload: WorkloadInfo{
				CurrentTasks:    8,
				MaxTasks:        10,
				UtilizationRate: 0.8,
			},
		},
	}

	req := &AssignmentRequest{
		TaskID:   1,
		Strategy: StrategyLoadBalance,
		Priority: "medium",
	}

	ctx := context.Background()

	// 执行分配算法
	result, err := algorithm.Assign(ctx, req, candidates)
	if err != nil {
		t.Fatalf("Assignment failed: %v", err)
	}

	// 应该选择负载较低的员工（employee2）
	if result.SelectedEmployee.ID != 2 {
		t.Errorf("Expected employee 2 (lower load), got employee %d", result.SelectedEmployee.ID)
	}
}

// TestSkillMatchAlgorithm 测试技能匹配分配算法
func TestSkillMatchAlgorithm(t *testing.T) {
	algorithm := NewSkillMatchAlgorithm()

	// 验证算法基本信息
	if algorithm.GetName() != "Skill Match Assignment" {
		t.Errorf("Expected name 'Skill Match Assignment', got %s", algorithm.GetName())
	}

	candidates := createTestCandidates()

	req := &AssignmentRequest{
		TaskID:   1,
		Strategy: StrategySkillMatch,
		Priority: "medium",
		RequiredSkills: []SkillRequirement{
			{SkillID: 1, MinLevel: 3},
			{SkillID: 2, MinLevel: 2},
		},
	}

	ctx := context.Background()

	// 执行分配算法
	result, err := algorithm.Assign(ctx, req, candidates)
	if err != nil {
		t.Fatalf("Assignment failed: %v", err)
	}

	// 验证结果包含技能匹配信息
	if result.Score <= 0 {
		t.Errorf("Expected positive score, got %f", result.Score)
	}

	if result.Reason == "" {
		t.Error("Expected non-empty reason")
	}
}

// TestComprehensiveAlgorithm 测试综合评分分配算法
func TestComprehensiveAlgorithm(t *testing.T) {
	algorithm := NewComprehensiveAlgorithm()

	// 验证算法基本信息
	if algorithm.GetName() != "Comprehensive Score Assignment" {
		t.Errorf("Expected name 'Comprehensive Score Assignment', got %s", algorithm.GetName())
	}

	candidates := createTestCandidates()

	req := &AssignmentRequest{
		TaskID:   1,
		Strategy: StrategyComprehensive,
		Priority: "high",
		RequiredSkills: []SkillRequirement{
			{SkillID: 1, MinLevel: 3},
		},
	}

	ctx := context.Background()

	// 执行分配算法
	result, err := algorithm.Assign(ctx, req, candidates)
	if err != nil {
		t.Fatalf("Assignment failed: %v", err)
	}

	// 验证综合评分结果
	if result.Score <= 0 || result.Score > 100 {
		t.Errorf("Expected score between 0-100, got %f", result.Score)
	}

	// 应该有备选方案
	if len(result.Alternatives) == 0 {
		t.Error("Expected alternatives in comprehensive algorithm")
	}
}

// TestEmptyCandidates 测试空候选人列表的处理
func TestEmptyCandidates(t *testing.T) {
	algorithm := NewRoundRobinAlgorithm()

	req := &AssignmentRequest{
		TaskID:   1,
		Strategy: StrategyRoundRobin,
	}

	ctx := context.Background()

	// 测试空候选人列表
	_, err := algorithm.Assign(ctx, req, []AssignmentCandidate{})
	if err == nil {
		t.Error("Expected error for empty candidates list")
	}
}

// createTestCandidates 创建测试用的候选人列表
func createTestCandidates() []AssignmentCandidate {
	return []AssignmentCandidate{
		{
			Employee: database.Employee{
				BaseModel: database.BaseModel{
					ID: 1,
				},
				UserID:       1,
				CurrentTasks: 3,
				MaxTasks:     10,
				Status:       "available",
				DepartmentID: func() *uint { id := uint(1); return &id }(), // 使用指针类型
				Department: database.Department{Name: "IT"}, // 添加关联的Department结构体,
				User: database.User{
					BaseModel: database.BaseModel{
						ID: 1,
					},
					Username: "employee1",
				},
			},
			Workload: WorkloadInfo{
				CurrentTasks:    3,
				MaxTasks:        10,
				UtilizationRate: 0.3,
				AvgTaskDuration: 24 * time.Hour,
			},
		},
		{
			Employee: database.Employee{
				BaseModel: database.BaseModel{
					ID: 2,
				},
				UserID:       2,
				CurrentTasks: 7,
				MaxTasks:     10,
				Status:       "busy",
				DepartmentID: func() *uint { id := uint(1); return &id }(),
				Department: database.Department{Name: "IT"},
				User: database.User{
					BaseModel: database.BaseModel{
						ID: 2,
					},
					Username: "employee2",
				},
			},
			Workload: WorkloadInfo{
				CurrentTasks:    7,
				MaxTasks:        10,
				UtilizationRate: 0.7,
				AvgTaskDuration: 48 * time.Hour,
			},
		},
		{
			Employee: database.Employee{
				BaseModel: database.BaseModel{
					ID: 3,
				},
				UserID:       3,
				CurrentTasks: 1,
				MaxTasks:     8,
				Status:       "available",
				DepartmentID: func() *uint { id := uint(2); return &id }(),
				Department: database.Department{Name: "HR"},
				User: database.User{
					BaseModel: database.BaseModel{
						ID: 3,
					},
					Username: "employee3",
				},
			},
			Workload: WorkloadInfo{
				CurrentTasks:    1,
				MaxTasks:        8,
				UtilizationRate: 0.125,
				AvgTaskDuration: 12 * time.Hour,
			},
		},
	}
}
