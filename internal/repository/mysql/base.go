package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// BaseRepositoryImpl 基础仓储实现
type BaseRepositoryImpl[T any] struct {
	db    *gorm.DB
	model T
}

// NewBaseRepository 创建基础仓储
func NewBaseRepository[T any](db *gorm.DB) *BaseRepositoryImpl[T] {
	return &BaseRepositoryImpl[T]{
		db: db,
	}
}

// Create 创建实体
func (r *BaseRepositoryImpl[T]) Create(ctx context.Context, entity *T) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		logger.Errorf("创建实体失败: %v", err)
		return fmt.Errorf("创建实体失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取实体
func (r *BaseRepositoryImpl[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var entity T
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		logger.Errorf("根据ID获取实体失败: %v", err)
		return nil, fmt.Errorf("根据ID获取实体失败: %w", err)
	}

	return &entity, nil
}

// Update 更新实体
func (r *BaseRepositoryImpl[T]) Update(ctx context.Context, entity *T) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		logger.Errorf("更新实体失败: %v", err)
		return fmt.Errorf("更新实体失败: %w", err)
	}

	return nil
}

// Delete 删除实体
func (r *BaseRepositoryImpl[T]) Delete(ctx context.Context, id uint) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var entity T
	if err := r.db.WithContext(ctx).Delete(&entity, id).Error; err != nil {
		logger.Errorf("删除实体失败: %v", err)
		return fmt.Errorf("删除实体失败: %w", err)
	}

	return nil
}

// List 获取实体列表
func (r *BaseRepositoryImpl[T]) List(ctx context.Context, filter repository.ListFilter) ([]*T, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var entities []*T
	var total int64

	// 构建查询
	query := r.db.WithContext(ctx).Model(r.model)

	// 应用过滤器
	query = r.applyFilters(query, filter.Filters)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("获取实体总数失败: %v", err)
		return nil, 0, fmt.Errorf("获取实体总数失败: %w", err)
	}

	// 应用分页和排序
	query = r.applyPagination(query, filter)
	query = r.applySorting(query, filter)

	// 获取数据
	if err := query.Find(&entities).Error; err != nil {
		logger.Errorf("获取实体列表失败: %v", err)
		return nil, 0, fmt.Errorf("获取实体列表失败: %w", err)
	}

	return entities, total, nil
}

// Exists 检查实体是否存在
func (r *BaseRepositoryImpl[T]) Exists(ctx context.Context, id uint) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var count int64
	if err := r.db.WithContext(ctx).Model(r.model).Where("id = ?", id).Count(&count).Error; err != nil {
		logger.Errorf("检查实体存在性失败: %v", err)
		return false, fmt.Errorf("检查实体存在性失败: %w", err)
	}

	return count > 0, nil
}

// applyFilters 应用过滤器
func (r *BaseRepositoryImpl[T]) applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	for key, value := range filters {
		switch key {
		case "status":
			query = query.Where("status = ?", value)
		case "created_after":
			if t, ok := value.(time.Time); ok {
				query = query.Where("created_at >= ?", t)
			}
		case "created_before":
			if t, ok := value.(time.Time); ok {
				query = query.Where("created_at <= ?", t)
			}
		case "search":
			// 通用搜索，可以根据具体模型重写
			if str, ok := value.(string); ok && str != "" {
				query = query.Where("name LIKE ? OR title LIKE ? OR description LIKE ?", 
					"%"+str+"%", "%"+str+"%", "%"+str+"%")
			}
		default:
			// 其他字段的精确匹配
			query = query.Where(fmt.Sprintf("%s = ?", key), value)
		}
	}
	return query
}

// applyPagination 应用分页
func (r *BaseRepositoryImpl[T]) applyPagination(query *gorm.DB, filter repository.ListFilter) *gorm.DB {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100 // 限制最大页面大小
	}

	offset := (filter.Page - 1) * filter.PageSize
	return query.Offset(offset).Limit(filter.PageSize)
}

// applySorting 应用排序
func (r *BaseRepositoryImpl[T]) applySorting(query *gorm.DB, filter repository.ListFilter) *gorm.DB {
	if filter.Sort == "" {
		filter.Sort = "created_at"
	}
	if filter.Order == "" {
		filter.Order = "desc"
	}

	// 验证排序方向
	if filter.Order != "asc" && filter.Order != "desc" {
		filter.Order = "desc"
	}

	// 验证排序字段（基础字段）
	allowedSortFields := []string{"id", "created_at", "updated_at", "name", "title", "status", "priority"}
	isAllowed := false
	for _, field := range allowedSortFields {
		if filter.Sort == field {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		filter.Sort = "created_at"
	}

	return query.Order(fmt.Sprintf("%s %s", filter.Sort, filter.Order))
}

// WithRetry 带重试的操作执行
func (r *BaseRepositoryImpl[T]) WithRetry(ctx context.Context, operation func() error, maxRetries int) error {
	var lastErr error
	
	for i := 0; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := operation(); err != nil {
			lastErr = err
			if i < maxRetries {
				// 指数退避
				backoff := time.Duration(i+1) * 100 * time.Millisecond
				logger.Warnf("操作失败，%v后重试 (第%d次): %v", backoff, i+1, err)
				
				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("操作在%d次重试后仍然失败: %w", maxRetries, lastErr)
}
