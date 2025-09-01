package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"taskmanage/internal/service"
)

// OnboardingHandler 入职工作流处理器
type OnboardingHandler struct {
	onboardingService service.OnboardingService
	logger            *logrus.Logger
}

// NewOnboardingHandler 创建入职工作流处理器
func NewOnboardingHandler(onboardingService service.OnboardingService, logger *logrus.Logger) *OnboardingHandler {
	return &OnboardingHandler{
		onboardingService: onboardingService,
		logger:            logger,
	}
}

// CreatePendingEmployee 创建待入职员工
func (h *OnboardingHandler) CreatePendingEmployee(c *gin.Context) {
	var req service.CreatePendingEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定创建待入职员工请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"real_name": req.RealName,
		"email":     req.Email,
	}).Info("处理创建待入职员工请求")

	result, err := h.onboardingService.CreatePendingEmployee(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("创建待入职员工失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建待入职员工失败", "details": err.Error()})
		return
	}

	h.logger.Info("创建待入职员工成功")
	c.JSON(http.StatusOK, gin.H{"message": "创建待入职员工成功", "data": result})
}

// ConfirmOnboarding 确认入职
func (h *OnboardingHandler) ConfirmOnboarding(c *gin.Context) {
	var req service.OnboardConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定确认入职请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"employee_id": req.EmployeeID,
		"department_id": req.DepartmentID,
	}).Info("处理确认入职请求")

	result, err := h.onboardingService.ConfirmOnboarding(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("确认入职失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "确认入职失败", "details": err.Error()})
		return
	}

	h.logger.Info("确认入职成功")
	c.JSON(http.StatusOK, gin.H{"message": "确认入职成功", "data": result})
}

// CompleteProbation 完成试用期
func (h *OnboardingHandler) CompleteProbation(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	employeeID, err := strconv.ParseUint(employeeIDStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).Error("员工ID无效")
		c.JSON(http.StatusBadRequest, gin.H{"error": "员工ID无效", "details": err.Error()})
		return
	}

	// 从JWT中获取操作员ID
	operatorID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("无法获取操作员ID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取操作员信息"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"employee_id": employeeID,
		"operator_id": operatorID,
	}).Info("处理完成试用期请求")

	result, err := h.onboardingService.CompleteProbation(c.Request.Context(), uint(employeeID), operatorID.(uint))
	if err != nil {
		h.logger.WithError(err).Error("完成试用期失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "完成试用期失败", "details": err.Error()})
		return
	}

	h.logger.Info("完成试用期成功")
	c.JSON(http.StatusOK, gin.H{"message": "完成试用期成功", "data": result})
}

// ConfirmEmployee 确认员工（试用期转正）
func (h *OnboardingHandler) ConfirmEmployee(c *gin.Context) {
	var req service.ProbationToActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定员工确认请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	// 从JWT中获取操作员ID
	operatorID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("无法获取操作员ID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取操作员信息"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"employee_id": req.EmployeeID,
		"operator_id": operatorID,
	}).Info("处理员工确认请求")

	result, err := h.onboardingService.ConfirmEmployee(c.Request.Context(), &req, operatorID.(uint))
	if err != nil {
		h.logger.WithError(err).Error("确认员工失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "确认员工失败", "details": err.Error()})
		return
	}

	h.logger.Info("确认员工成功")
	c.JSON(http.StatusOK, gin.H{"message": "确认员工成功", "data": result})
}

// ChangeEmployeeStatus 更改员工入职状态
func (h *OnboardingHandler) ChangeEmployeeStatus(c *gin.Context) {
	var req service.EmployeeStatusChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定状态更改请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	// 从JWT中获取操作员ID
	operatorID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("无法获取操作员ID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取操作员信息"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"employee_id": req.EmployeeID,
		"new_status": req.NewStatus,
		"operator_id": operatorID,
	}).Info("处理员工状态更改请求")

	result, err := h.onboardingService.ChangeEmployeeStatus(c.Request.Context(), &req, operatorID.(uint))
	if err != nil {
		h.logger.WithError(err).Error("更改员工状态失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更改员工状态失败", "details": err.Error()})
		return
	}

	h.logger.Info("更改员工状态成功")
	c.JSON(http.StatusOK, gin.H{"message": "更改员工状态成功", "data": result})
}

// GetOnboardingWorkflows 获取入职工作流列表
// @Summary 获取入职工作流列表
// @Description 获取入职工作流列表，支持分页和过滤
// @Tags 入职工作流
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "入职状态过滤"
// @Param department query string false "部门过滤"
// @Param date_from query string false "开始日期过滤"
// @Param date_to query string false "结束日期过滤"
// @Success 200 {object} response.Response{data=[]service.OnboardingWorkflowResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/onboarding/workflows [get]
func (h *OnboardingHandler) GetOnboardingWorkflows(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	filter := &service.OnboardingWorkflowFilter{
		Page:       page,
		PageSize:   pageSize,
		Status:     c.Query("status"),
		Department: c.Query("department"),
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
	}

	h.logger.WithFields(logrus.Fields{
		"page": page,
		"page_size": pageSize,
		"status": filter.Status,
	}).Info("处理获取入职工作流列表请求")

	workflows, err := h.onboardingService.GetOnboardingWorkflows(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("获取入职工作流列表失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取入职工作流列表失败", "details": err.Error()})
		return
	}

	h.logger.Info("获取入职工作流列表成功")
	c.JSON(http.StatusOK, gin.H{"message": "获取入职工作流列表成功", "data": workflows})
}

// GetOnboardingHistory 获取入职历史记录
func (h *OnboardingHandler) GetOnboardingHistory(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	employeeID, err := strconv.ParseUint(employeeIDStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).Error("员工ID无效")
		c.JSON(http.StatusBadRequest, gin.H{"error": "员工ID无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"employee_id": employeeID,
	}).Info("处理获取入职历史记录请求")

	history, err := h.onboardingService.GetOnboardingHistory(c.Request.Context(), uint(employeeID))
	if err != nil {
		h.logger.WithError(err).Error("获取入职历史记录失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取入职历史记录失败", "details": err.Error()})
		return
	}

	h.logger.Info("获取入职历史记录成功")
	c.JSON(http.StatusOK, gin.H{"message": "获取入职历史记录成功", "data": history})
}

// StartOnboardingApproval 启动入职审批流程
func (h *OnboardingHandler) StartOnboardingApproval(c *gin.Context) {
	var req service.OnboardingApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("解析启动审批请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 从JWT中获取操作员ID
	if userID, exists := c.Get("user_id"); exists {
		req.RequesterID = userID.(uint)
	}

	result, err := h.onboardingService.StartOnboardingApproval(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("启动入职审批失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "启动入职审批失败", "details": err.Error()})
		return
	}

	h.logger.Info("启动入职审批成功")
	c.JSON(http.StatusOK, gin.H{"message": "启动入职审批成功", "data": result})
}

// ProcessOnboardingApproval 处理入职审批决策
func (h *OnboardingHandler) ProcessOnboardingApproval(c *gin.Context) {
	// 先读取原始请求体用于调试
	body, _ := c.GetRawData()
	h.logger.Infof("收到入职审批请求，原始请求体: %s", string(body))
	
	// 重新设置请求体，因为GetRawData()会消耗掉
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
	
	var req service.ProcessOnboardingApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("解析审批处理请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}
	
	h.logger.Infof("解析后的请求: InstanceID=%s, NodeID=%s, Action=%s, Comment=%s", req.InstanceID, req.NodeID, req.Action, req.Comment)

	// 从JWT中获取审批员ID
	if userID, exists := c.Get("user_id"); exists {
		req.ApproverID = userID.(uint)
	}

	result, err := h.onboardingService.ProcessOnboardingApproval(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("处理入职审批失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理入职审批失败", "details": err.Error()})
		return
	}

	h.logger.Info("处理入职审批成功")
	c.JSON(http.StatusOK, gin.H{"message": "处理入职审批成功", "data": result})
}

// GetPendingOnboardingApprovals 获取待审批的入职申请
func (h *OnboardingHandler) GetPendingOnboardingApprovals(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("无法获取用户ID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	result, err := h.onboardingService.GetPendingOnboardingApprovals(c.Request.Context(), userID.(uint))
	if err != nil {
		h.logger.WithError(err).Error("获取待审批入职申请失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取待审批入职申请失败", "details": err.Error()})
		return
	}

	h.logger.Info("获取待审批入职申请成功")
	c.JSON(http.StatusOK, gin.H{"message": "获取待审批入职申请成功", "data": result})
}

// GetOnboardingApprovalHistory 获取入职审批历史
func (h *OnboardingHandler) GetOnboardingApprovalHistory(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	employeeID, err := strconv.ParseUint(employeeIDStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).Error("解析员工ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "员工ID格式错误"})
		return
	}

	result, err := h.onboardingService.GetOnboardingApprovalHistory(c.Request.Context(), uint(employeeID))
	if err != nil {
		h.logger.WithError(err).Error("获取入职审批历史失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取入职审批历史失败", "details": err.Error()})
		return
	}

	h.logger.Info("获取入职审批历史成功")
	c.JSON(http.StatusOK, gin.H{"message": "获取入职审批历史成功", "data": result})
}

// CancelOnboardingApproval 取消入职审批流程
func (h *OnboardingHandler) CancelOnboardingApproval(c *gin.Context) {
	instanceID := c.Param("instance_id")
	if instanceID == "" {
		h.logger.Error("工作流实例ID不能为空")
		c.JSON(http.StatusBadRequest, gin.H{"error": "工作流实例ID不能为空"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("解析取消请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 从JWT中获取操作员ID
	operatorID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("无法获取操作员ID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	err := h.onboardingService.CancelOnboardingApproval(c.Request.Context(), instanceID, req.Reason, operatorID.(uint))
	if err != nil {
		h.logger.WithError(err).Error("取消入职审批失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消入职审批失败", "details": err.Error()})
		return
	}

	h.logger.Info("取消入职审批成功")
	c.JSON(http.StatusOK, gin.H{"message": "取消入职审批成功"})
}
