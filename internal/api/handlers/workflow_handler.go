package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"taskmanage/internal/service"
	"taskmanage/internal/workflow"
	"taskmanage/pkg/response"
)

// WorkflowHandler 工作流处理器
type WorkflowHandler struct {
	workflowService service.WorkflowService
	logger          *logrus.Logger
}

// NewWorkflowHandler 创建工作流处理器
func NewWorkflowHandler(workflowService service.WorkflowService, logger *logrus.Logger) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
		logger:          logger,
	}
}

// StartTaskAssignmentApproval 启动任务分配审批流程
// @Summary 启动任务分配审批流程
// @Description 为任务分配启动审批流程
// @Tags workflow
// @Accept json
// @Produce json
// @Param request body workflow.TaskAssignmentApprovalRequest true "审批请求"
// @Success 200 {object} response.Response{data=workflow.WorkflowInstance}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/task-assignment/start [post]
func (h *WorkflowHandler) StartTaskAssignmentApproval(c *gin.Context) {
	var req workflow.TaskAssignmentApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("解析请求参数失败")
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	req.RequesterID = userID.(uint)

	instance, err := h.workflowService.StartTaskAssignmentApproval(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("启动任务分配审批流程失败")
		response.InternalError(c, "启动审批流程失败")
		return
	}

	response.SuccessWithMessage(c, "审批流程启动成功", instance)
}

// ProcessApproval 处理审批决策
// @Summary 处理审批决策
// @Description 处理审批决策（同意/拒绝/退回）
// @Tags workflow
// @Accept json
// @Produce json
// @Param request body workflow.ApprovalRequest true "审批决策"
// @Success 200 {object} response.Response{data=workflow.ApprovalResult}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/approvals/process [post]
func (h *WorkflowHandler) ProcessApproval(c *gin.Context) {
	var req workflow.ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("解析请求参数失败")
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	req.ApprovedBy = userID.(uint)

	result, err := h.workflowService.ProcessTaskAssignmentApproval(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("处理审批决策失败")
		response.InternalError(c, "处理审批失败")
		return
	}

	response.SuccessWithMessage(c, "审批处理成功", result)
}

// GetWorkflowInstance 获取流程实例
// @Summary 获取流程实例
// @Description 根据实例ID获取流程实例详情
// @Tags workflow
// @Accept json
// @Produce json
// @Param instance_id path string true "实例ID"
// @Success 200 {object} response.Response{data=workflow.WorkflowInstance}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/instances/{instance_id} [get]
func (h *WorkflowHandler) GetWorkflowInstance(c *gin.Context) {
	instanceID := c.Param("instance_id")
	if instanceID == "" {
		response.BadRequest(c, "实例ID不能为空")
		return
	}

	instance, err := h.workflowService.GetWorkflowInstance(c.Request.Context(), instanceID)
	if err != nil {
		h.logger.WithError(err).Error("获取流程实例失败")
		response.InternalError(c, "获取流程实例失败")
		return
	}

	response.SuccessWithMessage(c, "获取流程实例成功", instance)
}

// GetPendingApprovals 获取待审批任务
// @Summary 获取待审批任务
// @Description 获取当前用户的待审批任务列表
// @Tags workflow
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]workflow.PendingApproval}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/approvals/pending [get]
func (h *WorkflowHandler) GetPendingApprovals(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	approvals, err := h.workflowService.GetPendingApprovals(c.Request.Context(), userID.(uint))
	if err != nil {
		h.logger.WithError(err).Error("获取待审批任务失败")
		response.InternalError(c, "获取待审批任务失败")
		return
	}

	response.SuccessWithMessage(c, "获取待审批任务成功", approvals)
}

// CancelWorkflow 取消流程
// @Summary 取消流程
// @Description 取消指定的流程实例
// @Tags workflow
// @Accept json
// @Produce json
// @Param instance_id path string true "实例ID"
// @Param request body CancelWorkflowRequest true "取消原因"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/instances/{instance_id}/cancel [post]
func (h *WorkflowHandler) CancelWorkflow(c *gin.Context) {
	instanceID := c.Param("instance_id")
	if instanceID == "" {
		response.BadRequest(c, "实例ID不能为空")
		return
	}

	var req CancelWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("解析请求参数失败")
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	err := h.workflowService.CancelWorkflow(c.Request.Context(), instanceID, req.Reason)
	if err != nil {
		h.logger.WithError(err).Error("取消流程失败")
		response.InternalError(c, "取消流程失败")
		return
	}

	response.SuccessWithMessage(c, "流程取消成功", nil)
}

// GetWorkflowHistory 获取流程历史
// @Summary 获取流程历史
// @Description 获取指定流程实例的执行历史
// @Tags workflow
// @Accept json
// @Produce json
// @Param instance_id path string true "实例ID"
// @Success 200 {object} response.Response{data=[]workflow.ExecutionHistory}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/instances/{instance_id}/history [get]
func (h *WorkflowHandler) GetWorkflowHistory(c *gin.Context) {
	instanceID := c.Param("instance_id")
	if instanceID == "" {
		response.BadRequest(c, "实例ID不能为空")
		return
	}

	history, err := h.workflowService.GetWorkflowHistory(c.Request.Context(), instanceID)
	if err != nil {
		h.logger.WithError(err).Error("获取流程历史失败")
		response.InternalError(c, "获取流程历史失败")
		return
	}

	response.SuccessWithMessage(c, "获取流程历史成功", history)
}

// GetApprovalCount 获取待审批数量
// @Summary 获取待审批数量
// @Description 获取当前用户的待审批任务数量
// @Tags workflow
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=ApprovalCountResponse}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/approvals/count [get]
func (h *WorkflowHandler) GetApprovalCount(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	approvals, err := h.workflowService.GetPendingApprovals(c.Request.Context(), userID.(uint))
	if err != nil {
		h.logger.WithError(err).Error("获取待审批任务失败")
		response.InternalError(c, "获取待审批任务失败")
		return
	}

	countResp := ApprovalCountResponse{
		Total: len(approvals),
	}

	response.SuccessWithMessage(c, "获取待审批数量成功", countResp)
}

// CreateWorkflowDefinition 创建工作流定义
// @Summary 创建工作流定义
// @Description 创建新的工作流定义
// @Tags workflow
// @Accept json
// @Produce json
// @Param request body workflow.CreateWorkflowRequest true "工作流定义"
// @Success 200 {object} response.Response{data=workflow.WorkflowDefinition}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/definitions [post]
func (h *WorkflowHandler) CreateWorkflowDefinition(c *gin.Context) {
	var req workflow.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("解析工作流定义参数失败")
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	definition, err := h.workflowService.CreateWorkflowDefinition(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("创建工作流定义失败")
		response.InternalError(c, "创建工作流定义失败")
		return
	}

	response.SuccessWithMessage(c, "工作流定义创建成功", definition)
}

// GetWorkflowDefinitions 获取工作流定义列表
// @Summary 获取工作流定义列表
// @Description 获取所有工作流定义
// @Tags workflow
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]workflow.WorkflowDefinition}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/definitions [get]
func (h *WorkflowHandler) GetWorkflowDefinitions(c *gin.Context) {
	definitions, err := h.workflowService.GetWorkflowDefinitions(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("获取工作流定义列表失败")
		response.InternalError(c, "获取工作流定义列表失败")
		return
	}

	response.SuccessWithMessage(c, "获取工作流定义列表成功", definitions)
}

// GetWorkflowDefinition 获取工作流定义详情
// @Summary 获取工作流定义详情
// @Description 根据ID获取工作流定义详情
// @Tags workflow
// @Accept json
// @Produce json
// @Param id path string true "工作流定义ID"
// @Success 200 {object} response.Response{data=workflow.WorkflowDefinition}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/definitions/{id} [get]
func (h *WorkflowHandler) GetWorkflowDefinition(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "工作流定义ID不能为空")
		return
	}

	definition, err := h.workflowService.GetWorkflowDefinition(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("获取工作流定义失败")
		response.InternalError(c, "获取工作流定义失败")
		return
	}

	response.SuccessWithMessage(c, "获取工作流定义成功", definition)
}

// UpdateWorkflowDefinition 更新工作流定义
// @Summary 更新工作流定义
// @Description 更新指定的工作流定义
// @Tags workflow
// @Accept json
// @Produce json
// @Param id path string true "工作流定义ID"
// @Param request body workflow.UpdateWorkflowRequest true "更新请求"
// @Success 200 {object} response.Response{data=workflow.WorkflowDefinition}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/definitions/{id} [put]
func (h *WorkflowHandler) UpdateWorkflowDefinition(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "工作流定义ID不能为空")
		return
	}

	var req workflow.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("解析更新参数失败")
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	definition, err := h.workflowService.UpdateWorkflowDefinition(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.WithError(err).Error("更新工作流定义失败")
		response.InternalError(c, "更新工作流定义失败")
		return
	}

	response.SuccessWithMessage(c, "工作流定义更新成功", definition)
}

// DeleteWorkflowDefinition 删除工作流定义
// @Summary 删除工作流定义
// @Description 删除指定的工作流定义
// @Tags workflow
// @Accept json
// @Produce json
// @Param id path string true "工作流定义ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/definitions/{id} [delete]
func (h *WorkflowHandler) DeleteWorkflowDefinition(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "工作流定义ID不能为空")
		return
	}

	err := h.workflowService.DeleteWorkflowDefinition(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("删除工作流定义失败")
		response.InternalError(c, "删除工作流定义失败")
		return
	}

	response.SuccessWithMessage(c, "工作流定义删除成功", nil)
}

// ValidateWorkflowDefinition 验证工作流定义
// @Summary 验证工作流定义
// @Description 验证工作流定义的有效性
// @Tags workflow
// @Accept json
// @Produce json
// @Param request body workflow.CreateWorkflowRequest true "工作流定义"
// @Success 200 {object} response.Response{data=ValidationResult}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/definitions/validate [post]
func (h *WorkflowHandler) ValidateWorkflowDefinition(c *gin.Context) {
	var req workflow.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("解析验证参数失败")
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	err := h.workflowService.ValidateWorkflowDefinition(c.Request.Context(), &req)
	result := ValidationResult{
		Valid: err == nil,
	}
	if err != nil {
		result.Error = err.Error()
	}

	response.SuccessWithMessage(c, "工作流定义验证完成", result)
}

// CancelWorkflowRequest 取消流程请求
type CancelWorkflowRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// ApprovalCountResponse 待审批数量响应
type ApprovalCountResponse struct {
	Total int `json:"total"`
}

// GetPendingTaskAssignmentApprovals 获取待审批的任务分配
// @Summary 获取待审批的任务分配
// @Description 获取当前用户待审批的任务分配列表
// @Tags workflow
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]workflow.PendingApproval}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/workflows/approvals/task-assignments [get]
func (h *WorkflowHandler) GetPendingTaskAssignmentApprovals(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户信息缺失")
		return
	}

	approvals, err := h.workflowService.GetPendingTaskAssignmentApprovals(c.Request.Context(), userID.(uint))
	if err != nil {
		h.logger.WithError(err).Error("获取待审批任务分配列表失败")
		response.InternalError(c, "获取待审批任务分配列表失败")
		return
	}

	response.SuccessWithMessage(c, "获取待审批任务分配列表成功", approvals)
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}
