package service

import (
	"context"
	"fmt"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// EmployeeServiceImpl 员工服务实现
type EmployeeServiceImpl struct {
	employeeRepo repository.EmployeeRepository
	skillRepo    repository.SkillRepository
	userRepo     repository.UserRepository
}

// NewEmployeeService 创建员工服务实例
func NewEmployeeService(
	employeeRepo repository.EmployeeRepository,
	skillRepo repository.SkillRepository,
	userRepo repository.UserRepository,
) EmployeeService {
	return &EmployeeServiceImpl{
		employeeRepo: employeeRepo,
		skillRepo:    skillRepo,
		userRepo:     userRepo,
	}
}

// CreateEmployee 创建员工
func (s *EmployeeServiceImpl) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*EmployeeResponse, error) {
	logger.Infof("Creating employee: %s", req.Name)

	// 检查邮箱是否已存在
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("email already exists: %s", req.Email)
	}

	// 创建用户记录
	user := &database.User{
		Username: req.Email, // 使用邮箱作为用户名
		Email:    req.Email,
		RealName: req.Name,
		Status:   "active",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Errorf("Failed to create user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 生成员工编号（简单实现）
	employeeNo := fmt.Sprintf("EMP%06d", user.ID)

	// 创建员工记录
	employee := &database.Employee{
		UserID:       user.ID,
		EmployeeNo:   employeeNo,
		DepartmentID: &req.DepartmentID,
		PositionID:   &req.PositionID,
		OnboardingStatus: "pending_onboard", // 默认待入职状态
		Status:       "available",
		MaxTasks:     req.MaxConcurrentTasks,
		CurrentTasks: 0,
	}

	if err := s.employeeRepo.Create(ctx, employee); err != nil {
		logger.Errorf("Failed to create employee: %v", err)
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	logger.Infof("Employee created successfully: %d", employee.ID)

	return s.buildEmployeeResponse(employee, user), nil
}

// GetEmployee 获取员工信息
func (s *EmployeeServiceImpl) GetEmployee(ctx context.Context, employeeID uint) (*EmployeeResponse, error) {
	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, employee.UserID)
	if err != nil {
		logger.Errorf("Failed to get user info: %v", err)
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// 获取员工技能
	employeeWithSkills, err := s.employeeRepo.GetEmployeeWithSkills(ctx, employeeID)
	if err != nil {
		logger.Warnf("Failed to get employee skills: %v", err)
		employeeWithSkills = employee // 使用基本信息
	}

	return s.buildEmployeeResponse(employeeWithSkills, user), nil
}

// UpdateEmployee 更新员工信息
func (s *EmployeeServiceImpl) UpdateEmployee(ctx context.Context, employeeID uint, req *UpdateEmployeeRequest) (*EmployeeResponse, error) {
	logger.Infof("Updating employee: %d", employeeID)

	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	// 更新字段
	if req.DepartmentID != nil {
		employee.DepartmentID = req.DepartmentID
	}
	if req.PositionID != nil {
		employee.PositionID = req.PositionID
	}
	if req.MaxConcurrentTasks != nil {
		employee.MaxTasks = *req.MaxConcurrentTasks
	}

	// 更新用户信息
	if req.Name != nil || req.Email != nil {
		user, err := s.userRepo.GetByID(ctx, employee.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		if req.Name != nil {
			user.RealName = *req.Name
		}
		if req.Email != nil {
			user.Email = *req.Email
		}
		if err := s.userRepo.Update(ctx, user); err != nil {
			logger.Errorf("Failed to update user: %v", err)
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		logger.Errorf("Failed to update employee: %v", err)
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	// 获取更新后的用户信息
	user, err := s.userRepo.GetByID(ctx, employee.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	logger.Infof("Employee updated successfully: %d", employeeID)

	return s.buildEmployeeResponse(employee, user), nil
}

// DeleteEmployee 删除员工
func (s *EmployeeServiceImpl) DeleteEmployee(ctx context.Context, employeeID uint) error {
	logger.Infof("Deleting employee: %d", employeeID)

	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return fmt.Errorf("employee not found: %w", err)
	}

	// 检查是否有未完成的任务
	if employee.CurrentTasks > 0 {
		return fmt.Errorf("cannot delete employee with active tasks")
	}

	if err := s.employeeRepo.Delete(ctx, employeeID); err != nil {
		logger.Errorf("Failed to delete employee: %v", err)
		return fmt.Errorf("failed to delete employee: %w", err)
	}

	logger.Infof("Employee deleted successfully: %d", employeeID)
	return nil
}

// ListEmployees 获取员工列表
func (s *EmployeeServiceImpl) ListEmployees(ctx context.Context, filter EmployeeListFilter) ([]*EmployeeResponse, int64, error) {
	// 构建repository过滤器
	repoFilter := repository.ListFilter{
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Sort:     "created_at", // 默认按创建时间排序
		Order:    "desc",       // 默认降序
	}

	employees, total, err := s.employeeRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list employees: %w", err)
	}

	// 转换为响应格式

	responses := make([]*EmployeeResponse, 0, len(employees))

	for _, employee := range employees {
		// 获取用户信息
		user, err := s.userRepo.GetByID(ctx, employee.UserID)
		if err != nil {
			logger.Warnf("Failed to get user info for employee %d: %v", employee.ID, err)
			continue
		}

		responses = append(responses, s.buildEmployeeResponse(employee, user))
	}

	return responses, total, nil
}

// UpdateEmployeeStatus 更新员工状态
func (s *EmployeeServiceImpl) UpdateEmployeeStatus(ctx context.Context, employeeID uint, status string) error {
	logger.Infof("Updating employee %d status to: %s", employeeID, status)

	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return fmt.Errorf("employee not found: %w", err)
	}

	// 验证状态值
	validStatuses := map[string]bool{
		"available":   true,
		"busy":        true,
		"unavailable": true,
		"on_leave":    true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	employee.Status = status
	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		logger.Errorf("Failed to update employee status: %v", err)
		return fmt.Errorf("failed to update employee status: %w", err)
	}

	logger.Infof("Employee %d status updated to: %s", employeeID, status)
	return nil
}

// GetEmployeeWorkload 获取员工工作负载
func (s *EmployeeServiceImpl) GetEmployeeWorkload(ctx context.Context, employeeID uint) (*WorkloadResponse, error) {
	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	workload := &WorkloadResponse{
		EmployeeID:      employeeID,
		EmployeeName:    employee.User.RealName,
		Department:      employee.Department.Name, // 使用关联的部门名称
		ActiveTasks:     employee.CurrentTasks,
		PendingTasks:    0, // TODO: 从任务表统计
		CompletedTasks:  0, // TODO: 从任务表统计
		OverdueTasks:    0, // TODO: 从任务表统计
		MaxTasks:        employee.MaxTasks,
		WorkloadRate:    0,
		EfficiencyRate:  0, // TODO: 计算效率率
		AvgTaskDuration: 0, // TODO: 计算平均任务时长
		Status:          employee.Status,
		LastActiveTime:  employee.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if employee.MaxTasks > 0 {
		workload.WorkloadRate = float64(employee.CurrentTasks) / float64(employee.MaxTasks)
	}

	return workload, nil
}

// GetAvailableEmployees 获取可用员工列表
func (s *EmployeeServiceImpl) GetAvailableEmployees(ctx context.Context) ([]*EmployeeResponse, error) {
	employees, err := s.employeeRepo.GetAvailableEmployees(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available employees: %w", err)
	}

	responses := make([]*EmployeeResponse, 0, len(employees))
	for _, employee := range employees {
		// 获取用户信息
		user, err := s.userRepo.GetByID(ctx, employee.UserID)
		if err != nil {
			logger.Warnf("Failed to get user info for employee %d: %v", employee.ID, err)
			continue
		}

		responses = append(responses, s.buildEmployeeResponse(employee, user))
	}

	return responses, nil
}

// AddSkill 为员工添加技能
func (s *EmployeeServiceImpl) AddSkill(ctx context.Context, employeeID uint, req *SkillRequest) error {
	logger.Infof("Adding skill %d to employee %d with level %d", req.SkillID, employeeID, req.Level)

	// 检查员工是否存在
	_, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return fmt.Errorf("employee not found: %w", err)
	}

	// 检查技能是否存在
	_, err = s.skillRepo.GetByID(ctx, req.SkillID)
	if err != nil {
		return fmt.Errorf("skill not found: %w", err)
	}

	// 验证技能等级
	if req.Level < 1 || req.Level > 5 {
		return fmt.Errorf("invalid skill level: %d (must be 1-5)", req.Level)
	}

	if err := s.skillRepo.AssignToEmployee(ctx, employeeID, req.SkillID, req.Level); err != nil {
		logger.Errorf("Failed to assign skill to employee: %v", err)
		return fmt.Errorf("failed to assign skill: %w", err)
	}

	logger.Infof("Skill %d added to employee %d successfully", req.SkillID, employeeID)
	return nil
}

// RemoveSkill 移除员工技能
func (s *EmployeeServiceImpl) RemoveSkill(ctx context.Context, employeeID uint, req *RemoveSkillRequest) error {
	logger.Infof("Removing skill %d from employee %d", req.SkillID, employeeID)

	// 检查员工是否存在
	_, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return fmt.Errorf("employee not found: %w", err)
	}

	if err := s.skillRepo.RemoveFromEmployee(ctx, employeeID, req.SkillID); err != nil {
		logger.Errorf("Failed to remove skill from employee: %v", err)
		return fmt.Errorf("failed to remove skill: %w", err)
	}

	logger.Infof("Skill %d removed from employee %d successfully", req.SkillID, employeeID)
	return nil
}

// GetEmployeesByStatus 根据状态获取员工列表
func (s *EmployeeServiceImpl) GetEmployeesByStatus(ctx context.Context, status string) ([]*EmployeeResponse, error) {
	logger.Infof("Getting employees by status: %s", status)

	employees, err := s.employeeRepo.GetByStatus(ctx, status)
	if err != nil {
		logger.Errorf("Failed to get employees by status: %v", err)
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}

	responses := make([]*EmployeeResponse, 0, len(employees))
	for _, employee := range employees {
		user, err := s.userRepo.GetByID(ctx, employee.UserID)
		if err != nil {
			logger.Warnf("Failed to get user info for employee %d: %v", employee.ID, err)
			continue
		}
		responses = append(responses, s.buildEmployeeResponse(employee, user))
	}

	return responses, nil
}

// GetWorkloadStats 获取工作负载统计
func (s *EmployeeServiceImpl) GetWorkloadStats(ctx context.Context, req *WorkloadStatsRequest) ([]*WorkloadResponse, error) {
	logger.Infof("Getting workload stats for department: %s", req.Department)

	var employees []*database.Employee
	var err error

	if req.DepartmentID != 0 {
		// TODO: 需要实现GetByDepartmentID方法或使用现有方法
		employees, err = s.employeeRepo.GetAll(ctx)
	} else {
		employees, err = s.employeeRepo.GetAll(ctx)
	}

	if err != nil {
		logger.Errorf("Failed to get employees for workload stats: %v", err)
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}

	responses := make([]*WorkloadResponse, 0, len(employees))
	for _, employee := range employees {
		workload, err := s.GetEmployeeWorkload(ctx, employee.ID)
		if err != nil {
			logger.Warnf("Failed to get workload for employee %d: %v", employee.ID, err)
			continue
		}
		responses = append(responses, workload)
	}

	return responses, nil
}

// GetDepartmentWorkload 获取部门工作负载统计
func (s *EmployeeServiceImpl) GetDepartmentWorkload(ctx context.Context, departmentID uint) (*DepartmentWorkloadResponse, error) {
	logger.Infof("Getting department workload for: %d", departmentID)

	// TODO: 需要实现GetByDepartmentID方法或使用现有方法
	employees, err := s.employeeRepo.GetAll(ctx)
	if err != nil {
		logger.Errorf("Failed to get department employees: %v", err)
		return nil, fmt.Errorf("failed to get department employees: %w", err)
	}

	var totalEmployees, activeEmployees, totalTasks, completedTasks int
	var totalWorkloadRate float64

	for _, employee := range employees {
		totalEmployees++
		if employee.Status == "active" || employee.Status == "available" {
			activeEmployees++
		}

		workload, err := s.GetEmployeeWorkload(ctx, employee.ID)
		if err != nil {
			logger.Warnf("Failed to get workload for employee %d: %v", employee.ID, err)
			continue
		}

		totalTasks += workload.ActiveTasks + workload.PendingTasks
		completedTasks += workload.CompletedTasks
		totalWorkloadRate += workload.WorkloadRate
	}

	avgWorkloadRate := float64(0)
	if totalEmployees > 0 {
		avgWorkloadRate = totalWorkloadRate / float64(totalEmployees)
	}

	// 获取部门名称
	departmentName := "Unknown"
	if len(employees) > 0 {
		departmentName = employees[0].Department.Name
	}

	return &DepartmentWorkloadResponse{
		Department:      departmentName,
		TotalEmployees:  totalEmployees,
		ActiveEmployees: activeEmployees,
		TotalTasks:      totalTasks,
		CompletedTasks:  completedTasks,
		AvgWorkloadRate: avgWorkloadRate,
		OverloadedCount: 0, // TODO: 计算超负荷员工数量
	}, nil
}

// buildEmployeeResponse 构建员工响应对象
func (s *EmployeeServiceImpl) buildEmployeeResponse(employee *database.Employee, user *database.User) *EmployeeResponse {
	response := &EmployeeResponse{
		ID:                 employee.ID,
		Name:               user.RealName,
		Email:              user.Email,
		Department:         employee.Department.Name,
		Position:           employee.Position.Name,
		Status:             employee.Status,
		MaxConcurrentTasks: employee.MaxTasks,
		CurrentTasks:       employee.CurrentTasks,
		CreatedAt:          employee.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          employee.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// 获取员工技能信息（包括级别）
	ctx := context.Background()
	employeeSkills, err := s.skillRepo.GetEmployeeSkills(ctx, employee.ID)
	if err != nil {
		logger.Warnf("Failed to get employee skills for employee %d: %v", employee.ID, err)
		response.Skills = []SkillResponse{}
	} else {
		response.Skills = make([]SkillResponse, 0, len(employeeSkills))
		for _, skill := range employeeSkills {
			// 获取技能级别
			level, err := s.skillRepo.GetEmployeeSkillLevel(ctx, employee.ID, skill.ID)
			if err != nil {
				logger.Warnf("Failed to get skill level for employee %d, skill %d: %v", employee.ID, skill.ID, err)
				level = 1 // 默认级别
			}

			response.Skills = append(response.Skills, SkillResponse{
				ID:          skill.ID,
				Name:        skill.Name,
				Category:    skill.Category,
				Description: skill.Description,
				//Tags:        []string{}, // TODO: 解析JSON格式的tags
				Level:     level, // 员工技能级别
				CreatedAt: skill.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt: skill.UpdatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	return response
}
