package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"taskmanage/internal/database"
	"taskmanage/internal/repository"
)

// positionService 职位服务实现
type positionService struct {
	repoManager repository.RepositoryManager
	logger      *logrus.Logger
}

// NewPositionService 创建职位服务实例
func NewPositionService(repoManager repository.RepositoryManager, logger *logrus.Logger) PositionService {
	return &positionService{
		repoManager: repoManager,
		logger:      logger,
	}
}

// CreatePosition 创建职位
func (s *positionService) CreatePosition(ctx context.Context, req *CreatePositionRequest) (*PositionResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"name":     req.Name,
		"category": req.Category,
		"level":    req.Level,
	}).Info("创建职位")

	position := &database.Position{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Level:       req.Level,
	}

	repo := s.repoManager.PositionRepository()
	if err := repo.Create(ctx, position); err != nil {
		s.logger.WithError(err).Error("创建职位失败")
		return nil, fmt.Errorf("创建职位失败: %w", err)
	}

	return s.positionToResponse(position), nil
}

// UpdatePosition 更新职位
func (s *positionService) UpdatePosition(ctx context.Context, id uint, req *UpdatePositionRequest) (*PositionResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"id":   id,
		"name": req.Name,
	}).Info("更新职位")

	repo := s.repoManager.PositionRepository()
	position, err := repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("获取职位失败")
		return nil, fmt.Errorf("获取职位失败: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		position.Name = *req.Name
	}
	if req.Description != nil {
		position.Description = *req.Description
	}
	if req.Category != nil {
		position.Category = *req.Category
	}
	if req.Level != nil {
		position.Level = *req.Level
	}

	if err := repo.Update(ctx, position); err != nil {
		s.logger.WithError(err).Error("更新职位失败")
		return nil, fmt.Errorf("更新职位失败: %w", err)
	}

	return s.positionToResponse(position), nil
}

// DeletePosition 删除职位
func (s *positionService) DeletePosition(ctx context.Context, id uint) error {
	s.logger.WithField("id", id).Info("删除职位")

	repo := s.repoManager.PositionRepository()
	if err := repo.Delete(ctx, id); err != nil {
		s.logger.WithError(err).Error("删除职位失败")
		return fmt.Errorf("删除职位失败: %w", err)
	}

	return nil
}

// GetPosition 获取职位详情
func (s *positionService) GetPosition(ctx context.Context, id uint) (*PositionResponse, error) {
	s.logger.WithField("id", id).Debug("获取职位详情")

	repo := s.repoManager.PositionRepository()
	position, err := repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("获取职位失败")
		return nil, fmt.Errorf("获取职位失败: %w", err)
	}

	return s.positionToResponse(position), nil
}

// ListPositions 获取职位列表
func (s *positionService) ListPositions(ctx context.Context, req *ListRequest) (*ListResponse[*PositionResponse], error) {
	s.logger.WithFields(logrus.Fields{
		"page":      req.Page,
		"page_size": req.PageSize,
	}).Debug("获取职位列表")

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

	repo := s.repoManager.PositionRepository()
	positions, total, err := repo.List(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("获取职位列表失败")
		return nil, fmt.Errorf("获取职位列表失败: %w", err)
	}

	responses := make([]*PositionResponse, len(positions))
	for i, pos := range positions {
		responses[i] = s.positionToResponse(pos)
	}

	return &ListResponse[*PositionResponse]{
		Items: responses,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

// GetPositionsByCategory 根据类别获取职位
func (s *positionService) GetPositionsByCategory(ctx context.Context, category string) ([]*PositionResponse, error) {
	s.logger.WithField("category", category).Debug("根据类别获取职位")

	repo := s.repoManager.PositionRepository()
	positions, err := repo.GetByCategory(ctx, category)
	if err != nil {
		s.logger.WithError(err).Error("根据类别获取职位失败")
		return nil, fmt.Errorf("根据类别获取职位失败: %w", err)
	}

	responses := make([]*PositionResponse, len(positions))
	for i, pos := range positions {
		responses[i] = s.positionToResponse(pos)
	}

	return responses, nil
}

// GetPositionsByLevel 根据级别获取职位
func (s *positionService) GetPositionsByLevel(ctx context.Context, level int) ([]*PositionResponse, error) {
	s.logger.WithField("level", level).Debug("根据级别获取职位")

	repo := s.repoManager.PositionRepository()
	positions, err := repo.GetByLevel(ctx, level)
	if err != nil {
		s.logger.WithError(err).Error("根据级别获取职位失败")
		return nil, fmt.Errorf("根据级别获取职位失败: %w", err)
	}

	responses := make([]*PositionResponse, len(positions))
	for i, pos := range positions {
		responses[i] = s.positionToResponse(pos)
	}

	return responses, nil
}

// GetAllCategories 获取所有职位类别
func (s *positionService) GetAllCategories(ctx context.Context) ([]string, error) {
	s.logger.Debug("获取所有职位类别")

	repo := s.repoManager.PositionRepository()
	categories, err := repo.GetAllCategories(ctx)
	if err != nil {
		s.logger.WithError(err).Error("获取职位类别失败")
		return nil, fmt.Errorf("获取职位类别失败: %w", err)
	}

	return categories, nil
}

// positionToResponse 转换职位模型为响应DTO
func (s *positionService) positionToResponse(pos *database.Position) *PositionResponse {
	return &PositionResponse{
		ID:          pos.ID,
		Name:        pos.Name,
		Description: pos.Description,
		Category:    pos.Category,
		Level:       pos.Level,
		CreatedAt:   pos.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   pos.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
