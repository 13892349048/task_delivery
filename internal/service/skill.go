package service

import (
	"context"
	"fmt"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// SkillServiceImpl 技能服务实现
type SkillServiceImpl struct {
	skillRepo    repository.SkillRepository
	employeeRepo repository.EmployeeRepository
}

// NewSkillService 创建技能服务实例
func NewSkillService(
	skillRepo repository.SkillRepository,
	employeeRepo repository.EmployeeRepository,
) SkillService {
	return &SkillServiceImpl{
		skillRepo:    skillRepo,
		employeeRepo: employeeRepo,
	}
}

// CreateSkill 创建技能
func (s *SkillServiceImpl) CreateSkill(ctx context.Context, req *CreateSkillRequest) (*SkillResponse, error) {
	logger.Infof("Creating skill: %s", req.Name)

	// 检查技能名称是否已存在
	existing, _ := s.skillRepo.GetByName(ctx, req.Name)
	if existing != nil {
		return nil, fmt.Errorf("skill name already exists: %s", req.Name)
	}

	// 创建技能记录
	skill := &database.Skill{
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
	}

	if err := s.skillRepo.Create(ctx, skill); err != nil {
		logger.Errorf("Failed to create skill: %v", err)
		return nil, fmt.Errorf("failed to create skill: %w", err)
	}

	logger.Infof("Skill created successfully: %d", skill.ID)

	return s.buildSkillResponse(skill), nil
}

// GetSkill 获取技能详情
func (s *SkillServiceImpl) GetSkill(ctx context.Context, skillID uint) (*SkillResponse, error) {
	skill, err := s.skillRepo.GetByID(ctx, skillID)
	if err != nil {
		return nil, fmt.Errorf("skill not found: %w", err)
	}

	return s.buildSkillResponse(skill), nil
}

// UpdateSkill 更新技能信息
func (s *SkillServiceImpl) UpdateSkill(ctx context.Context, skillID uint, req *UpdateSkillRequest) (*SkillResponse, error) {
	logger.Infof("Updating skill: %d", skillID)

	skill, err := s.skillRepo.GetByID(ctx, skillID)
	if err != nil {
		return nil, fmt.Errorf("skill not found: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		// 检查新名称是否已存在
		if *req.Name != skill.Name {
			existing, _ := s.skillRepo.GetByName(ctx, *req.Name)
			if existing != nil {
				return nil, fmt.Errorf("skill name already exists: %s", *req.Name)
			}
		}
		skill.Name = *req.Name
	}
	if req.Category != nil {
		skill.Category = *req.Category
	}
	if req.Description != nil {
		skill.Description = *req.Description
	}

	if err := s.skillRepo.Update(ctx, skill); err != nil {
		logger.Errorf("Failed to update skill: %v", err)
		return nil, fmt.Errorf("failed to update skill: %w", err)
	}

	logger.Infof("Skill updated successfully: %d", skillID)

	return s.buildSkillResponse(skill), nil
}

// DeleteSkill 删除技能
func (s *SkillServiceImpl) DeleteSkill(ctx context.Context, skillID uint) error {
	logger.Infof("Deleting skill: %d", skillID)

	// 检查技能是否存在
	_, err := s.skillRepo.GetByID(ctx, skillID)
	if err != nil {
		return fmt.Errorf("skill not found: %w", err)
	}

	// 检查是否有员工使用该技能
	// 这里可以添加检查逻辑，暂时允许删除

	// 删除技能
	err = s.skillRepo.Delete(ctx, skillID)
	if err != nil {
		logger.Error("删除技能失败", "skill_id", skillID, "error", err)
		return err
	}

	logger.Info("技能删除成功", "skill_id", skillID)
	return nil
}

// ListSkills 获取技能列表
func (s *SkillServiceImpl) ListSkills(ctx context.Context, req *ListSkillsRequest) (*ListSkillsResponse, error) {
	// 构建repository过滤器
	repoFilter := repository.ListFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Sort:     "created_at",
		Order:    "desc",
	}

	var skills []*database.Skill
	var total int64
	var err error

	// 根据分类过滤
	if req.Category != "" {
		skills, err = s.skillRepo.GetByCategory(ctx, req.Category)
		if err != nil {
			return nil, fmt.Errorf("failed to list skills by category: %w", err)
		}
		total = int64(len(skills))
	} else {
		skills, total, err = s.skillRepo.List(ctx, repoFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to list skills: %w", err)
		}
	}

	// 转换为响应格式
	responses := make([]*SkillResponse, 0, len(skills))
	for _, skill := range skills {
		responses = append(responses, &SkillResponse{
			ID:          skill.ID,
			Name:        skill.Name,
			Category:    skill.Category,
			Description: skill.Description,
			CreatedAt:   skill.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   skill.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &ListSkillsResponse{
		ListResponse: ListResponse[*SkillResponse]{
			Items: responses,
			Total: total,
			Page:  req.Page,
			Size:  req.PageSize,
		},
	}, nil
}

// GetSkillCategories 获取技能分类列表（兼容性方法）
func (s *SkillServiceImpl) GetSkillCategories(ctx context.Context) ([]string, error) {
	return s.GetAllCategories(ctx)
}

// GetAllCategories 获取所有技能分类
func (s *SkillServiceImpl) GetAllCategories(ctx context.Context) ([]string, error) {
	categories, err := s.skillRepo.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill categories: %w", err)
	}

	return categories, nil
}

// GetSkillsByCategory 根据分类获取技能列表
func (s *SkillServiceImpl) GetSkillsByCategory(ctx context.Context, category string) ([]*SkillResponse, error) {
	skills, err := s.skillRepo.GetByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get skills by category: %w", err)
	}

	// 转换为响应格式
	responses := make([]*SkillResponse, 0, len(skills))
	for _, skill := range skills {
		responses = append(responses, &SkillResponse{
			ID:          skill.ID,
			Name:        skill.Name,
			Category:    skill.Category,
			Description: skill.Description,
			CreatedAt:   skill.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   skill.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

// AssignSkillToEmployee 为员工分配技能
func (s *SkillServiceImpl) AssignSkillToEmployee(ctx context.Context, employeeID, skillID uint, level int) error {
	logger.Infof("Assigning skill %d to employee %d with level %d", skillID, employeeID, level)

	// 检查员工是否存在
	_, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return fmt.Errorf("employee not found: %w", err)
	}

	// 检查技能是否存在
	_, err = s.skillRepo.GetByID(ctx, skillID)
	if err != nil {
		return fmt.Errorf("skill not found: %w", err)
	}

	if err := s.skillRepo.AssignToEmployee(ctx, employeeID, skillID, level); err != nil {
		logger.Errorf("Failed to assign skill to employee: %v", err)
		return fmt.Errorf("failed to assign skill: %w", err)
	}

	logger.Infof("Skill %d assigned to employee %d successfully", skillID, employeeID)
	return nil
}

// RemoveSkillFromEmployee 移除员工技能
func (s *SkillServiceImpl) RemoveSkillFromEmployee(ctx context.Context, employeeID uint, skillID uint) error {
	logger.Infof("Removing skill %d from employee %d", skillID, employeeID)

	// 检查员工是否存在
	_, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return fmt.Errorf("employee not found: %w", err)
	}

	if err := s.skillRepo.RemoveFromEmployee(ctx, employeeID, skillID); err != nil {
		logger.Errorf("Failed to remove skill from employee: %v", err)
		return fmt.Errorf("failed to remove skill: %w", err)
	}

	logger.Infof("Skill %d removed from employee %d successfully", skillID, employeeID)
	return nil
}

// GetEmployeeSkills 获取员工技能列表
func (s *SkillServiceImpl) GetEmployeeSkills(ctx context.Context, employeeID uint) ([]*SkillResponse, error) {
	// 检查员工是否存在
	_, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	skills, err := s.skillRepo.GetEmployeeSkills(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee skills: %w", err)
	}

	// 转换为响应格式
	responses := make([]*SkillResponse, 0, len(skills))
	for _, skill := range skills {
		level, _ := s.skillRepo.GetEmployeeSkillLevel(ctx, employeeID, skill.ID)
		responses = append(responses, &SkillResponse{
			ID:          skill.ID,
			Name:        skill.Name,
			Category:    skill.Category,
			Description: skill.Description,
			Level:       level,
			CreatedAt:   skill.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   skill.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

// GetSkillByName 根据名称获取技能
func (s *SkillServiceImpl) GetSkillByName(ctx context.Context, name string) (*SkillResponse, error) {
	skill, err := s.skillRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("skill not found: %w", err)
	}

	return s.buildSkillResponse(skill), nil
}

// buildSkillResponse 构建技能响应对象
func (s *SkillServiceImpl) buildSkillResponse(skill *database.Skill) *SkillResponse {
	return &SkillResponse{
		ID:          skill.ID,
		Name:        skill.Name,
		Category:    skill.Category,
		Description: skill.Description,
		CreatedAt:   skill.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   skill.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
