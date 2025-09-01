package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"taskmanage/internal/service"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	projectService service.ProjectService
	logger         *logrus.Logger
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(projectService service.ProjectService, logger *logrus.Logger) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		logger:         logger,
	}
}

// CreateProject 创建项目
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req service.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定创建项目请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"name":          req.Name,
		"department_id": req.DepartmentID,
		"manager_id":    req.ManagerID,
	}).Info("处理创建项目请求")

	project, err := h.projectService.CreateProject(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("创建项目失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建项目失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "项目创建成功",
		"data":    project,
	})
}

// UpdateProject 更新项目
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析项目ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	var req service.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定更新项目请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"id":   id,
		"name": req.Name,
	}).Info("处理更新项目请求")

	project, err := h.projectService.UpdateProject(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.WithError(err).Error("更新项目失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新项目失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "项目更新成功",
		"data":    project,
	})
}

// DeleteProject 删除项目
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析项目ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	h.logger.WithField("id", id).Info("处理删除项目请求")

	if err := h.projectService.DeleteProject(c.Request.Context(), uint(id)); err != nil {
		h.logger.WithError(err).Error("删除项目失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除项目失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "项目删除成功"})
}

// GetProject 获取项目详情
func (h *ProjectHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析项目ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	h.logger.WithField("id", id).Debug("处理获取项目详情请求")

	project, err := h.projectService.GetProject(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).Error("获取项目详情失败")
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": project})
}

// ListProjects 获取项目列表
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	var req service.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Error("绑定项目列表请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"page":      req.Page,
		"page_size": req.PageSize,
	}).Debug("处理获取项目列表请求")

	response, err := h.projectService.ListProjects(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("获取项目列表失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取项目列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProjectsByDepartment 根据部门获取项目
func (h *ProjectHandler) GetProjectsByDepartment(c *gin.Context) {
	idStr := c.Param("department_id")
	departmentID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("department_id", idStr).Error("解析部门ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的部门ID"})
		return
	}

	h.logger.WithField("department_id", departmentID).Debug("处理根据部门获取项目请求")

	projects, err := h.projectService.GetProjectsByDepartment(c.Request.Context(), uint(departmentID))
	if err != nil {
		h.logger.WithError(err).Error("根据部门获取项目失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取项目失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": projects})
}

// GetProjectsByManager 根据管理者获取项目
func (h *ProjectHandler) GetProjectsByManager(c *gin.Context) {
	idStr := c.Param("manager_id")
	managerID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("manager_id", idStr).Error("解析管理者ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的管理者ID"})
		return
	}

	h.logger.WithField("manager_id", managerID).Debug("处理根据管理者获取项目请求")

	projects, err := h.projectService.GetProjectsByManager(c.Request.Context(), uint(managerID))
	if err != nil {
		h.logger.WithError(err).Error("根据管理者获取项目失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取项目失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": projects})
}

// GetProjectsByStatus 根据状态获取项目
func (h *ProjectHandler) GetProjectsByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		h.logger.Error("项目状态参数为空")
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目状态不能为空"})
		return
	}

	h.logger.WithField("status", status).Debug("处理根据状态获取项目请求")

	projects, err := h.projectService.GetProjectsByStatus(c.Request.Context(), status)
	if err != nil {
		h.logger.WithError(err).Error("根据状态获取项目失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取项目失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": projects})
}

// AddProjectMember 添加项目成员
func (h *ProjectHandler) AddProjectMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析项目ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	var req service.AddProjectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定添加项目成员请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"project_id":  id,
		"employee_id": req.EmployeeID,
	}).Info("处理添加项目成员请求")

	if err := h.projectService.AddProjectMember(c.Request.Context(), uint(id), &req); err != nil {
		h.logger.WithError(err).Error("添加项目成员失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加项目成员失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "项目成员添加成功"})
}

// RemoveProjectMember 移除项目成员
func (h *ProjectHandler) RemoveProjectMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析项目ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	var req service.RemoveProjectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定移除项目成员请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"project_id":  id,
		"employee_id": req.EmployeeID,
	}).Info("处理移除项目成员请求")

	if err := h.projectService.RemoveProjectMember(c.Request.Context(), uint(id), &req); err != nil {
		h.logger.WithError(err).Error("移除项目成员失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除项目成员失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "项目成员移除成功"})
}

// GetProjectMembers 获取项目成员
func (h *ProjectHandler) GetProjectMembers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析项目ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	h.logger.WithField("id", id).Debug("处理获取项目成员请求")

	members, err := h.projectService.GetProjectMembers(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).Error("获取项目成员失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取项目成员失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": members})
}

// UpdateProjectManager 更新项目管理者
func (h *ProjectHandler) UpdateProjectManager(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析项目ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
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
		"project_id": id,
		"manager_id": req.ManagerID,
	}).Info("处理更新项目管理者请求")

	if err := h.projectService.UpdateProjectManager(c.Request.Context(), uint(id), req.ManagerID); err != nil {
		h.logger.WithError(err).Error("更新项目管理者失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新项目管理者失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "项目管理者更新成功"})
}
