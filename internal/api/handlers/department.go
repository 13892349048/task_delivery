package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"taskmanage/internal/service"
)

// DepartmentHandler 部门处理器
type DepartmentHandler struct {
	departmentService service.DepartmentService
	logger            *logrus.Logger
}

// NewDepartmentHandler 创建部门处理器
func NewDepartmentHandler(departmentService service.DepartmentService, logger *logrus.Logger) *DepartmentHandler {
	return &DepartmentHandler{
		departmentService: departmentService,
		logger:            logger,
	}
}

// CreateDepartment 创建部门
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req service.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定创建部门请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"name":      req.Name,
		"parent_id": req.ParentID,
	}).Info("处理创建部门请求")

	department, err := h.departmentService.CreateDepartment(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("创建部门失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建部门失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "部门创建成功",
		"data":    department,
	})
}

// UpdateDepartment 更新部门
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析部门ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的部门ID"})
		return
	}

	var req service.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定更新部门请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"id":   id,
		"name": req.Name,
	}).Info("处理更新部门请求")

	department, err := h.departmentService.UpdateDepartment(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.WithError(err).Error("更新部门失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新部门失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "部门更新成功",
		"data":    department,
	})
}

// DeleteDepartment 删除部门
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析部门ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的部门ID"})
		return
	}

	h.logger.WithField("id", id).Info("处理删除部门请求")

	if err := h.departmentService.DeleteDepartment(c.Request.Context(), uint(id)); err != nil {
		h.logger.WithError(err).Error("删除部门失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除部门失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "部门删除成功"})
}

// GetDepartment 获取部门详情
func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析部门ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的部门ID"})
		return
	}

	h.logger.WithField("id", id).Debug("处理获取部门详情请求")

	department, err := h.departmentService.GetDepartment(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).Error("获取部门详情失败")
		c.JSON(http.StatusNotFound, gin.H{"error": "部门不存在", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": department})
}

// ListDepartments 获取部门列表
func (h *DepartmentHandler) ListDepartments(c *gin.Context) {
	var req service.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Error("绑定部门列表请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"page":      req.Page,
		"page_size": req.PageSize,
	}).Debug("处理获取部门列表请求")

	response, err := h.departmentService.ListDepartments(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("获取部门列表失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取部门列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDepartmentTree 获取部门树结构
func (h *DepartmentHandler) GetDepartmentTree(c *gin.Context) {
	h.logger.Debug("处理获取部门树请求")

	tree, err := h.departmentService.GetDepartmentTree(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("获取部门树失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取部门树失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tree})
}

// GetRootDepartments 获取根部门
func (h *DepartmentHandler) GetRootDepartments(c *gin.Context) {
	h.logger.Debug("处理获取根部门请求")

	departments, err := h.departmentService.GetRootDepartments(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("获取根部门失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取根部门失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": departments})
}

// GetSubDepartments 获取子部门
func (h *DepartmentHandler) GetSubDepartments(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析部门ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的部门ID"})
		return
	}

	h.logger.WithField("id", id).Debug("处理获取子部门请求")

	departments, err := h.departmentService.GetSubDepartments(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).Error("获取子部门失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取子部门失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": departments})
}

// UpdateDepartmentManager 更新部门管理者
func (h *DepartmentHandler) UpdateDepartmentManager(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析部门ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的部门ID"})
		return
	}

	var req struct {
		ManagerID uint `json:"manager_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定更新管理者请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"department_id": id,
		"manager_id":    req.ManagerID,
	}).Info("处理更新部门管理者请求")

	if err := h.departmentService.UpdateManager(c.Request.Context(), uint(id), req.ManagerID); err != nil {
		h.logger.WithError(err).Error("更新部门管理者失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新部门管理者失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "部门管理者更新成功"})
}
