package handlers

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"taskmanage/internal/service"
	"taskmanage/pkg/response"
)

// PermissionAssignmentHandler 权限分配处理器
type PermissionAssignmentHandler struct {
	permissionAssignmentService service.PermissionAssignmentService
	logger                      *logrus.Logger
}

// NewPermissionAssignmentHandler 创建权限分配处理器
func NewPermissionAssignmentHandler(permissionAssignmentService service.PermissionAssignmentService, logger *logrus.Logger) *PermissionAssignmentHandler {
	return &PermissionAssignmentHandler{
		permissionAssignmentService: permissionAssignmentService,
		logger:                      logger,
	}
}

// CreatePermissionTemplate 创建权限模板
func (h *PermissionAssignmentHandler) CreatePermissionTemplate(c *gin.Context) {
	var req service.CreatePermissionTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("创建权限模板参数绑定失败: %v", err)
		response.BadRequest(c, "参数错误")
		return
	}

	template, err := h.permissionAssignmentService.CreatePermissionTemplate(c.Request.Context(), &req)
	if err != nil {
		h.logger.Errorf("创建权限模板失败: %v", err)
		response.InternalError(c, "创建权限模板失败")
		return
	}

	h.logger.Infof("成功创建权限模板: %s", template.Name)
	response.Success(c, template)
}

// GetPermissionTemplate 获取权限模板
func (h *PermissionAssignmentHandler) GetPermissionTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	template, err := h.permissionAssignmentService.GetPermissionTemplate(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Errorf("获取权限模板失败: %v", err)
		response.NotFound(c, "权限模板不存在")
		return
	}

	response.Success(c, template)
}

// ListPermissionTemplates 获取权限模板列表
func (h *PermissionAssignmentHandler) ListPermissionTemplates(c *gin.Context) {
	var req service.ListPermissionTemplatesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warnf("权限模板列表参数绑定失败: %v", err)
		response.BadRequest(c, "参数错误")
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	templates, err := h.permissionAssignmentService.ListPermissionTemplates(c.Request.Context(), &req)
	if err != nil {
		h.logger.Errorf("获取权限模板列表失败: %v", err)
		response.InternalError(c, "获取权限模板列表失败")
		return
	}

	response.Success(c, templates)
}

// UpdatePermissionTemplate 更新权限模板
func (h *PermissionAssignmentHandler) UpdatePermissionTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	var req service.UpdatePermissionTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("更新权限模板参数绑定失败: %v", err)
		response.BadRequest(c, "参数错误")
		return
	}

	template, err := h.permissionAssignmentService.UpdatePermissionTemplate(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Errorf("更新权限模板失败: %v", err)
		response.InternalError(c, "更新权限模板失败")
		return
	}

	h.logger.Infof("成功更新权限模板: %d", id)
	response.Success(c, template)
}

// DeletePermissionTemplate 删除权限模板
func (h *PermissionAssignmentHandler) DeletePermissionTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}

	if err := h.permissionAssignmentService.DeletePermissionTemplate(c.Request.Context(), uint(id)); err != nil {
		h.logger.Errorf("删除权限模板失败: %v", err)
		response.InternalError(c, "删除权限模板失败")
		return
	}

	h.logger.Infof("成功删除权限模板: %d", id)
	response.Success(c, gin.H{"message": "权限模板删除成功"})
}

// AssignPermissions 分配权限
func (h *PermissionAssignmentHandler) AssignPermissions(c *gin.Context) {
	var req service.AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("分配权限参数绑定失败: %v", err)
		response.BadRequest(c, "参数错误")
		return
	}

	assignment, err := h.permissionAssignmentService.AssignPermissions(c.Request.Context(), &req)
	if err != nil {
		h.logger.Errorf("分配权限失败: %v", err)
		response.InternalError(c, "分配权限失败")
		return
	}

	h.logger.Infof("成功分配权限: user=%d", req.UserID)
	response.Success(c, assignment)
}

// GetPermissionAssignment 获取权限分配
func (h *PermissionAssignmentHandler) GetPermissionAssignment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的分配ID")
		return
	}

	assignment, err := h.permissionAssignmentService.GetPermissionAssignment(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Errorf("获取权限分配失败: %v", err)
		response.StatusNotFound(c, "权限分配不存在")
		return
	}

	response.Success(c, assignment)
}

// ListPermissionAssignments 获取权限分配列表
func (h *PermissionAssignmentHandler) ListPermissionAssignments(c *gin.Context) {
	var req service.ListPermissionAssignmentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warnf("权限分配列表参数绑定失败: %v", err)
		response.BadRequest(c, "参数错误")
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	assignments, err := h.permissionAssignmentService.ListPermissionAssignments(c.Request.Context(), &req)
	if err != nil {
		h.logger.Errorf("获取权限分配列表失败: %v", err)
		response.InternalError(c, "获取权限分配列表失败")
		return
	}

	response.Success(c, assignments)
}

// RevokePermissionAssignment 撤销权限分配
func (h *PermissionAssignmentHandler) RevokePermissionAssignment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的分配ID")
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("撤销权限分配参数绑定失败: %v", err)
		response.BadRequest(c, "参数错误")
		return
	}

	// 从JWT中获取操作员ID
	operatorID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	if err := h.permissionAssignmentService.RevokePermissionAssignment(c.Request.Context(), uint(id), req.Reason, operatorID.(uint)); err != nil {
		h.logger.Errorf("撤销权限分配失败: %v", err)
		response.InternalError(c, "撤销权限分配失败")
		return
	}

	h.logger.Infof("成功撤销权限分配: %d", id)
	response.Success(c, gin.H{"message": "权限分配撤销成功"})
}

// GetPermissionAssignmentHistory 获取权限分配历史
func (h *PermissionAssignmentHandler) GetPermissionAssignmentHistory(c *gin.Context) {
	var req service.GetPermissionAssignmentHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warnf("权限分配历史参数绑定失败: %v", err)
		response.BadRequest(c, "参数错误")
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	history, err := h.permissionAssignmentService.GetPermissionAssignmentHistory(c.Request.Context(), &req)
	if err != nil {
		h.logger.Errorf("获取权限分配历史失败: %v", err)
		response.InternalError(c, "获取权限分配历史失败")
		return
	}

	response.Success(c, history)
}

// CreateOnboardingPermissionConfig 创建入职权限配置
func (h *PermissionAssignmentHandler) CreateOnboardingPermissionConfig(c *gin.Context) {
	var req service.CreateOnboardingPermissionConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("创建入职权限配置参数绑定失败: %v", err)
		response.BadRequest(c, "参数错误")
		return
	}

	config, err := h.permissionAssignmentService.CreateOnboardingPermissionConfig(c.Request.Context(), &req)
	if err != nil {
		h.logger.Errorf("创建入职权限配置失败: %v", err)
		response.InternalError(c, "创建入职权限配置失败")
		return
	}

	h.logger.Infof("成功创建入职权限配置: %s", req.OnboardingStatus)
	response.Success(c, config)
}

// GetOnboardingPermissionConfig 获取入职权限配置
func (h *PermissionAssignmentHandler) GetOnboardingPermissionConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的配置ID")
		return
	}

	config, err := h.permissionAssignmentService.GetOnboardingPermissionConfig(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Errorf("获取入职权限配置失败: %v", err)
		response.StatusNotFound(c, "入职权限配置不存在")
		return
	}

	response.Success(c, config)
}

// ListOnboardingPermissionConfigs 获取入职权限配置列表
func (h *PermissionAssignmentHandler) ListOnboardingPermissionConfigs(c *gin.Context) {
	var req service.ListOnboardingPermissionConfigsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warnf("入职权限配置列表参数绑定失败: %v", err)
		response.BadRequest(c, "参数错误")
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	configs, err := h.permissionAssignmentService.ListOnboardingPermissionConfigs(c.Request.Context(), &req)
	if err != nil {
		h.logger.Errorf("获取入职权限配置列表失败: %v", err)
		response.InternalError(c, "获取入职权限配置列表失败")
		return
	}

	response.Success(c, configs)
}

// UpdateOnboardingPermissionConfig 更新入职权限配置
func (h *PermissionAssignmentHandler) UpdateOnboardingPermissionConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的配置ID")
		return
	}

	var req service.UpdateOnboardingPermissionConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("更新入职权限配置参数绑定失败: %v", err)
		response.BadRequest(c, "参数错误")
		return
	}

	config, err := h.permissionAssignmentService.UpdateOnboardingPermissionConfig(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Errorf("更新入职权限配置失败: %v", err)
		response.InternalError(c, "更新入职权限配置失败")
		return
	}

	h.logger.Infof("成功更新入职权限配置: %d", id)
	response.Success(c, config)
}

// DeleteOnboardingPermissionConfig 删除入职权限配置
func (h *PermissionAssignmentHandler) DeleteOnboardingPermissionConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的配置ID")
		return
	}

	if err := h.permissionAssignmentService.DeleteOnboardingPermissionConfig(c.Request.Context(), uint(id)); err != nil {
		h.logger.Errorf("删除入职权限配置失败: %v", err)
		response.InternalError(c, "删除入职权限配置失败")
		return
	}

	h.logger.Infof("成功删除入职权限配置: %d", id)
	response.Success(c, gin.H{"message": "入职权限配置删除成功"})
}

// ProcessPermissionApproval 处理权限审批
func (h *PermissionAssignmentHandler) ProcessPermissionApproval(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的分配ID")
		return
	}

	var req struct {
		Approved bool   `json:"approved" binding:"required"`
		Comments string `json:"comments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("处理权限审批参数绑定失败: %v", err)
		response.BadRequest(c, "参数错误")
		return
	}

	// 从JWT中获取审批员ID
	approverID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	if err := h.permissionAssignmentService.ProcessPermissionApproval(c.Request.Context(), uint(id), req.Approved, approverID.(uint), req.Comments); err != nil {
		h.logger.Errorf("处理权限审批失败: %v", err)
		response.InternalError(c, "处理权限审批失败")
		return
	}

	action := "拒绝"
	if req.Approved {
		action = "批准"
	}
	h.logger.Infof("成功%s权限分配: %d", action, id)
	response.Success(c, gin.H{"message": fmt.Sprintf("权限分配%s成功", action)})
}

// GetPendingPermissionApprovals 获取待审批权限
func (h *PermissionAssignmentHandler) GetPendingPermissionApprovals(c *gin.Context) {
	// 从JWT中获取审批员ID
	approverID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	approvals, err := h.permissionAssignmentService.GetPendingPermissionApprovals(c.Request.Context(), approverID.(uint))
	if err != nil {
		h.logger.Errorf("获取待审批权限失败: %v", err)
		response.InternalError(c, "获取待审批权限失败")
		return
	}

	response.Success(c, approvals)
}
