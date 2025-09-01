package assignment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// AssignmentEngineImpl 分配引擎实现
type AssignmentEngineImpl struct {
	algorithms        map[AssignmentStrategy]AssignmentAlgorithm
	candidateProvider CandidateProvider
	history           AssignmentHistory
	mutex             sync.RWMutex
}

// NewAssignmentEngine 创建分配引擎实例
func NewAssignmentEngine(candidateProvider CandidateProvider, history AssignmentHistory) AssignmentEngine {
	engine := &AssignmentEngineImpl{
		algorithms:        make(map[AssignmentStrategy]AssignmentAlgorithm),
		candidateProvider: candidateProvider,
		history:           history,
	}

	return engine
}

// RegisterAlgorithm 注册分配算法
func (e *AssignmentEngineImpl) RegisterAlgorithm(algorithm AssignmentAlgorithm) error {
	if algorithm == nil {
		return fmt.Errorf("算法不能为空")
	}

	strategy := algorithm.GetStrategy()
	if strategy == "" {
		return fmt.Errorf("算法策略不能为空")
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.algorithms[strategy] = algorithm
	logger.Infof("注册分配算法: %s (%s)", algorithm.GetName(), strategy)

	return nil
}

// GetAlgorithm 获取分配算法
func (e *AssignmentEngineImpl) GetAlgorithm(strategy AssignmentStrategy) (AssignmentAlgorithm, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	algorithm, exists := e.algorithms[strategy]
	if !exists {
		return nil, fmt.Errorf("未找到分配算法: %s", strategy)
	}

	return algorithm, nil
}

// ListAlgorithms 列出所有可用算法
func (e *AssignmentEngineImpl) ListAlgorithms() []AssignmentAlgorithm {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	algorithms := make([]AssignmentAlgorithm, 0, len(e.algorithms))
	for _, algorithm := range e.algorithms {
		algorithms = append(algorithms, algorithm)
	}

	return algorithms
}

// ExecuteAssignment 执行任务分配
func (e *AssignmentEngineImpl) ExecuteAssignment(ctx context.Context, req *AssignmentRequest) (*AssignmentResult, error) {
	logger.Infof("开始执行任务分配: TaskID=%d, Strategy=%s", req.TaskID, req.Strategy)

	// 获取分配算法
	algorithm, err := e.GetAlgorithm(req.Strategy)
	if err != nil {
		return nil, fmt.Errorf("获取分配算法失败: %w", err)
	}

	// 获取候选人
	candidates, err := e.GetCandidates(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取候选人失败: %w", err)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用的候选人")
	}

	logger.Infof("找到 %d 个候选人", len(candidates))

	// 执行分配算法
	result, err := algorithm.Assign(ctx, req, candidates)
	if err != nil {
		return nil, fmt.Errorf("执行分配算法失败: %w", err)
	}

	// 记录分配历史
	if e.history != nil {
		if err := e.history.RecordAssignment(ctx, result); err != nil {
			logger.Warnf("记录分配历史失败: %v", err)
		}
	}

	logger.Infof("任务分配完成: TaskID=%d, SelectedEmployee=%d, Score=%.2f",
		result.TaskID, result.SelectedEmployee.ID, result.Score)

	return result, nil
}

// GetCandidates 获取分配候选人
func (e *AssignmentEngineImpl) GetCandidates(ctx context.Context, req *AssignmentRequest) ([]AssignmentCandidate, error) {
	// 获取可用员工
	employees, err := e.candidateProvider.GetAvailableEmployees(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取可用员工失败: %w", err)
	}

	candidates := make([]AssignmentCandidate, 0, len(employees))

	for _, employee := range employees {
		// 获取员工工作负载
		workload, err := e.candidateProvider.GetEmployeeWorkload(ctx, employee.ID)
		if err != nil {
			logger.Warnf("获取员工 %d 工作负载失败: %v", employee.ID, err)
			// 使用默认工作负载信息
			workload = &WorkloadInfo{
				CurrentTasks:    employee.CurrentTasks,
				MaxTasks:        employee.MaxTasks,
				UtilizationRate: float64(employee.CurrentTasks) / float64(employee.MaxTasks),
			}
		}

		// 检查员工可用性（如果有截止时间）
		if req.Deadline != nil {
			available, err := e.candidateProvider.CheckEmployeeAvailability(ctx, employee.ID, req.Deadline)
			if err != nil {
				logger.Warnf("检查员工 %d 可用性失败: %v", employee.ID, err)
			} else if !available {
				logger.Debugf("员工 %d 在截止时间前不可用", employee.ID)
				continue
			}
		}

		candidate := AssignmentCandidate{
			Employee: *employee,
			Workload: *workload,
		}

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

// PreviewAssignment 预览分配结果（不实际分配）
func (e *AssignmentEngineImpl) PreviewAssignment(ctx context.Context, req *AssignmentRequest) (*AssignmentResult, error) {
	logger.Infof("预览任务分配: TaskID=%d, Strategy=%s", req.TaskID, req.Strategy)

	// 获取分配算法
	algorithm, err := e.GetAlgorithm(req.Strategy)
	if err != nil {
		return nil, fmt.Errorf("获取分配算法失败: %w", err)
	}

	// 获取候选人
	candidates, err := e.GetCandidates(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取候选人失败: %w", err)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用的候选人")
	}

	// 执行分配算法（预览模式）
	result, err := algorithm.Assign(ctx, req, candidates)
	if err != nil {
		return nil, fmt.Errorf("执行分配算法失败: %w", err)
	}

	// 标记为预览结果
	result.Reason = "[预览] " + result.Reason

	return result, nil
}

// CandidateProviderImpl 候选人提供者实现
type CandidateProviderImpl struct {
	employeeRepo repository.EmployeeRepository
	skillRepo    repository.SkillRepository
	taskRepo     repository.TaskRepository
}

// NewCandidateProvider 创建候选人提供者实例
func NewCandidateProvider(employeeRepo repository.EmployeeRepository, skillRepo repository.SkillRepository, taskRepo repository.TaskRepository) CandidateProvider {
	return &CandidateProviderImpl{
		employeeRepo: employeeRepo,
		skillRepo:    skillRepo,
		taskRepo:     taskRepo,
	}
}

// GetAvailableEmployees 获取可用员工
func (c *CandidateProviderImpl) GetAvailableEmployees(ctx context.Context, req *AssignmentRequest) ([]*database.Employee, error) {
	// 获取所有可用员工
	employees, err := c.employeeRepo.GetAvailableEmployees(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取可用员工失败: %w", err)
	}

	// 根据部门过滤（如果指定）
	if req.Department != "" {
		filteredEmployees := make([]*database.Employee, 0)
		for _, employee := range employees {
			if employee.Department.Name == req.Department {
				filteredEmployees = append(filteredEmployees, employee)
			}
		}
		employees = filteredEmployees
	}

	// 排除指定员工
	if len(req.ExcludeEmployees) > 0 {
		excludeMap := make(map[uint]bool)
		for _, id := range req.ExcludeEmployees {
			excludeMap[id] = true
		}

		filteredEmployees := make([]*database.Employee, 0)
		for _, employee := range employees {
			if !excludeMap[employee.ID] {
				filteredEmployees = append(filteredEmployees, employee)
			}
		}
		employees = filteredEmployees
	}

	return employees, nil
}

// GetEmployeeSkills 获取员工技能
func (c *CandidateProviderImpl) GetEmployeeSkills(ctx context.Context, employeeID uint) ([]database.Skill, error) {
	skills, err := c.skillRepo.GetEmployeeSkills(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("获取员工技能失败: %w", err)
	}

	// 转换为值类型切片
	result := make([]database.Skill, len(skills))
	for i, skill := range skills {
		result[i] = *skill
	}

	return result, nil
}

// GetEmployeeWorkload 获取员工工作负载
func (c *CandidateProviderImpl) GetEmployeeWorkload(ctx context.Context, employeeID uint) (*WorkloadInfo, error) {
	employee, err := c.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("获取员工信息失败: %w", err)
	}

	workload := &WorkloadInfo{
		CurrentTasks: employee.CurrentTasks,
		MaxTasks:     employee.MaxTasks,
	}

	// 计算利用率
	if employee.MaxTasks > 0 {
		workload.UtilizationRate = float64(employee.CurrentTasks) / float64(employee.MaxTasks)
	}

	// 计算平均任务完成时间（这里需要实际的统计数据，暂时使用模拟值）
	workload.AvgTaskDuration = 24 * time.Hour // 默认24小时

	return workload, nil
}

// CheckEmployeeAvailability 检查员工可用性
func (c *CandidateProviderImpl) CheckEmployeeAvailability(ctx context.Context, employeeID uint, deadline *time.Time) (bool, error) {
	if deadline == nil {
		return true, nil
	}

	employee, err := c.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return false, fmt.Errorf("获取员工信息失败: %w", err)
	}

	// 检查员工状态
	if employee.Status != "available" && employee.Status != "busy" {
		return false, nil
	}

	// 检查工作负载
	if employee.CurrentTasks >= employee.MaxTasks {
		return false, nil
	}

	// 检查是否有足够时间完成任务（简单检查）
	now := time.Now()
	if deadline.Before(now.Add(time.Hour)) { // 至少需要1小时
		return false, nil
	}

	return true, nil
}

// AssignmentHistoryImpl 分配历史记录实现
type AssignmentHistoryImpl struct {
	// 这里可以使用数据库或缓存存储历史记录
	// 暂时使用内存存储
	records map[uint][]*AssignmentResult // taskID -> results
	mutex   sync.RWMutex
}

// NewAssignmentHistory 创建分配历史记录实例
func NewAssignmentHistory() AssignmentHistory {
	return &AssignmentHistoryImpl{
		records: make(map[uint][]*AssignmentResult),
	}
}

// RecordAssignment 记录分配结果
func (h *AssignmentHistoryImpl) RecordAssignment(ctx context.Context, result *AssignmentResult) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	taskID := result.TaskID
	if h.records[taskID] == nil {
		h.records[taskID] = make([]*AssignmentResult, 0)
	}

	h.records[taskID] = append(h.records[taskID], result)
	logger.Infof("记录分配历史: TaskID=%d, Strategy=%s", taskID, result.Strategy)

	return nil
}

// GetAssignmentHistory 获取分配历史
func (h *AssignmentHistoryImpl) GetAssignmentHistory(ctx context.Context, taskID uint) ([]*AssignmentResult, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	records := h.records[taskID]
	if records == nil {
		return []*AssignmentResult{}, nil
	}

	// 返回副本
	result := make([]*AssignmentResult, len(records))
	copy(result, records)

	return result, nil
}

// GetEmployeeAssignmentStats 获取员工分配统计
func (h *AssignmentHistoryImpl) GetEmployeeAssignmentStats(ctx context.Context, employeeID uint, period time.Duration) (*AssignmentStats, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	stats := &AssignmentStats{
		EmployeeID: employeeID,
	}

	cutoff := time.Now().Add(-period)
	var totalScore float64
	var scoreCount int

	// 遍历所有记录
	for _, records := range h.records {
		for _, record := range records {
			if record.SelectedEmployee.ID == employeeID && record.ExecutedAt.After(cutoff) {
				stats.TotalAssignments++
				if record.Score > 0 {
					totalScore += record.Score
					scoreCount++
				}
			}
		}
	}

	// 计算平均分数
	if scoreCount > 0 {
		stats.AverageScore = totalScore / float64(scoreCount)
	}

	// 这里可以添加更多统计逻辑，如完成率、平均完成时间等
	// 暂时使用模拟数据
	stats.CompletedTasks = stats.TotalAssignments // 假设都完成了
	stats.SuccessRate = 1.0                       // 100%成功率
	stats.AverageCompletion = 24 * time.Hour      // 平均24小时完成

	return stats, nil
}
