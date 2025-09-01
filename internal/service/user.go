package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"taskmanage/internal/config"
	"taskmanage/internal/database"
	"taskmanage/internal/repository"
	"taskmanage/pkg/logger"
)

// userService 用户服务实现
type userService struct {
	userRepo     repository.UserRepository
	employeeRepo repository.EmployeeRepository
	repoManager  repository.RepositoryManager
	config       *config.Config
	logger       *logrus.Logger
}

// NewUserService 创建用户服务实例
func NewUserService(repoManager repository.RepositoryManager, cfg *config.Config) UserService {
	return &userService{
		userRepo:     repoManager.UserRepository(),
		employeeRepo: repoManager.EmployeeRepository(),
		repoManager:  repoManager,
		config:       cfg,
		logger:       logger.GetLogger(),
	}
}

// CreateUser 创建用户
func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	// 验证用户名是否已存在
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		logger.Errorf("检查用户名失败: %v", err)
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	// 验证邮箱是否已存在
	existingUser, err = s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		logger.Errorf("检查邮箱失败: %v", err)
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("邮箱已被注册")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("密码加密失败: %v", err)
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户对象
	user := &database.User{
		Username:     req.Username,
		Email:        req.Email,
		Password:     string(hashedPassword), // 兼容数据库中的password字段
		PasswordHash: string(hashedPassword),
		RealName:     req.RealName,
		Status:       "active",
		Role:         "employee", // 默认角色
	}

	// 保存用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Errorf("创建用户失败: %v", err)
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 自动创建对应的员工记录
	defaultDeptID := uint(1)
	defaultPosID := uint(1)
	employee := &database.Employee{
		UserID:           user.ID,
		EmployeeNo:       fmt.Sprintf("EMP%06d", user.ID), // 生成员工编号
		DepartmentID:     &defaultDeptID, // TODO: 需要创建默认部门或从请求中获取
		PositionID:       &defaultPosID,  // TODO: 需要创建默认职位或从请求中获取
		OnboardingStatus: "pending_onboard", // 默认待入职状态
		Status:           "available",
		MaxTasks:         5,
		CurrentTasks:     0,
	}

	if err := s.employeeRepo.Create(ctx, employee); err != nil {
		logger.Warnf("创建员工记录失败: %v", err)
		// 员工记录创建失败不影响用户创建，只记录警告
	} else {
		logger.Infof("员工记录创建成功: EmployeeID=%d, EmployeeNo=%s", employee.ID, employee.EmployeeNo)
	}

	logger.Infof("用户创建成功: ID=%d, Username=%s", user.ID, user.Username)

	return UserToResponse(user), nil
}

// GetUser 获取用户信息
func (s *userService) GetUser(ctx context.Context, userID uint) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return UserToResponse(user), nil
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(ctx context.Context, userID uint, req *UpdateUserRequest) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 更新字段
	if req.Email != nil {
		// 检查邮箱是否被其他用户使用
		existingUser, err := s.userRepo.GetByEmail(ctx, *req.Email)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("检查邮箱失败: %w", err)
		}
		if existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("邮箱已被其他用户使用")
		}
		user.Email = *req.Email
	}
	if req.RealName != nil {
		user.RealName = *req.RealName
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Errorf("更新用户失败: %v", err)
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	logger.Infof("用户更新成功: ID=%d, Username=%s", user.ID, user.Username)

	return UserToResponse(user), nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, userID uint) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errors.New("用户不存在")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		logger.Errorf("删除用户失败: %v", err)
		return fmt.Errorf("删除用户失败: %w", err)
	}

	logger.Infof("用户删除成功: ID=%d, Username=%s", user.ID, user.Username)
	return nil
}

// ListUsers 获取用户列表
func (s *userService) ListUsers(ctx context.Context, filter repository.ListFilter) ([]*UserResponse, int64, error) {
	users, total, err := s.userRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = UserToResponse(user)
	}

	return responses, total, nil
}

// AuthenticateUser 用户认证
func (s *userService) AuthenticateUser(ctx context.Context, username, password string) (*UserResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 验证密码 - 优先使用PasswordHash，如果为空则使用Password字段
	passwordHash := user.PasswordHash
	if passwordHash == "" {
		passwordHash = user.Password
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("用户账号已被禁用")
	}

	return UserToResponse(user), nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 认证用户
	user, err := s.AuthenticateUser(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	// TODO: 生成JWT令牌
	// 这里暂时返回空令牌，等JWT管理器实现后再完善
	return &LoginResponse{
		AccessToken:  "",
		RefreshToken: "",
		ExpiresIn:    3600,
		User:         *user,
	}, nil
}

// RefreshToken 刷新令牌
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// TODO: 实现刷新令牌逻辑
	return nil, errors.New("刷新令牌功能尚未实现")
}

// Logout 用户登出
func (s *userService) Logout(ctx context.Context, userID uint) error {
	// TODO: 实现登出逻辑（如令牌黑名单）
	logger.Infof("用户登出: ID=%d", userID)
	return nil
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(ctx context.Context, userID uint, req *ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errors.New("用户不存在")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 验证旧密码 - 优先使用PasswordHash，如果为空则使用Password字段
	passwordHash := user.PasswordHash
	if passwordHash == "" {
		passwordHash = user.Password
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.OldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("新密码加密失败: %v", err)
		return fmt.Errorf("新密码加密失败: %w", err)
	}

	// 更新密码 - 同时更新两个字段以保持兼容性
	user.Password = string(hashedPassword)
	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Errorf("更新密码失败: %v", err)
		return fmt.Errorf("更新密码失败: %w", err)
	}

	logger.Infof("用户密码修改成功: ID=%d", userID)
	return nil
}

// HasPermission 检查用户权限
func (s *userService) HasPermission(ctx context.Context, userID uint, resource, action string) (bool, error) {
	permRepo := s.repoManager.PermissionRepository()

	// 获取用户的所有权限
	permissions, err := permRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("获取用户权限失败")
		return false, err
	}

	// 检查是否有匹配的权限
	for _, perm := range permissions {
		if perm.Resource == resource && perm.Action == action {
			s.logger.WithFields(map[string]interface{}{
				"user_id":    userID,
				"resource":   resource,
				"action":     action,
				"permission": perm.Name,
			}).Debug("用户权限检查通过")
			return true, nil
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":  userID,
		"resource": resource,
		"action":   action,
	}).Debug("用户权限检查失败")

	return false, nil
}

// AssignRoles 分配角色
func (s *userService) AssignRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	// TODO: 实现角色分配逻辑
	return nil
}

// RemoveRoles 移除角色
func (s *userService) RemoveRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	// TODO: 实现角色移除逻辑
	return nil
}

// GetUserPermissions 获取用户权限
func (s *userService) GetUserPermissions(ctx context.Context, userID uint) ([]*PermissionResponse, error) {
	// TODO: 实现获取用户权限逻辑
	return nil, nil
}
