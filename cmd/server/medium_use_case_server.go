package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"risk-decision-engine/internal/engine/flow"
	"risk-decision-engine/internal/engine/model"
	"risk-decision-engine/internal/engine/rule"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	flowEngine *flow.SimpleFlowEngine
	once       sync.Once
	mockMode   string // "low", "medium", "high"
)

func initEngine() {
	once.Do(func() {
		// 创建规则
		ageRule := rule.NewAgeRule()
		incomeRule := rule.NewIncomeRule()
		rules := []*rule.SimpleRule{ageRule, incomeRule}

		// 创建模型客户端（指向mock服务）
		modelClient := model.NewModelClient("http://localhost:8081/mock/model/M001", "test-api-key")

		// 创建决策流引擎
		flowEngine = flow.NewSimpleFlowEngine(rules, modelClient)

		fmt.Printf("✓ 规则加载成功: %s, %s\n", ageRule.Name, incomeRule.Name)
		fmt.Printf("  规则条件: %s; %s\n", ageRule.Expression, incomeRule.Expression)
	})
}

func setMockMode(c *gin.Context) {
	mode := c.Param("mode")
	if mode != "low" && mode != "medium" && mode != "high" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "1001",
			"message": "无效的mock模式，支持: low, medium, high",
		})
		return
	}
	mockMode = mode
	c.JSON(http.StatusOK, gin.H{
		"code":    "0000",
		"message": "success",
		"data": gin.H{
			"mockMode": mockMode,
		},
	})
}

func mockModelService(c *gin.Context) {
	var req model.ModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "400",
			"message": "参数错误",
		})
		return
	}

	fmt.Printf("[Mock模型] 请求: contractId=%s, mockMode=%s\n", req.ContractID, mockMode)

	var resp gin.H
	switch mockMode {
	case "low":
		resp = gin.H{
			"code":    0,
			"message": "success",
			"data": gin.H{
				"default_probability": 0.03,
				"risk_score":          750,
				"risk_level":          "LOW",
			},
		}
	case "medium":
		resp = gin.H{
			"code":    0,
			"message": "success",
			"data": gin.H{
				"default_probability": 0.15,
				"risk_score":          650,
				"risk_level":          "MEDIUM",
			},
		}
	case "high":
		resp = gin.H{
			"code":    0,
			"message": "success",
			"data": gin.H{
				"default_probability": 0.35,
				"risk_score":          550,
				"risk_level":          "HIGH",
			},
		}
	default:
		resp = gin.H{
			"code":    0,
			"message": "success",
			"data": gin.H{
				"default_probability": 0.05,
				"risk_score":          720,
				"risk_level":          "LOW",
			},
		}
	}

	respJSON, _ := json.Marshal(resp)
	fmt.Printf("[Mock模型] 响应: %s\n", string(respJSON))

	c.JSON(http.StatusOK, resp)
}

func executeDecision(c *gin.Context) {
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
	fmt.Printf("\n[决策请求] requestId=%s, decisionId=%s\n", req.RequestID, decisionID)

	// 执行决策流
	ctx := &flow.DecisionContext{
		Input: req.Data,
	}

	if err := flowEngine.Execute(ctx); err != nil {
		fmt.Printf("[决策错误] %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "4001",
			"message": "决策执行失败",
		})
		return
	}

	fmt.Printf("[决策结果] decisionId=%s, decision=%s, code=%s, reason=%s\n",
		decisionID, ctx.Decision, ctx.DecisionCode, ctx.DecisionReason)

	c.JSON(http.StatusOK, gin.H{
		"code":      "0000",
		"message":   "成功",
		"requestId": req.RequestID,
		"data": gin.H{
			"decisionId":     decisionID,
			"businessId":     req.BusinessID,
			"decision":       ctx.Decision,
			"decisionCode":   ctx.DecisionCode,
			"decisionReason": ctx.DecisionReason,
		},
	})
}

func main() {
	fmt.Println("========================================")
	fmt.Println("   风险决策引擎 - 中等用例服务")
	fmt.Println("   规则 + 模型联合决策")
	fmt.Println("========================================")
	fmt.Println()

	initEngine()
	mockMode = "low"

	fmt.Println()
	fmt.Println("服务说明:")
	fmt.Println("  - 决策服务端口: 8080")
	fmt.Println("  - Mock模型服务端口: 8081")
	fmt.Println()
	fmt.Println("决策服务接口:")
	fmt.Println("  - 健康检查: http://192.168.1.27:8080/health")
	fmt.Println("  - 决策接口: http://192.168.1.27:8080/api/v1/decision/execute")
	fmt.Println("  - 设置Mock模式: http://192.168.1.27:8080/mock/mode/:mode (low/medium/high)")
	fmt.Println()
	fmt.Println("Mock模型服务接口:")
	fmt.Println("  - 模型预测: http://192.168.1.27:8081/mock/model/M001")
	fmt.Println()
	fmt.Println("测试用例:")
	fmt.Println("  1. 规则通过 + 模型低分 -> 通过 (设置mock mode=low)")
	fmt.Println("  2. 规则通过 + 模型中分 -> 人工复核 (设置mock mode=medium)")
	fmt.Println("  3. 规则通过 + 模型高分 -> 拒绝 (设置mock mode=high)")
	fmt.Println("  4. 年龄规则拒绝 (年龄<21或>60)")
	fmt.Println("  5. 收入规则拒绝 (月收入<5000)")
	fmt.Println()
	fmt.Println("按 Ctrl+C 停止服务")
	fmt.Println("========================================")
	fmt.Println()

	gin.SetMode(gin.ReleaseMode)

	// 启动Mock模型服务 (8081端口)
	go func() {
		mockRouter := gin.Default()
		mockRouter.POST("/mock/model/M001", mockModelService)
		fmt.Println("Mock模型服务启动: 0.0.0.0:8081")
		if err := mockRouter.Run("0.0.0.0:8081"); err != nil {
			fmt.Printf("Mock模型服务启动失败: %v\n", err)
		}
	}()

	// 启动决策服务 (8080端口)
	mainRouter := gin.Default()

	mainRouter.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "risk-decision-engine-medium",
			"version": "0.4.0",
		})
	})

	mainRouter.POST("/api/v1/decision/execute", executeDecision)
	mainRouter.POST("/mock/mode/:mode", setMockMode)

	fmt.Println("决策服务启动: 0.0.0.0:8080")
	if err := mainRouter.Run("0.0.0.0:8080"); err != nil {
		fmt.Printf("决策服务启动失败: %v\n", err)
	}
}
