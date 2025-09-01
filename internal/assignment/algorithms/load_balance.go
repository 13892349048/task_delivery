package algorithms

import (
	"context"
	"fmt"
	"sort"
	"time"

	"taskmanage/pkg/logger"
)

// LoadBalanceAlgorithm 负载均衡分配算法
type LoadBalanceAlgorithm struct {
	name     string
	strategy AssignmentStrategy
}

// NewLoadBalanceAlgorithm 创建负载均衡分配算法实例
func NewLoadBalanceAlgorithm() AssignmentAlgorithm {
	return &LoadBalanceAlgorithm{
		name:     "Load Balance Assignment",
		strategy: StrategyLoadBalance,
	}
}

// GetName 获取算法名称
func (a *LoadBalanceAlgorithm) GetName() string {
	return a.name
}

// GetStrategy 获取算法策略
func (a *LoadBalanceAlgorithm) GetStrategy() AssignmentStrategy {
	return a.strategy
}

// Assign 执行负载均衡分配算法
func (a *LoadBalanceAlgorithm) Assign(ctx context.Context, req *AssignmentRequest, candidates []AssignmentCandidate) (*AssignmentResult, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用的候选人")
	}

	logger.Infof("开始负载均衡分配算法，候选人数量: %d", len(candidates))

	// 按工作负载排序（升序，负载最低的在前面）
	sortedCandidates := make([]AssignmentCandidate, len(candidates))
	copy(sortedCandidates, candidates)

	sort.Slice(sortedCandidates, func(i, j int) bool {
		// 优先比较利用率
		if sortedCandidates[i].Workload.UtilizationRate != sortedCandidates[j].Workload.UtilizationRate {
			return sortedCandidates[i].Workload.UtilizationRate < sortedCandidates[j].Workload.UtilizationRate
		}
		// 利用率相同时，比较当前任务数
		return sortedCandidates[i].Workload.CurrentTasks < sortedCandidates[j].Workload.CurrentTasks
	})

	selectedCandidate := sortedCandidates[0]

	// 计算负载均衡评分
	score := a.calculateLoadBalanceScore(selectedCandidate)

	// 准备备选方案（负载第二低的候选人）
	alternatives := make([]AssignmentCandidate, 0)
	if len(sortedCandidates) > 1 {
		alt := sortedCandidates[1]
		alt.Score = a.calculateLoadBalanceScore(alt)
		alternatives = append(alternatives, alt)
	}

	result := &AssignmentResult{
		TaskID:           req.TaskID,
		Strategy:         a.strategy,
		SelectedEmployee: selectedCandidate.Employee,
		Score:            score,
		Reason: fmt.Sprintf("负载均衡分配 (当前负载: %d/%d, 利用率: %.1f%%, %s)",
			selectedCandidate.Workload.CurrentTasks,
			selectedCandidate.Workload.MaxTasks,
			selectedCandidate.Workload.UtilizationRate*100,
			selectedCandidate.Employee.User.Username),
		Alternatives: alternatives,
		ExecutedAt:   time.Now(),
	}

	logger.Infof("负载均衡分配完成，选择员工: %d (%s), 当前负载: %d/%d",
		result.SelectedEmployee.ID, result.SelectedEmployee.User.Username,
		selectedCandidate.Workload.CurrentTasks, selectedCandidate.Workload.MaxTasks)

	return result, nil
}

// calculateLoadBalanceScore 计算负载均衡评分
func (a *LoadBalanceAlgorithm) calculateLoadBalanceScore(candidate AssignmentCandidate) float64 {
	if candidate.Workload.MaxTasks == 0 {
		return 50.0 // 无负载信息时给中等评分
	}

	// 基于利用率计算评分（利用率越低，评分越高）
	utilizationRate := candidate.Workload.UtilizationRate
	score := (1.0 - utilizationRate) * 100

	// 确保评分在0-100之间
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}
