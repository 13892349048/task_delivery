package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"taskmanage/internal/database"
	"taskmanage/internal/repository"
)

// PositionRepositoryImpl 职位仓储MySQL实现
type PositionRepositoryImpl struct {
	db *gorm.DB
}

// NewPositionRepository 创建职位仓储实例
func NewPositionRepository(db *gorm.DB) repository.PositionRepository {
	return &PositionRepositoryImpl{db: db}
}

// Create 创建职位
func (r *PositionRepositoryImpl) Create(ctx context.Context, position *database.Position) error {
	return r.db.WithContext(ctx).Create(position).Error
}

// GetByID 根据ID获取职位
func (r *PositionRepositoryImpl) GetByID(ctx context.Context, id uint) (*database.Position, error) {
	var position database.Position
	err := r.db.WithContext(ctx).First(&position, id).Error
	if err != nil {
		return nil, err
	}
	return &position, nil
}

// Update 更新职位
func (r *PositionRepositoryImpl) Update(ctx context.Context, position *database.Position) error {
	return r.db.WithContext(ctx).Save(position).Error
}

// Delete 删除职位
func (r *PositionRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&database.Position{}, id).Error
}

// List 获取职位列表
func (r *PositionRepositoryImpl) List(ctx context.Context, filter repository.ListFilter) ([]*database.Position, int64, error) {
	var positions []*database.Position
	var total int64

	query := r.db.WithContext(ctx).Model(&database.Position{})

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	err := query.
		Offset(offset).
		Limit(filter.PageSize).
		Order(fmt.Sprintf("%s %s", filter.Sort, filter.Order)).
		Find(&positions).Error

	return positions, total, err
}

// Exists 检查职位是否存在
func (r *PositionRepositoryImpl) Exists(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&database.Position{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// GetByName 根据名称获取职位
func (r *PositionRepositoryImpl) GetByName(ctx context.Context, name string) (*database.Position, error) {
	var position database.Position
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&position).Error
	if err != nil {
		return nil, err
	}
	return &position, nil
}

// GetByCategory 根据类别获取职位
func (r *PositionRepositoryImpl) GetByCategory(ctx context.Context, category string) ([]*database.Position, error) {
	var positions []*database.Position
	err := r.db.WithContext(ctx).Where("category = ?", category).Find(&positions).Error
	return positions, err
}

// GetByLevel 根据级别获取职位
func (r *PositionRepositoryImpl) GetByLevel(ctx context.Context, level int) ([]*database.Position, error) {
	var positions []*database.Position
	err := r.db.WithContext(ctx).Where("level = ?", level).Find(&positions).Error
	return positions, err
}

// GetAllCategories 获取所有职位类别
func (r *PositionRepositoryImpl) GetAllCategories(ctx context.Context) ([]string, error) {
	var categories []string
	err := r.db.WithContext(ctx).
		Model(&database.Position{}).
		Distinct("category").
		Pluck("category", &categories).Error
	return categories, err
}
