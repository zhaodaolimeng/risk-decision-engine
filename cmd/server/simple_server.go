package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("   风险决策引擎服务")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("本机访问地址:")
	fmt.Println("  - 健康检查: http://192.168.1.27:8080/health")
	fmt.Println("  - 决策接口: http://192.168.1.27:8080/api/v1/decision/execute")
	fmt.Println()
	fmt.Println("局域网访问地址:")
	fmt.Println("  - 健康检查: http://192.168.1.27:8080/health")
	fmt.Println("  - 决策接口: http://192.168.1.27:8080/api/v1/decision/execute")
	fmt.Println()
	fmt.Println("服务监听: 0.0.0.0:8080")
	fmt.Println("按 Ctrl+C 停止服务")
	fmt.Println("========================================")
	fmt.Println()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "risk-decision-engine",
			"version": "0.1.0",
		})
	})

	// 简单的决策接口
	r.POST("/api/v1/decision/execute", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "1001",
				"message": "参数错误",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":      "0000",
			"message":   "成功",
			"requestId": req["requestId"],
			"timestamp": getTimestamp(),
			"data": gin.H{
				"decisionId":     "dec_" + randString(8),
				"businessId":     req["businessId"],
				"decision":       "APPROVE",
				"decisionCode":   "APPROVE_001",
				"decisionReason": "测试通过",
				"riskScore":      750,
				"riskLevel":      "LOW",
			},
		})
	})

	if err := r.Run("0.0.0.0:8080"); err != nil {
		fmt.Printf("启动失败: %v\n", err)
	}
}

func getTimestamp() int64 {
	return 1713600000000
}

func randString(length int) string {
	return "abc123xyz"
}
