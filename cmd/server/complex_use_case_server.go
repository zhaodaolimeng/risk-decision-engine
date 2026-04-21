package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"risk-decision-engine/internal/engine/datasource"
	"risk-decision-engine/internal/engine/flow"
	"risk-decision-engine/internal/engine/model"
	"risk-decision-engine/internal/engine/rule"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	complexFlowEngine *flow.ComplexFlowEngine
	complexOnce       sync.Once
	complexMockMode   struct {
		userType     string // "normal", "blacklist"
		creditType   string // "good", "bad"
		multiType    string // "low", "medium", "high"
		modelRisk    string // "low", "medium", "high"
	}
)

func initComplexEngine() {
	complexOnce.Do(func() {
		// 创建Mock数据源客户端
		dsClient1 := datasource.NewDataSourceClient("http://localhost:8082/mock/datasource", "test-api-key")
		dsClient2 := datasource.NewDataSourceClient("http://localhost:8082/mock/datasource", "test-api-key")
		dsClient3 := datasource.NewDataSourceClient("http://localhost:8082/mock/datasource", "test-api-key")

		// 配置数据源
		dataSources := []*flow.DataSourceConfig{
			{ID: "DS001", Client: dsClient1, Required: true},
			{ID: "DS002", Client: dsClient2, Required: true},
			{ID: "DS003", Client: dsClient3, Required: true},
		}

		// 创建规则
		blacklistRule := rule.NewBlacklistRule()
		ageRule := rule.NewAgeRule()
		multiRule := rule.NewMultiQueryRule()
		rules := []*rule.SimpleRule{blacklistRule, ageRule, multiRule}

		// 创建模型客户端（指向mock服务）
		modelClient := model.NewModelClient("http://localhost:8082/mock/model/M001", "test-api-key")

		// 创建复杂决策流引擎
		complexFlowEngine = flow.NewComplexFlowEngine(dataSources, rules, modelClient)

		fmt.Printf("✓ 数据源加载成功: DS001, DS002, DS003\n")
		fmt.Printf("✓ 规则加载成功: %s, %s, %s\n", blacklistRule.Name, ageRule.Name, multiRule.Name)
	})
}

func setComplexMockMode(c *gin.Context) {
	var req struct {
		UserType   string `json:"userType"`
		CreditType string `json:"creditType"`
		MultiType  string `json:"multiType"`
		ModelRisk  string `json:"modelRisk"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "1001",
			"message": "参数错误",
		})
		return
	}

	complexMockMode.userType = req.UserType
	complexMockMode.creditType = req.CreditType
	complexMockMode.multiType = req.MultiType
	complexMockMode.modelRisk = req.ModelRisk

	c.JSON(http.StatusOK, gin.H{
		"code":    "0000",
		"message": "success",
		"data": gin.H{
			"userType":   complexMockMode.userType,
			"creditType": complexMockMode.creditType,
			"multiType":  complexMockMode.multiType,
			"modelRisk":  complexMockMode.modelRisk,
		},
	})
}

func mockComplexServices(c *gin.Context) {
	dsID := c.Param("dsId")
	modelID := c.Param("modelId")

	if dsID != "" {
		mockDataSource(c, dsID)
		return
	}

	if modelID != "" {
		mockComplexModel(c, modelID)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{
		"code":    "404",
		"message": "not found",
	})
}

func mockDataSource(c *gin.Context, dsID string) {
	var req datasource.DataSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "400",
			"message": "参数错误",
		})
		return
	}

	fmt.Printf("[Mock数据源] dsId=%s, userId=%s\n", dsID, req.UserID)

	var resp gin.H

	switch dsID {
	case "DS001":
		// 用户基本信息
		if complexMockMode.userType == "blacklist" {
			resp = gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"userId":      req.UserID,
					"isBlacklist": true,
					"name":        "黑名单用户",
				},
			}
		} else {
			resp = gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"userId":      req.UserID,
					"isBlacklist": false,
					"name":        "正常用户",
				},
			}
		}
	case "DS002":
		// 征信数据
		if complexMockMode.creditType == "bad" {
			resp = gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"creditScore":     500,
					"hasOverdue":      true,
					"overdueDays":     90,
					"creditLevel":     "D",
				},
			}
		} else {
			resp = gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"creditScore":     750,
					"hasOverdue":      false,
					"overdueDays":     0,
					"creditLevel":     "A",
				},
			}
		}
	case "DS003":
		// 多头借贷数据
		var multiCount int
		switch complexMockMode.multiType {
		case "low":
			multiCount = 1
		case "medium":
			multiCount = 4
		case "high":
			multiCount = 8
		default:
			multiCount = 1
		}
		resp = gin.H{
			"code":    0,
			"message": "success",
			"data": gin.H{
				"multiQueryCount7d":  multiCount,
				"multiQueryCount30d": multiCount * 3,
				"approvalCount7d":    multiCount / 2,
			},
		}
	default:
		resp = gin.H{
			"code":    0,
			"message": "success",
			"data":    gin.H{},
		}
	}

	respJSON, _ := json.Marshal(resp)
	fmt.Printf("[Mock数据源响应] dsId=%s, data=%s\n", dsID, string(respJSON))

	c.JSON(http.StatusOK, resp)
}

func mockComplexModel(c *gin.Context, modelID string) {
	var req model.ModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "400",
			"message": "参数错误",
		})
		return
	}

	fmt.Printf("[Mock模型] modelId=%s, contractId=%s, risk=%s\n",
		modelID, req.ContractID, complexMockMode.modelRisk)

	var resp gin.H
	switch complexMockMode.modelRisk {
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
	fmt.Printf("[Mock模型响应] %s\n", string(respJSON))

	c.JSON(http.StatusOK, resp)
}

func executeComplexDecision(c *gin.Context) {
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
	fmt.Printf("\n[复杂决策请求] requestId=%s, decisionId=%s\n", req.RequestID, decisionID)

	// 执行复杂决策流
	ctx := &flow.ComplexDecisionContext{
		Input: req.Data,
	}

	if err := complexFlowEngine.Execute(ctx); err != nil {
		fmt.Printf("[复杂决策错误] %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "4001",
			"message": "决策执行失败",
		})
		return
	}

	fmt.Printf("[复杂决策结果] decisionId=%s, decision=%s, code=%s, reason=%s\n",
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
	fmt.Println("   风险决策引擎 - 复杂用例服务")
	fmt.Println("   数据源 + 规则 + 模型 完整流程")
	fmt.Println("========================================")
	fmt.Println()

	initComplexEngine()

	// 默认mock模式
	complexMockMode.userType = "normal"
	complexMockMode.creditType = "good"
	complexMockMode.multiType = "low"
	complexMockMode.modelRisk = "low"

	fmt.Println()
	fmt.Println("服务说明:")
	fmt.Println("  - 决策服务端口: 8080")
	fmt.Println("  - Mock服务端口: 8082")
	fmt.Println()
	fmt.Println("决策服务接口:")
	fmt.Println("  - 健康检查: http://localhost:8080/health")
	fmt.Println("  - 决策接口: http://localhost:8080/api/v1/decision/execute")
	fmt.Println("  - 设置Mock模式: POST http://localhost:8080/mock/mode")
	fmt.Println()
	fmt.Println("Mock服务接口:")
	fmt.Println("  - 数据源: http://localhost:8082/mock/datasource/:dsId")
	fmt.Println("  - 模型: http://localhost:8082/mock/model/:modelId")
	fmt.Println()
	fmt.Println("测试场景:")
	fmt.Println("  1. 黑名单用户 -> 直接拒绝 (userType=blacklist)")
	fmt.Println("  2. 正常用户+低多头+低风险 -> 通过 (userType=normal, multiType=low, modelRisk=low)")
	fmt.Println("  3. 高多头+中风险 -> 人工复核 (multiType=high, modelRisk=medium)")
	fmt.Println("  4. 征信不良+高风险 -> 拒绝 (creditType=bad, modelRisk=high)")
	fmt.Println()
	fmt.Println("按 Ctrl+C 停止服务")
	fmt.Println("========================================")
	fmt.Println()

	gin.SetMode(gin.ReleaseMode)

	// 启动Mock服务 (8082端口)
	go func() {
		mockRouter := gin.Default()
		mockRouter.POST("/mock/datasource/:dsId", mockComplexServices)
		mockRouter.POST("/mock/model/:modelId", mockComplexServices)
		fmt.Println("Mock服务启动: 0.0.0.0:8082")
		if err := mockRouter.Run("0.0.0.0:8082"); err != nil {
			fmt.Printf("Mock服务启动失败: %v\n", err)
		}
	}()

	// 启动决策服务 (8080端口)
	mainRouter := gin.Default()

	mainRouter.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "risk-decision-engine-complex",
			"version": "0.5.0",
		})
	})

	mainRouter.POST("/api/v1/decision/execute", executeComplexDecision)
	mainRouter.POST("/mock/mode", setComplexMockMode)

	fmt.Println("决策服务启动: 0.0.0.0:8080")
	if err := mainRouter.Run("0.0.0.0:8080"); err != nil {
		fmt.Printf("决策服务启动失败: %v\n", err)
	}
}
