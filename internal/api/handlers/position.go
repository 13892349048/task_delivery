package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"taskmanage/internal/service"
)

// PositionHandler 职位处理器
type PositionHandler struct {
	positionService service.PositionService
	logger          *logrus.Logger
}

// NewPositionHandler 创建职位处理器
func NewPositionHandler(positionService service.PositionService, logger *logrus.Logger) *PositionHandler {
	return &PositionHandler{
		positionService: positionService,
		logger:          logger,
	}
}

// CreatePosition 创建职位
func (h *PositionHandler) CreatePosition(c *gin.Context) {
	var req service.CreatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定创建职位请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"name":     req.Name,
		"category": req.Category,
		"level":    req.Level,
	}).Info("处理创建职位请求")

	position, err := h.positionService.CreatePosition(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("创建职位失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建职位失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "职位创建成功",
		"data":    position,
	})
}

// UpdatePosition 更新职位
func (h *PositionHandler) UpdatePosition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析职位ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的职位ID"})
		return
	}

	var req service.UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("绑定更新职位请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"id":   id,
		"name": req.Name,
	}).Info("处理更新职位请求")

	position, err := h.positionService.UpdatePosition(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.WithError(err).Error("更新职位失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新职位失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "职位更新成功",
		"data":    position,
	})
}

// DeletePosition 删除职位
func (h *PositionHandler) DeletePosition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析职位ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的职位ID"})
		return
	}

	h.logger.WithField("id", id).Info("处理删除职位请求")

	if err := h.positionService.DeletePosition(c.Request.Context(), uint(id)); err != nil {
		h.logger.WithError(err).Error("删除职位失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除职位失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "职位删除成功"})
}

// GetPosition 获取职位详情
func (h *PositionHandler) GetPosition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("解析职位ID失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的职位ID"})
		return
	}

	h.logger.WithField("id", id).Debug("处理获取职位详情请求")

	position, err := h.positionService.GetPosition(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.WithError(err).Error("获取职位详情失败")
		c.JSON(http.StatusNotFound, gin.H{"error": "职位不存在", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": position})
}

// ListPositions 获取职位列表
func (h *PositionHandler) ListPositions(c *gin.Context) {
	var req service.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Error("绑定职位列表请求失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效", "details": err.Error()})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"page":      req.Page,
		"page_size": req.PageSize,
	}).Debug("处理获取职位列表请求")

	response, err := h.positionService.ListPositions(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("获取职位列表失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取职位列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPositionsByCategory 根据类别获取职位
func (h *PositionHandler) GetPositionsByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		h.logger.Error("职位类别参数为空")
		c.JSON(http.StatusBadRequest, gin.H{"error": "职位类别不能为空"})
		return
	}

	h.logger.WithField("category", category).Debug("处理根据类别获取职位请求")

	positions, err := h.positionService.GetPositionsByCategory(c.Request.Context(), category)
	if err != nil {
		h.logger.WithError(err).Error("根据类别获取职位失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取职位失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": positions})
}

// GetPositionsByLevel 根据级别获取职位
func (h *PositionHandler) GetPositionsByLevel(c *gin.Context) {
	levelStr := c.Param("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil {
		h.logger.WithError(err).WithField("level", levelStr).Error("解析职位级别失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的职位级别"})
		return
	}

	h.logger.WithField("level", level).Debug("处理根据级别获取职位请求")

	positions, err := h.positionService.GetPositionsByLevel(c.Request.Context(), level)
	if err != nil {
		h.logger.WithError(err).Error("根据级别获取职位失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取职位失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": positions})
}

// GetPositionCategories 获取所有职位类别
func (h *PositionHandler) GetPositionCategories(c *gin.Context) {
	h.logger.Debug("处理获取职位类别请求")

	categories, err := h.positionService.GetAllCategories(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("获取职位类别失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取职位类别失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}
