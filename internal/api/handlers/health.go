package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	
	"taskmanage/internal/container"
	"taskmanage/pkg/response"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	container *container.ApplicationContainer
	logger    *logrus.Logger
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(container *container.ApplicationContainer, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		container: container,
		logger:    logger,
	}
}

// HealthCheck 健康检查
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := gin.H{
		"status":    "healthy",
		"timestamp": "2025-08-28T17:00:00Z",
		"version":   "v1.0.0",
		"services":  gin.H{},
	}
	
	// 检查数据库连接
	if db := h.container.GetDB(); db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			if err := sqlDB.Ping(); err == nil {
				status["services"].(gin.H)["database"] = "healthy"
			} else {
				status["services"].(gin.H)["database"] = "unhealthy"
				status["status"] = "degraded"
			}
		} else {
			status["services"].(gin.H)["database"] = "unhealthy"
			status["status"] = "degraded"
		}
	} else {
		status["services"].(gin.H)["database"] = "unavailable"
		status["status"] = "unhealthy"
	}
	
	// 检查缓存连接 (暂时跳过，因为缓存管理器未实现)
	status["services"].(gin.H)["cache"] = "not_implemented"
	
	// 根据整体状态返回相应的HTTP状态码
	httpStatus := http.StatusOK
	if status["status"] == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	} else if status["status"] == "degraded" {
		httpStatus = http.StatusPartialContent
	}
	
	c.JSON(httpStatus, status)
}

// ReadinessCheck 就绪检查
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ready := true
	checks := gin.H{}
	
	// 检查数据库
	if db := h.container.GetDB(); db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			if err := sqlDB.Ping(); err == nil {
				checks["database"] = "ready"
			} else {
				checks["database"] = "not_ready"
				ready = false
			}
		} else {
			checks["database"] = "not_ready"
			ready = false
		}
	} else {
		checks["database"] = "not_available"
		ready = false
	}
	
	// 检查Repository管理器
	if repoManager := h.container.GetRepositoryManager(); repoManager != nil {
		checks["repository"] = "ready"
	} else {
		checks["repository"] = "not_ready"
		ready = false
	}
	
	status := gin.H{
		"ready":  ready,
		"checks": checks,
	}
	
	if ready {
		c.JSON(http.StatusOK, status)
	} else {
		c.JSON(http.StatusServiceUnavailable, status)
	}
}

// LivenessCheck 存活检查
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"alive":     true,
		"timestamp": "2025-08-28T17:00:00Z",
		"uptime":    "running",
	})
}
