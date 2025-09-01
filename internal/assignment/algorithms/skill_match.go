package algorithms

import (
	"context"
	"fmt"
	"sort"
	"time"

	"taskmanage/pkg/logger"
)

// SkillMatchAlgorithm 技能匹配分配算法
type SkillMatchAlgorithm struct {
	name     string
	strategy AssignmentStrategy
}

// NewSkillMatchAlgorithm 创建技能匹配分配算法实例
func NewSkillMatchAlgorithm() AssignmentAlgorithm {
	return &SkillMatchAlgorithm{
		name:     "Skill Match Assignment",
		strategy: StrategySkillMatch,
	}
}

// GetName 获取算法名称
func (a *SkillMatchAlgorithm) GetName() string {
	return a.name
}

// GetStrategy 获取算法策略
func (a *SkillMatchAlgorithm) GetStrategy() AssignmentStrategy {
	return a.strategy
}

// Assign 执行技能匹配分配算法
func (a *SkillMatchAlgorithm) Assign(ctx context.Context, req *AssignmentRequest, candidates []AssignmentCandidate) (*AssignmentResult, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用的候选人")
	}

	logger.Infof("开始技能匹配分配算法，候选人数量: %d", len(candidates))

	// 为每个候选人计算技能匹配评分
	scoredCandidates := make([]ScoredCandidate, 0, len(candidates))

	for _, candidate := range candidates {
		score := a.calculateSkillMatchScore(req, candidate)

		scoredCandidate := ScoredCandidate{
			Candidate: candidate,
			Score:     score,
		}

		scoredCandidates = append(scoredCandidates, scoredCandidate)

		logger.Debugf("员工 %d 技能匹配评分: %.2f", candidate.Employee.ID, score)
	}

	// 按评分排序（降序）
	sort.Slice(scoredCandidates, func(i, j int) bool {
		return scoredCandidates[i].Score > scoredCandidates[j].Score
	})

	// 选择评分最高的候选人
	selectedCandidate := scoredCandidates[0]

	// 准备备选方案（前2名）
	alternatives := make([]AssignmentCandidate, 0)
	for i := 1; i < len(scoredCandidates) && i < 3; i++ {
		alt := scoredCandidates[i].Candidate
		alt.Score = scoredCandidates[i].Score
		alternatives = append(alternatives, alt)
	}

	result := &AssignmentResult{
		TaskID:           req.TaskID,
		Strategy:         a.strategy,
		SelectedEmployee: selectedCandidate.Candidate.Employee,
		Score:            selectedCandidate.Score,
		Reason: fmt.Sprintf("技能匹配分配 (匹配度: %.1f%%, %s)",
			selectedCandidate.Score, selectedCandidate.Candidate.Employee.User.Username),
		Alternatives: alternatives,
		ExecutedAt:   time.Now(),
	}

	logger.Infof("技能匹配分配完成，选择员工: %d (%s), 匹配度: %.2f",
		result.SelectedEmployee.ID, result.SelectedEmployee.User.Username, result.Score)

	return result, nil
}

// calculateSkillMatchScore 计算技能匹配评分
func (a *SkillMatchAlgorithm) calculateSkillMatchScore(req *AssignmentRequest, candidate AssignmentCandidate) float64 {
	// 如果没有技能要求，给予基础评分
	if len(req.RequiredSkills) == 0 {
		return 75.0 // 基础评分
	}

	// 这里需要实际的技能匹配逻辑
	// 暂时使用模拟逻辑，实际应该查询员工技能数据

	totalRequiredSkills := len(req.RequiredSkills)
	matchedSkills := 0
	totalSkillLevel := 0
	requiredSkillLevel := 0

	// 模拟技能匹配计算
	for _, requiredSkill := range req.RequiredSkills {
		requiredSkillLevel += requiredSkill.MinLevel

		// 简单的模拟逻辑：根据员工ID和技能ID计算匹配度
		employeeSkillLevel := a.simulateEmployeeSkillLevel(candidate.Employee.ID, requiredSkill.SkillID)

		if employeeSkillLevel >= requiredSkill.MinLevel {
			matchedSkills++
			totalSkillLevel += employeeSkillLevel
		}
	}

	// 如果没有匹配的技能，返回低分
	if matchedSkills == 0 {
		return 20.0
	}

	// 计算匹配率
	matchRate := float64(matchedSkills) / float64(totalRequiredSkills)

	// 计算技能水平评分
	avgSkillLevel := float64(totalSkillLevel) / float64(matchedSkills)
	avgRequiredLevel := float64(requiredSkillLevel) / float64(totalRequiredSkills)

	skillLevelScore := (avgSkillLevel / avgRequiredLevel) * 100
	if skillLevelScore > 100 {
		skillLevelScore = 100
	}

	// 综合评分：匹配率权重70%，技能水平权重30%
	finalScore := matchRate*70 + skillLevelScore*0.3

	return finalScore
}

// simulateEmployeeSkillLevel 模拟员工技能水平
// 实际应该从数据库查询员工技能
func (a *SkillMatchAlgorithm) simulateEmployeeSkillLevel(employeeID, skillID uint) int {
	// 简单的模拟逻辑
	seed := (employeeID * skillID) % 10

	if seed < 2 {
		return 0 // 20% 概率没有这个技能
	} else if seed < 4 {
		return 2 // 20% 概率技能水平为2
	} else if seed < 7 {
		return 3 // 30% 概率技能水平为3
	} else if seed < 9 {
		return 4 // 20% 概率技能水平为4
	} else {
		return 5 // 10% 概率技能水平为5
	}
}
