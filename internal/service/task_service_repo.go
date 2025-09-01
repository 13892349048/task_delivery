package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"taskmanage/internal/assignment"
	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/internal/workflow"
	"taskmanage/pkg/logger"
)

// getUserIDFromContext 从上下文中获取用户ID
func getUserIDFromContext(ctx context.Context) (uint, error) {
	userID := ctx.Value("user_id")
	if userID == nil {
		return 0, fmt.Errorf("用户ID不存在于上下文中")
	}

	if id, ok := userID.(uint); ok {
		return id, nil
	}

	return 0, fmt.Errorf("用户ID类型错误")
}

// taskServiceRepo 基于Repository层的任务服务实现
type taskServiceRepo struct {
	taskRepo          repository.TaskRepository
	employeeRepo      repository.EmployeeRepository
	userRepo          repository.UserRepository
	assignmentRepo    repository.AssignmentRepository
	assignmentService *assignment.AssignmentService
	workflowService   WorkflowService
}

// NewTaskServiceRepo 创建基于Repository的任务服务实例
func NewTaskService(taskRepo repository.TaskRepository, employeeRepo repository.EmployeeRepository, userRepo repository.UserRepository, assignmentRepo repository.AssignmentRepository, assignmentService *assignment.AssignmentService, workflowService WorkflowService) TaskService {
	return &taskServiceRepo{
		taskRepo:          taskRepo,
		employeeRepo:      employeeRepo,
		userRepo:          userRepo,
		assignmentRepo:    assignmentRepo,
		assignmentService: assignmentService,
		workflowService:   workflowService,
	}
}

// CreateTask 创建任务
func (s *taskServiceRepo) CreateTask(ctx context.Context, req *CreateTaskRequest) (*TaskResponse, error) {
	// 验证请求参数
	if req.Title == "" {
		return nil, fmt.Errorf("任务标题不能为空")
	}
	if req.Priority != "" && !isValidPriority(req.Priority) {
		return nil, fmt.Errorf("无效的优先级")
	}

	// 创建任务对象
	var dueDate *time.Time
	if !req.DueDate.IsZero() {
		dueDate = &req.DueDate
	}

	task := &database.Task{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      "pending", // 默认状态为待处理
		DueDate:     dueDate,
		CreatorID:   1, // 暂时硬编码，后续从JWT中获取
	}

	// 保存任务
	if err := s.taskRepo.Create(ctx, task); err != nil {
		logger.Errorf("创建任务失败: %v", err)
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	logger.Infof("任务创建成功: ID=%d, Title=%s", task.ID, task.Title)

	// 转换为响应格式
	return &TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Priority:    task.Priority,
		Status:      task.Status,
		DueDate:     task.DueDate,
		CreatedBy:   task.CreatorID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

// GetTask 获取任务详情
func (s *taskServiceRepo) GetTask(ctx context.Context, taskID uint) (*TaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("任务不存在")
		}
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}

	return &TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Priority:    task.Priority,
		Status:      task.Status,
		DueDate:     task.DueDate,
		CreatedBy:   task.CreatorID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

// UpdateTask 更新任务
func (s *taskServiceRepo) UpdateTask(ctx context.Context, taskID uint, req *UpdateTaskRequest) (*TaskResponse, error) {
	// 验证请求参数
	if req.Title != nil && *req.Title == "" {
		return nil, fmt.Errorf("任务标题不能为空")
	}
	if req.Priority != nil && !isValidPriority(*req.Priority) {
		return nil, fmt.Errorf("无效的优先级")
	}
	if req.Status != nil && !isValidTaskStatus(*req.Status) {
		return nil, fmt.Errorf("无效的任务状态")
	}

	// 获取现有任务
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("任务不存在")
		}
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}

	// 更新任务字段
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}

	// 保存更新
	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Errorf("更新任务失败: %v", err)
		return nil, fmt.Errorf("更新任务失败: %w", err)
	}

	logger.Infof("任务更新成功: ID=%d, Title=%s", task.ID, task.Title)

	return &TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Priority:    task.Priority,
		Status:      task.Status,
		DueDate:     task.DueDate,
		CreatedBy:   task.CreatorID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

// DeleteTask 删除任务
func (s *taskServiceRepo) DeleteTask(ctx context.Context, taskID uint) error {
	// 检查任务是否存在
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("任务不存在")
		}
		return fmt.Errorf("查询任务失败: %w", err)
	}

	// 检查任务是否可以删除
	if task.Status == "in_progress" || task.Status == "completed" {
		return fmt.Errorf("进行中或已完成的任务不能删除")
	}

	// 删除任务
	if err := s.taskRepo.Delete(ctx, taskID); err != nil {
		logger.Errorf("删除任务失败: %v", err)
		return fmt.Errorf("删除任务失败: %w", err)
	}

	logger.Infof("任务删除成功: ID=%d, Title=%s", task.ID, task.Title)
	return nil
}

// ListTasks 获取任务列表
func (s *taskServiceRepo) ListTasks(ctx context.Context, filter TaskListFilter) ([]*TaskResponse, int64, error) {
	// 构建仓库过滤器
	repoFilter := repository.ListFilter{
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Sort:     "created_at",
		Order:    "desc",
	}

	// 添加任务特定的过滤条件
	conditions := make(map[string]interface{})
	if filter.Status != "" {
		conditions["status"] = filter.Status
	}
	if filter.Priority != "" {
		conditions["priority"] = filter.Priority
	}
	if filter.CreatedBy != nil && *filter.CreatedBy != 0 {
		conditions["created_by"] = *filter.CreatedBy
	}

	repoFilter.Filters = conditions

	// 查询任务列表
	tasks, total, err := s.taskRepo.List(ctx, repoFilter)
	if err != nil {
		logger.Errorf("查询任务列表失败: %v", err)
		return nil, 0, fmt.Errorf("查询任务列表失败: %w", err)
	}

	// 转换为响应格式
	responses := make([]*TaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = &TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Priority:    task.Priority,
			Status:      task.Status,
			DueDate:     task.DueDate,
			CreatedBy:   task.CreatorID,
			CreatedAt:   task.CreatedAt,
			UpdatedAt:   task.UpdatedAt,
		}
	}

	return responses, total, nil
}

// 分配任务
func (s *taskServiceRepo) AssignTask(ctx context.Context, req *AssignTaskRequest) (*AssignmentResponse, error) {
	// 获取任务
	task, err := s.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		logger.Errorf("获取任务失败: %v", err)
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	// 验证任务状态 - 只有pending状态的任务可以分配
	if task.Status != "pending" {
		return nil, errors.New("只有待分配状态的任务才能进行分配")
	}

	// 验证员工是否存在
	employee, err := s.employeeRepo.GetByID(ctx, req.AssigneeID)
	if err != nil {
		logger.Errorf("获取员工信息失败: %v", err)
		return nil, fmt.Errorf("员工不存在或获取失败: %w", err)
	}

	// 检查员工工作负载
	if employee.CurrentTasks >= employee.MaxTasks {
		return nil, fmt.Errorf("员工当前任务已达上限(%d/%d)", employee.CurrentTasks, employee.MaxTasks)
	}

	// 启动任务分配审批工作流
	if s.workflowService != nil {
		workflowReq := &workflow.TaskAssignmentApprovalRequest{
			TaskID:      req.TaskID,
			AssigneeID:  req.AssigneeID,
			RequesterID: 1, // TODO: 从上下文获取当前用户ID
			Priority:    task.Priority,
			Reason:      req.Reason,
		}

		instance, err := s.workflowService.StartTaskAssignmentApproval(ctx, workflowReq)
		if err != nil {
			logger.Errorf("启动任务分配审批工作流失败: %v", err)
			return nil, fmt.Errorf("启动审批流程失败: %w", err)
		}

		logger.Infof("任务分配审批工作流已启动: TaskID=%d, WorkflowInstanceID=%s", req.TaskID, instance.ID)

		// 从上下文获取当前用户ID
		currentUserID, err := getUserIDFromContext(ctx)
		if err != nil {
			logger.Warnf("无法从上下文获取用户ID，使用默认值: %v", err)
			currentUserID = 1 // 兜底值
		}

		// 创建待审批的分配记录，用于跟踪工作流状态
		now := time.Now()
		assignment := &database.Assignment{
			TaskID:             req.TaskID,
			AssigneeID:         req.AssigneeID,
			AssignerID:         currentUserID,
			Method:             req.Method,
			Status:             "pending_approval",
			AssignedAt:         now,
			Reason:             req.Reason,
			WorkflowInstanceID: &instance.ID, // 关联工作流实例ID
		}

		// 这里需要保存Assignment记录，但当前没有Assignment repository
		// 暂时记录日志，实际部署时需要保存到数据库
		logger.Infof("创建待审批分配记录: TaskID=%d, AssigneeID=%d, WorkflowInstanceID=%s",
			assignment.TaskID, assignment.AssigneeID, instance.ID)

		// 返回审批中状态的响应，包含工作流实例ID
		return &AssignmentResponse{
			ID:                 0, // Assignment记录ID，需要从数据库获取
			TaskID:             req.TaskID,
			EmployeeID:         req.AssigneeID,
			Status:             "pending_approval",
			AssignedBy:         currentUserID,
			AssignedAt:         now,
			WorkflowInstanceID: instance.ID, // 返回工作流实例ID
			Comment:            fmt.Sprintf("任务分配审批流程已启动，工作流实例ID: %s", instance.ID),
		}, nil
	}

	// 如果没有工作流服务，则直接分配（向后兼容）
	logger.Warnf("工作流服务不可用，直接执行任务分配")

	// 更新任务状态和分配信息
	task.Status = "assigned"
	task.AssigneeID = &employee.UserID

	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Errorf("更新任务分配失败: %v", err)
		return nil, fmt.Errorf("更新任务分配失败: %w", err)
	}

	// 更新员工当前任务数
	employee.CurrentTasks++
	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		logger.Warnf("更新员工任务数失败: %v", err)
	}

	logger.Infof("任务直接分配成功: TaskID=%d, EmployeeID=%d", req.TaskID, req.AssigneeID)

	return &AssignmentResponse{
		ID:         0,
		TaskID:     req.TaskID,
		EmployeeID: req.AssigneeID,
		Status:     "assigned",
		AssignedBy: 1, // TODO: 从上下文获取当前用户ID
		AssignedAt: time.Now(),
		Comment:    "任务已直接分配",
	}, nil
}

func (s *taskServiceRepo) ReassignTask(ctx context.Context, taskID uint, req *ReassignTaskRequest) (*AssignmentResponse, error) {
	return nil, errors.New("功能暂未实现")
}

func (s *taskServiceRepo) ApproveAssignment(ctx context.Context, assignmentID uint, req *ApproveAssignmentRequest) error {
	return errors.New("功能暂未实现")
}

// CompleteTaskAssignmentWorkflow 完成任务分配工作流
// 当工作流审批通过时调用此方法完成实际的任务分配
func (s *taskServiceRepo) CompleteTaskAssignmentWorkflow(ctx context.Context, workflowInstanceID string, approved bool, approverID uint) error {
	logger.Infof("完成任务分配工作流: InstanceID=%s, Approved=%v, ApproverID=%d", workflowInstanceID, approved, approverID)

	// 根据工作流实例ID查找对应的Assignment记录
	var assignment *database.Assignment
	if s.assignmentRepo != nil {
		// 查找工作流实例ID对应的分配记录，不过滤状态
		assignments, total, err := s.assignmentRepo.List(ctx, repository.ListFilter{
			Page:     1,
			PageSize: 20,
			Filters: map[string]interface{}{
				"workflow_instance_id": workflowInstanceID,
			},
		})
		if err != nil {
			return fmt.Errorf("查找分配记录失败: %w", err)
		}
		if total == 0 {
			return fmt.Errorf("未找到对应的分配记录")
		}
		assignment = assignments[0]
	} else {
		return fmt.Errorf("分配仓库未配置")
	}

	// 获取任务信息
	task, err := s.taskRepo.GetByID(ctx, assignment.TaskID)
	if err != nil {
		return fmt.Errorf("获取任务失败: %w", err)
	}

	now := time.Now()

	if approved {
		// 审批通过：更新任务状态为已分配
		task.Status = "assigned"
		task.AssigneeID = &assignment.AssigneeID

		if err := s.taskRepo.Update(ctx, task); err != nil {
			return fmt.Errorf("更新任务分配失败: %w", err)
		}

		// 更新分配记录状态
		assignment.Status = "approved"
		assignment.ApprovedAt = &now
		assignment.ApproverID = &approverID

		// 更新员工工作负载
		employee, err := s.employeeRepo.GetByID(ctx, assignment.AssigneeID)
		if err == nil && employee != nil {
			employee.CurrentTasks++
			if employee.CurrentTasks >= employee.MaxTasks {
				employee.Status = "busy"
			}
			s.employeeRepo.Update(ctx, employee)
		}

		logger.Infof("任务分配审批通过: TaskID=%d, AssigneeID=%d", assignment.TaskID, assignment.AssigneeID)
	} else {
		// 审批拒绝：重置任务状态为待分配
		task.Status = "pending"
		task.AssigneeID = nil

		if err := s.taskRepo.Update(ctx, task); err != nil {
			return fmt.Errorf("重置任务状态失败: %w", err)
		}

		// 更新分配记录状态
		assignment.Status = "rejected"
		assignment.ApprovedAt = &now
		assignment.ApproverID = &approverID

		logger.Infof("任务分配审批拒绝: TaskID=%d, AssigneeID=%d", assignment.TaskID, assignment.AssigneeID)
	}

	// 保存分配记录更新
	if err := s.assignmentRepo.Update(ctx, assignment); err != nil {
		return fmt.Errorf("更新分配记录失败: %w", err)
	}

	return nil
}

func (s *taskServiceRepo) RejectAssignment(ctx context.Context, assignmentID uint, req *RejectAssignmentRequest) error {
	return errors.New("功能暂未实现")
}

func (s *taskServiceRepo) StartTask(ctx context.Context, taskID uint, userID uint) error {
	// 获取任务
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		logger.Errorf("获取任务失败: %v", err)
		return fmt.Errorf("获取任务失败: %w", err)
	}

	// 验证任务状态
	if task.Status != "assigned" {
		return errors.New("只有已分配的任务才能开始")
	}

	// 验证用户权限 - 只有被分配者才能开始任务
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return errors.New("只有任务被分配者才能开始任务")
	}

	// 更新任务状态
	task.Status = "in_progress"
	task.StartedAt = &time.Time{}
	*task.StartedAt = time.Now()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Errorf("更新任务状态失败: %v", err)
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	logger.Infof("任务开始成功: TaskID=%d, UserID=%d", taskID, userID)
	return nil
}

func (s *taskServiceRepo) CompleteTask(ctx context.Context, taskID uint, userID uint, req *CompleteTaskRequest) error {
	// 获取任务
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		logger.Errorf("获取任务失败: %v", err)
		return fmt.Errorf("获取任务失败: %w", err)
	}

	// 验证任务状态
	if task.Status != "in_progress" {
		return errors.New("只有进行中的任务才能完成")
	}

	// 验证用户权限 - 只有被分配者才能完成任务
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return errors.New("只有任务被分配者才能完成任务")
	}

	// 更新任务状态
	task.Status = "completed"
	task.CompletedAt = &time.Time{}
	*task.CompletedAt = time.Now()

	// 如果提供了实际工时，更新实际工时
	if req.Comment != "" {
		// 这里可以添加评论逻辑，暂时跳过
		logger.Infof("任务完成评论: %s", req.Comment)
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Errorf("更新任务状态失败: %v", err)
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	// 更新员工当前任务数
	if task.AssigneeID != nil {
		employee, err := s.employeeRepo.GetByUserID(ctx, *task.AssigneeID)
		if err == nil && employee != nil {
			if employee.CurrentTasks > 0 {
				employee.CurrentTasks--
				s.employeeRepo.Update(ctx, employee)
			}
		}
	}

	logger.Infof("任务完成成功: TaskID=%d, UserID=%d", taskID, userID)
	return nil
}

func (s *taskServiceRepo) CancelTask(ctx context.Context, taskID uint, userID uint, reason string) error {
	// 获取任务
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		logger.Errorf("获取任务失败: %v", err)
		return fmt.Errorf("获取任务失败: %w", err)
	}

	// 验证任务状态 - 只有pending, assigned, in_progress状态的任务可以取消
	if task.Status == "completed" || task.Status == "cancelled" {
		return errors.New("已完成或已取消的任务不能再次取消")
	}

	// 验证用户权限 - 创建者或被分配者都可以取消任务
	canCancel := task.CreatorID == userID
	if task.AssigneeID != nil && *task.AssigneeID == userID {
		canCancel = true
	}
	if !canCancel {
		return errors.New("只有任务创建者或被分配者才能取消任务")
	}

	// 更新任务状态
	task.Status = "cancelled"
	task.CompletedAt = &time.Time{}
	*task.CompletedAt = time.Now()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Errorf("更新任务状态失败: %v", err)
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	// 如果任务已分配，减少员工当前任务数
	if task.AssigneeID != nil {
		employee, err := s.employeeRepo.GetByUserID(ctx, *task.AssigneeID)
		if err == nil && employee != nil {
			if employee.CurrentTasks > 0 {
				employee.CurrentTasks--
				s.employeeRepo.Update(ctx, employee)
			}
		}
	}

	logger.Infof("任务取消成功: TaskID=%d, UserID=%d, Reason=%s", taskID, userID, reason)
	return nil
}

func (s *taskServiceRepo) AutoAssignTask(ctx context.Context, taskID uint, strategy AssignmentStrategy) (*AssignmentResponse, error) {
	if s.assignmentService == nil {
		return nil, errors.New("分配服务未初始化")
	}

	// 获取任务信息
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	// 验证任务状态
	if task.Status != "pending" {
		return nil, errors.New("只有待分配状态的任务才能进行自动分配")
	}

	// 构建分配请求
	req := &assignment.AssignmentRequest{
		TaskID:   taskID,
		Strategy: assignment.AssignmentStrategy(strategy),
		Priority: task.Priority,
	}

	if task.DueDate != nil {
		req.Deadline = task.DueDate
	}

	// 执行自动分配
	result, err := s.assignmentService.AssignTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("自动分配失败: %w", err)
	}

	// 更新任务状态和分配信息
	task.Status = "assigned"
	task.AssigneeID = &result.SelectedEmployee.UserID

	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Errorf("更新任务分配失败: %v", err)
		return nil, fmt.Errorf("更新任务分配失败: %w", err)
	}

	// 更新员工当前任务数
	employee, err := s.employeeRepo.GetByID(ctx, result.SelectedEmployee.ID)
	if err == nil && employee != nil {
		employee.CurrentTasks++
		s.employeeRepo.Update(ctx, employee)
	}

	// 返回分配响应
	return &AssignmentResponse{
		ID:         0, // 这里需要实际的分配记录ID
		TaskID:     taskID,
		EmployeeID: result.SelectedEmployee.ID,
		Status:     "approved",
		AssignedBy: 1, // TODO: 从上下文获取当前用户ID
		AssignedAt: result.ExecutedAt,
		Comment:    result.Reason,
	}, nil
}

func (s *taskServiceRepo) GetAssignmentSuggestions(ctx context.Context, taskID uint) ([]*AssignmentSuggestion, error) {
	if s.assignmentService == nil {
		return nil, errors.New("分配服务未初始化")
	}

	// 获取任务信息
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	// 验证任务状态
	if task.Status != "pending" {
		return nil, errors.New("只有待分配状态的任务才能获取分配建议")
	}

	suggestions := make([]*AssignmentSuggestion, 0)

	// 获取所有可用策略
	fmt.Println("---")
	strategies := s.assignmentService.GetAvailableStrategies()
	fmt.Println("---", strategies)
	for _, strategyInfo := range strategies {
		// 构建分配请求
		req := &assignment.AssignmentRequest{
			TaskID:   taskID,
			Strategy: strategyInfo.Strategy,
			Priority: task.Priority,
		}

		if task.DueDate != nil {
			req.Deadline = task.DueDate
		}

		// 预览分配结果
		result, err := s.assignmentService.PreviewAssignment(ctx, req)
		if err != nil {
			logger.Warnf("预览分配策略 %s 失败: %v", strategyInfo.Strategy, err)
			continue
		}

		// 创建分配建议
		suggestion := &AssignmentSuggestion{
			Employee:   &result.SelectedEmployee,
			Score:      result.Score,
			Reason:     result.Reason,
			Confidence: result.Score,
			Workload: WorkloadInfo{
				CurrentTasks:    result.SelectedEmployee.CurrentTasks,
				MaxTasks:        10, // 默认最大任务数
				UtilizationRate: float64(result.SelectedEmployee.CurrentTasks) / 10.0 * 100,
				AvgTaskDuration: 0,
			},
			SkillMatch:   result.Score,
			Availability: 100.0 - (float64(result.SelectedEmployee.CurrentTasks) / 10.0 * 100),
		}

		suggestions = append(suggestions, suggestion)
	}

	// 按评分排序（降序）
	for i := 0; i < len(suggestions)-1; i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[i].Score < suggestions[j].Score {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	return suggestions, nil
}

// calculateConfidence 根据评分计算置信度
func (s *taskServiceRepo) calculateConfidence(score float64) string {
	if score >= 90 {
		return "high"
	} else if score >= 70 {
		return "medium"
	} else {
		return "low"
	}
}

// 验证优先级
func isValidPriority(priority string) bool {
	validPriorities := []string{"low", "medium", "high", "urgent"}
	for _, p := range validPriorities {
		if p == priority {
			return true
		}
	}
	return false
}

// 验证任务状态
func isValidTaskStatus(status string) bool {
	validStatuses := []string{"pending", "assigned", "in_progress", "completed", "cancelled"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// StartTaskAssignmentApproval 启动任务分配审批流程
func (s *taskServiceRepo) StartTaskAssignmentApproval(ctx context.Context, req *StartTaskAssignmentApprovalRequest) (*TaskAssignmentApprovalResponse, error) {
	if s.workflowService == nil {
		// 如果没有工作流服务，直接执行分配
		logger.Warn("工作流服务未配置，直接执行任务分配")
		return s.directAssignTask(ctx, req)
	}

	// 启动审批流程
	workflowReq := &workflow.TaskAssignmentApprovalRequest{
		TaskID:         req.TaskID,
		AssigneeID:     req.AssigneeID,
		AssignmentType: assignment.AssignmentStrategy(req.AssignmentType),
		Priority:       req.Priority,
		RequesterID:    req.RequesterID,
		Reason:         req.Reason,
	}

	instance, err := s.workflowService.StartTaskAssignmentApproval(ctx, workflowReq)
	if err != nil {
		return nil, fmt.Errorf("启动任务分配审批流程失败: %w", err)
	}

	return &TaskAssignmentApprovalResponse{
		WorkflowInstanceID: instance.ID,
		Status:             "pending_approval",
		Message:            "任务分配审批流程已启动",
		CreatedAt:          instance.StartedAt,
	}, nil
}

// ProcessTaskAssignmentApproval 处理任务分配审批
func (s *taskServiceRepo) ProcessTaskAssignmentApproval(ctx context.Context, req *ProcessTaskAssignmentApprovalRequest) (*TaskAssignmentApprovalResponse, error) {
	if s.workflowService == nil {
		return nil, fmt.Errorf("工作流服务未配置")
	}

	processReq := &workflow.ApprovalRequest{
		InstanceID: req.InstanceID,
		NodeID:     req.NodeID,
		Action:     workflow.ApprovalAction(req.Action),
		Comment:    req.Comment,
		Variables:  req.Variables,
		ApprovedBy: req.ApprovedBy,
	}

	result, err := s.workflowService.ProcessTaskAssignmentApproval(ctx, processReq)
	if err != nil {
		return nil, fmt.Errorf("处理任务分配审批失败: %w", err)
	}

	status := "pending_approval"
	if result.IsCompleted {
		if result.Action == workflow.ActionApprove {
			status = "approved"
		} else if result.Action == workflow.ActionReject {
			status = "rejected"
		}
	}

	return &TaskAssignmentApprovalResponse{
		WorkflowInstanceID: req.InstanceID,
		Status:             status,
		Message:            result.Message,
		CreatedAt:          result.ExecutedAt,
	}, nil
}

// GetPendingTaskAssignmentApprovals 获取待审批的任务分配
func (s *taskServiceRepo) GetPendingTaskAssignmentApprovals(ctx context.Context, userID uint) ([]*PendingTaskAssignmentApproval, error) {
	if s.workflowService == nil {
		return []*PendingTaskAssignmentApproval{}, nil
	}

	approvals, err := s.workflowService.GetPendingTaskAssignmentApprovals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取待审批任务失败: %w", err)
	}

	var result []*PendingTaskAssignmentApproval
	for _, approval := range approvals {
		// 转换RequiredAction从[]workflow.ApprovalAction到[]string
		var requiredActions []string
		for _, action := range approval.RequiredAction {
			requiredActions = append(requiredActions, string(action))
		}

		result = append(result, &PendingTaskAssignmentApproval{
			InstanceID:     approval.InstanceID,
			WorkflowName:   approval.WorkflowName,
			NodeID:         approval.NodeID,
			NodeName:       approval.NodeName,
			TaskID:         approval.BusinessID,
			Priority:       approval.Priority,
			AssignedTo:     approval.AssignedTo,
			CreatedAt:      approval.CreatedAt,
			Deadline:       approval.Deadline,
			CanDelegate:    approval.CanDelegate,
			RequiredAction: requiredActions,
		})
	}

	return result, nil
}

// directAssignTask 直接分配任务（无审批流程）
func (s *taskServiceRepo) directAssignTask(ctx context.Context, req *StartTaskAssignmentApprovalRequest) (*TaskAssignmentApprovalResponse, error) {
	// 获取任务信息
	task, err := s.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	// 更新任务分配
	task.AssigneeID = &req.AssigneeID
	task.Status = "assigned"

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("更新任务分配失败: %w", err)
	}

	// 更新员工工作负载
	employee, err := s.employeeRepo.GetByID(ctx, req.AssigneeID)
	if err == nil && employee != nil {
		employee.CurrentTasks++
		s.employeeRepo.Update(ctx, employee)
	}

	return &TaskAssignmentApprovalResponse{
		WorkflowInstanceID: "",
		Status:             "approved",
		Message:            "任务直接分配成功",
		CreatedAt:          time.Now(),
	}, nil
}
