package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"taskmanage/internal/service"
	"taskmanage/pkg/logger"
	"taskmanage/pkg/response"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService       service.TaskService
	assignmentService service.AssignmentService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(container interface{}, _ interface{}) *TaskHandler {
	// 从容器中获取TaskService和AssignmentService
	if c, ok := container.(interface{ 
		GetServiceManager() service.ServiceManager
		GetAssignmentManagementService() service.AssignmentService
	}); ok {
		return &TaskHandler{
			taskService:       c.GetServiceManager().TaskService(),
			assignmentService: c.GetAssignmentManagementService(),
		}
	}
	panic("无法从容器中获取服务")
}

// CreateTask 创建任务
// @Summary 创建任务
// @Description 创建新任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param request body service.CreateTaskRequest true "创建任务请求"
// @Success 201 {object} response.Response{data=service.TaskResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/tasks [post]
// @Security BearerAuth
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req service.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("创建任务请求参数绑定失败: %v", err)
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 创建任务
	task, err := h.taskService.CreateTask(c.Request.Context(), &req)
	if err != nil {
		logger.Errorf("创建任务失败: %v", err)
		response.InternalError(c, "创建任务失败")
		return
	}

	logger.Infof("任务创建成功: ID=%d, Title=%s", task.ID, task.Title)
	c.JSON(http.StatusCreated, response.Response{
		Code:    response.ErrCodeSuccess,
		Message: "任务创建成功",
		Data:    task,
	})
}

// GetTask 获取任务详情
// @Summary 获取任务详情
// @Description 根据ID获取任务详情
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response{data=service.TaskResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "任务不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/tasks/{id} [get]
// @Security BearerAuth
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	task, err := h.taskService.GetTask(c.Request.Context(), uint(taskID))
	if err != nil {
		if err.Error() == "任务不存在" {
			response.NotFound(c, "任务不存在")
			return
		}
		logger.Errorf("获取任务失败: %v", err)
		response.InternalError(c, "获取任务失败")
		return
	}

	response.Success(c, task)
}

// UpdateTask 更新任务
// @Summary 更新任务
// @Description 更新任务信息
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Param request body service.UpdateTaskRequest true "更新任务请求"
// @Success 200 {object} response.Response{data=service.TaskResponse} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response "任务不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/tasks/{id} [put]
// @Security BearerAuth
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	var req service.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("更新任务请求参数绑定失败: %v", err)
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	task, err := h.taskService.UpdateTask(c.Request.Context(), uint(taskID), &req)
	if err != nil {
		if err.Error() == "任务不存在" {
			response.NotFound(c, "任务不存在")
			return
		}
		logger.Errorf("更新任务失败: %v", err)
		response.InternalError(c, "更新任务失败")
		return
	}

	logger.Infof("任务更新成功: ID=%d, Title=%s", task.ID, task.Title)
	response.Success(c, task)
}

// DeleteTask 删除任务
// @Summary 删除任务
// @Description 删除任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response "任务不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/tasks/{id} [delete]
// @Security BearerAuth
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	err = h.taskService.DeleteTask(c.Request.Context(), uint(taskID))
	if err != nil {
		if err.Error() == "任务不存在" {
			response.NotFound(c, "任务不存在")
			return
		}
		logger.Errorf("删除任务失败: %v", err)
		response.InternalError(c, "删除任务失败")
		return
	}

	logger.Infof("任务删除成功: ID=%d", taskID)
	response.Success(c, nil)
}

// ListTasks 获取任务列表
// @Summary 获取任务列表
// @Description 获取任务列表，支持分页和过滤
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param search query string false "搜索关键词"
// @Param status query string false "任务状态" Enums(pending,assigned,in_progress,completed,cancelled)
// @Param priority query string false "优先级" Enums(low,medium,high,urgent)
// @Param type query string false "任务类型" Enums(development,testing,design,documentation,maintenance,research)
// @Param category query string false "任务分类"
// @Param created_by query int false "创建者ID"
// @Param assigned_to query int false "分配给用户ID"
// @Param sort_by query string false "排序字段" default(created_at)
// @Param sort_desc query bool false "是否降序" default(true)
// @Success 200 {object} response.Response{data=response.ListResponse{items=[]service.TaskResponse}} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/tasks [get]
// @Security BearerAuth
func (h *TaskHandler) ListTasks(c *gin.Context) {
	// 解析查询参数
	filter := service.TaskListFilter{
		Page:     1,
		PageSize: 10,
	}

	// 分页参数
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = p
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			filter.PageSize = ps
		}
	}

	// 过滤参数
	filter.Status = c.Query("status")
	filter.Priority = c.Query("priority")

	// 创建者过滤
	if createdBy := c.Query("created_by"); createdBy != "" {
		if id, err := strconv.ParseUint(createdBy, 10, 32); err == nil {
			createdBy := uint(id)
		filter.CreatedBy = &createdBy
		}
	}

	// 分配者过滤
	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		if id, err := strconv.ParseUint(assignedTo, 10, 32); err == nil {
			idPtr := uint(id)
			filter.AssignedTo = &idPtr
		}
	}

	// 获取任务列表
	tasks, total, err := h.taskService.ListTasks(c.Request.Context(), filter)
	if err != nil {
		logger.Errorf("获取任务列表失败: %v", err)
		response.InternalError(c, "获取任务列表失败")
		return
	}

	// 使用分页响应
	response.SuccessWithPagination(c, tasks, filter.Page, filter.PageSize, total)
}

// GetTaskStats 获取任务统计信息
// @Summary 获取任务统计信息
// @Description 获取任务状态统计
// @Tags 任务管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]int64} "获取成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/tasks/stats [get]
// @Security BearerAuth
func (h *TaskHandler) GetTaskStats(c *gin.Context) {
	// 获取各状态的任务数量统计
	stats := make(map[string]int64)

	statuses := []string{"pending", "assigned", "in_progress", "completed", "cancelled"}
	for _, status := range statuses {
		filter := service.TaskListFilter{
			Status:   status,
			Page:     1,
			PageSize: 1, // 只需要获取总数，不需要实际数据
		}

		_, total, err := h.taskService.ListTasks(c.Request.Context(), filter)
		if err != nil {
			logger.Errorf("获取任务统计失败: %v", err)
			response.InternalError(c, "获取任务统计失败")
			return
		}

		stats[status] = total
	}

	// 计算总数
	var totalTasks int64
	for _, count := range stats {
		totalTasks += count
	}
	stats["total"] = totalTasks

	response.Success(c, stats)
}

// AssignTask 分配任务
func (h *TaskHandler) AssignTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		response.BadRequest(c, "任务ID不能为空")
		return
	}

	var req service.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("分配任务请求参数绑定失败: %v", err)
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 转换任务ID
	id, err := strconv.ParseUint(taskID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	// 获取当前用户ID并添加到上下文
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		logger.Warnf("无法获取用户ID: %v", err)
		response.Unauthorized(c, "用户信息缺失")
		return
	}

	// 将用户ID添加到标准context中
	ctx := SetUserIDInContext(c.Request.Context(), userID)

	// 执行任务分配
	req.TaskID = uint(id) // 设置任务ID
	result, err := h.taskService.AssignTask(ctx, &req)
	if err != nil {
		logger.Errorf("分配任务失败: %v", err)
		response.InternalError(c, "分配任务失败")
		return
	}

	response.Success(c, result)
}

// ReassignTask 重新分配任务
func (h *TaskHandler) ReassignTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		response.BadRequest(c, "任务ID不能为空")
		return
	}

	var req service.ReassignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("重新分配任务请求参数绑定失败: %v", err)
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 转换任务ID
	id, err := strconv.ParseUint(taskID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	// 执行任务重新分配
	_, err = h.taskService.ReassignTask(c.Request.Context(), uint(id), &service.ReassignTaskRequest{
		FromEmployeeID: req.FromEmployeeID,
		ToEmployeeID:   req.ToEmployeeID,
		Reason:         req.Reason,
	})
	if err != nil {
		logger.Errorf("重新分配任务失败: %v", err)
		response.InternalError(c, "重新分配任务失败")
		return
	}

	response.Success(c, gin.H{"message": "任务重新分配成功"})
}

// StartTask 开始任务
func (h *TaskHandler) StartTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		response.BadRequest(c, "任务ID不能为空")
		return
	}

	// 转换任务ID
	id, err := strconv.ParseUint(taskID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户信息缺失")
		return
	}

	// 执行开始任务
	err = h.taskService.StartTask(c.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		logger.Errorf("开始任务失败: %v", err)
		response.InternalError(c, "开始任务失败")
		return
	}

	response.Success(c, gin.H{"message": "任务已开始"})
}

// CompleteTask 完成任务
func (h *TaskHandler) CompleteTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		response.BadRequest(c, "任务ID不能为空")
		return
	}

	var req service.CompleteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("完成任务请求参数绑定失败: %v", err)
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 转换任务ID
	id, err := strconv.ParseUint(taskID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户信息缺失")
		return
	}

	// 执行完成任务
	err = h.taskService.CompleteTask(c.Request.Context(), uint(id), userID.(uint), &service.CompleteTaskRequest{
		Comment: req.Comment,
		Files:   req.Files,
	})
	if err != nil {
		logger.Errorf("完成任务失败: %v", err)
		response.InternalError(c, "完成任务失败")
		return
	}

	response.Success(c, gin.H{"message": "任务已完成"})
}

// CancelTask 取消任务
func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		response.BadRequest(c, "任务ID不能为空")
		return
	}

	var req service.CancelTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("取消任务请求参数绑定失败: %v", err)
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 转换任务ID
	id, err := strconv.ParseUint(taskID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户信息缺失")
		return
	}

	// 执行取消任务
	err = h.taskService.CancelTask(c.Request.Context(), uint(id), userID.(uint), req.Reason)
	if err != nil {
		logger.Errorf("取消任务失败: %v", err)
		response.InternalError(c, "取消任务失败")
		return
	}

	response.Success(c, gin.H{"message": "任务已取消"})
}

// AutoAssignTask 自动分配任务
func (h *TaskHandler) AutoAssignTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		response.BadRequest(c, "任务ID不能为空")
		return
	}

	var req service.AutoAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("自动分配任务请求参数绑定失败: %v", err)
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 转换任务ID
	id, err := strconv.ParseUint(taskID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	// 执行自动分配
	assignment, err := h.assignmentService.AutoAssign(c.Request.Context(), uint(id), req.Strategy)
	if err != nil {
		logger.Errorf("自动分配任务失败: %v", err)
		response.InternalError(c, "自动分配任务失败")
		return
	}

	response.Success(c, gin.H{
		"message":    "任务自动分配成功",
		"assignment": assignment,
	})
}

// GetAssignmentSuggestions 获取分配建议
func (h *TaskHandler) GetAssignmentSuggestions(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		response.BadRequest(c, "任务ID不能为空")
		return
	}

	// 转换任务ID
	id, err := strconv.ParseUint(taskID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	// 构建建议请求
	req := &service.AssignmentSuggestionRequest{
		TaskID:         uint(id),
		Strategy:       c.DefaultQuery("strategy", "comprehensive"),
		MaxSuggestions: 5,
	}

	// 获取分配建议
	suggestions, err := h.assignmentService.GetAssignmentSuggestions(c.Request.Context(), req)
	if err != nil {
		logger.Errorf("获取分配建议失败: %v", err)
		response.InternalError(c, "获取分配建议失败")
		return
	}

	response.Success(c, suggestions)
}
