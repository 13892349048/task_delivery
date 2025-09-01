package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"taskmanage/internal/database"
	"taskmanage/internal/repository"
)

// departmentService 部门服务实现
type departmentService struct {
	repoManager repository.RepositoryManager
	logger      *logrus.Logger
}

// NewDepartmentService 创建部门服务实例
func NewDepartmentService(repoManager repository.RepositoryManager, logger *logrus.Logger) DepartmentService {
	return &departmentService{
		repoManager: repoManager,
		logger:      logger,
	}
}

// CreateDepartment 创建部门
func (s *departmentService) CreateDepartment(ctx context.Context, req *CreateDepartmentRequest) (*DepartmentResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"name":      req.Name,
		"parent_id": req.ParentID,
	}).Info("创建部门")

	// 构建部门路径
	path := req.Path
	if path == "" {
		if req.ParentID != nil {
			parent, err := s.repoManager.DepartmentRepository().GetByID(ctx, *req.ParentID)
			if err != nil {
				s.logger.WithError(err).Error("获取父部门失败")
				return nil, fmt.Errorf("获取父部门失败: %w", err)
			}
			path = parent.Path + "/" + req.Name
		} else {
			path = "/" + req.Name
		}
	}

	department := &database.Department{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		ManagerID:   req.ManagerID,
		Path:        path,
	}

	repo := s.repoManager.DepartmentRepository()
	if err := repo.Create(ctx, department); err != nil {
		s.logger.WithError(err).Error("创建部门失败")
		return nil, fmt.Errorf("创建部门失败: %w", err)
	}

	return s.departmentToResponse(department), nil
}

// UpdateDepartment 更新部门
func (s *departmentService) UpdateDepartment(ctx context.Context, id uint, req *UpdateDepartmentRequest) (*DepartmentResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"id":   id,
		"name": req.Name,
	}).Info("更新部门")

	repo := s.repoManager.DepartmentRepository()
	department, err := repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("获取部门失败")
		return nil, fmt.Errorf("获取部门失败: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		department.Name = *req.Name
	}
	if req.Description != nil {
		department.Description = *req.Description
	}
	if req.ParentID != nil {
		department.ParentID = req.ParentID
	}
	if req.ManagerID != nil {
		department.ManagerID = req.ManagerID
	}
	if req.Path != nil {
		department.Path = *req.Path
	}

	if err := repo.Update(ctx, department); err != nil {
		s.logger.WithError(err).Error("更新部门失败")
		return nil, fmt.Errorf("更新部门失败: %w", err)
	}

	return s.departmentToResponse(department), nil
}

// DeleteDepartment 删除部门
func (s *departmentService) DeleteDepartment(ctx context.Context, id uint) error {
	s.logger.WithField("id", id).Info("删除部门")

	repo := s.repoManager.DepartmentRepository()
	
	// 检查是否有子部门
	subDepartments, err := repo.GetByParentID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("检查子部门失败")
		return fmt.Errorf("检查子部门失败: %w", err)
	}
	
	if len(subDepartments) > 0 {
		return fmt.Errorf("不能删除有子部门的部门")
	}

	if err := repo.Delete(ctx, id); err != nil {
		s.logger.WithError(err).Error("删除部门失败")
		return fmt.Errorf("删除部门失败: %w", err)
	}

	return nil
}

// GetDepartment 获取部门详情
func (s *departmentService) GetDepartment(ctx context.Context, id uint) (*DepartmentResponse, error) {
	s.logger.WithField("id", id).Debug("获取部门详情")

	repo := s.repoManager.DepartmentRepository()
	department, err := repo.GetDepartmentWithManager(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("获取部门失败")
		return nil, fmt.Errorf("获取部门失败: %w", err)
	}

	return s.departmentToResponse(department), nil
}

// ListDepartments 获取部门列表
func (s *departmentService) ListDepartments(ctx context.Context, req *ListRequest) (*ListResponse[*DepartmentResponse], error) {
	s.logger.WithFields(logrus.Fields{
		"page":      req.Page,
		"page_size": req.PageSize,
	}).Debug("获取部门列表")

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

	repo := s.repoManager.DepartmentRepository()
	departments, total, err := repo.List(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("获取部门列表失败")
		return nil, fmt.Errorf("获取部门列表失败: %w", err)
	}

	responses := make([]*DepartmentResponse, len(departments))
	for i, dept := range departments {
		responses[i] = s.departmentToResponse(dept)
	}

	return &ListResponse[*DepartmentResponse]{
		Items: responses,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

// GetDepartmentTree 获取部门树结构
func (s *departmentService) GetDepartmentTree(ctx context.Context) ([]*DepartmentResponse, error) {
	s.logger.Debug("获取部门树结构")

	repo := s.repoManager.DepartmentRepository()
	departments, err := repo.GetDepartmentTree(ctx)
	if err != nil {
		s.logger.WithError(err).Error("获取部门树失败")
		return nil, fmt.Errorf("获取部门树失败: %w", err)
	}

	// 构建树结构
	departmentMap := make(map[uint]*DepartmentResponse)
	var roots []*DepartmentResponse

	// 第一遍遍历，创建所有节点
	for _, dept := range departments {
		response := s.departmentToResponse(dept)
		departmentMap[dept.ID] = response
	}

	// 第二遍遍历，建立父子关系
	for _, dept := range departments {
		response := departmentMap[dept.ID]
		if dept.ParentID != nil {
			if parent, exists := departmentMap[*dept.ParentID]; exists {
				if parent.Children == nil {
					parent.Children = make([]*DepartmentResponse, 0)
				}
				parent.Children = append(parent.Children, response)
			}
		} else {
			roots = append(roots, response)
		}
	}

	return roots, nil
}

// GetRootDepartments 获取根部门
func (s *departmentService) GetRootDepartments(ctx context.Context) ([]*DepartmentResponse, error) {
	s.logger.Debug("获取根部门")

	repo := s.repoManager.DepartmentRepository()
	departments, err := repo.GetRootDepartments(ctx)
	if err != nil {
		s.logger.WithError(err).Error("获取根部门失败")
		return nil, fmt.Errorf("获取根部门失败: %w", err)
	}

	responses := make([]*DepartmentResponse, len(departments))
	for i, dept := range departments {
		responses[i] = s.departmentToResponse(dept)
	}

	return responses, nil
}

// GetSubDepartments 获取子部门
func (s *departmentService) GetSubDepartments(ctx context.Context, departmentID uint) ([]*DepartmentResponse, error) {
	s.logger.WithField("department_id", departmentID).Debug("获取子部门")

	repo := s.repoManager.DepartmentRepository()
	departments, err := repo.GetSubDepartments(ctx, departmentID)
	if err != nil {
		s.logger.WithError(err).Error("获取子部门失败")
		return nil, fmt.Errorf("获取子部门失败: %w", err)
	}

	responses := make([]*DepartmentResponse, len(departments))
	for i, dept := range departments {
		responses[i] = s.departmentToResponse(dept)
	}

	return responses, nil
}

// UpdateManager 更新部门管理者
func (s *departmentService) UpdateManager(ctx context.Context, departmentID, managerID uint) error {
	s.logger.WithFields(logrus.Fields{
		"department_id": departmentID,
		"manager_id":    managerID,
	}).Info("更新部门管理者")

	repo := s.repoManager.DepartmentRepository()
	if err := repo.UpdateManager(ctx, departmentID, managerID); err != nil {
		s.logger.WithError(err).Error("更新部门管理者失败")
		return fmt.Errorf("更新部门管理者失败: %w", err)
	}

	return nil
}

// departmentToResponse 转换部门模型为响应DTO
func (s *departmentService) departmentToResponse(dept *database.Department) *DepartmentResponse {
	response := &DepartmentResponse{
		ID:          dept.ID,
		Name:        dept.Name,
		Description: dept.Description,
		ParentID:    dept.ParentID,
		ManagerID:   dept.ManagerID,
		Path:        dept.Path,
		CreatedAt:   dept.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   dept.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 添加管理者信息
	if dept.Manager != nil {
		response.Manager = &EmployeeResponse{
			ID:     dept.Manager.ID,
			Name:   dept.Manager.User.RealName,
			Email:  dept.Manager.User.Email,
			Status: dept.Manager.Status,
		}
	}

	// 添加父部门信息
	if dept.Parent != nil {
		response.Parent = &DepartmentResponse{
			ID:          dept.Parent.ID,
			Name:        dept.Parent.Name,
			Description: dept.Parent.Description,
			Path:        dept.Parent.Path,
			CreatedAt:   dept.Parent.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   dept.Parent.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return response
}
