package middleware

import (
	"time"

	"risk-decision-engine/internal/metrics"

	"github.com/gin-gonic/gin"
)

// MetricsMiddleware 指标采集中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		// 继续处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start)
		hasError := len(c.Errors) > 0

		metrics.Get().RecordAPI(path, duration, hasError)
	}
}
