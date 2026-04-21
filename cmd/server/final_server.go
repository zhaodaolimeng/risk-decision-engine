package main

import (
	"fmt"
	"net/http"
	"sync"

	"risk-decision-engine/internal/engine/rule"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	ageRule *rule.SimpleRule
	once    sync.Once
)

func initRule() {
	once.Do(func() {
		ageRule = rule.NewAgeRule()
		fmt.Printf("✓ 规则加载成功: %s\n", ageRule.Name)
		fmt.Printf("  条件: %s\n", ageRule.Expression)
	})
}

func main() {
	fmt.Println("========================================")
	fmt.Println("   风险决策引擎服务")
	fmt.Println("========================================")
	fmt.Println()

	initRule()

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
			"version": "1.0.0",
		})
	})

	// 决策接口
	r.POST("/api/v1/decision/execute", func(c *gin.Context) {
		var req struct {
			RequestID  string                 `json:"requestId"`
			BusinessID string                 `json:"businessId"`
			Data       map[string]interface{} `json:"data"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "1001",
				"message": "参数错误",
			})
			return
		}

		decisionID := "dec_" + uuid.NewString()[:8]
		fmt.Printf("[决策请求] requestId=%s\n", req.RequestID)

		// 提取年龄
		age := extractAge(req.Data)
		fact := map[string]interface{}{
			"age": age,
		}

		// 执行规则
		decision := "APPROVE"
		reason := "通过"
		decisionCode := "APPROVE_001"

		ruleResult, err := ageRule.Execute(fact)
		if err != nil {
			fmt.Printf("[规则错误] %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "4001",
				"message": "规则执行失败",
			})
			return
		}

		if ruleResult.Result == "REJECT" {
			decision = "REJECT"
			decisionCode = "REJECT_R001"
			reason = ruleResult.Reason
		} else {
			reason = "年龄符合要求"
		}

		fmt.Printf("[决策结果] decisionId=%s, decision=%s, reason=%s\n",
			decisionID, decision, reason)

		c.JSON(http.StatusOK, gin.H{
			"code":      "0000",
			"message":   "成功",
			"requestId": req.RequestID,
			"data": gin.H{
				"decisionId":     decisionID,
				"businessId":     req.BusinessID,
				"decision":       decision,
				"decisionCode":   decisionCode,
				"decisionReason": reason,
			},
		})
	})

	if err := r.Run("0.0.0.0:8080"); err != nil {
		fmt.Printf("启动失败: %v\n", err)
	}
}

func extractAge(data map[string]interface{}) int {
	if data == nil {
		return 0
	}

	if applicant, ok := data["applicant"].(map[string]interface{}); ok {
		if age, ok := applicant["age"].(float64); ok {
			return int(age)
		}
		if age, ok := applicant["age"].(int); ok {
			return age
		}
	}

	if age, ok := data["age"].(float64); ok {
		return int(age)
	}
	if age, ok := data["age"].(int); ok {
		return age
	}

	return 0
}
