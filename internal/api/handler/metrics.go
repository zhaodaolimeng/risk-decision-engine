package handler

import (
	"risk-decision-engine/internal/metrics"

	"github.com/gin-gonic/gin"
)

// MetricsHandler 指标处理器
type MetricsHandler struct{}

// NewMetricsHandler 创建指标处理器
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// RegisterRoutes 注册路由
func (h *MetricsHandler) RegisterRoutes(r *gin.Engine) {
	metricsGroup := r.Group("/api/v1/metrics")
	{
		metricsGroup.GET("", h.GetMetrics)
		metricsGroup.GET("/health", h.Health)
		metricsGroup.POST("/reset", h.ResetMetrics)
	}
}

// GetMetrics 获取指标
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	stats := metrics.Get().GetStats()

	c.JSON(200, gin.H{
		"code": "0000",
		"message": "success",
		"data": stats,
	})
}

// Health 健康检查（带指标）
func (h *MetricsHandler) Health(c *gin.Context) {
	stats := metrics.Get().GetStats()

	c.JSON(200, gin.H{
		"status": "ok",
		"decisionTotal": stats.DecisionTotal,
		"cacheHitRate": stats.CacheHitRate,
		"approveRate": stats.ApproveRate,
	})
}

// ResetMetrics 重置指标
func (h *MetricsHandler) ResetMetrics(c *gin.Context) {
	metrics.Get().Reset()

	c.JSON(200, gin.H{
		"code": "0000",
		"message": "metrics reset successfully",
	})
}
