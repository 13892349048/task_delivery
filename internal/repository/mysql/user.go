package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// UserRepositoryImpl 用户仓储实现
type UserRepositoryImpl struct {
	*BaseRepositoryImpl[database.User]
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &UserRepositoryImpl{
		BaseRepositoryImpl: NewBaseRepository[database.User](db),
	}
}

// GetByUsername 根据用户名获取用户
func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*database.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var user database.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		logger.Errorf("根据用户名获取用户失败: %v", err)
		return nil, fmt.Errorf("根据用户名获取用户失败: %w", err)
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*database.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var user database.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		logger.Errorf("根据邮箱获取用户失败: %v", err)
		return nil, fmt.Errorf("根据邮箱获取用户失败: %w", err)
	}

	return &user, nil
}

// UpdateLastLogin 更新最后登录信息
func (r *UserRepositoryImpl) UpdateLastLogin(ctx context.Context, userID uint, ip string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&database.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"last_login_ip": ip,
		}).Error; err != nil {
		logger.Errorf("更新用户最后登录信息失败: %v", err)
		return fmt.Errorf("更新用户最后登录信息失败: %w", err)
	}

	return nil
}

// GetUserWithRoles 获取用户及其角色信息
func (r *UserRepositoryImpl) GetUserWithRoles(ctx context.Context, userID uint) (*database.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user database.User
	if err := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Permissions").
		First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		logger.Errorf("获取用户角色信息失败: %v", err)
		return nil, fmt.Errorf("获取用户角色信息失败: %w", err)
	}

	return &user, nil
}

// BatchUpdateStatus 批量更新用户状态
func (r *UserRepositoryImpl) BatchUpdateStatus(ctx context.Context, userIDs []uint, status string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if len(userIDs) == 0 {
		return repository.ErrInvalidInput
	}

	if err := r.db.WithContext(ctx).Model(&database.User{}).
		Where("id IN ?", userIDs).
		Update("status", status).Error; err != nil {
		logger.Errorf("批量更新用户状态失败: %v", err)
		return fmt.Errorf("批量更新用户状态失败: %w", err)
	}

	return nil
}

// ListUsers 分页查询用户列表
func (r *UserRepositoryImpl) ListUsers(ctx context.Context, page, limit int, conditions map[string]interface{}, keyword string) ([]*database.User, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var users []*database.User
	var total int64

	query := r.db.WithContext(ctx).Model(&database.User{})

	// 应用条件过滤
	for key, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}

	// 应用关键字搜索
	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR real_name LIKE ?", 
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("查询用户总数失败: %v", err)
		return nil, 0, fmt.Errorf("查询用户总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Preload("Roles").Preload("Employee").
		Offset(offset).Limit(limit).
		Find(&users).Error; err != nil {
		logger.Errorf("查询用户列表失败: %v", err)
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}

	return users, total, nil
}

// AssignRoles 为用户分配角色
func (r *UserRepositoryImpl) AssignRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if len(roleIDs) == 0 {
		return repository.ErrInvalidInput
	}

	// 获取用户
	var user database.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.ErrNotFound
		}
		logger.Errorf("获取用户失败: %v", err)
		return fmt.Errorf("获取用户失败: %w", err)
	}

	// 获取角色
	var roles []database.Role
	if err := r.db.WithContext(ctx).Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		logger.Errorf("获取角色失败: %v", err)
		return fmt.Errorf("获取角色失败: %w", err)
	}

	if len(roles) != len(roleIDs) {
		return fmt.Errorf("部分角色不存在")
	}

	// 分配角色
	if err := r.db.WithContext(ctx).Model(&user).Association("Roles").Append(roles); err != nil {
		logger.Errorf("分配角色失败: %v", err)
		return fmt.Errorf("分配角色失败: %w", err)
	}

	return nil
}

// RemoveRoles 移除用户角色
func (r *UserRepositoryImpl) RemoveRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if len(roleIDs) == 0 {
		return repository.ErrInvalidInput
	}

	// 获取用户
	var user database.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.ErrNotFound
		}
		logger.Errorf("获取用户失败: %v", err)
		return fmt.Errorf("获取用户失败: %w", err)
	}

	// 获取要移除的角色
	var roles []database.Role
	if err := r.db.WithContext(ctx).Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		logger.Errorf("获取角色失败: %v", err)
		return fmt.Errorf("获取角色失败: %w", err)
	}

	// 移除角色
	if err := r.db.WithContext(ctx).Model(&user).Association("Roles").Delete(roles); err != nil {
		logger.Errorf("移除角色失败: %v", err)
		return fmt.Errorf("移除角色失败: %w", err)
	}

	return nil
}

// GetUsersByRole 根据角色获取用户列表
func (r *UserRepositoryImpl) GetUsersByRole(ctx context.Context, role string) ([]*database.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var users []*database.User
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("roles.name = ?", role).
		Find(&users).Error

	if err != nil {
		logger.Errorf("根据角色查找用户失败: role=%s, error=%v", role, err)
		return nil, fmt.Errorf("根据角色查找用户失败: %w", err)
	}

	return users, nil
}
