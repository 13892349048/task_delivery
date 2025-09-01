package algorithms

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"taskmanage/pkg/logger"
)

// ComprehensiveAlgorithm 综合评分分配算法
type ComprehensiveAlgorithm struct {
	name     string
	strategy AssignmentStrategy
	weights  *ComprehensiveWeights
}

// ComprehensiveWeights 综合评分权重配置
type ComprehensiveWeights struct {
	SkillMatch   float64 `json:"skill_match"`  // 技能匹配权重
	LoadBalance  float64 `json:"load_balance"` // 负载均衡权重
	Availability float64 `json:"availability"` // 可用性权重
	Performance  float64 `json:"performance"`  // 历史表现权重
	Department   float64 `json:"department"`   // 部门匹配权重
	Priority     float64 `json:"priority"`     // 任务优先级权重
}

// DefaultWeights 默认权重配置
var DefaultWeights = &ComprehensiveWeights{
	SkillMatch:   0.35, // 35% - 技能匹配最重要
	LoadBalance:  0.25, // 25% - 负载均衡
	Availability: 0.20, // 20% - 可用性
	Performance:  0.10, // 10% - 历史表现
	Department:   0.05, // 5% - 部门匹配
	Priority:     0.05, // 5% - 优先级适应
}

// NewComprehensiveAlgorithm 创建综合评分分配算法实例
func NewComprehensiveAlgorithm() AssignmentAlgorithm {
	return &ComprehensiveAlgorithm{
		name:     "Comprehensive Score Assignment",
		strategy: StrategyComprehensive,
		weights:  DefaultWeights,
	}
}

// NewComprehensiveAlgorithmWithWeights 创建带自定义权重的综合评分算法
func NewComprehensiveAlgorithmWithWeights(weights *ComprehensiveWeights) AssignmentAlgorithm {
	return &ComprehensiveAlgorithm{
		name:     "Comprehensive Score Assignment",
		strategy: StrategyComprehensive,
		weights:  weights,
	}
}

// GetName 获取算法名称
func (a *ComprehensiveAlgorithm) GetName() string {
	return a.name
}

// GetStrategy 获取算法策略
func (a *ComprehensiveAlgorithm) GetStrategy() AssignmentStrategy {
	return a.strategy
}

// Assign 执行综合评分分配算法
func (a *ComprehensiveAlgorithm) Assign(ctx context.Context, req *AssignmentRequest, candidates []AssignmentCandidate) (*AssignmentResult, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用的候选人")
	}

	logger.Infof("开始综合评分分配算法，候选人数量: %d", len(candidates))

	// 为每个候选人计算综合评分
	scoredCandidates := make([]ScoredCandidate, 0, len(candidates))

	for _, candidate := range candidates {
		score, reasons := a.calculateComprehensiveScore(req, candidate)

		scoredCandidate := ScoredCandidate{
			Candidate: candidate,
			Score:     score,
			Reasons:   reasons,
		}

		scoredCandidates = append(scoredCandidates, scoredCandidate)

		logger.Debugf("员工 %d 综合评分: %.2f, 原因: %v",
			candidate.Employee.ID, score, reasons)
	}

	// 按评分排序（降序）
	sort.Slice(scoredCandidates, func(i, j int) bool {
		return scoredCandidates[i].Score > scoredCandidates[j].Score
	})

	// 选择评分最高的候选人
	selectedCandidate := scoredCandidates[0]

	// 准备备选方案（前3名）
	alternatives := make([]AssignmentCandidate, 0)
	for i := 1; i < len(scoredCandidates) && i < 4; i++ {
		alt := scoredCandidates[i].Candidate
		alt.Score = scoredCandidates[i].Score
		alternatives = append(alternatives, alt)
	}

	result := &AssignmentResult{
		TaskID:           req.TaskID,
		Strategy:         a.strategy,
		SelectedEmployee: selectedCandidate.Candidate.Employee,
		Score:            selectedCandidate.Score,
		Reason:           a.generateReason(selectedCandidate),
		Alternatives:     alternatives,
		ExecutedAt:       time.Now(),
	}

	logger.Infof("综合评分分配完成，选择员工: %d (%s), 评分: %.2f",
		result.SelectedEmployee.ID, result.SelectedEmployee.User.Username, result.Score)

	return result, nil
}


// calculateComprehensiveScore 计算综合评分
func (a *ComprehensiveAlgorithm) calculateComprehensiveScore(req *AssignmentRequest, candidate AssignmentCandidate) (float64, []string) {
	var totalScore float64
	var reasons []string

	// 1. 技能匹配评分
	skillScore, skillReason := a.calculateSkillScore(req, candidate)
	totalScore += skillScore * a.weights.SkillMatch
	if skillReason != "" {
		reasons = append(reasons, skillReason)
	}

	// 2. 负载均衡评分
	loadScore, loadReason := a.calculateLoadScore(candidate)
	totalScore += loadScore * a.weights.LoadBalance
	if loadReason != "" {
		reasons = append(reasons, loadReason)
	}

	// 3. 可用性评分
	availScore, availReason := a.calculateAvailabilityScore(req, candidate)
	totalScore += availScore * a.weights.Availability
	if availReason != "" {
		reasons = append(reasons, availReason)
	}

	// 4. 历史表现评分（暂时使用模拟数据）
	perfScore, perfReason := a.calculatePerformanceScore(candidate)
	totalScore += perfScore * a.weights.Performance
	if perfReason != "" {
		reasons = append(reasons, perfReason)
	}

	// 5. 部门匹配评分
	deptScore, deptReason := a.calculateDepartmentScore(req, candidate)
	totalScore += deptScore * a.weights.Department
	if deptReason != "" {
		reasons = append(reasons, deptReason)
	}

	// 6. 优先级适应评分
	priorityScore, priorityReason := a.calculatePriorityScore(req, candidate)
	totalScore += priorityScore * a.weights.Priority
	if priorityReason != "" {
		reasons = append(reasons, priorityReason)
	}

	// 确保评分在0-100之间
	totalScore = math.Max(0, math.Min(100, totalScore))

	return totalScore, reasons
}

// calculateSkillScore 计算技能匹配评分
func (a *ComprehensiveAlgorithm) calculateSkillScore(req *AssignmentRequest, candidate AssignmentCandidate) (float64, string) {
	if len(req.RequiredSkills) == 0 {
		return 80.0, "无特定技能要求"
	}

	// 这里需要实际的技能匹配逻辑
	// 暂时使用简化逻辑
	matchedSkills := 0
	totalSkills := len(req.RequiredSkills)

	// 模拟技能匹配（实际应该查询数据库）
	for range req.RequiredSkills {
		// 假设70%的概率匹配技能
		if candidate.Employee.ID%3 != 0 { // 简单的模拟逻辑
			matchedSkills++
		}
	}

	if matchedSkills == 0 {
		return 20.0, "技能不匹配"
	}

	matchRate := float64(matchedSkills) / float64(totalSkills)
	score := matchRate * 100

	return score, fmt.Sprintf("技能匹配度: %d/%d (%.1f%%)", matchedSkills, totalSkills, matchRate*100)
}

// calculateLoadScore 计算负载均衡评分
func (a *ComprehensiveAlgorithm) calculateLoadScore(candidate AssignmentCandidate) (float64, string) {
	if candidate.Workload.MaxTasks == 0 {
		return 50.0, "无负载信息"
	}

	// 计算负载率（越低越好）
	loadRate := candidate.Workload.UtilizationRate

	// 转换为评分（负载率越低，评分越高）
	score := (1.0 - loadRate) * 100
	score = math.Max(0, math.Min(100, score))

	return score, fmt.Sprintf("当前负载: %d/%d (%.1f%%)",
		candidate.Workload.CurrentTasks, candidate.Workload.MaxTasks, loadRate*100)
}

// calculateAvailabilityScore 计算可用性评分
func (a *ComprehensiveAlgorithm) calculateAvailabilityScore(req *AssignmentRequest, candidate AssignmentCandidate) (float64, string) {
	// 检查员工状态
	switch candidate.Employee.Status {
	case "available":
		return 100.0, "完全可用"
	case "busy":
		return 60.0, "忙碌但可接受任务"
	case "leave":
		return 10.0, "请假中"
	case "inactive":
		return 0.0, "不活跃"
	default:
		return 50.0, "状态未知"
	}
}

// calculatePerformanceScore 计算历史表现评分
func (a *ComprehensiveAlgorithm) calculatePerformanceScore(candidate AssignmentCandidate) (float64, string) {
	// 这里应该基于历史数据计算
	// 暂时使用模拟数据
	baseScore := 75.0

	// 根据员工ID模拟不同的表现评分
	variation := float64(candidate.Employee.ID%20) - 10.0 // -10 到 +9 的变化
	score := baseScore + variation
	score = math.Max(0, math.Min(100, score))

	return score, fmt.Sprintf("历史表现评分: %.1f", score)
}

// calculateDepartmentScore 计算部门匹配评分
func (a *ComprehensiveAlgorithm) calculateDepartmentScore(req *AssignmentRequest, candidate AssignmentCandidate) (float64, string) {
	if req.Department == "" {
		return 80.0, "无部门要求"
	}

	if candidate.Employee.Department.Name == req.Department {
		return 100.0, "部门完全匹配"
	}

	return 30.0, "部门不匹配"
}

// calculatePriorityScore 计算优先级适应评分
func (a *ComprehensiveAlgorithm) calculatePriorityScore(req *AssignmentRequest, candidate AssignmentCandidate) (float64, string) {
	if req.Priority == "" {
		return 80.0, "无优先级要求"
	}

	// 根据任务优先级和员工当前负载调整评分
	switch req.Priority {
	case "urgent":
		// 紧急任务优先分配给负载较低的员工
		if candidate.Workload.UtilizationRate < 0.5 {
			return 100.0, "适合处理紧急任务"
		}
		return 60.0, "可处理紧急任务但负载较高"
	case "high":
		if candidate.Workload.UtilizationRate < 0.7 {
			return 90.0, "适合处理高优先级任务"
		}
		return 70.0, "可处理高优先级任务"
	case "medium":
		return 80.0, "适合处理中等优先级任务"
	case "low":
		return 85.0, "适合处理低优先级任务"
	default:
		return 75.0, "优先级未知"
	}
}

// generateReason 生成分配原因
func (a *ComprehensiveAlgorithm) generateReason(candidate ScoredCandidate) string {
	if len(candidate.Reasons) == 0 {
		return fmt.Sprintf("综合评分分配 (分数: %.1f)", candidate.Score)
	}

	return fmt.Sprintf("综合评分分配 (%s, 分数: %.1f)",
		fmt.Sprintf("%s", candidate.Reasons[0]), candidate.Score)
}
