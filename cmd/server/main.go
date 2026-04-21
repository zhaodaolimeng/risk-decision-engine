package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("=== 风险决策引擎服务 ===")

	// 获取本机IP
	ips := getLocalIPs()
	fmt.Println("\n本机IP地址:")
	for _, ip := range ips {
		fmt.Printf("  - http://%s:8080\n", ip)
	}

	// 创建 Gin 引擎
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "risk-decision-engine",
		})
	})

	// 简单的决策接口（测试用）
	r.POST("/api/v1/decision/execute", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "1001",
				"message": "参数错误",
			})
			return
		}

		// 简单的响应
		c.JSON(http.StatusOK, gin.H{
			"code":    "0000",
			"message": "成功",
			"data": gin.H{
				"decisionId": "dec_" + randomString(8),
				"decision":   "APPROVE",
				"reason":     "测试通过",
			},
		})
	})

	fmt.Println("\n服务启动中，监听 0.0.0.0:8080")
	fmt.Println("按 Ctrl+C 停止服务")

	// 启动服务
	if err := r.Run("0.0.0.0:8080"); err != nil {
		fmt.Printf("启动失败: %v\n", err)
	}
}

func getLocalIPs() []string {
	var ips []string

	ifaces, err := net.Interfaces()
	if err != nil {
		return []string{"127.0.0.1"}
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}

	if len(ips) == 0 {
		ips = append(ips, "127.0.0.1")
	}

	return ips
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
