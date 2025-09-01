package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"taskmanage/internal/container"
	"taskmanage/internal/service"
	"taskmanage/pkg/response"
)

// EmployeeHandler 员工处理器
type EmployeeHandler struct {
	*BaseHandler
}

func NewEmployeeHandler(container *container.ApplicationContainer, logger *logrus.Logger) *EmployeeHandler {
	return &EmployeeHandler{
		BaseHandler: &BaseHandler{container: container, logger: logger},
	}
}

// ListEmployees 获取员工列表
func (h *EmployeeHandler) ListEmployees(c *gin.Context) {
	// 解析查询参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")
	department := c.Query("department")
	position := c.Query("position")
	status := c.Query("status")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(sizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	h.logger.WithFields(logrus.Fields{
		"page":       page,
		"page_size":  pageSize,
		"department": department,
		"position":   position,
		"status":     status,
	}).Info("Listing employees")

	// 构建过滤器
	filter := service.EmployeeListFilter{
		Page:       page,
		PageSize:   pageSize,
		Department: department,
		Position:   position,
		Status:     status,
	}

	// 调用服务层
	employeeService := h.container.GetServiceManager().EmployeeService()
	employees, total, err := employeeService.ListEmployees(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list employees")
		response.InternalError(c, "获取员工列表失败")
		return
	}

	response.SuccessWithPagination(c, employees, page, pageSize, total)
}

// CreateEmployee 创建员工
func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	var req service.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid create employee request")
		response.BadRequest(c, "请求参数无效")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"name":          req.Name,
		"email":         req.Email,
		"department_id": req.DepartmentID,
		"position_id":   req.PositionID,
	}).Info("Creating employee")

	// 调用服务层
	employeeService := h.container.GetServiceManager().EmployeeService()
	employee, err := employeeService.CreateEmployee(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create employee")
		response.InternalError(c, "创建员工失败")
		return
	}

	response.Success(c, employee)
}

// GetEmployee 获取员工详情
func (h *EmployeeHandler) GetEmployee(c *gin.Context) {
	employeeID := c.Param("id")
	if employeeID == "" {
		response.BadRequest(c, "员工ID不能为空")
		return
	}

	// 转换员工ID
	id, err := strconv.ParseUint(employeeID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	h.logger.WithField("employee_id", id).Info("Getting employee")

	// 调用服务层
	employeeService := h.container.GetServiceManager().EmployeeService()
	employee, err := employeeService.GetEmployee(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).Error("Failed to get employee")
		response.NotFound(c, "员工不存在")
		return
	}

	response.Success(c, employee)
}

// UpdateEmployee 更新员工信息
func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	employeeID := c.Param("id")
	if employeeID == "" {
		response.BadRequest(c, "员工ID不能为空")
		return
	}

	// 转换员工ID
	id, err := strconv.ParseUint(employeeID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	var req service.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid update employee request")
		response.BadRequest(c, "请求参数无效")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"employee_id":   id,
		"name":          req.Name,
		"department_id": req.DepartmentID,
		"position_id":   req.PositionID,
	}).Info("Updating employee")

	// 调用服务层
	employeeService := h.container.GetServiceManager().EmployeeService()
	employee, err := employeeService.UpdateEmployee(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update employee")
		response.InternalError(c, "更新员工失败")
		return
	}

	response.Success(c, employee)
}

// DeleteEmployee 删除员工
func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	employeeID := c.Param("id")
	if employeeID == "" {
		response.BadRequest(c, "员工ID不能为空")
		return
	}

	// 转换员工ID
	id, err := strconv.ParseUint(employeeID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	h.logger.WithField("employee_id", id).Info("Deleting employee")

	// 调用服务层
	employeeService := h.container.GetServiceManager().EmployeeService()
	err = employeeService.DeleteEmployee(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete employee")
		response.InternalError(c, "删除员工失败")
		return
	}

	response.Success(c, gin.H{"message": "员工删除成功"})
}

// GetAvailableEmployees 获取可用员工列表
func (h *EmployeeHandler) GetAvailableEmployees(c *gin.Context) {
	// 获取查询参数
	skills := c.QueryArray("skills")
	maxLoad := c.Query("max_load")
	department := c.Query("department")

	h.logger.WithFields(logrus.Fields{
		"skills":     skills,
		"max_load":   maxLoad,
		"department": department,
	}).Info("Getting available employees")

	// 调用服务层
	employeeService := h.container.GetServiceManager().EmployeeService()
	employees, err := employeeService.GetAvailableEmployees(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get available employees")
		response.InternalError(c, "获取可用员工失败")
		return
	}

	response.Success(c, employees)
}

// GetEmployeeWorkload 获取员工工作负载
func (h *EmployeeHandler) GetEmployeeWorkload(c *gin.Context) {
	employeeID := c.Param("id")
	if employeeID == "" {
		response.BadRequest(c, "员工ID不能为空")
		return
	}

	// 转换员工ID
	id, err := strconv.ParseUint(employeeID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	h.logger.WithField("employee_id", id).Info("Getting employee workload")

	// 调用服务层
	employeeService := h.container.GetServiceManager().EmployeeService()
	workload, err := employeeService.GetEmployeeWorkload(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).Error("Failed to get employee workload")
		response.NotFound(c, "员工不存在")
		return
	}

	response.Success(c, workload)
}

// AddSkill 添加员工技能
func (h *EmployeeHandler) AddSkill(c *gin.Context) {
	employeeID := c.Param("id")
	if employeeID == "" {
		response.BadRequest(c, "员工ID不能为空")
		return
	}

	// 转换员工ID
	id, err := strconv.ParseUint(employeeID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	var req service.AddSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid add skill request")
		response.BadRequest(c, "请求参数无效")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"employee_id": id,
		"skill_name":  req.Name,
		"skill_level": req.Level,
	}).Info("Adding skill to employee")

	// 首先根据技能名称查找技能ID
	// 这里需要实现技能查找逻辑，暂时返回错误提示
	response.BadRequest(c, "技能管理功能尚未完全实现，请先实现技能查找功能")
}

// RemoveSkill 移除员工技能
func (h *EmployeeHandler) RemoveSkill(c *gin.Context) {
	employeeID := c.Param("id")
	if employeeID == "" {
		response.BadRequest(c, "员工ID不能为空")
		return
	}

	// 转换员工ID
	id, err := strconv.ParseUint(employeeID, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	var req service.RemoveSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid remove skill request")
		response.BadRequest(c, "请求参数无效")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"employee_id": id,
		"skill_id":    req.SkillID,
	}).Info("Removing skill from employee")

	// 调用服务删除技能
	employeeService := h.container.GetEmployeeService()
	err = employeeService.RemoveSkill(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to remove skill from employee")
		response.InternalError(c, "删除员工技能失败")
		return
	}

	response.Success(c, gin.H{"message": "员工技能删除成功"})
}

// UpdateEmployeeStatus 更新员工状态
func (h *EmployeeHandler) UpdateEmployeeStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.WithError(err).Warn("Invalid employee ID")
		response.BadRequest(c, "无效的员工ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid update status request")
		response.BadRequest(c, "请求参数无效")
		return
	}

	employeeService := h.container.GetEmployeeService()
	err = employeeService.UpdateEmployeeStatus(c.Request.Context(), uint(id), req.Status)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update employee status")
		response.InternalError(c, "更新员工状态失败")
		return
	}

	response.Success(c, gin.H{"message": "员工状态更新成功"})
}

// GetEmployeesByStatus 根据状态获取员工列表
func (h *EmployeeHandler) GetEmployeesByStatus(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		response.BadRequest(c, "状态参数不能为空")
		return
	}

	employeeService := h.container.GetEmployeeService()
	employees, err := employeeService.GetEmployeesByStatus(c.Request.Context(), status)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get employees by status")
		response.InternalError(c, "获取员工列表失败")
		return
	}

	response.Success(c, employees)
}

// GetWorkloadStats 获取工作负载统计
func (h *EmployeeHandler) GetWorkloadStats(c *gin.Context) {
	var req service.WorkloadStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid workload stats request")
		response.BadRequest(c, "请求参数无效")
		return
	}

	employeeService := h.container.GetEmployeeService()
	stats, err := employeeService.GetWorkloadStats(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get workload stats")
		response.InternalError(c, "获取工作负载统计失败")
		return
	}

	response.Success(c, stats)
}

// GetDepartmentWorkload 获取部门工作负载统计
func (h *EmployeeHandler) GetDepartmentWorkload(c *gin.Context) {
	departmentIDStr := c.Param("department")
	if departmentIDStr == "" {
		response.BadRequest(c, "部门ID参数不能为空")
		return
	}

	// 将字符串转换为uint
	departmentID, err := strconv.ParseUint(departmentIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的部门ID")
		return
	}

	employeeService := h.container.GetEmployeeService()
	workload, err := employeeService.GetDepartmentWorkload(c.Request.Context(), uint(departmentID))
	if err != nil {
		h.logger.WithError(err).Error("Failed to get department workload")
		response.InternalError(c, "获取部门工作负载统计失败")
		return
	}

	response.Success(c, workload)
}
