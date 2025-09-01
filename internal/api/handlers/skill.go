package handlers

import (
	"strconv"

	"taskmanage/internal/container"
	"taskmanage/internal/service"
	"taskmanage/pkg/logger"
	"taskmanage/pkg/response"

	"github.com/gin-gonic/gin"
)

// SkillHandler 技能处理器
type SkillHandler struct {
	container *container.ApplicationContainer
}

// NewSkillHandler 创建技能处理器
func NewSkillHandler(c *container.ApplicationContainer) *SkillHandler {
	return &SkillHandler{
		container: c,
	}
}

// CreateSkill 创建技能
func (h *SkillHandler) CreateSkill(c *gin.Context) {
	var req service.CreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("Invalid create skill request: %v", err)
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	skillService := h.container.GetSkillService()
	skill, err := skillService.CreateSkill(c.Request.Context(), &req)
	if err != nil {
		logger.Errorf("Failed to create skill: %v", err)
		response.InternalError(c, "Failed to create skill")
		return
	}

	response.Success(c, skill)
}

// GetSkill 获取技能详情
func (h *SkillHandler) GetSkill(c *gin.Context) {
	skillIDStr := c.Param("id")
	skillID, err := strconv.ParseUint(skillIDStr, 10, 32)
	if err != nil {
		logger.Warnf("Invalid skill ID: %s", skillIDStr)
		response.BadRequest(c, "Invalid skill ID")
		return
	}

	skillService := h.container.GetSkillService()
	skill, err := skillService.GetSkill(c.Request.Context(), uint(skillID))
	if err != nil {
		logger.Errorf("Failed to get skill: %v", err)
		response.NotFound(c, "Skill not found")
		return
	}

	response.Success(c, skill)
}

// UpdateSkill 更新技能
func (h *SkillHandler) UpdateSkill(c *gin.Context) {
	skillIDStr := c.Param("id")
	skillID, err := strconv.ParseUint(skillIDStr, 10, 32)
	if err != nil {
		logger.Warnf("Invalid skill ID: %s", skillIDStr)
		response.BadRequest(c, "Invalid skill ID")
		return
	}

	var req service.UpdateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("Invalid update skill request: %v", err)
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	skillService := h.container.GetSkillService()
	skill, err := skillService.UpdateSkill(c.Request.Context(), uint(skillID), &req)
	if err != nil {
		logger.Errorf("Failed to update skill: %v", err)
		response.InternalError(c, "Failed to update skill")
		return
	}

	response.Success(c, skill)
}

// DeleteSkill 删除技能
func (h *SkillHandler) DeleteSkill(c *gin.Context) {
	skillIDStr := c.Param("id")
	skillID, err := strconv.ParseUint(skillIDStr, 10, 32)
	if err != nil {
		logger.Warnf("Invalid skill ID: %s", skillIDStr)
		response.BadRequest(c, "Invalid skill ID")
		return
	}

	skillService := h.container.GetSkillService()
	err = skillService.DeleteSkill(c.Request.Context(), uint(skillID))
	if err != nil {
		logger.Errorf("Failed to delete skill: %v", err)
		response.InternalError(c, "Failed to delete skill")
		return
	}

	response.Success(c, skillID)
}

// ListSkills 获取技能列表
func (h *SkillHandler) ListSkills(c *gin.Context) {
	var filter service.SkillListFilter

	// 解析分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if filter.Page == 0 {
		filter.Page = 1
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			filter.PageSize = pageSize
		}
	}
	if filter.PageSize == 0 {
		filter.PageSize = 20
	}

	// 解析分类过滤
	filter.Category = c.Query("category")

	skillService := h.container.GetSkillService()
	
	// Convert filter to ListSkillsRequest
	req := &service.ListSkillsRequest{
		ListRequest: service.ListRequest{
			Page:     filter.Page,
			PageSize: filter.PageSize,
		},
		Category: filter.Category,
	}
	
	result, err := skillService.ListSkills(c.Request.Context(), req)
	if err != nil {
		logger.Errorf("Failed to list skills: %v", err)
		response.InternalError(c, "Failed to list skills")
		return
	}

	response.Success(c, gin.H{
		"skills":    result.Items,
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.Size,
	})
}

// GetSkillCategories 获取技能分类列表
func (h *SkillHandler) GetSkillCategories(c *gin.Context) {
	skillService := h.container.GetSkillService()
	categories, err := skillService.GetAllCategories(c.Request.Context())
	if err != nil {
		logger.Errorf("Failed to get skill categories: %v", err)
		response.InternalError(c, "Failed to get skill categories")
		return
	}

	response.Success(c, gin.H{
		"categories": categories,
	})
}

// AssignSkillToEmployee 为员工分配技能
func (h *SkillHandler) AssignSkillToEmployee(c *gin.Context) {
	var req service.AssignSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("Invalid assign skill request: %v", err)
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	skillService := h.container.GetSkillService()
	err := skillService.AssignSkillToEmployee(c.Request.Context(), req.EmployeeID, req.SkillID, req.Level)
	if err != nil {
		logger.Errorf("Failed to assign skill to employee: %v", err)
		response.InternalError(c, "Failed to assign skill")
		return
	}

	response.Success(c, gin.H{"message": "success assign"})
}

// RemoveSkillFromEmployee 移除员工技能
func (h *SkillHandler) RemoveSkillFromEmployee(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	employeeID, err := strconv.ParseUint(employeeIDStr, 10, 32)
	if err != nil {
		logger.Warnf("Invalid employee ID: %s", employeeIDStr)
		response.BadRequest(c, "Invalid employee ID")
		return
	}

	skillIDStr := c.Param("skill_id")
	skillID, err := strconv.ParseUint(skillIDStr, 10, 32)
	if err != nil {
		logger.Warnf("Invalid skill ID: %s", skillIDStr)
		response.BadRequest(c, "Invalid skill ID")
		return
	}

	skillService := h.container.GetSkillService()
	err = skillService.RemoveSkillFromEmployee(c.Request.Context(), uint(employeeID), uint(skillID))
	if err != nil {
		logger.Errorf("Failed to remove skill from employee: %v", err)
		response.InternalError(c, "Failed to remove skill")
		return
	}

	response.Success(c, skillID)
}

// GetEmployeeSkills 获取员工技能列表
func (h *SkillHandler) GetEmployeeSkills(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	employeeID, err := strconv.ParseUint(employeeIDStr, 10, 32)
	if err != nil {
		logger.Warnf("Invalid employee ID: %s", employeeIDStr)
		response.BadRequest(c, "Invalid employee ID")
		return
	}

	skillService := h.container.GetSkillService()
	skills, err := skillService.GetEmployeeSkills(c.Request.Context(), uint(employeeID))
	if err != nil {
		logger.Errorf("Failed to get employee skills: %v", err)
		response.InternalError(c, "Failed to get employee skills")
		return
	}

	response.Success(c, gin.H{
		"skills": skills,
	})
}
