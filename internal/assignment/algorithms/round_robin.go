package algorithms

import (
	"context"
	"fmt"
	"sync"
	"time"

	"taskmanage/pkg/logger"
)

// RoundRobinAlgorithm 轮询分配算法
type RoundRobinAlgorithm struct {
	name     string
	strategy AssignmentStrategy
	counter  int
	mutex    sync.Mutex
}

// NewRoundRobinAlgorithm 创建轮询分配算法实例
func NewRoundRobinAlgorithm() AssignmentAlgorithm {
	return &RoundRobinAlgorithm{
		name:     "Round Robin Assignment",
		strategy: StrategyRoundRobin,
		counter:  0,
	}
}

// GetName 获取算法名称
func (a *RoundRobinAlgorithm) GetName() string {
	return a.name
}

// GetStrategy 获取算法策略
func (a *RoundRobinAlgorithm) GetStrategy() AssignmentStrategy {
	return a.strategy
}

// Assign 执行轮询分配算法
func (a *RoundRobinAlgorithm) Assign(ctx context.Context, req *AssignmentRequest, candidates []AssignmentCandidate) (*AssignmentResult, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用的候选人")
	}

	logger.Infof("开始轮询分配算法，候选人数量: %d", len(candidates))

	// 使用互斥锁确保线程安全
	a.mutex.Lock()
	selectedIndex := a.counter % len(candidates)
	a.counter++
	a.mutex.Unlock()

	selectedCandidate := candidates[selectedIndex]

	// 准备备选方案（下一个候选人）
	alternatives := make([]AssignmentCandidate, 0)
	if len(candidates) > 1 {
		nextIndex := (selectedIndex + 1) % len(candidates)
		alternatives = append(alternatives, candidates[nextIndex])
	}

	result := &AssignmentResult{
		TaskID:           req.TaskID,
		Strategy:         a.strategy,
		SelectedEmployee: selectedCandidate.Employee,
		Score:            80.0, // 轮询算法固定评分
		Reason:           fmt.Sprintf("轮询分配 (第%d个候选人: %s)", selectedIndex+1, selectedCandidate.Employee.User.Username),
		Alternatives:     alternatives,
		ExecutedAt:       time.Now(),
	}

	logger.Infof("轮询分配完成，选择员工: %d (%s), 索引: %d",
		result.SelectedEmployee.ID, result.SelectedEmployee.User.Username, selectedIndex)

	return result, nil
}
