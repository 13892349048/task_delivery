package assignment

import (
	"context"
	"fmt"

	"taskmanage/internal/assignment/algorithms"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// AssignmentService 分配服务
type AssignmentService struct {
	engine            AssignmentEngine
	candidateProvider CandidateProvider
	history           AssignmentHistory
}

// NewAssignmentService 创建分配服务实例
func NewAssignmentService(repoManager repository.RepositoryManager) *AssignmentService {
	// 创建候选人提供者
	candidateProvider := NewCandidateProvider(
		repoManager.EmployeeRepository(),
		repoManager.SkillRepository(),
		repoManager.TaskRepository(),
	)

	// 创建分配历史记录
	history := NewAssignmentHistory()

	// 创建分配引擎
	engine := NewAssignmentEngine(candidateProvider, history)

	// 注册所有算法
	service := &AssignmentService{
		engine:            engine,
		candidateProvider: candidateProvider,
		history:           history,
	}

	service.registerAlgorithms()

	return service
}

// AlgorithmAdapter 算法适配器，用于适配不同包中的类型定义
type AlgorithmAdapter struct {
	algorithm algorithms.AssignmentAlgorithm
}

// GetName 获取算法名称
func (a *AlgorithmAdapter) GetName() string {
	return a.algorithm.GetName()
}

// GetStrategy 获取算法策略
func (a *AlgorithmAdapter) GetStrategy() AssignmentStrategy {
	return AssignmentStrategy(a.algorithm.GetStrategy())
}

// Assign 执行分配算法
func (a *AlgorithmAdapter) Assign(ctx context.Context, req *AssignmentRequest, candidates []AssignmentCandidate) (*AssignmentResult, error) {
	// 转换请求类型
	algoReq := &algorithms.AssignmentRequest{
		TaskID:           req.TaskID,
		Strategy:         algorithms.AssignmentStrategy(req.Strategy),
		RequiredSkills:   make([]algorithms.SkillRequirement, len(req.RequiredSkills)),
		Department:       req.Department,
		Priority:         req.Priority,
		Deadline:         req.Deadline,
		ExcludeEmployees: req.ExcludeEmployees,
		Preferences:      req.Preferences,
	}

	// 转换技能要求
	for i, skill := range req.RequiredSkills {
		algoReq.RequiredSkills[i] = algorithms.SkillRequirement{
			SkillID:  skill.SkillID,
			MinLevel: skill.MinLevel,
		}
	}

	// 转换候选人类型
	algoCandidates := make([]algorithms.AssignmentCandidate, len(candidates))
	for i, candidate := range candidates {
		algoCandidates[i] = algorithms.AssignmentCandidate{
			Employee: candidate.Employee,
			Workload: algorithms.WorkloadInfo{
				CurrentTasks:    candidate.Workload.CurrentTasks,
				MaxTasks:        candidate.Workload.MaxTasks,
				UtilizationRate: candidate.Workload.UtilizationRate,
				AvgTaskDuration: candidate.Workload.AvgTaskDuration,
			},
			Score: candidate.Score,
		}
	}

	// 调用算法
	result, err := a.algorithm.Assign(ctx, algoReq, algoCandidates)
	if err != nil {
		return nil, err
	}

	// 转换结果类型
	alternatives := make([]AssignmentCandidate, len(result.Alternatives))
	for i, alt := range result.Alternatives {
		alternatives[i] = AssignmentCandidate{
			Employee: alt.Employee,
			Workload: WorkloadInfo{
				CurrentTasks:    alt.Workload.CurrentTasks,
				MaxTasks:        alt.Workload.MaxTasks,
				UtilizationRate: alt.Workload.UtilizationRate,
				AvgTaskDuration: alt.Workload.AvgTaskDuration,
			},
			Score: alt.Score,
		}
	}

	return &AssignmentResult{
		TaskID:           result.TaskID,
		Strategy:         AssignmentStrategy(result.Strategy),
		SelectedEmployee: result.SelectedEmployee,
		Score:            result.Score,
		Reason:           result.Reason,
		Alternatives:     alternatives,
		ExecutedAt:       result.ExecutedAt,
	}, nil
}

// registerAlgorithms 注册所有分配算法
func (s *AssignmentService) registerAlgorithms() {
	algorithmList := []AssignmentAlgorithm{
		&AlgorithmAdapter{algorithms.NewRoundRobinAlgorithm()},
		&AlgorithmAdapter{algorithms.NewLoadBalanceAlgorithm()},
		&AlgorithmAdapter{algorithms.NewSkillMatchAlgorithm()},
		&AlgorithmAdapter{algorithms.NewComprehensiveAlgorithm()},
	}

	for _, algorithm := range algorithmList {
		if err := s.engine.RegisterAlgorithm(algorithm); err != nil {
			logger.Errorf("注册分配算法失败: %v", err)
		}
	}

	logger.Infof("已注册 %d 个分配算法", len(algorithmList))
}

// AssignTask 分配任务
func (s *AssignmentService) AssignTask(ctx context.Context, req *AssignmentRequest) (*AssignmentResult, error) {
	logger.Infof("开始分配任务: TaskID=%d, Strategy=%s", req.TaskID, req.Strategy)

	// 验证请求
	if err := s.validateAssignmentRequest(req); err != nil {
		return nil, fmt.Errorf("分配请求验证失败: %w", err)
	}

	// 执行分配
	result, err := s.engine.ExecuteAssignment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("执行任务分配失败: %w", err)
	}

	logger.Infof("任务分配成功: TaskID=%d, EmployeeID=%d, Strategy=%s, Score=%.2f",
		result.TaskID, result.SelectedEmployee.ID, result.Strategy, result.Score)

	return result, nil
}

// PreviewAssignment 预览分配结果
func (s *AssignmentService) PreviewAssignment(ctx context.Context, req *AssignmentRequest) (*AssignmentResult, error) {
	logger.Infof("预览任务分配: TaskID=%d, Strategy=%s", req.TaskID, req.Strategy)

	// 验证请求
	if err := s.validateAssignmentRequest(req); err != nil {
		return nil, fmt.Errorf("分配请求验证失败: %w", err)
	}

	// 预览分配
	result, err := s.engine.PreviewAssignment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("预览任务分配失败: %w", err)
	}

	return result, nil
}

// GetCandidates 获取分配候选人
func (s *AssignmentService) GetCandidates(ctx context.Context, req *AssignmentRequest) ([]AssignmentCandidate, error) {
	candidates, err := s.engine.GetCandidates(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取候选人失败: %w", err)
	}

	logger.Infof("获取到 %d 个候选人", len(candidates))
	return candidates, nil
}

// GetAvailableStrategies 获取可用的分配策略
func (s *AssignmentService) GetAvailableStrategies() []StrategyInfo {
	algorithms := s.engine.ListAlgorithms()
	strategies := make([]StrategyInfo, len(algorithms))

	for i, algorithm := range algorithms {
		strategies[i] = StrategyInfo{
			Strategy:    algorithm.GetStrategy(),
			Name:        algorithm.GetName(),
			Description: s.getStrategyDescription(algorithm.GetStrategy()),
		}
	}

	return strategies
}

// GetAssignmentHistory 获取分配历史
func (s *AssignmentService) GetAssignmentHistory(ctx context.Context, taskID uint) ([]*AssignmentResult, error) {
	return s.history.GetAssignmentHistory(ctx, taskID)
}

// GetEmployeeAssignmentStats 获取员工分配统计
func (s *AssignmentService) GetEmployeeAssignmentStats(ctx context.Context, employeeID uint) (*AssignmentStats, error) {
	// 获取最近30天的统计
	return s.history.GetEmployeeAssignmentStats(ctx, employeeID, 30*24*3600*1000000000) // 30天的纳秒
}

// validateAssignmentRequest 验证分配请求
func (s *AssignmentService) validateAssignmentRequest(req *AssignmentRequest) error {
	if req == nil {
		return fmt.Errorf("分配请求不能为空")
	}

	if req.TaskID == 0 {
		return fmt.Errorf("任务ID不能为空")
	}

	if req.Strategy == "" {
		return fmt.Errorf("分配策略不能为空")
	}

	// 检查策略是否存在
	_, err := s.engine.GetAlgorithm(req.Strategy)
	if err != nil {
		return fmt.Errorf("不支持的分配策略: %s", req.Strategy)
	}

	// 验证技能要求
	if req.Strategy == StrategySkillMatch || req.Strategy == StrategyComprehensive {
		for _, skill := range req.RequiredSkills {
			if skill.SkillID == 0 {
				return fmt.Errorf("技能ID不能为空")
			}
			if skill.MinLevel < 1 || skill.MinLevel > 5 {
				return fmt.Errorf("技能级别必须在1-5之间")
			}
		}
	}

	return nil
}

// getStrategyDescription 获取策略描述
func (s *AssignmentService) getStrategyDescription(strategy AssignmentStrategy) string {
	descriptions := map[AssignmentStrategy]string{
		StrategyRoundRobin:    "轮询分配 - 按顺序轮流分配给可用员工",
		StrategyLoadBalance:   "负载均衡 - 优先分配给工作负载较低的员工",
		StrategySkillMatch:    "技能匹配 - 根据技能要求匹配最合适的员工",
		StrategyComprehensive: "综合评分 - 综合考虑技能、负载、可用性等因素",
		StrategyManual:        "手动分配 - 管理员手动指定员工",
	}

	if desc, exists := descriptions[strategy]; exists {
		return desc
	}

	return "未知策略"
}

// StrategyInfo 策略信息
type StrategyInfo struct {
	Strategy    AssignmentStrategy `json:"strategy"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
}

// AssignmentServiceManager 分配服务管理器
type AssignmentServiceManager struct {
	service *AssignmentService
}

// NewAssignmentServiceManager 创建分配服务管理器
func NewAssignmentServiceManager(repoManager repository.RepositoryManager) *AssignmentServiceManager {
	return &AssignmentServiceManager{
		service: NewAssignmentService(repoManager),
	}
}

// GetService 获取分配服务
func (m *AssignmentServiceManager) GetService() *AssignmentService {
	return m.service
}

// HealthCheck 健康检查
func (m *AssignmentServiceManager) HealthCheck(ctx context.Context) error {
	// 检查算法是否正常注册
	strategies := m.service.GetAvailableStrategies()
	if len(strategies) == 0 {
		return fmt.Errorf("没有可用的分配算法")
	}

	logger.Infof("分配服务健康检查通过，可用策略数量: %d", len(strategies))
	return nil
}
