package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"taskmanage/internal/database"
	"taskmanage/internal/repository"
)

// projectService 项目服务实现
type projectService struct {
	repoManager repository.RepositoryManager
	logger      *logrus.Logger
}

// NewProjectService 创建项目服务实例
func NewProjectService(repoManager repository.RepositoryManager, logger *logrus.Logger) ProjectService {
	return &projectService{
		repoManager: repoManager,
		logger:      logger,
	}
}

// CreateProject 创建项目
func (s *projectService) CreateProject(ctx context.Context, req *CreateProjectRequest) (*ProjectResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"name":          req.Name,
		"department_id": req.DepartmentID,
		"manager_id":    req.ManagerID,
	}).Info("创建项目")

	project := &database.Project{
		Name:         req.Name,
		Description:  req.Description,
		Status:       req.Status,
		Priority:     req.Priority,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Budget:       req.Budget,
		DepartmentID: req.DepartmentID,
		ManagerID:    req.ManagerID,
	}

	repo := s.repoManager.ProjectRepository()
	if err := repo.Create(ctx, project); err != nil {
		s.logger.WithError(err).Error("创建项目失败")
		return nil, fmt.Errorf("创建项目失败: %w", err)
	}

	return s.projectToResponse(project), nil
}

// UpdateProject 更新项目
func (s *projectService) UpdateProject(ctx context.Context, id uint, req *UpdateProjectRequest) (*ProjectResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"id":   id,
		"name": req.Name,
	}).Info("更新项目")

	repo := s.repoManager.ProjectRepository()
	project, err := repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("获取项目失败")
		return nil, fmt.Errorf("获取项目失败: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.Status != nil {
		project.Status = *req.Status
	}
	if req.Priority != nil {
		project.Priority = *req.Priority
	}
	if req.StartDate != nil {
		project.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		project.EndDate = req.EndDate
	}
	if req.Budget != nil {
		project.Budget = *req.Budget
	}
	if req.DepartmentID != nil {
		project.DepartmentID = *req.DepartmentID
	}
	if req.ManagerID != nil {
		project.ManagerID = *req.ManagerID
	}

	if err := repo.Update(ctx, project); err != nil {
		s.logger.WithError(err).Error("更新项目失败")
		return nil, fmt.Errorf("更新项目失败: %w", err)
	}

	return s.projectToResponse(project), nil
}

// DeleteProject 删除项目
func (s *projectService) DeleteProject(ctx context.Context, id uint) error {
	s.logger.WithField("id", id).Info("删除项目")

	repo := s.repoManager.ProjectRepository()
	if err := repo.Delete(ctx, id); err != nil {
		s.logger.WithError(err).Error("删除项目失败")
		return fmt.Errorf("删除项目失败: %w", err)
	}

	return nil
}

// GetProject 获取项目详情
func (s *projectService) GetProject(ctx context.Context, id uint) (*ProjectResponse, error) {
	s.logger.WithField("id", id).Debug("获取项目详情")

	repo := s.repoManager.ProjectRepository()
	project, err := repo.GetProjectWithMembers(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("获取项目失败")
		return nil, fmt.Errorf("获取项目失败: %w", err)
	}

	return s.projectToResponse(project), nil
}

// ListProjects 获取项目列表
func (s *projectService) ListProjects(ctx context.Context, req *ListRequest) (*ListResponse[*ProjectResponse], error) {
	s.logger.WithFields(logrus.Fields{
		"page":      req.Page,
		"page_size": req.PageSize,
	}).Debug("获取项目列表")

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.Sort == "" {
		req.Sort = "id"
	}
	if req.Order == "" {
		req.Order = "asc"
	}

	filter := repository.ListFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Sort:     req.Sort,
		Order:    req.Order,
	}

	repo := s.repoManager.ProjectRepository()
	projects, total, err := repo.List(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("获取项目列表失败")
		return nil, fmt.Errorf("获取项目列表失败: %w", err)
	}

	responses := make([]*ProjectResponse, len(projects))
	for i, proj := range projects {
		responses[i] = s.projectToResponse(proj)
	}

	return &ListResponse[*ProjectResponse]{
		Items: responses,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

// GetProjectsByDepartment 根据部门获取项目
func (s *projectService) GetProjectsByDepartment(ctx context.Context, departmentID uint) ([]*ProjectResponse, error) {
	s.logger.WithField("department_id", departmentID).Debug("根据部门获取项目")

	repo := s.repoManager.ProjectRepository()
	projects, err := repo.GetByDepartmentID(ctx, departmentID)
	if err != nil {
		s.logger.WithError(err).Error("根据部门获取项目失败")
		return nil, fmt.Errorf("根据部门获取项目失败: %w", err)
	}

	responses := make([]*ProjectResponse, len(projects))
	for i, proj := range projects {
		responses[i] = s.projectToResponse(proj)
	}

	return responses, nil
}

// GetProjectsByManager 根据管理者获取项目
func (s *projectService) GetProjectsByManager(ctx context.Context, managerID uint) ([]*ProjectResponse, error) {
	s.logger.WithField("manager_id", managerID).Debug("根据管理者获取项目")

	repo := s.repoManager.ProjectRepository()
	projects, err := repo.GetByManagerID(ctx, managerID)
	if err != nil {
		s.logger.WithError(err).Error("根据管理者获取项目失败")
		return nil, fmt.Errorf("根据管理者获取项目失败: %w", err)
	}

	responses := make([]*ProjectResponse, len(projects))
	for i, proj := range projects {
		responses[i] = s.projectToResponse(proj)
	}

	return responses, nil
}

// GetProjectsByStatus 根据状态获取项目
func (s *projectService) GetProjectsByStatus(ctx context.Context, status string) ([]*ProjectResponse, error) {
	s.logger.WithField("status", status).Debug("根据状态获取项目")

	repo := s.repoManager.ProjectRepository()
	projects, err := repo.GetByStatus(ctx, status)
	if err != nil {
		s.logger.WithError(err).Error("根据状态获取项目失败")
		return nil, fmt.Errorf("根据状态获取项目失败: %w", err)
	}

	responses := make([]*ProjectResponse, len(projects))
	for i, proj := range projects {
		responses[i] = s.projectToResponse(proj)
	}

	return responses, nil
}

// AddProjectMember 添加项目成员
func (s *projectService) AddProjectMember(ctx context.Context, projectID uint, req *AddProjectMemberRequest) error {
	s.logger.WithFields(logrus.Fields{
		"project_id":  projectID,
		"employee_id": req.EmployeeID,
	}).Info("添加项目成员")

	repo := s.repoManager.ProjectRepository()
	if err := repo.AddMember(ctx, projectID, req.EmployeeID); err != nil {
		s.logger.WithError(err).Error("添加项目成员失败")
		return fmt.Errorf("添加项目成员失败: %w", err)
	}

	return nil
}

// RemoveProjectMember 移除项目成员
func (s *projectService) RemoveProjectMember(ctx context.Context, projectID uint, req *RemoveProjectMemberRequest) error {
	s.logger.WithFields(logrus.Fields{
		"project_id":  projectID,
		"employee_id": req.EmployeeID,
	}).Info("移除项目成员")

	repo := s.repoManager.ProjectRepository()
	if err := repo.RemoveMember(ctx, projectID, req.EmployeeID); err != nil {
		s.logger.WithError(err).Error("移除项目成员失败")
		return fmt.Errorf("移除项目成员失败: %w", err)
	}

	return nil
}

// GetProjectMembers 获取项目成员
func (s *projectService) GetProjectMembers(ctx context.Context, projectID uint) ([]*EmployeeResponse, error) {
	s.logger.WithField("project_id", projectID).Debug("获取项目成员")

	repo := s.repoManager.ProjectRepository()
	members, err := repo.GetProjectMembers(ctx, projectID)
	if err != nil {
		s.logger.WithError(err).Error("获取项目成员失败")
		return nil, fmt.Errorf("获取项目成员失败: %w", err)
	}

	responses := make([]*EmployeeResponse, len(members))
	for i, member := range members {
		responses[i] = &EmployeeResponse{
			ID:           member.ID,
			Name:       member.User.RealName,
			Email:      member.User.Email,
			Department: member.Department.Name,
			Position:   member.Position.Name,
			Status:       member.Status,
			CreatedAt:    member.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    member.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return responses, nil
}

// UpdateProjectManager 更新项目管理者
func (s *projectService) UpdateProjectManager(ctx context.Context, projectID, managerID uint) error {
	s.logger.WithFields(logrus.Fields{
		"project_id": projectID,
		"manager_id": managerID,
	}).Info("更新项目管理者")

	repo := s.repoManager.ProjectRepository()
	if err := repo.UpdateManager(ctx, projectID, managerID); err != nil {
		s.logger.WithError(err).Error("更新项目管理者失败")
		return fmt.Errorf("更新项目管理者失败: %w", err)
	}

	return nil
}

// projectToResponse 转换项目模型为响应DTO
func (s *projectService) projectToResponse(proj *database.Project) *ProjectResponse {
	response := &ProjectResponse{
		ID:           proj.ID,
		Name:         proj.Name,
		Description:  proj.Description,
		Status:       proj.Status,
		Priority:     proj.Priority,
		StartDate:    proj.StartDate,
		EndDate:      proj.EndDate,
		Budget:       proj.Budget,
		DepartmentID: proj.DepartmentID,
		ManagerID:    proj.ManagerID,
		CreatedAt:    proj.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    proj.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 添加部门信息
	if proj.DepartmentID != 0 {
		response.Department = &DepartmentResponse{
			ID:          proj.Department.ID,
			Name:        proj.Department.Name,
			Description: proj.Department.Description,
			Path:        proj.Department.Path,
			CreatedAt:   proj.Department.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   proj.Department.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// 添加管理者信息
	if proj.ManagerID != 0 {
		response.Manager = &EmployeeResponse{
			ID:           proj.Manager.ID,
			Name:       proj.Manager.User.RealName,
			Email:      proj.Manager.User.Email,
			Department: proj.Manager.Department.Name,
			Position:   proj.Manager.Position.Name,
			Status:       proj.Manager.Status,
			CreatedAt:    proj.Manager.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    proj.Manager.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// 添加成员信息
	if len(proj.Members) > 0 {
		response.Members = make([]*EmployeeResponse, len(proj.Members))
		for i, member := range proj.Members {
			response.Members[i] = &EmployeeResponse{
				ID:           member.ID,
				Name:       member.User.RealName,
				Email:      member.User.Email,
				Department: member.Department.Name,
				Position:   member.Position.Name,
				Status:       member.Status,
				CreatedAt:    member.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt:    member.UpdatedAt.Format("2006-01-02 15:04:05"),
			}
		}
	}

	return response
}
