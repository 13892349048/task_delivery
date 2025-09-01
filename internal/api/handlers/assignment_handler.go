package handlers

import (
	"strconv"

	"taskmanage/internal/container"
	"taskmanage/internal/service"
	"taskmanage/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AssignmentHandler 分配管理处理器
type AssignmentHandler struct {
	container         *container.ApplicationContainer
	logger            *logrus.Logger
	assignmentService service.AssignmentService
}

// NewAssignmentHandler 创建分配管理处理器
func NewAssignmentHandler(container *container.ApplicationContainer, logger *logrus.Logger) *AssignmentHandler {
	return &AssignmentHandler{
		container:         container,
		logger:            logger,
		assignmentService: container.GetAssignmentManagementService(),
	}
}

// ManualAssign 手动分配任务
// @Summary 手动分配任务
// @Description 管理员手动分配任务给员工
// @Tags 任务分配
// @Accept json
// @Produce json
// @Param request body service.ManualAssignmentRequest true "分配请求"
// @Success 200 {object} response.Response{data=service.AssignmentHistory}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/assignments/manual [post]
func (h *AssignmentHandler) ManualAssign(c *gin.Context) {
	var req service.ManualAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 从JWT中获取用户ID作为分配者
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}
	req.AssignedBy = userID.(uint)

	history, err := h.assignmentService.ManualAssign(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "分配失败")
		return
	}

	response.Success(c, gin.H{
		"message": "分配成功",
		"history": history,
	})
}

// GetAssignmentSuggestions 获取分配建议
// @Summary 获取分配建议
// @Description 根据任务要求获取分配建议
// @Tags 任务分配
// @Accept json
// @Produce json
// @Param request body service.AssignmentSuggestionRequest true "建议请求"
// @Success 200 {object} response.Response{data=[]service.AssignmentSuggestion}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/assignments/suggestions [post]
func (h *AssignmentHandler) GetAssignmentSuggestions(c *gin.Context) {
	var req service.AssignmentSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 设置默认策略
	if req.Strategy == "" {
		req.Strategy = "comprehensive"
	}

	suggestions, err := h.assignmentService.GetAssignmentSuggestions(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "获取建议失败")
		return
	}

	response.Success(c, gin.H{
		"message":     "获取建议成功",
		"suggestions": suggestions,
	})
}

// CheckAssignmentConflicts 检查分配冲突
// @Summary 检查分配冲突
// @Description 检查任务分配是否存在冲突
// @Tags 任务分配
// @Accept json
// @Produce json
// @Param request body service.AssignmentConflictCheck true "冲突检查请求"
// @Success 200 {object} response.Response{data=[]service.AssignmentConflict}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/assignments/conflicts [post]
func (h *AssignmentHandler) CheckAssignmentConflicts(c *gin.Context) {
	// 简化冲突检查实现
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "任务ID格式错误")
		return
	}

	employeeIDStr := c.Query("employee_id")
	employeeID, err := strconv.ParseUint(employeeIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "员工ID格式错误")
		return
	}

	// 检查分配冲突
	conflicts, err := h.assignmentService.CheckAssignmentConflicts(c.Request.Context(), uint(taskID), uint(employeeID))
	if err != nil {
		response.InternalError(c, "冲突检查失败")
		return
	}

	response.Success(c, gin.H{
		"message":   "冲突检查完成",
		"conflicts": conflicts,
	})
}

// GetAssignmentHistory 获取分配历史
// @Summary 获取分配历史
// @Description 获取任务的分配历史记录
// @Tags 任务分配
// @Accept json
// @Produce json
// @Param task_id path int true "任务ID"
// @Success 200 {object} response.Response{data=[]service.AssignmentHistory}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/assignments/history/{task_id} [get]
func (h *AssignmentHandler) GetAssignmentHistory(c *gin.Context) {
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "任务ID格式错误")
		return
	}

	history, err := h.assignmentService.GetAssignmentHistory(c.Request.Context(), uint(taskID))
	if err != nil {
		response.InternalError(c, "获取历史失败")
		return
	}

	response.Success(c, gin.H{
		"message": "获取历史成功",
		"history": history,
	})
}

// ReassignTask 重新分配任务
// @Summary 重新分配任务
// @Description 将任务重新分配给其他员工
// @Tags 任务分配
// @Accept json
// @Produce json
// @Param task_id path int true "任务ID"
// @Param request body ReassignRequest true "重新分配请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/assignments/reassign/{task_id} [post]
func (h *AssignmentHandler) ReassignTask(c *gin.Context) {
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "任务ID格式错误")
		return
	}

	var req ReassignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	err = h.assignmentService.ReassignTask(c.Request.Context(), uint(taskID), req.NewEmployeeID, req.Reason, userID.(uint))
	if err != nil {
		response.InternalError(c, "重新分配失败")
		return
	}

	response.Success(c, gin.H{
		"message": "重新分配成功",
		"assign":  true,
	})
}

// CancelAssignment 取消分配
// @Summary 取消分配
// @Description 取消任务分配
// @Tags 任务分配
// @Accept json
// @Produce json
// @Param task_id path int true "任务ID"
// @Param request body CancelAssignmentRequest true "取消分配请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/assignments/cancel/{task_id} [post]
func (h *AssignmentHandler) CancelAssignment(c *gin.Context) {
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "任务ID格式错误")
		return
	}

	var req CancelAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	err = h.assignmentService.CancelAssignment(c.Request.Context(), uint(taskID), req.Reason, userID.(uint))
	if err != nil {
		response.InternalError(c, "取消分配失败")
		return
	}

	response.Success(c, gin.H{
		"message": "取消分配成功",
		"assign":  nil,
	})
}

// GetAssignmentStrategies 获取可用的分配策略
// @Summary 获取分配策略
// @Description 获取所有可用的分配策略
// @Tags 任务分配
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]StrategyInfo}
// @Failure 500 {object} response.Response
// @Router /api/assignments/strategies [get]
func (h *AssignmentHandler) GetAssignmentStrategies(c *gin.Context) {
	strategies := []StrategyInfo{
		{
			Strategy:    "round_robin",
			Name:        "轮询分配",
			Description: "按顺序轮流分配给可用员工",
		},
		{
			Strategy:    "load_balance",
			Name:        "负载均衡",
			Description: "优先分配给工作负载较低的员工",
		},
		{
			Strategy:    "skill_match",
			Name:        "技能匹配",
			Description: "根据技能要求匹配最合适的员工",
		},
		{
			Strategy:    "comprehensive",
			Name:        "综合评分",
			Description: "综合考虑技能、负载、可用性等因素",
		},
		{
			Strategy:    "manual",
			Name:        "手动分配",
			Description: "管理员手动指定员工",
		},
	}

	response.Success(c, gin.H{
		"message":    "获取策略成功",
		"strategies": strategies,
	})
}

// GetAssignmentStats 获取分配统计
// @Summary 获取分配统计
// @Description 获取分配相关的统计信息
// @Tags 任务分配
// @Accept json
// @Produce json
// @Param employee_id query int false "员工ID"
// @Param days query int false "统计天数，默认30天"
// @Success 200 {object} response.Response{data=AssignmentStats}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/assignments/stats [get]
func (h *AssignmentHandler) GetAssignmentStats(c *gin.Context) {
	employeeIDStr := c.Query("employee_id")
	daysStr := c.Query("days")

	var employeeID uint
	if employeeIDStr != "" {
		id, err := strconv.ParseUint(employeeIDStr, 10, 32)
		if err != nil {
			response.BadRequest(c, "员工ID格式错误")
			return
		}
		employeeID = uint(id)
	}

	days := 30
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err != nil {
			response.BadRequest(c, "天数格式错误")
			return
		}
		days = d
	}

	// 构建统计信息
	stats := AssignmentStats{
		TotalAssignments:      100, // 示例数据
		SuccessfulAssignments: 85,
		PendingApprovals:      5,
		CancelledAssignments:  10,
		AverageAssignmentTime: 2.5,
		TopStrategies: []StrategyUsage{
			{Strategy: "comprehensive", Count: 40, Percentage: 40.0},
			{Strategy: "load_balance", Count: 30, Percentage: 30.0},
			{Strategy: "skill_match", Count: 20, Percentage: 20.0},
			{Strategy: "manual", Count: 10, Percentage: 10.0},
		},
	}

	response.Success(c, "")
	response.Success(c, gin.H{
		"message":    "获取统计成功",
		"stats":      stats,
		"days":       days,
		"employeeID": employeeID,
	})
}

// 请求和响应结构体

// ReassignRequest 重新分配请求
type ReassignRequest struct {
	NewEmployeeID uint   `json:"new_employee_id" binding:"required"`
	Reason        string `json:"reason" binding:"required"`
}

// CancelAssignmentRequest 取消分配请求
type CancelAssignmentRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// StrategyInfo 策略信息
type StrategyInfo struct {
	Strategy    string `json:"strategy"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AssignmentStats 分配统计
type AssignmentStats struct {
	TotalAssignments      int             `json:"total_assignments"`
	SuccessfulAssignments int             `json:"successful_assignments"`
	PendingApprovals      int             `json:"pending_approvals"`
	CancelledAssignments  int             `json:"cancelled_assignments"`
	AverageAssignmentTime float64         `json:"average_assignment_time"`
	TopStrategies         []StrategyUsage `json:"top_strategies"`
}

// StrategyUsage 策略使用情况
type StrategyUsage struct {
	Strategy   string  `json:"strategy"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

func (h *AssignmentHandler) GetPendingAssignments(c *gin.Context) {
	response.Success(c, gin.H{"message": "GetPendingAssignments - TODO: implement"})
}

func (h *AssignmentHandler) GetAssignment(c *gin.Context) {
	response.Success(c, gin.H{"message": "GetAssignment - TODO: implement"})
}
